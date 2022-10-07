package ltd

import (
	"fmt"
	"github.com/jacohend/ltd/lightning-terminal"
	"github.com/lightninglabs/taro/chanutils"
	"github.com/lightninglabs/taro/tarocfg"
	"github.com/lightningnetwork/lnd"
	"github.com/lightningnetwork/lnd/signal"
	"github.com/spf13/viper"
	"os"
	sig "os/signal"
)

func main() {

	var config Config
	err := viper.Unmarshal(&config)
	if err != nil {
		panic(err)
	}
	if config.LnHome == "" {
		homedirname, err := os.UserHomeDir()
		if err != nil {
			panic(err)
		}
		lndHome := homedirname + "/.lnd"
		config.LnHome = lndHome
	}

	loadLndConfig := make(chan GoroutineNotifier)
	go Lnd(config, loadLndConfig)
	result := <-loadLndConfig
	if result.err != nil {
		panic(result.err)
	}

	os.Args = os.Args[:1]

	loadTaroConfig := make(chan GoroutineNotifier)
	go Taro(config, loadTaroConfig)
	result = <-loadTaroConfig
	if result.err != nil {
		panic(result.err)
	}

	done := make(chan os.Signal)
	sig.Notify(done, os.Interrupt)
	<-done
}

// Lnd : We pass in commandline arguments because of undefined DefaultConfig unmarshalling behavior in subRPCServers
func Lnd(config Config, loadComplete chan GoroutineNotifier) {
	osArgs := []string{
		"--lnddir=" + config.LnHome,
		"--logdir=" + config.LnHome + "/logs",
		"--datadir=" + config.LnHome + "/data",
		"--bitcoin.active",
		"--bitcoin.node=neutrino",
		"--externalip=" + config.LnIp + ":9735",
		"--externalhosts=" + config.LnDomain + ":9735",
		"--listen=0.0.0.0:9735",
		"--restlisten=0.0.0.0:8080",
		"--rpclisten=0.0.0.0:10009",
		"--bitcoin.defaultchanconfs=3",
		"--tlsextradomain=" + config.LnDomain,
		"--tlsextraip=" + config.LnIp,
		"--debuglevel=" + config.LogLevel,
		"--accept-amp",
		"--accept-keysend",
	}
	if config.BitcoinNetwork == "mainnet" {
		osArgs = append(osArgs, "--feeurl=https://nodes.lightning.computer/fees/v1/btc-fee-estimates.json")
		osArgs = append(osArgs, []string{"btcd-mainnet.lightning.computer", "mainnet1-btcd.zaphq.io", "mainnet2-btcd.zaphq.io", "24.155.196.246:8333", "75.103.209.147:8333"}...)
		osArgs = append(osArgs, "--bitcoin.mainnet")
		osArgs = append(osArgs, "--routing.assumechanvalid")
	} else if config.BitcoinNetwork == "testnet" {
		osArgs = append(osArgs, "--bitcoin.testnet")
		osArgs = append(osArgs, []string{"faucet.lightning.community:18333", "btcd-testnet.lightning.computer", "testnet1-btcd.zaphq.io", "testnet2-btcd.zaphq.io"}...)
	}
	os.Args = append(os.Args, osArgs...)
	shutdownInterceptor, err := signal.Intercept()
	if err != nil {
		loadComplete <- GoroutineNotifier{result: 1, err: err}
		return
	}
	loadedConfig, err := lnd.LoadConfig(shutdownInterceptor)
	if err != nil {
		loadComplete <- GoroutineNotifier{result: 1, err: err}
		return
	}
	loadComplete <- GoroutineNotifier{result: 0, err: nil}
	if err := lnd.Main(
		loadedConfig, lnd.ListenerCfg{}, loadedConfig.ImplementationConfig(shutdownInterceptor), shutdownInterceptor,
	); err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
		loadComplete <- GoroutineNotifier{result: 1, err: err}
		return
	}
}

func Taro(config Config, loadComplete chan GoroutineNotifier) {
	// Hook interceptor for os signals.
	shutdownInterceptor, err := signal.Intercept()
	if err != nil {
		loadComplete <- GoroutineNotifier{result: 1, err: err}
		return
	}

	osArgs := []string{
		"--network=" + config.BitcoinNetwork,
		"--debuglevel=" + config.LogLevel,
		"--lnd.host=" + "localhost:10009",
		"--lnd.macaroonpath=" + fmt.Sprintf("%s/data/chain/bitcoin/%s/admin.macaroon", config.LnHome, config.BitcoinNetwork),
		"--lnd.tlspath=" + fmt.Sprintf("%s/tls.cert", config.LnHome),
	}
	os.Args = append(os.Args, osArgs...)

	// Load the configuration, and parse any command line options. This
	// function will also set up logging properly.
	cfg, cfgLogger, err := tarocfg.LoadConfig(shutdownInterceptor)
	if err != nil {
		loadComplete <- GoroutineNotifier{result: 1, err: err}
		return
	}

	// This concurrent error queue can be used by every component that can
	// raise runtime errors. Using a queue will prevent us from blocking on
	// sending errors to it, as long as the queue is running.
	errQueue := chanutils.NewConcurrentQueue[error](
		chanutils.DefaultQueueSize,
	)
	errQueue.Start()
	defer errQueue.Stop()

	server, err := tarocfg.CreateServerFromConfig(
		cfg, cfgLogger, shutdownInterceptor, errQueue.ChanIn(),
	)
	if err != nil {
		err := fmt.Errorf("error creating server: %v", err)
		_, _ = fmt.Fprintln(os.Stderr, err)
		loadComplete <- GoroutineNotifier{result: 1, err: err}
		return
	}
	loadComplete <- GoroutineNotifier{result: 0, err: nil}

	err = server.RunUntilShutdown(errQueue.ChanOut())
	if err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
		loadComplete <- GoroutineNotifier{result: 1, err: err}
		return
	}
}

// Run starts everything and then blocks until either the application is shut
// down or a critical error happens.
func Terminal() {
	terminal.New().Run()
}

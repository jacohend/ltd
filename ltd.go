package main

import (
	_ "embed"
	"fmt"
	"github.com/lightninglabs/taro/chanutils"
	"github.com/lightninglabs/taro/tarocfg"
	"github.com/lightningnetwork/lnd"
	"os"
	sig "os/signal"
	"time"
)

//go:embed litd-debug
var LightningTerminal []byte

func main() {

	config := LoadConfig()
	fmt.Printf("LTD starting with config: %+v\n", config)

	os.Args = os.Args[:1]

	loadLndConfig := make(chan GoroutineNotifier)
	go Lnd(config, loadLndConfig)
	result := <-loadLndConfig
	if result.err != nil {
		panic(result.err)
	}
	fmt.Printf("LND is starting, waiting for macaroon...\n")
	config.LndClient.WaitForMacaroon(time.Minute * 5)
	fmt.Printf("LND RPC Servers are starting, waiting for connection...\n")
	config.LndClient.WaitForConnection(time.Hour * 1)

	os.Args = os.Args[:1]

	loadTaroConfig := make(chan GoroutineNotifier)
	go Taro(config, loadTaroConfig)
	result = <-loadTaroConfig
	if result.err != nil {
		panic(result.err)
	}

	os.Args = os.Args[:1]

	loadTerminalConfig := make(chan GoroutineNotifier)
	go Terminal(config, loadTerminalConfig)
	result = <-loadTerminalConfig
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
		fmt.Sprintf("--bitcoin.%s", config.BitcoinNetwork),
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
	loadedConfig, err := lnd.LoadConfig(config.Interceptor)
	if err != nil {
		loadComplete <- GoroutineNotifier{result: 1, err: err}
		return
	}
	loadComplete <- GoroutineNotifier{result: 0, err: nil}
	if err := lnd.Main(
		loadedConfig, lnd.ListenerCfg{}, loadedConfig.ImplementationConfig(config.Interceptor), config.Interceptor,
	); err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
		loadComplete <- GoroutineNotifier{result: 1, err: err}
		return
	}
}

func Taro(config Config, loadComplete chan GoroutineNotifier) {
	// Hook interceptor for os signals.
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
	cfg, cfgLogger, err := tarocfg.LoadConfig(config.Interceptor)
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
		cfg, cfgLogger, config.Interceptor, errQueue.ChanIn(),
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

// Terminal starts everything and then blocks until either the application is shut
// down or a critical error happens.
func Terminal(config Config, loadComplete chan GoroutineNotifier) {
	fd, err := MemfdCreate("/litd-debug")
	if err != nil {
		loadComplete <- GoroutineNotifier{result: 1, err: err}
		return
	}

	err = CopyToMem(fd, LightningTerminal)
	if err != nil {
		loadComplete <- GoroutineNotifier{result: 1, err: err}
		return
	}

	err = ExecveAt(fd)
	if err != nil {
		loadComplete <- GoroutineNotifier{result: 1, err: err}
		return
	}
	loadComplete <- GoroutineNotifier{result: 0, err: nil}
}

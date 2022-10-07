package main

import (
	"fmt"
	lightning "github.com/chainpoint/lightning-go"
	"github.com/jessevdk/go-flags"
	"github.com/lightningnetwork/lnd/signal"
	"os"
)

type Config struct {
	BitcoinNetwork string `long:"netork" description:"bitcoin network to use; mainnet or testnet"`
	LnIp           string `long:"ip" description:"public server ip"`
	LnDomain       string `long:"domain" description:"public server domain"`
	LnHome         string `long:"ln_dir" description:"Path to lightning directory"`
	LogLevel       string `long:"log_level" description:"log level for lnd and taro"`
	UIPassword     string `long:"uipassword" description:"password for your lightning-terminal web ui"`
	Interceptor    signal.Interceptor
	LndClient      lightning.LightningClient
}

type GoroutineNotifier struct {
	result int
	err    error
}

func LoadConfig() Config {
	var err error
	var config Config
	// Finally, parse the remaining command line options again to ensure
	// they take precedence.
	flagParser := flags.NewParser(&config, flags.IgnoreUnknown)
	if _, err := flagParser.Parse(); err != nil {
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
	if config.BitcoinNetwork == "" {
		config.BitcoinNetwork = "testnet"
	}
	lndClient := lightning.LightningClient{
		TlsPath:        fmt.Sprintf("%s/tls.cert", config.LnHome),
		MacPath:        fmt.Sprintf("%s/data/chain/bitcoin/%s/admin.macaroon", config.LnHome, config.BitcoinNetwork),
		ServerHostPort: config.LnIp + ":10009",
		LndLogLevel:    "error",
		MinConfs:       3,
		Testnet:        config.BitcoinNetwork == "testnet",
	}
	config.LndClient = lndClient
	config.Interceptor, err = signal.Intercept()
	if err != nil {
		panic(err)
	}
	return config
}

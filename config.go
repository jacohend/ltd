package main

import (
	"fmt"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btclog"
	lightning "github.com/chainpoint/lightning-go"
	"github.com/jessevdk/go-flags"
	"github.com/lightninglabs/taro/tarocfg"
	"github.com/lightningnetwork/lnd"
	"github.com/lightningnetwork/lnd/chainreg"
	"github.com/lightningnetwork/lnd/signal"
	"os"
)

type Config struct {
	BitcoinNetwork string `long:"network" description:"bitcoin network to use; mainnet or testnet"`
	LnIp           string `long:"ip" description:"public server ip"`
	LnDomain       string `long:"domain" description:"public server domain"`
	LnHome         string `long:"ln_dir" description:"Path to lightning directory"`
	LogLevel       string `long:"log_level" description:"log level for lnd and taro"`
	UIPassword     string `long:"uipassword" description:"password for your lightning-terminal web ui"`
	Interceptor    signal.Interceptor
	LndClient      lightning.LightningClient
	Lnd            *lnd.Config     `group:"lnd" namespace:"lnd"`
	Taro           *tarocfg.Config `group:"taro" namespace:"taro"`
	Logger         btclog.Logger
}

type GoroutineNotifier struct {
	result int
	err    error
}

func LoadConfig() Config {
	var err error
	var config Config
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
	if config.LogLevel == "" {
		config.LogLevel = "warn"
	}
	if config.LnDomain == "" {
		config.LnDomain = "lightning.tabulon.io"
	}
	taroConfig := tarocfg.DefaultConfig()
	config.Taro = &taroConfig
	lndConfig := lnd.DefaultConfig()
	config.Lnd = &lndConfig

	flagParser := flags.NewParser(&config, flags.IgnoreUnknown)
	if _, err := flagParser.Parse(); err != nil {
		panic(err)
	}
	configFilePath := lnd.CleanAndExpandPath(lndConfig.ConfigFile)
	fileParser := flags.NewParser(&config, flags.IgnoreUnknown)
	flags.NewIniParser(fileParser).ParseFile(configFilePath)

	config.Lnd.LndDir = config.LnHome
	config.Lnd.LogDir = config.LnHome + "/logs"
	config.Lnd.DataDir = config.LnHome + "/data"
	config.Lnd.Bitcoin.Active = true
	config.Lnd.Bitcoin.Node = "neutrino"
	config.Lnd.RawExternalIPs = []string{config.LnIp + ":9735"}
	config.Lnd.ExternalHosts = []string{config.LnDomain + ":9735"}
	config.Lnd.RawListeners = []string{"0.0.0.0:9735"}
	config.Lnd.RawRESTListeners = []string{"0.0.0.0:8080"}
	config.Lnd.RawRPCListeners = []string{"0.0.0.0:10009"}
	config.Lnd.Bitcoin.DefaultNumChanConfs = 3
	config.Lnd.TLSExtraDomains = []string{config.LnDomain}
	config.Lnd.TLSExtraIPs = []string{config.LnIp}
	config.Lnd.DebugLevel = config.LogLevel
	config.Lnd.AcceptAMP = true
	config.Lnd.AcceptKeySend = true
	config.Taro.ChainConf.Network = config.BitcoinNetwork
	switch config.BitcoinNetwork {
	case "mainnet":
		config.Lnd.Bitcoin.MainNet = true
		config.Lnd.ActiveNetParams = chainreg.BitcoinMainNetParams
		config.Taro.ActiveNetParams = chaincfg.MainNetParams
		config.Lnd.FeeURL = "https://nodes.lightning.computer/fees/v1/btc-fee-estimates.json"
		config.Lnd.NeutrinoMode.AddPeers = []string{"btcd-mainnet.lightning.computer", "mainnet1-btcd.zaphq.io", "mainnet2-btcd.zaphq.io", "24.155.196.246:8333", "75.103.209.147:8333"}
		config.Lnd.Routing.AssumeChannelValid = true
	case "testnet":
		config.Lnd.Bitcoin.TestNet3 = true
		config.Lnd.ActiveNetParams = chainreg.BitcoinTestNetParams
		config.Taro.ActiveNetParams = chaincfg.TestNet3Params
		config.Lnd.NeutrinoMode.AddPeers = []string{"faucet.lightning.community", "btcd-testnet.lightning.computer"}
	case "simnet":
		config.Lnd.Bitcoin.SimNet = true
		config.Lnd.ActiveNetParams = chainreg.BitcoinSimNetParams
		config.Taro.ActiveNetParams = chaincfg.SimNetParams
	default:
	}
	config.Taro.DebugLevel = config.LogLevel
	config.Taro.Lnd = &tarocfg.LndConfig{
		Host:         "127.0.0.1:10009",
		MacaroonPath: fmt.Sprintf("%s/data/chain/bitcoin/%s/admin.macaroon", config.LnHome, config.BitcoinNetwork),
		TLSPath:      fmt.Sprintf("%s/tls.cert", config.LnHome),
	}

	flagParser = flags.NewParser(&config, flags.IgnoreUnknown)
	if _, err := flagParser.Parse(); err != nil {
		panic(err)
	}

	fmt.Printf("LTD Config: %+v\n\n", config)
	fmt.Printf("LND Config: %+v\n\n", config.Lnd)
	fmt.Printf("Bitcoin Config: %+v\n\n", config.Lnd.BitcoindMode)
	fmt.Printf("Neutrino Config: %+v\n\n", config.Lnd.NeutrinoMode)
	fmt.Printf("Taro Config: %+v\n\n", config.Taro)

	config.Lnd, err = lnd.ValidateConfig(*config.Lnd, config.Interceptor, fileParser, flagParser)
	if err != nil {
		panic(err)
	}
	config.Taro, config.Logger, err = tarocfg.ValidateConfig(*config.Taro, config.Interceptor)
	if err != nil {
		panic(err)
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

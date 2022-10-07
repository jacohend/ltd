package ltd

type Config struct {
	BitcoinNetwork string `long:"netork" description:"bitcoin network to use; mainnet or testnet"`
	LnIp string `long:"ip" description:"public server ip"`
	LnDomain string `long:"domain" description:"public server domain"`
	LnHome string `long:"ln_dir" description:"Path to lightning directory"`
	LogLevel string `long:"log_level" description:"log level for lnd and taro"`

}

type GoroutineNotifier struct {
	result int
	err    error
}
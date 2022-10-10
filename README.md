# Lightning-Taro Daemon (ltd)

Integrated Lnd and Taro daemon, with Terminal on the side. Basically Lightning Voltron. Currently restricted to testnet mode. 

## Installation

git, make, and go 1.18+ are required.

```bash
git clone https://github.com/jacohend/ltd.git

cd ltd && make install-go

source ~/.bashrc  # reload shell for go install

make build-all
```

## Usage

`lncli` and `tarocli` required for wallet creation and interaction. 
The Makefile should build these binaries for you.

First run `ltd` in one shell: 
```bash
./ltd --network=testnet --uipassword=changethis --ip=127.0.0.1 --log_level=debug
```

Then open another shell and create the wallet with `lncli`:

```bash
# create wallet. Be sure to save your passsword and seed phrase!
./lncli --rpcserver 127.0.0.1:10009 --lnddir=~/.lnd --macaroonpath=~/.lnd/data/chain/bitcoin/testnet/admin.macaroon --tlscertpath=~/.lnd/tls.cert create

# create address
./lncli --rpcserver 127.0.0.1:10009 --lnddir=~/.lnd --macaroonpath=~/.lnd/data/chain/bitcoin/testnet/admin.macaroon --tlscertpath=~/.lnd/tls.cert newaddress p2tr

# later you can unlock the wallet when you need to: 
./lncli --rpcserver 127.0.0.1:10009 --lnddir=~/.lnd --macaroonpath=~/.lnd/data/chain/bitcoin/testnet/admin.macaroon --tlscertpath=~/.lnd/tls.cert unlock
```

You can interact with taro as well, such as minting new assets:

```bash
./tarocli assets mint --type=normal --name=waynechain --supply=2100000000000000 --meta=wayne --enable_emission
{
    "batch_key": "0385b55688b9170d2d4f613a5fa7fcb988ae0a0f6e61a7e72184e77236611b556e"
}
```


The Lightning Terminal UI will be accessible at http://localhost:8443

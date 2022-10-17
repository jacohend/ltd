LND=dev autopilotrpc signrpc walletrpc chainrpc invoicesrpc watchtowerrpc neutrinorpc monitoring peersrpc kvdb_postgres kvdb_etcd
comma=,
noop=
space=$(noop) $(noop)
TAGS=$(subst $(space),$(comma),$(LND))
VERSION=v0.15.99

.PHONY : build-all
build-all: init-submodules build-ui build-lncli build-tarocli build

.PHONY : build-tarocli
build-tarocli:
	cd taro && make build && cd .. && cp taro/tarocli-debug ./tarocli

.PHONY : build-lncli
build-lncli:
	cd lnd && make build && cd .. && cp lnd/lncli-debug ./lncli

.PHONY : init-submodules
init-submodules:
	git submodule update --init

.PHONY : build-ui
build-ui:
	cd lightning-terminal && make build && cd .. && cp lightning-terminal/litd-debug ./litd

.PHONY : build
build:
	GO111MODULE=on CGO_ENABLED=0 go build -tags "$(LND)" -ldflags " -X github.com/lightningnetwork/lnd/build.Commit=$(VERSION) -X github.com/lightningnetwork/lnd/build.RawTags=$(TAGS)"

.PHONY : build-dev-all
build-dev-all: build-ui-dev build-dev

.PHONY : build-ui
build-ui-dev:
	cd lightning-terminal && go mod tidy && go mod vendor && make build && cd .. && cp lightning-terminal/litd-debug ./litd

.PHONY : build
build-dev:
	go mod tidy && go mod vendor
	GO111MODULE=on CGO_ENABLED=0 go build -tags "$(LND)" -ldflags " -X github.com/lightningnetwork/lnd/build.Commit=$(VERSION) -X github.com/lightningnetwork/lnd/build.RawTags=$(TAGS)"

.PHONY : install-all
install-all: init-submodules install-lncli install-tarocli install

.PHONY : install-tarocli
install-tarocli:
	cd taro && make install

.PHONY : install-lncli
install-lncli:
	cd lnd && make install

.PHONY : install
install:
	GO111MODULE=on CGO_ENABLED=0 go install -tags "$(LND)" -ldflags " -X github.com/lightningnetwork/lnd/build.Commit=$(VERSION) -X github.com/lightningnetwork/lnd/build.RawTags=$(TAGS)"

.PHONE : install-go
install-go:
	curl -L https://git.io/vQhTU | bash -s -- --version 1.18.7

.PHONY : install-daemon
install-daemon:
	envsubst < ./config/ltd.service.template > ./config/ltd.service
	sudo cp ./config/ltd.service /lib/systemd/system
	sudo systemctl daemon-reload
	sudo systemctl enable ltd

.PHONY : start-daemon
start-daemon:
	sudo systemctl start ltd

.PHONY : stop-daemon
stop-daemon:
	sudo systemctl stop ltd

.PHONY : status-daemon
status-daemon:
	sudo systemctl status ltd

.PHONY : log-daemon
log-daemon:
	journalctl --unit ltd --follow
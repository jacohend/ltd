LND=dev autopilotrpc signrpc walletrpc chainrpc invoicesrpc watchtowerrpc neutrinorpc monitoring peersrpc kvdb_postgres kvdb_etcd
comma=,
noop=
space=$(noop) $(noop)
TAGS=$(subst $(space),$(comma),$(LND))
VERSION=v0.15.99

.PHONY : build-all
build-all: build-ui build

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

.PHONY : install
install:
	GO111MODULE=on CGO_ENABLED=0 go install -tags "$(LND)" -ldflags " -X github.com/lightningnetwork/lnd/build.Commit=$(VERSION) -X github.com/lightningnetwork/lnd/build.RawTags=$(TAGS)"

.PHONE : install-go
install-go:
	curl -L https://git.io/vQhTU | bash -s -- --version 1.18.7
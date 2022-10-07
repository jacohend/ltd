LND=litd autopilotrpc signrpc walletrpc chainrpc invoicesrpc watchtowerrpc neutrinorpc monitoring peersrpc kvdb_postgres kvdb_etcd

.PHONY : build-all
build-all: build-ui build

.PHONY : build-ui
build-ui:
	cd lightning-terminal && make build && cd .. && cp lightning-terminal/litd-debug .

.PHONY : build
build:
	GO111MODULE=on CGO_ENABLED=0 go build -tags "$(LND)"

.PHONY : install
install:
	GO111MODULE=on CGO_ENABLED=0 go install -tags "$(LND)"

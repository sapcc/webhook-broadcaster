PKG:=github.com/sapcc/webhook-ingester
build:
	go build -o bin/webhook-ingester $(PKG)

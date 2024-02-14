//This file is only used to speedup the docker build
//We use this to download and compile the go module dependencies before adding our own source code.
//See Dockerfile for more details

package main

import (
	_ "flag"
	_ "log"
	_ "net"
	_ "net/http"
	_ "os"
	_ "os/signal"
	_ "syscall"
	_ "time"

	_ "github.com/oklog/run"
	_ "github.com/prometheus/client_golang/prometheus"
	_ "github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {

}

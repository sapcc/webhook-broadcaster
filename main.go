package main

import (
	"flag"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/oklog/run"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sapcc/go-api-declarations/bininfo"
)

type resource struct {
	team     string
	pipeline string
	name     string
	token    string
}

var (
	listenAddr         string
	concourseURL       string
	authUser           string
	authPassword       string
	refreshInterval    time.Duration
	webhookConcurrency int
	flags              *flag.FlagSet
	debug              bool
	printVersion       bool
)

func init() {
	//we go with our own flagset to get rid of crap added by glog to the default flagset
	flags = flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	flags.StringVar(&listenAddr, "listen-addr", ":8080", "Listen address of webhook ingester")
	flags.StringVar(&concourseURL, "concourse-url", "", "External URL of the concourse api")
	flags.StringVar(&authUser, "auth-user", "", "Basic auth concourse username")
	flags.StringVar(&authPassword, "auth-password", "", "Basic auth concourse password")
	flags.DurationVar(&refreshInterval, "refresh-interval", 5*time.Minute, "Resource refresh interval")
	flags.IntVar(&webhookConcurrency, "webhook-concurrency", 20, "How many resources to notify in parallel")
	flags.BoolVar(&debug, "dry-run", false, "Dry-run. Don't call webhooks")
	flags.BoolVar(&printVersion, "version", false, "Print version info")
}

func main() {
	flags.Parse(os.Args[1:])

	if printVersion {
		log.Printf("webhook-broadcaster version %s", bininfo.Version())
		os.Exit(0)
	}

	if concourseURL == "" || authUser == "" || authPassword == "" {
		log.Fatal("Missing one or more of required flags: -concourse-url -auth-user -auth-password")
	}

	client, err := NewConcourseClient(concourseURL, authUser, authPassword)
	if err != nil {
		log.Fatalf("Failed to create Concourse client")
	}

	var group run.Group

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, os.Interrupt, syscall.SIGTERM) // Push signals into channel

	//setup signal Handler
	cancelSignal := make(chan struct{})
	group.Add(func() error {
		select {
		case sig := <-sigs:
			log.Printf("Received %s signal, shutting down", sig)
		case <-cancelSignal:
		}
		return nil
	}, func(_ error) {
		close(cancelSignal)
	})

	//setup resource cache
	cancelCache := make(chan struct{})
	group.Add(func() error {
		defer logend(logstart("resource cache"))
		tick := time.NewTicker(refreshInterval)
		defer tick.Stop()
		for {
			if err := UpdateCache(*client); err != nil {
				log.Printf("Failed to update cache: %s", err)
			}
			select {
			case <-tick.C:
			case <-cancelCache:
				return nil
			}
		}
	}, func(_ error) {
		close(cancelCache)
	})

	//setup workqueue
	requestQueue := NewRequestWorkqueue(webhookConcurrency)
	cancelQueue := make(chan struct{})
	group.Add(func() error {
		defer logend(logstart("request workqueue"))
		requestQueue.Run(cancelQueue)
		return nil
	}, func(_ error) {
		close(cancelQueue)
	})

	//setup http server
	ln, err := net.Listen("tcp", listenAddr)
	if err != nil {
		log.Fatalf("Failed to listen on %s: %s", listenAddr, err)
	}
	group.Add(func() error {
		defer logend(logstart("http server"))
		log.Printf("Listening for incoming webhooks on %s", ln.Addr())
		mux := http.NewServeMux()

		requestCounter := prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "http_requests_total",
				Help: "Total number of incoming HTTP requests",
			},
			[]string{"code", "method"},
		)
		prometheus.Register(requestCounter)
		ghHandler := promhttp.InstrumentHandlerCounter(requestCounter, &GithubWebhookHandler{requestQueue})
		mux.Handle("/github", ghHandler)
		mux.Handle("/metrics", promhttp.Handler())
		return http.Serve(ln, mux)
	}, func(_ error) {
		ln.Close()
	})

	group.Run()

}

func logstart(what string) string {
	log.Println("Starting ", what)
	return what
}
func logend(what string) {
	log.Println("Stopped ", what)
}

func debugf(format string, args ...interface{}) {
	if debug {
		log.Printf(format, args...)
	}
}

package main

import (
	"fmt"
	"log"
	"net/http"
	"regexp"
	"time"

	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/util/workqueue"

	"github.com/prometheus/client_golang/prometheus"
	_ "github.com/sapcc/webhook-broadcaster/prometheus"
)

const (
	WORKER_BASE_DELAY = 5 * time.Second
	WORKER_MAX_DELAY  = 60 * time.Second
)

type RequestWorkqueue struct {
	queue       workqueue.RateLimitingInterface
	threadiness int

	webhooksSuccess prometheus.Counter
	webhooksErrors  prometheus.Counter
}

func NewRequestWorkqueue(threadiness int) *RequestWorkqueue {
	wq := &RequestWorkqueue{
		queue:       workqueue.NewNamedRateLimitingQueue(workqueue.NewItemExponentialFailureRateLimiter(WORKER_BASE_DELAY, WORKER_MAX_DELAY), "webhook"),
		threadiness: threadiness,

		webhooksSuccess: prometheus.NewCounter(prometheus.CounterOpts{
			Subsystem: "webhook",
			Name:      "success_total",
			Help:      "Total number of successfully delivered webhooks",
		}),
		webhooksErrors: prometheus.NewCounter(prometheus.CounterOpts{
			Subsystem: "webhook",
			Name:      "errors_total",
			Help:      "Total number of successfully delivered webhooks",
		}),
	}

	prometheus.Register(wq.webhooksSuccess)
	prometheus.Register(wq.webhooksErrors)
	return wq

}

func (c *RequestWorkqueue) Add(url string) {
	c.queue.Add(url)
}

func (c *RequestWorkqueue) Run(stopCh <-chan struct{}) {

	defer c.queue.ShutDown()

	for i := 0; i < c.threadiness; i++ {
		go wait.Until(c.worker, time.Second, stopCh)
	}

	<-stopCh

}

func (c *RequestWorkqueue) worker() {
	for c.processNextWorkItem() {
	}
}

func (c *RequestWorkqueue) processNextWorkItem() bool {
	key, quit := c.queue.Get()
	if quit {
		return false
	}
	defer c.queue.Done(key)

	err := c.perform(key.(string))
	if err != nil {
		c.webhooksErrors.Inc()
	} else {
		c.webhooksSuccess.Inc()
	}
	if err != nil && c.queue.NumRequeues(key) < 5 {
		// Re-enqueue the key rate limited. Based on the rate limiter on the
		// queue and the re-enqueue history, the key will be processed later again.
		c.queue.AddRateLimited(key)
		return true
	}
	//clear the rate limit history on successful processing
	c.queue.Forget(key)
	return true

}

var tokenRegexp = regexp.MustCompile(`webhook_token=[^&]+`)

func (c *RequestWorkqueue) perform(url string) error {
	redactedURL := tokenRegexp.ReplaceAllString(url, "webhook_token=[REDACTED]")
	if debug {
		log.Printf("DRY RUN: Calling POST %s", redactedURL)
		return nil
	}

	log.Printf("Calling POST %s", redactedURL)
	response, err := http.Post(url, "", nil)
	if err != nil || response.StatusCode >= 400 {
		return fmt.Errorf("Request failed. URL: %s, response: %s Error: %v",
			redactedURL,
			response.Status,
			err,
		)
	}
	return nil
}

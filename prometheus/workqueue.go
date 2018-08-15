/*
Copied from https://github.com/kubernetes/kubernetes/blob/master/pkg/util/workqueue/prometheus/prometheus.go
Changes: Converted summaries to histgramms

Copyright 2016 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package prometheus

import (
	"k8s.io/client-go/util/workqueue"

	"github.com/prometheus/client_golang/prometheus"
)

// Package prometheus sets the workqueue DefaultMetricsFactory to produce
// prometheus metrics. To use this package, you just have to import it.

func init() {
	workqueue.SetProvider(prometheusMetricsProvider{})
}

type prometheusMetricsProvider struct{}

//convert stupid microsecond intervals to seconds
type summaryWrapper struct {
	metric workqueue.SummaryMetric
}

func (s *summaryWrapper) Observe(o float64) {
	s.metric.Observe(o / 1000000.0)
}

func (_ prometheusMetricsProvider) NewDepthMetric(name string) workqueue.GaugeMetric {
	depth := prometheus.NewGauge(prometheus.GaugeOpts{
		Subsystem: name,
		Name:      "depth",
		Help:      "Current depth of workqueue: " + name,
	})
	prometheus.Register(depth)
	return depth
}

func (_ prometheusMetricsProvider) NewAddsMetric(name string) workqueue.CounterMetric {
	adds := prometheus.NewCounter(prometheus.CounterOpts{
		Subsystem: name,
		Name:      "adds",
		Help:      "Total number of adds handled by workqueue: " + name,
	})
	prometheus.Register(adds)
	return adds
}

func (_ prometheusMetricsProvider) NewLatencyMetric(name string) workqueue.SummaryMetric {
	latency := prometheus.NewHistogram(prometheus.HistogramOpts{
		Subsystem: name,
		Name:      "queue_latency_seconds",
		Help:      "How long an item stays in workqueue" + name + " before being requested.",
		Buckets:   []float64{.5, 1, 2.5, 5, 10, 30, 60, 120, 300},
	})
	prometheus.Register(latency)
	return &summaryWrapper{latency}
}

func (_ prometheusMetricsProvider) NewWorkDurationMetric(name string) workqueue.SummaryMetric {
	workDuration := prometheus.NewHistogram(prometheus.HistogramOpts{
		Subsystem: name,
		Name:      "work_duration_seconds",
		Help:      "How long processing an item from workqueue " + name + " takes.",
		Buckets:   []float64{.5, 1, 2.5, 5, 10, 20, 40, 60, 120},
	})
	prometheus.Register(workDuration)
	return &summaryWrapper{workDuration}
}

func (_ prometheusMetricsProvider) NewRetriesMetric(name string) workqueue.CounterMetric {
	retries := prometheus.NewCounter(prometheus.CounterOpts{
		Subsystem: name,
		Name:      "retries",
		Help:      "Total number of retries handled by workqueue: " + name,
	})
	prometheus.Register(retries)
	return retries
}

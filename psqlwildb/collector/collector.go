package collector

import (
	"github.com/prometheus/client_golang/prometheus"
)

var TotalReq = prometheus.NewCounterVec(prometheus.CounterOpts{
	Name: "http_request_total",
	Help: "Total number of request by HTTP code.",
}, []string{"code", "url", "method"})

var ReqDuration = prometheus.NewHistogramVec(prometheus.HistogramOpts{
	Name: "http_request_duration",
	Help: "Duration of requests by HTTP code.",
}, []string{"code", "url", "method"})

func init() {
	prometheus.MustRegister(TotalReq)
	prometheus.MustRegister(ReqDuration)
}

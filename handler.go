package metrics

import (
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"
)

func (m *Metrics) ServeHTTP(w http.ResponseWriter, r *http.Request) (int, error) {
	next := m.next
	host, err := host(r)
	if err != nil {
		host = "-"
	}
	start := time.Now()

	status, err := next.ServeHTTP(w, r)

	requestCount.WithLabelValues(host).Inc()
	requestDuration.WithLabelValues(host).Observe(float64(time.Since(start)) / float64(time.Second))
	// responseSize.WithLabelValues(host).Observe(rlen) // TODO(miek): how to get the length?
	responseStatus.WithLabelValues(host, strconv.Itoa(status)).Inc()

	return status, err
}

func host(r *http.Request) (string, error) {
	host, _, err := net.SplitHostPort(r.Host)
	if err != nil {
		if !strings.Contains(r.Host, ":") {
			return r.Host, nil
		}
		return "", err
	}
	return host, nil
}

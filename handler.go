package metrics

import (
	"net/http"
	"strconv"
	"time"

	"github.com/mholt/caddy/middleware"
)

func (m *Metrics) ServeHTTP(w http.ResponseWriter, r *http.Request) (int, error) {
	next := m.next
	host, err := middleware.Context.Host()
	if err != nil {
		host = "-"
	}
	start := time.Now()

	status, err := next.ServeHTTP(w, r)

	requestCount.WithLabelValues(host).Inc()
	requestDuration.WithLabelValues(host).Observe(float64(time.Since(start)) / float64(time.Second))
	// responseSize.WithLabelValues(host).Observe(rlen) // TODO(miek): how to get the length?
	responseStatus.WithLabelValues(host, strconv.Itoa(status)).Inc()

	return code, err
}

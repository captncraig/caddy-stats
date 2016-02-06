package stats

import (
	"fmt"
	"net/http"
	"time"

	"github.com/mholt/caddy/middleware"
	"github.com/rcrowley/go-metrics"
)

func (m *metricsModule) ServeHTTP(w http.ResponseWriter, r *http.Request) (int, error) {
	next := m.next
	path := ""
	if m.uiPath != "" && middleware.Path(r.URL.Path).Matches(m.uiPath) {
		next = statsHandler{}
		path = "getStats"
	}
	if path == "" {
		path = m.pathName(r.URL.Path, r.Method)
	}
	//every datapoint gets tagged with server and path. A few get some extra.
	tags := func(extra ...string) map[string]string {
		m := map[string]string{"path": path, "server": m.serverName}
		if len(extra)%2 == 0 {
			for i := 1; i < len(extra); i += 2 {
				m[extra[i-1]] = extra[i]
			}
		}
		return m
	}
	start := time.Now()
	code, err := next.ServeHTTP(w, r)
	duration := time.Now().Sub(start)

	mname := MetricName("caddy.requests", tags("status", fmt.Sprint(code)))
	counter := metrics.GetOrRegisterCounter(mname, metrics.DefaultRegistry)
	counter.Inc(1)

	mname = MetricName("caddy.errors", tags())
	counter = metrics.GetOrRegisterCounter(mname, metrics.DefaultRegistry)
	if err != nil {
		counter.Inc(1)
	}

	mname = MetricName("caddy.response_time", tags())
	timer := metrics.GetOrRegisterTimer(mname, metrics.DefaultRegistry)
	timer.Update(duration)

	return code, err
}

type statsHandler struct{}

func (s statsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) (int, error) {
	snapshotLock.RLock()
	defer snapshotLock.RUnlock()
	w.Write(currentJSON)
	w.Header().Set("Content-Type", "application/json")
	return 200, nil
}

func (m *metricsModule) pathName(url string, method string) string {
	for _, pth := range m.paths {
		if middleware.Path(url).Matches(pth.path) {
			if pth.methods == nil {
				return pth.name
			}
			for _, m := range pth.methods {
				if m == method {
					return pth.name
				}
			}
		}
	}
	return "/"
}

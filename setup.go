package metrics

import (
	"fmt"
	"net/http"
	"sync"

	"github.com/mholt/caddy/caddy/setup"
	"github.com/mholt/caddy/middleware"
	"github.com/prometheus/client_golang/prometheus"
)

const (
	path = "/metrics"
	addr = "localhost:9180"
)

// Metrics holds the prometheus configuration. The metrics' path is fixed to be /metrics
type Metrics struct {
	next middleware.Handler
	addr string // where to we listen
	// subsystem?
}

func (m *metrics) start() error {
	define("")

	prometheus.MustRegister(requestCount)
	prometheus.MustRegister(requestDuration)
	prometheus.MustRegister(responseSize)
	prometheus.MustRegister(responseStatus)

	http.Handle(path, prometheus.Handler())
	go func() {
		fmt.Errorf("%s", http.ListenAndServe(m.addr, nil))
	}()
	return nil
}

func Setup(c *setup.Controller) (middleware.Middleware, error) {
	metrics, err := parse(c)
	if err != nil {
		return nil, err
	}
	if metrics.addr == "" {
		metrcs.addr = addr
	}
	sync.Once.Do(func() {
		c.Startup = append(c.Startup, start)
	})

	return func(next middleware.Handler) middleware.Handler {
		module.next = next
		return module
	}, nil
}

// prometheus {
//	address localhost:9180
// }
// Or just: prometheus localhost:9180
func parse(c *setup.Controller) (*Metrics, error) {
	metrics := &Metrics{}
	var err error

	for c.Next() {
		if metrics != nil {
			return nil, c.Err("prometheus: can only have one metrics module per server")
		}
		args := c.RemainingArgs()

		switch len(args) {
		case 0:
		case 1:
			metrics.addr = args[0]
		default:
			return nil, c.ArgErr()
		}
		for c.NextBlock() {
			switch c.Val() {
			case "address":
				args = c.RemainingArgs()
				if len(args) != 1 {
					return nil, c.ArgErr()
				}
				module.addr = args[0]
			default:
				return nil, c.Errf("prometheus: unknown item: %s", c.Val())
			}

		}
	}
	return metrics, err
}

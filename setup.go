package stats

import (
	"sync"
	"time"

	"github.com/mholt/caddy/caddy/setup"
	"github.com/mholt/caddy/middleware"
	"github.com/rcrowley/go-metrics"
)

type metricsModule struct {
	next       middleware.Handler
	uiPath     string
	serverName string
	paths      []pathMatch
}

type pathMatch struct {
	path    string
	name    string
	methods []string
}

var interval = 15 * time.Second
var once sync.Once

// start all collectors and gatherers
func start() error {
	metrics.RegisterRuntimeMemStats(metrics.DefaultRegistry)
	go metrics.CaptureRuntimeMemStats(metrics.DefaultRegistry, interval)
	time.Sleep(time.Second)
	go func() {
		for {
			time.Sleep(interval)
			snapshot()
		}
	}()
	return nil
}

func Setup(c *setup.Controller) (middleware.Middleware, error) {

	once.Do(func() {
		c.Startup = append(c.Startup, start)
	})

	module, err := parse(c)
	if err != nil {
		return nil, err
	}
	if module.serverName == "" {
		module.serverName = c.Address()
	}

	return func(next middleware.Handler) middleware.Handler {
		module.next = next
		return module
	}, nil
}

func parse(c *setup.Controller) (*metricsModule, error) {
	var module *metricsModule

	var err error
	for c.Next() {
		if module != nil {
			return nil, c.Err("Can only create one stats module per server")
		}
		module = &metricsModule{}
		args := c.RemainingArgs()

		switch len(args) {
		case 0:
		case 1:
			module.uiPath = args[0]
		default:
			return nil, c.ArgErr()
		}
		for c.NextBlock() {
			switch c.Val() {
			case "path":
				//path /foo
				args = c.RemainingArgs()
				if len(args) < 2 {
					return nil, c.ArgErr()
				}
				pth := pathMatch{
					path: args[0],
					name: args[1],
				}
				for _, meth := range args[2:] {
					pth.methods = append(pth.methods, meth)
				}
				module.paths = append(module.paths, pth)
			case "server":
				args = c.RemainingArgs()
				if len(args) != 1 {
					return nil, c.ArgErr()
				}
				module.serverName = args[0]
			case "send":
				// send dbtype args
				args = c.RemainingArgs()
				l := len(args)
				if l < 1 {
					return nil, c.ArgErr()
				}
				switch args[0] {
				case "influx":
					// influx server db uname password
					if l != 3 && l != 5 {
						return nil, c.ArgErr()
					}
					pub := &influxPublisher{url: args[1], database: args[2]}
					if l == 5 {
						pub.username = args[3]
						pub.password = args[4]
					}
					publishers["influx-"+args[1]] = pub
				default:
					return nil, c.Errf("Unknown send db: %s", args[0])
				}
			default:
				return nil, c.Errf("Unknown stats config item: %s", c.Val())
			}

		}
	}
	return module, err
}

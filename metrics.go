package stats

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/rcrowley/go-metrics"
)

var currentSnapshot []*Measurement
var currentJSON = []byte("[]")
var snapshotLock = sync.RWMutex{}
var publishers = map[string]publisher{}
var host string

func init() {
	host, _ = os.Hostname()
}

//Our internal representation of a measurement. Served directly to json endpoint, or adapted to each backend
type Measurement struct {
	Name      string
	Tags      map[string]string
	Values    map[string]float64
	Timestamp time.Time
}

type publisher interface {
	SendData([]*Measurement) error
}

//simple wrapper to allow metrics.Timer and metrics.Historgram to share a case.
//just the fields I use
type timerhistogram interface {
	Max() int64
	Mean() float64
	StdDev() float64
	Percentiles([]float64) []float64
}

func snapshot() {
	snapshotLock.RLock()
	initLen := len(currentSnapshot) + 128 //optimistic sizing
	snapshotLock.RUnlock()
	newSnap := make([]*Measurement, 0, initLen)

	now := time.Now()
	metrics.DefaultRegistry.Each(func(name string, val interface{}) {
		name, tags, err := parseTags(name)
		if err != nil {
			fmt.Printf("error parsing metric name/tags: %s\n", err)
		}
		if _, ok := tags["host"]; !ok {
			tags["host"] = host
		}
		m := &Measurement{
			Name:      name,
			Tags:      tags,
			Values:    map[string]float64{},
			Timestamp: now,
		}
		switch v := val.(type) {
		case metrics.Gauge:
			m.Values["value"] = float64(v.Value())
		case metrics.GaugeFloat64:
			m.Values["value"] = v.Value()
		case timerhistogram:
			m.Values["max"] = float64(v.Max())
			m.Values["avg"] = v.Mean()
			m.Values["stdev"] = v.StdDev()
			pcts := v.Percentiles([]float64{.9, .95, .99})
			m.Values["90th"] = pcts[0]
			m.Values["95th"] = pcts[1]
			m.Values["99th"] = pcts[2]
		case metrics.Counter:
			m.Values["value"] = float64(v.Count())
		default:
			fmt.Printf("UNKNOWN METRIC TYPE: %T\n", val)
			return
		}
		newSnap = append(newSnap, m)
	})

	snapshotLock.Lock()
	currentSnapshot = newSnap
	currentJSON, _ = json.MarshalIndent(currentSnapshot, "", "  ")
	snapshotLock.Unlock()

	wg := sync.WaitGroup{}
	for name, pub := range publishers {
		wg.Add(1)
		go func(name string, pub publisher) {
			defer wg.Done()
			if err := pub.SendData(newSnap); err != nil {
				log.Printf("Error sending data to %s: %s", pub, err)
			}
		}(name, pub)
	}
	wg.Wait()
}

func parseTags(name string) (string, map[string]string, error) {
	firstSplit := strings.SplitN(string(name), "{", 2)
	tags := map[string]string{}
	if len(firstSplit) == 1 {
		return name, tags, nil
	}
	name = firstSplit[0]
	t := firstSplit[1]
	if t[len(t)-1] != '}' {
		return "", nil, fmt.Errorf("metric name must end in } if tags used")
	}
	t = t[:len(t)-1]
	for _, pair := range strings.Split(t, ",") {
		parts := strings.Split(pair, "=")
		if len(parts) != 2 {
			return "", nil, fmt.Errorf("each metrig tag must be of 'key=value' form")
		}
		tags[parts[0]] = parts[1]
	}
	return name, tags, nil
}

func MetricName(name string, tags map[string]string) string {
	return fmt.Sprintf("%s{%s}", name, joinTags(tags))
}

type tagKSort []string

func (a tagKSort) Len() int      { return len(a) }
func (a tagKSort) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a tagKSort) Less(i, j int) bool {
	if a[i] == "host" {
		return true
	} else if a[j] == "host" {
		return false
	}
	return a[i] < a[j]
}

func joinTags(m map[string]string) string {
	keys := make(tagKSort, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Sort(keys)
	for i := range keys {
		keys[i] = fmt.Sprintf("%s=%s", keys[i], m[keys[i]])
	}
	return strings.Join([]string(keys), ",")
}

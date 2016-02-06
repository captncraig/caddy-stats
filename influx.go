package stats

import (
	"fmt"

	client "github.com/influxdata/influxdb/client/v2"
)

type influxPublisher struct {
	url                string
	database           string
	username, password string
}

func (i *influxPublisher) SendData(pts []*Measurement) error {
	c, err := client.NewHTTPClient(client.HTTPConfig{
		Addr:     i.url,
		Username: i.username,
		Password: i.password,
	})

	bp, err := client.NewBatchPoints(client.BatchPointsConfig{
		Database:  i.database,
		Precision: "s",
	})
	if err != nil {
		return err
	}
	for _, pt := range pts {
		vals := map[string]interface{}{}
		for k, v := range pt.Values {
			vals[k] = v
		}
		dp, err := client.NewPoint(pt.Name, pt.Tags, vals, pt.Timestamp)
		if err != nil {
			return err
		}
		bp.AddPoint(dp)
	}
	err = c.Write(bp)
	fmt.Println(err, len(pts), "FLLLLLL")
	return err
}

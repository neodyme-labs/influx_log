package influx_log

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/caddyserver/caddy/v2"
	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"github.com/influxdata/influxdb-client-go/v2/api"
	"go.uber.org/zap"
)

func init() {
	caddy.RegisterModule(InfluxLog{})
}

type InfluxLog struct {
	Host        string            `json:"host,omitempty"`
	Token       string            `json:"token,omitempty"`
	Org         string            `json:"org,omitempty"`
	Bucket      string            `json:"bucket,omitempty"`
	Measurement string            `json:"measurement,omitempty"`
	Tags        map[string]string `json:"tags,omitempty"`

	logger *zap.Logger
}

// CaddyModule returns the Caddy module information.
func (InfluxLog) CaddyModule() caddy.ModuleInfo {
	return caddy.ModuleInfo{
		ID:  "caddy.logging.writers.influx_log",
		New: func() caddy.Module { return new(InfluxLog) },
	}
}

func (l *InfluxLog) String() string {
	return "influx_log"
}

func (l *InfluxLog) WriterKey() string {
	return "influx_log"
}

func (l *InfluxLog) OpenWriter() (io.WriteCloser, error) {
	writer := &InfluxWriter{
		logger: l.logger,
	}

	go func() {
		writer.Open(l)
	}()

	return writer, nil
}

func (l *InfluxLog) Provision(ctx caddy.Context) error {
	l.logger = ctx.Logger(l)

	return nil
}

func (l *InfluxLog) Validate() error {
	if l.Host == "" {
		return fmt.Errorf("NO HOST SET")
	}

	if l.Token == "" {
		return fmt.Errorf("NO TOKEN SET")
	}

	if l.Org == "" {
		return fmt.Errorf("NO ORG SET")
	}

	if l.Bucket == "" {
		return fmt.Errorf("NO BUCKET SET")
	}

	if l.Measurement == "" {
		return fmt.Errorf("NO Measurement SET")
	}

	if l.Tags == nil {
		return fmt.Errorf("NO Tags SET")
	}

	return nil
}

func flatten(m map[string]interface{}, fields map[string]interface{}, prefix string) map[string]interface{} {
	for k, v := range m {
		key := prefix + k

		if v2, ok := v.(map[string]interface{}); ok {
			flatten(v2, fields, key+"_")
		} else {
			fields[key] = v
		}
	}
	return m
}

type InfluxWriter struct {
	logger      *zap.Logger
	measurement string
	tags        map[string]string
	client      influxdb2.Client
	writeAPI    api.WriteAPI
}

func (prom *InfluxWriter) Write(p []byte) (n int, err error) {
	f := map[string]interface{}{}
	if err := json.Unmarshal(p, &f); err != nil {
		prom.logger.Error("Unmarshal failed on log", zap.Error((err)))
	}

	fields := map[string]interface{}{}
	flatten(f, fields, "")

	tags := map[string]string{}
	for key, element := range prom.tags {
		val := element
		if strings.HasPrefix(element, "{") && strings.HasSuffix(element, "}") {
			if v, ok := fields[element[1:len(element)-1]]; ok {
				val = v.(string)
			}
		}

		tags[key] = val
	}

	point := influxdb2.NewPoint(
		prom.measurement,
		tags,
		fields,
		time.Now())
	prom.writeAPI.WritePoint(point)

	return
}

func (prom *InfluxWriter) Close() error {
	prom.writeAPI.Flush()
	prom.client.Close()
	return nil
}

func (prom *InfluxWriter) Open(i *InfluxLog) error {
	client := influxdb2.NewClient(i.Host, i.Token)
	writeAPI := client.WriteAPI(i.Org, i.Bucket)

	prom.client = client
	prom.writeAPI = writeAPI
	prom.measurement = i.Measurement
	prom.tags = i.Tags

	return nil
}

// Interface guards.
var (
	_ caddy.Provisioner  = (*InfluxLog)(nil)
	_ caddy.Validator    = (*InfluxLog)(nil)
	_ caddy.WriterOpener = (*InfluxLog)(nil)
)

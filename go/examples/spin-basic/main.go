package main

import (
	"context"
	"fmt"
	"net/http"

	wasiMetrics "github.com/calebschoepp/opentelemetry-wasi/metrics"
	spinhttp "github.com/spinframework/spin-go-sdk/v3/http"
	"github.com/spinframework/spin-go-sdk/v3/kv"
	"go.opentelemetry.io/otel/attribute"
	api "go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/sdk/metric"
)

func init() {
	spinhttp.Handle(func(w http.ResponseWriter, r *http.Request) {
		ctx := context.Background()

		exporter := wasiMetrics.NewWasiMetricExporter()
		defer exporter.Export(ctx) // Export metrics to the host

		meterProvider := metric.NewMeterProvider(metric.WithReader(exporter))
		meter := meterProvider.Meter("spin-metrics")

		attrs := api.WithAttributes(
			attribute.Key("spinkey1").String("spinvalue1"),
			attribute.Key("spinkey2").String("spinvalue2"),
		)

		counter, err := meter.Int64Counter("spin-counter")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		counter.Add(ctx, 10, attrs)

		upDownCounter, err := meter.Int64UpDownCounter("spin-up-down-counter")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		upDownCounter.Add(ctx, -1, attrs)

		histogram, err := meter.Int64Histogram("spin-histogram")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		histogram.Record(ctx, 9, attrs)
		histogram.Record(ctx, 15, attrs)

		gauge, err := meter.Float64Gauge("spin-gauge")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		gauge.Record(ctx, 123.456, attrs)

		store, err := kv.OpenDefault()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if err := store.Set("foo", []byte("bar")); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "text/plain")
		fmt.Fprintln(w, "Hello World!")
	})
}

func main() {}

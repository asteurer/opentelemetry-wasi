package metrics

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/calebschoepp/opentelemetry-wasi/wit_component/wasi_clocks_wall_clock"
	"github.com/calebschoepp/opentelemetry-wasi/wit_component/wasi_otel_metrics"
	"github.com/calebschoepp/opentelemetry-wasi/wit_component/wasi_otel_types"
	"github.com/calebschoepp/opentelemetry-wasi/wit_component/wit_types"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/sdk/instrumentation"
	"go.opentelemetry.io/otel/sdk/metric/metricdata"
)

func toWasiResourceMetrics(rm metricdata.ResourceMetrics) wasi_otel_metrics.ResourceMetrics {
	return wasi_otel_metrics.ResourceMetrics{
		Resource: wasi_otel_types.Resource{
			Attributes: toWasiAttributes(rm.Resource.Attributes()),
			SchemaUrl:  toWasiOptStr(rm.Resource.SchemaURL()),
		},
		ScopeMetrics: toWasiScopeMetrics(rm.ScopeMetrics),
	}
}

func toWasiAttributes(attrs []attribute.KeyValue) []wasi_otel_types.KeyValue {
	result := make([]wasi_otel_types.KeyValue, len(attrs))
	for i, attr := range attrs {
		result[i] = wasi_otel_types.KeyValue{
			Key:   string(attr.Key),
			Value: otelValueToJson(attr.Value),
		}

	}

	return result
}

func otelValueToJson(v attribute.Value) string {
	switch v.Type() {
	case attribute.BOOL:
		return fmt.Sprint(v.AsBool())
	case attribute.BOOLSLICE:
		bytes, err := json.Marshal(v.AsBoolSlice())
		if err != nil {
			panic(err)
		}
		return string(bytes)
	case attribute.INT64:
		return fmt.Sprint(v.AsInt64())
	case attribute.INT64SLICE:
		bytes, err := json.Marshal(v.AsInt64Slice())
		if err != nil {
			panic(err)
		}
		return string(bytes)
	case attribute.FLOAT64:
		return fmt.Sprint(v.AsFloat64())
	case attribute.FLOAT64SLICE:
		bytes, err := json.Marshal(v.AsFloat64Slice())
		if err != nil {
			panic(err)
		}
		return string(bytes)
	case attribute.STRING:
		bytes, err := json.Marshal(v.AsString())
		if err != nil {
			panic(err)
		}
		return string(bytes)
	case attribute.STRINGSLICE:
		bytes, err := json.Marshal(v.AsStringSlice())
		if err != nil {
			panic(err)
		}
		return string(bytes)
	case attribute.INVALID:
		panic("invalid type")
	default:
		panic("unsupported type")
	}
}

func toWasiOptStr(s string) wit_types.Option[string] {
	if s == "" {
		return wit_types.None[string]()
	}

	return wit_types.Some(s)
}

func toWasiScopeMetrics(sm []metricdata.ScopeMetrics) []wasi_otel_metrics.ScopeMetrics {
	result := make([]wasi_otel_metrics.ScopeMetrics, len(sm))

	for i, m := range sm {
		result[i] = wasi_otel_metrics.ScopeMetrics{
			Scope:   toWasiInstrumentationScope(m.Scope),
			Metrics: toWasiMetrics(m.Metrics),
		}
	}

	return result
}

func toWasiInstrumentationScope(s instrumentation.Scope) wasi_otel_types.InstrumentationScope {
	return wasi_otel_types.InstrumentationScope{
		Name:       s.Name,
		Version:    toWasiOptStr(s.Version),
		SchemaUrl:  toWasiOptStr(s.SchemaURL),
		Attributes: toWasiAttributes(s.Attributes.ToSlice()),
	}
}

func toWasiMetrics(metrics []metricdata.Metrics) []wasi_otel_metrics.Metric {
	result := make([]wasi_otel_metrics.Metric, len(metrics))
	for i, m := range metrics {
		result[i] = wasi_otel_metrics.Metric{
			Name:        m.Name,
			Description: m.Description,
			Unit:        m.Unit,
			Data:        toWasiMetricData(m.Data),
		}
	}

	return result
}

func toWasiMetricData(data metricdata.Aggregation) wasi_otel_metrics.MetricData {
	switch v := data.(type) {
	case metricdata.Gauge[int64]:
		var startOpt wit_types.Option[wasi_clocks_wall_clock.Datetime]
		start, timeVal := extractTimestamps(v.DataPoints)
		if start == nil {
			startOpt = wit_types.None[wasi_clocks_wall_clock.Datetime]()
		} else {
			startOpt = wit_types.Some(*start)
		}

		return wasi_otel_metrics.MakeMetricDataS64Gauge(wasi_otel_metrics.Gauge{
			DataPoints: toWasiGaugeDataPoint(v.DataPoints),
			StartTime:  startOpt,
			Time:       timeVal,
		})
	case metricdata.Gauge[float64]:
		var startOpt wit_types.Option[wasi_clocks_wall_clock.Datetime]
		start, timeVal := extractTimestamps(v.DataPoints)
		if start == nil {
			startOpt = wit_types.None[wasi_clocks_wall_clock.Datetime]()
		} else {
			startOpt = wit_types.Some(*start)
		}
		return wasi_otel_metrics.MakeMetricDataF64Gauge(wasi_otel_metrics.Gauge{
			DataPoints: toWasiGaugeDataPoint(v.DataPoints),
			StartTime:  startOpt,
			Time:       timeVal,
		})
	case metricdata.Sum[int64]:
		start, timeVal := extractTimestamps(v.DataPoints)
		return wasi_otel_metrics.MakeMetricDataS64Sum(wasi_otel_metrics.Sum{
			DataPoints:  toWasiSumDataPoint(v.DataPoints),
			StartTime:   *start,
			Time:        timeVal,
			Temporality: toWasiTemporality(v.Temporality),
			IsMonotonic: v.IsMonotonic,
		})
	case metricdata.Sum[float64]:
		start, timeVal := extractTimestamps(v.DataPoints)
		return wasi_otel_metrics.MakeMetricDataF64Sum(wasi_otel_metrics.Sum{
			DataPoints:  toWasiSumDataPoint(v.DataPoints),
			StartTime:   *start,
			Time:        timeVal,
			Temporality: toWasiTemporality(v.Temporality),
			IsMonotonic: v.IsMonotonic,
		})
	case metricdata.ExponentialHistogram[int64]:
		start, timeVal := extractTimestamps(v.DataPoints)
		return wasi_otel_metrics.MakeMetricDataS64ExponentialHistogram(wasi_otel_metrics.ExponentialHistogram{
			DataPoints:  toWasiExponentialHistogram(v.DataPoints),
			StartTime:   *start,
			Time:        timeVal,
			Temporality: toWasiTemporality(v.Temporality),
		})
	case metricdata.ExponentialHistogram[float64]:
		start, timeVal := extractTimestamps(v.DataPoints)
		return wasi_otel_metrics.MakeMetricDataF64ExponentialHistogram(wasi_otel_metrics.ExponentialHistogram{
			DataPoints:  toWasiExponentialHistogram(v.DataPoints),
			StartTime:   *start,
			Time:        timeVal,
			Temporality: toWasiTemporality(v.Temporality),
		})
	case metricdata.Histogram[int64]:
		start, timeVal := extractTimestamps(v.DataPoints)
		return wasi_otel_metrics.MakeMetricDataS64Histogram(wasi_otel_metrics.Histogram{
			DataPoints:  toWasiHistogramDataPoint(v.DataPoints),
			StartTime:   *start,
			Time:        timeVal,
			Temporality: toWasiTemporality(v.Temporality),
		})
	case metricdata.Histogram[float64]:
		start, timeVal := extractTimestamps(v.DataPoints)
		return wasi_otel_metrics.MakeMetricDataF64Histogram(wasi_otel_metrics.Histogram{
			DataPoints:  toWasiHistogramDataPoint(v.DataPoints),
			StartTime:   *start,
			Time:        timeVal,
			Temporality: toWasiTemporality(v.Temporality),
		})
	case metricdata.Summary:
		panic("The metricdata.Summary metric type is not implemented")
	default:
		panic("unimplemented type")
	}
}

func toWasiGaugeDataPoint[T float64 | int64](dataPoints []metricdata.DataPoint[T]) []wasi_otel_metrics.GaugeDataPoint {
	result := make([]wasi_otel_metrics.GaugeDataPoint, len(dataPoints))
	for i, dp := range dataPoints {
		result[i] = wasi_otel_metrics.GaugeDataPoint{
			Attributes: toWasiAttributes(dp.Attributes.ToSlice()),
			Value:      toWasiMetricNumber(dp.Value),
			Exemplars:  toWasiExemplar(dp.Exemplars),
		}
	}

	return result
}

func toWasiSumDataPoint[T float64 | int64](dataPoints []metricdata.DataPoint[T]) []wasi_otel_metrics.SumDataPoint {
	result := make([]wasi_otel_metrics.SumDataPoint, len(dataPoints))
	for i, dp := range dataPoints {
		result[i] = wasi_otel_metrics.SumDataPoint{
			Attributes: toWasiAttributes(dp.Attributes.ToSlice()),
			Value:      toWasiMetricNumber(dp.Value),
			Exemplars:  toWasiExemplar(dp.Exemplars),
		}
	}

	return result
}

func toWasiHistogramDataPoint[T float64 | int64](dataPoints []metricdata.HistogramDataPoint[T]) []wasi_otel_metrics.HistogramDataPoint {
	result := make([]wasi_otel_metrics.HistogramDataPoint, len(dataPoints))
	for i, dp := range dataPoints {
		result[i] = wasi_otel_metrics.HistogramDataPoint{
			Attributes:   toWasiAttributes(dp.Attributes.ToSlice()),
			Count:        dp.Count,
			Bounds:       dp.Bounds,
			BucketCounts: dp.BucketCounts,
			Min:          toWasiOptMetricNumber(dp.Min),
			Max:          toWasiOptMetricNumber(dp.Max),
			Sum:          toWasiMetricNumber(dp.Sum),
			Exemplars:    toWasiExemplar(dp.Exemplars),
		}
	}

	return result
}

func toWasiExponentialHistogram[T float64 | int64](dataPoints []metricdata.ExponentialHistogramDataPoint[T]) []wasi_otel_metrics.ExponentialHistogramDataPoint {
	result := make([]wasi_otel_metrics.ExponentialHistogramDataPoint, len(dataPoints))
	for i, dp := range dataPoints {
		result[i] = wasi_otel_metrics.ExponentialHistogramDataPoint{
			Attributes: toWasiAttributes(dp.Attributes.ToSlice()),
			Count:      dp.Count,
			Min:        toWasiOptMetricNumber(dp.Min),
			Max:        toWasiOptMetricNumber(dp.Max),
			Sum:        toWasiMetricNumber(dp.Sum),
			Scale:      int8(dp.Scale),
			ZeroCount:  dp.ZeroCount,
			PositiveBucket: wasi_otel_metrics.ExponentialBucket{
				Offset: dp.PositiveBucket.Offset,
				Counts: dp.PositiveBucket.Counts,
			},
			NegativeBucket: wasi_otel_metrics.ExponentialBucket{
				Offset: dp.NegativeBucket.Offset,
				Counts: dp.NegativeBucket.Counts,
			},
			ZeroThreshold: dp.ZeroThreshold,
			Exemplars:     toWasiExemplar(dp.Exemplars),
		}
	}

	return result
}

func toWasiOptMetricNumber[T float64 | int64](n metricdata.Extrema[T]) wit_types.Option[wasi_otel_metrics.MetricNumber] {
	num, exists := n.Value()
	if exists {
		return wit_types.Some(toWasiMetricNumber(num))
	} else {
		return wit_types.None[wasi_otel_metrics.MetricNumber]()
	}
}

func toWasiTemporality(t metricdata.Temporality) wasi_otel_metrics.Temporality {
	switch t {
	case metricdata.CumulativeTemporality:
		return wasi_otel_metrics.TemporalityCumulative
	case metricdata.DeltaTemporality:
		return wasi_otel_metrics.TemporalityDelta
	default:
		return wasi_otel_metrics.TemporalityCumulative
	}
}

func toWasiMetricNumber[T float64 | int64](n T) wasi_otel_metrics.MetricNumber {
	switch v := any(n).(type) {
	case int64:
		return wasi_otel_metrics.MakeMetricNumberS64(v)
	case float64:
		return wasi_otel_metrics.MakeMetricNumberF64(v)
	default:
		panic("unsupported type")
	}
}

func toWasiExemplar[T float64 | int64](exemplars []metricdata.Exemplar[T]) []wasi_otel_metrics.Exemplar {
	result := make([]wasi_otel_metrics.Exemplar, len(exemplars))
	for i, e := range exemplars {
		result[i] = wasi_otel_metrics.Exemplar{
			FilteredAttributes: toWasiAttributes(e.FilteredAttributes),
			Time:               toWasiTime(e.Time),
			Value:              toWasiMetricNumber(e.Value),
			SpanId:             string(e.SpanID),
			TraceId:            string(e.TraceID),
		}
	}

	return result
}

func toWasiTime(t time.Time) wasi_clocks_wall_clock.Datetime {
	return wasi_clocks_wall_clock.Datetime{
		Seconds:     uint64(t.Unix()),
		Nanoseconds: uint32(t.Nanosecond()),
	}
}

// timestampProvider is a constraint for types that have StartTime and Time fields
type timestampProvider interface {
	metricdata.DataPoint[int64] | metricdata.DataPoint[float64] |
		metricdata.HistogramDataPoint[int64] | metricdata.HistogramDataPoint[float64] |
		metricdata.ExponentialHistogramDataPoint[int64] | metricdata.ExponentialHistogramDataPoint[float64]
}

// extractTimestamps extracts StartTime and Time from the first data point in a slice
func extractTimestamps[T timestampProvider](dataPoints []T) (startTime *wasi_clocks_wall_clock.Datetime, timeRecorded wasi_clocks_wall_clock.Datetime) {
	if len(dataPoints) == 0 {
		return nil, toWasiTime(time.Now())
	}

	// Use type assertion to access the fields
	var start, timeVal time.Time
	switch dp := any(dataPoints[0]).(type) {
	case metricdata.DataPoint[int64]:
		start, timeVal = dp.StartTime, dp.Time
	case metricdata.DataPoint[float64]:
		start, timeVal = dp.StartTime, dp.Time
	case metricdata.HistogramDataPoint[int64]:
		start, timeVal = dp.StartTime, dp.Time
	case metricdata.HistogramDataPoint[float64]:
		start, timeVal = dp.StartTime, dp.Time
	case metricdata.ExponentialHistogramDataPoint[int64]:
		start, timeVal = dp.StartTime, dp.Time
	case metricdata.ExponentialHistogramDataPoint[float64]:
		start, timeVal = dp.StartTime, dp.Time
	default:
		panic("unsupported types")
	}

	resultStart := toWasiTime(start)

	return &resultStart, toWasiTime(timeVal)
}

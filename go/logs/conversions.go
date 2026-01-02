package logs

import (
	"encoding/base64"
	"encoding/json"
	"fmt"

	"github.com/calebschoepp/opentelemetry-wasi/types"
	"github.com/calebschoepp/opentelemetry-wasi/wit_component/wasi_clocks_wall_clock"
	"github.com/calebschoepp/opentelemetry-wasi/wit_component/wasi_otel_logs"
	"github.com/calebschoepp/opentelemetry-wasi/wit_component/wasi_otel_tracing"
	"github.com/calebschoepp/opentelemetry-wasi/wit_component/wasi_otel_types"
	"github.com/calebschoepp/opentelemetry-wasi/wit_component/wit_types"
	logApi "go.opentelemetry.io/otel/log"
	"go.opentelemetry.io/otel/sdk/log"
)

func toWasiLogRecord(r log.Record) wasi_otel_logs.LogRecord {
	var ts wit_types.Option[wasi_otel_logs.Datetime]
	if r.Timestamp().IsZero() {
		ts = wit_types.None[wasi_clocks_wall_clock.Datetime]()
	} else {
		ts = wit_types.Some(types.ToWasiTime(r.Timestamp()))
	}

	var ots wit_types.Option[wasi_otel_logs.Datetime]
	if r.ObservedTimestamp().IsZero() {
		ots = wit_types.None[wasi_clocks_wall_clock.Datetime]()
	} else {
		ots = wit_types.Some(types.ToWasiTime(r.ObservedTimestamp()))
	}

	var sn wit_types.Option[uint8]
	if r.Severity() == logApi.SeverityUndefined {
		sn = wit_types.None[uint8]()
	} else {
		sn = wit_types.Some(uint8(r.Severity()))
	}

	var st wit_types.Option[string]
	if r.SeverityText() == "" {
		st = wit_types.None[string]()
	} else {
		st = wit_types.Some(r.SeverityText())
	}

	var attrs wit_types.Option[[]wasi_otel_types.KeyValue]
	if r.AttributesLen() == 0 {
		attrs = wit_types.None[[]wasi_otel_types.KeyValue]()
	} else {
		attrList := make([]wasi_otel_types.KeyValue, r.AttributesLen())
		r.WalkAttributes(func(attr logApi.KeyValue) bool {
			attrList = append(attrList, wasi_otel_types.KeyValue{
				Key:   attr.Key,
				Value: otelLogValueToJson(attr.Value),
			})

			return true
		})

		attrs = wit_types.Some(attrList)
	}

	var res wit_types.Option[wasi_otel_types.Resource]
	if r.Resource() == nil {
		res = wit_types.None[wasi_otel_logs.Resource]()
	} else {
		res = wit_types.Some(types.ToWasiResource(*r.Resource()))
	}

	var is wit_types.Option[wasi_otel_types.InstrumentationScope]
	if r.InstrumentationScope().Name == "" {
		is = wit_types.None[wasi_otel_logs.InstrumentationScope]()
	} else {
		is = wit_types.Some(types.ToWasiInstrumentationScope(r.InstrumentationScope()))
	}

	var tf wit_types.Option[wasi_otel_tracing.TraceFlags]
	if !r.TraceFlags().IsSampled() {
		tf = wit_types.None[wasi_otel_tracing.TraceFlags]()
	} else {
		tf = wit_types.Some(wasi_otel_tracing.TraceFlagsSampled)
	}

	return wasi_otel_logs.LogRecord{
		Timestamp:            ts,
		ObservedTimestamp:    ots,
		SeverityNumber:       sn,
		SeverityText:         st,
		Body:                 types.ToWasiOptStr(otelLogValueToJson(r.Body())),
		Attributes:           attrs,
		EventName:            types.ToWasiOptStr(r.EventName()),
		Resource:             res,
		InstrumentationScope: is,
		TraceId:              types.ToWasiOptStr(r.TraceID().String()),
		SpanId:               types.ToWasiOptStr(r.SpanID().String()),
		TraceFlags:           tf,
	}
}

func otelLogValueToJson(v logApi.Value) string {
	switch v.Kind() {
	case logApi.KindBool:
		bytes, _ := json.Marshal(v.AsBool())
		return string(bytes)
	case logApi.KindFloat64:
		bytes, _ := json.Marshal(v.AsFloat64())
		return string(bytes)
	case logApi.KindInt64:
		bytes, _ := json.Marshal(v.AsInt64())
		return string(bytes)
	case logApi.KindEmpty:
		bytes, _ := json.Marshal("")
		return string(bytes)
	case logApi.KindBytes:
		bytes, _ := json.Marshal(fmt.Sprintf("{{base64}}:%s", base64.StdEncoding.EncodeToString(v.AsBytes())))
		return string(bytes)
	case logApi.KindString:
		bytes, _ := json.Marshal(v.AsString())
		return string(bytes)
	case logApi.KindSlice:
		bytes, _ := json.Marshal(v.AsSlice())
		return string(bytes)
	case logApi.KindMap:
		bytes, _ := json.Marshal(v.AsMap())
		return string(bytes)
	default:
		panic("unsupported type")
	}
}

package pkg

type TestCase struct {
	Path string    `json:"path" yaml:"path"`
	Data DataTypes `json:"data" yaml:"data"`
}

type DataTypes struct {
	Metrics []ResourceMetric `json:"metrics" yaml:"metrics"`
	Logs    struct{}         // TODO
	Traces  struct{}         // TODO
}

// This is a wrapper to help parse the JSON from the collector
type MetricsPayload struct {
	ResourceMetrics []ResourceMetric `json:"resourceMetrics" yaml:"resourceMetrics"`
}

type ResourceMetric struct {
	Resource     Resource       `json:"resource" yaml:"resource"`
	ScopeMetrics []ScopeMetrics `json:"scopeMetrics" yaml:"scopeMetrics"`
}

type Resource struct {
	Attributes []KeyValue `json:"attributes" yaml:"attributes"`
	// There are intentionally omitted fields from this struct
}

type ScopeMetrics struct {
	Scope   Scope    `json:"scope" yaml:"scope"`
	Metrics []Metric `json:"metrics" yaml:"metrics"`
}

type Scope struct {
	Name string `json:"name" yaml:"name"`
}

type Metric struct {
	Name      string     `json:"name" yaml:"name"`
	Sum       *Sum       `json:"sum,omitempty" yaml:"sum,omitempty"`
	Histogram *Histogram `json:"histogram,omitempty" yaml:"histogram,omitempty"`
	Gauge     *Gauge     `json:"gauge,omitempty" yaml:"gauge,omitempty"`
}

type Sum struct {
	DataPoints  []DataPoint `json:"dataPoints" yaml:"dataPoints"`
	IsMonotonic bool        `json:"isMonotonic,omitempty" yaml:"isMonotonic,omitempty"`
	// There are intentionally omitted fields from this struct
}

type Histogram struct {
	DataPoints []HistogramDataPoint `json:"dataPoints" yaml:"dataPoints"`
	// There are intentionally omitted fields from this struct
}

type Gauge struct {
	DataPoints []DataPoint `json:"dataPoints" yaml:"dataPoints"`
	// There are intentionally omitted fields from this struct
}

type DataPoint struct {
	Attributes []KeyValue `json:"attributes" yaml:"attributes"`
	AsInt      string     `json:"asInt,omitempty" yaml:"asInt,omitempty"`
	// There are intentionally omitted fields from this struct
}

type HistogramDataPoint struct {
	Attributes []KeyValue `json:"attributes" yaml:"attributes"`
	Count      string     `json:"count" yaml:"count"`
	Sum        float64    `json:"sum" yaml:"sum"`
	Min        float64    `json:"min" yaml:"min"`
	Max        float64    `json:"max" yaml:"max"`
	// There are intentionally omitted fields from this struct
}

type KeyValue struct {
	Key   string     `json:"key" yaml:"key"`
	Value ValueField `json:"value" yaml:"value"`
}

type ValueField struct {
	StringValue string `json:"stringValue,omitempty" yaml:"stringValue,omitempty"`
}

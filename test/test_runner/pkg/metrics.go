package pkg

import (
	"encoding/json"
	"fmt"
	"os"
)

func TestMetrics(expected []ResourceMetric) error {
	// Note: This test assumes that the collector will export only one line of JSON. If there are issues
	// with marshalling the JSON, double check that the 'metrics.jsonl' file for multiple lines.
	fileData, err := os.ReadFile("metrics.jsonl")
	if err != nil {
		return fmt.Errorf("Failed to read file: %v", err)
	}

	var payload MetricsPayload
	if err := json.Unmarshal(fileData, &payload); err != nil {
		return fmt.Errorf("Failed to unmarshal JSON: %v", err)
	}
	actual := payload.ResourceMetrics

	if len(expected) != len(actual) {
		return fmt.Errorf("Metrics count mismatch: num expected metrics %d, num actual metrics %d", len(expected), len(actual))
	}

	for i, expRM := range expected {
		actRM := actual[i]

		// Compare Resource Attributes
		if err := compareAttributes("Resource", expRM.Resource.Attributes, actRM.Resource.Attributes); err != nil {
			return fmt.Errorf("ResourceMetric[%d]: %v", i, err)
		}

		// Compare ScopeMetrics
		if len(expRM.ScopeMetrics) != len(actRM.ScopeMetrics) {
			return fmt.Errorf("ResourceMetric[%d]: ScopeMetrics count mismatch: expected %d, got %d",
				i, len(expRM.ScopeMetrics), len(actRM.ScopeMetrics))
		}

		actSM := actRM.ScopeMetrics[0]
		expSM := expRM.ScopeMetrics[0]

		if expSM.Scope.Name != actSM.Scope.Name {
			return fmt.Errorf("Scope name mismatch: expected %q, got %q",
				expSM.Scope.Name, actSM.Scope.Name)
		}

		actMetricMap := make(map[string]Metric)
		for _, m := range actSM.Metrics {
			actMetricMap[m.Name] = m
		}

		expMetricMap := make(map[string]Metric)
		for _, m := range expSM.Metrics {
			expMetricMap[m.Name] = m
		}

		for metricName, expVal := range expMetricMap {
			actVal, exists := actMetricMap[metricName]
			if !exists {
				return fmt.Errorf("missing metric: %s", metricName)
			}

			if expVal.Gauge != nil && actVal.Gauge != nil {
				if err := compareDataPoints(metricName, expVal.Gauge.DataPoints, actVal.Gauge.DataPoints); err != nil {
					return err
				}
			} else if expVal.Sum != nil && actVal.Sum != nil {
				if expVal.Sum.IsMonotonic != actVal.Sum.IsMonotonic {
					return fmt.Errorf("IsMonotonic mismatch: expected value: %v, actual value: %v", expVal.Sum.IsMonotonic, actVal.Sum.IsMonotonic)
				}
				if err := compareDataPoints(metricName, expVal.Sum.DataPoints, actVal.Sum.DataPoints); err != nil {
					return err
				}
			} else if expVal.Histogram != nil && actVal.Histogram != nil {
				if err := compareHistogramDataPoints(metricName, expVal.Histogram.DataPoints, actVal.Histogram.DataPoints); err != nil {
					return err
				}
			} else {
				return fmt.Errorf("both actual and expected instruments are null")
			}
		}
	}
	return nil
}

func compareAttributes(source string, expected, actual []KeyValue) error {
	// Note: This function is designed to only evaluate the expected attributes, which
	// may end up being fewer in number than the actual attributes.
	if len(actual) < len(expected) {
		return fmt.Errorf("%s, Missing attributes: num expected attrs %d, num actual attrs %d",
			source, len(expected), len(actual))
	}

	expMap := make(map[string]string)
	for _, kv := range expected {
		expMap[kv.Key] = kv.Value.StringValue
	}

	actMap := make(map[string]string)
	for _, kv := range actual {
		actMap[kv.Key] = kv.Value.StringValue
	}

	for key, expVal := range expMap {
		actVal, exists := actMap[key]
		if !exists {
			return fmt.Errorf("%s: Missing attribute key %q", source, key)
		}
		if expVal != actVal {
			return fmt.Errorf("%s: Attribute %q value mismatch: expected %q, got %q",
				source, key, expVal, actVal)
		}
	}

	return nil
}

func compareDataPoints(source string, expected, actual []DataPoint) error {
	if len(expected) != len(actual) {
		return fmt.Errorf("%s: DataPoints count mismatch: expected %d, got %d",
			source, len(expected), len(actual))
	}

	for i, expDP := range expected {
		actDP := actual[i]

		if err := compareAttributes(fmt.Sprintf("%s.DataPoints[%d]", source, i),
			expDP.Attributes, actDP.Attributes); err != nil {
			return err
		}

		if expDP.AsInt != actDP.AsInt {
			return fmt.Errorf("%s.DataPoints[%d]: AsInt mismatch: expected %q, got %q",
				source, i, expDP.AsInt, actDP.AsInt)
		}
	}

	return nil
}

func compareHistogramDataPoints(source string, expected, actual []HistogramDataPoint) error {
	if len(expected) != len(actual) {
		return fmt.Errorf("%s: DataPoints count mismatch: expected %d, got %d",
			source, len(expected), len(actual))
	}

	for i, expDP := range expected {
		actDP := actual[i]

		if err := compareAttributes(fmt.Sprintf("%s.DataPoints[%d]", source, i),
			expDP.Attributes, actDP.Attributes); err != nil {
			return err
		}

		if expDP.Count != actDP.Count {
			return fmt.Errorf("%s.DataPoints[%d]: Count mismatch: expected %q, got %q",
				source, i, expDP.Count, actDP.Count)
		}

		if expDP.Sum != actDP.Sum {
			return fmt.Errorf("%s.DataPoints[%d]: Sum mismatch: expected %f, got %f",
				source, i, expDP.Sum, actDP.Sum)
		}

		if expDP.Min != actDP.Min {
			return fmt.Errorf("%s.DataPoints[%d]: Min mismatch: expected %f, got %f",
				source, i, expDP.Min, actDP.Min)
		}

		if expDP.Max != actDP.Max {
			return fmt.Errorf("%s.DataPoints[%d]: Max mismatch: expected %f, got %f",
				source, i, expDP.Max, actDP.Max)
		}
	}

	return nil
}

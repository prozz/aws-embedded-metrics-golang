// Package emf implements the spec available here: https://docs.aws.amazon.com/AmazonCloudWatch/latest/monitoring/CloudWatch_Embedded_Metric_Format_Specification.html
package emf

// Metadata struct as defined in AWS Embedded Metrics Format spec.
type Metadata struct {
	Timestamp int64             `json:"Timestamp"`
	Metrics   []MetricDirective `json:"CloudWatchMetrics"`
}

// MetricDirective struct as defined in AWS Embedded Metrics Format spec.
type MetricDirective struct {
	Namespace  string             `json:"Namespace"`
	Dimensions []DimensionSet     `json:"Dimensions"`
	Metrics    []MetricDefinition `json:"Metrics"`
}

// DimensionSet as defined in AWS Embedded Metrics Format spec.
type DimensionSet []string

// MetricDefinition struct as defined in AWS Embedded Metrics Format spec.
type MetricDefinition struct {
	Name string     `json:"Name"`
	Unit MetricUnit `json:"Unit,omitempty"`
}

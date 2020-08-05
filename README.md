# aws-embedded-metrics-golang

![test](https://github.com/prozz/aws-embedded-metrics-golang/workflows/test/badge.svg?branch=master)
![golangci-lint](https://github.com/prozz/aws-embedded-metrics-golang/workflows/lint/badge.svg?branch=master)

Go implementation of AWS CloudWatch [Embedded Metric Format](https://docs.aws.amazon.com/AmazonCloudWatch/latest/monitoring/CloudWatch_Embedded_Metric_Format_Specification.html)

It's aim is to simplify reporting metrics to CloudWatch:
- using EMF avoids additional HTTP API calls to CloudWatch as metrics are logged in JSON format to stdout
- no need for additional dependencies in your services (or mocks in tests) to report metrics from inside your code
- built in support for default dimensions and properties for Lambda functions
- TODO support for default dimensions and properties for EC2 (please send pull requests)

Supports namespaces, setting dimensions and properties as well as different contexts (at least partially).

Usage:
```
emf.New().Namespace("mtg").Metric("totalWins", 1500).Log()

emf.New().Dimension("colour", "red").
    MetricAs("gameLength", 2, emf.Seconds).Log()

emf.New().DimensionSet(
        emf.NewDimension("format", "edh"), 
        emf.NewDimension("commander", "Muldrotha")).
    MetricAs("wins", 1499, emf.Count).Log()
```

You may also use the lib together with `defer`.
```
m := emf.New() // sets up whatever you fancy here
defer m.Log()

// any reporting metrics calls
```

Functions for reporting metrics:
```
func Metric(name string, value int)
func Metrics(m map[string]int)
func MetricAs(name string, value int, unit MetricUnit)
func MetricsAs(m map[string]int, unit MetricUnit)

func MetricFloat(name string, value float64)
func MetricsFloat(m map[string]float64)
func MetricFloatAs(name string, value float64, unit MetricUnit)
func MetricsFloatAs(m map[string]float64, unit MetricUnit)
```

Functions for setting up dimensions:
```
func Dimension(key, value string)
func DimensionSet(dimensions ...Dimension) // use `func NewDimension` for creating one
```

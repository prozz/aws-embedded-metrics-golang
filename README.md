# aws-embedded-metrics-golang

Go implementation of AWS CloudWatch [Embedded Metric Format](https://docs.aws.amazon.com/AmazonCloudWatch/latest/monitoring/CloudWatch_Embedded_Metric_Format_Specification.html)

It's aim is to simplify reporting metrics to CloudWatch:
- using EMF avoids additional HTTP calls as it just logs JSON to stdout
- built in support for Lambda functions default dimensions and properties (happy to accept PR for EC2 support)
- no need for defining additional dependencies (or mocks in tests) to report metrics from inside your code

Examples:
```
emf.New().Namespace("mtg").Metric("totalWins", 1500).Log()

emf.New().Dimension("colour", "red").
    MetricAs("gameLength", 2, emf.Seconds).Log()

emf.New().DimensionSet(emf.NewDimension("format", "edh"), emf.NewDimension("commander", "Muldrotha")).
    MetricAs("wins", 1499, emf.Count).Log()
```


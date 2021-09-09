package emf

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"time"
)

// Logger for metrics with default Context.
type Logger struct {
	out               io.Writer
	timestamp         int64
	defaultContext    Context
	contexts          []*Context
	values            map[string]interface{}
	withoutDimensions bool
}

// Context gives ability to add another MetricDirective section for Logger.
type Context struct {
	metricDirective MetricDirective
	values          map[string]interface{}
}

// LoggerOption defines a function that can be used to customize a logger.
type LoggerOption func(l *Logger)

// WithWriter customizes the writer used by a logger.
func WithWriter(w io.Writer) LoggerOption {
	return func(l *Logger) {
		l.out = w
	}
}

// WithTimestamp customizes the timestamp used by a logger.
func WithTimestamp(t time.Time) LoggerOption {
	return func(l *Logger) {
		l.timestamp = t.UnixNano() / int64(time.Millisecond)
	}
}

// WithoutDimensions ignores default AWS Lambda related properties and dimensions.
func WithoutDimensions() LoggerOption {
	return func(l *Logger) {
		l.withoutDimensions = true
	}
}

// New creates logger with reasonable defaults for Lambda functions:
// - Prints to os.Stdout.
// - Context based on Lambda environment variables.
// - Timestamp set to the time when New was called.
// Specify LoggerOptions to customize the logger.
func New(opts ...LoggerOption) *Logger {
	l := Logger{
		out:       os.Stdout,
		timestamp: time.Now().UnixNano() / int64(time.Millisecond),
	}

	// apply any options
	for _, opt := range opts {
		opt(&l)
	}

	values := make(map[string]interface{})

	if !l.withoutDimensions {
		// set default properties for lambda function
		fnName := os.Getenv("AWS_LAMBDA_FUNCTION_NAME")
		if fnName != "" {
			values["executionEnvironment"] = os.Getenv("AWS_EXECUTION_ENV")
			values["memorySize"] = os.Getenv("AWS_LAMBDA_FUNCTION_MEMORY_SIZE")
			values["functionVersion"] = os.Getenv("AWS_LAMBDA_FUNCTION_VERSION")
			values["logStreamId"] = os.Getenv("AWS_LAMBDA_LOG_STREAM_NAME")
		}
	}

	// only collect traces which have been sampled
	amznTraceID := os.Getenv("_X_AMZN_TRACE_ID")
	if strings.Contains(amznTraceID, "Sampled=1") {
		values["traceId"] = amznTraceID
	}

	l.values = values
	l.defaultContext = newContext(values, l.withoutDimensions)

	return &l
}

// Dimension helps builds DimensionSet.
type Dimension struct {
	Key, Value string
}

// NewDimension creates Dimension from key/value pair.
func NewDimension(key, value string) Dimension {
	return Dimension{
		Key:   key,
		Value: value,
	}
}

// Namespace sets namespace on default context.
func (l *Logger) Namespace(namespace string) *Logger {
	l.defaultContext.Namespace(namespace)
	return l
}

// Property sets property.
func (l *Logger) Property(key, value string) *Logger {
	l.values[key] = value
	return l
}

// Dimension adds single dimension on default context.
func (l *Logger) Dimension(key, value string) *Logger {
	l.defaultContext.metricDirective.Dimensions = append(
		l.defaultContext.metricDirective.Dimensions, DimensionSet{key})
	l.values[key] = value
	return l
}

// DimensionSet adds multiple dimensions on default context.
func (l *Logger) DimensionSet(dimensions ...Dimension) *Logger {
	var set DimensionSet
	for _, d := range dimensions {
		set = append(set, d.Key)
		l.values[d.Key] = d.Value
	}
	l.defaultContext.metricDirective.Dimensions = append(
		l.defaultContext.metricDirective.Dimensions, set)
	return l
}

// Metric puts int metric on default context.
func (l *Logger) Metric(name string, value int) *Logger {
	l.defaultContext.put(name, value, None)
	return l
}

// Metrics puts all of the int metrics on default context.
func (l *Logger) Metrics(m map[string]int) *Logger {
	return l.MetricsAs(m, None)
}

// MetricFloat puts float metric on default context.
func (l *Logger) MetricFloat(name string, value float64) *Logger {
	l.defaultContext.put(name, value, None)
	return l
}

// MetricsFloat puts all of the float metrics on default context.
func (l *Logger) MetricsFloat(m map[string]float64) *Logger {
	return l.MetricsFloatAs(m, None)
}

// MetricAs puts int metric with MetricUnit on default context.
func (l *Logger) MetricAs(name string, value int, unit MetricUnit) *Logger {
	l.defaultContext.put(name, value, unit)
	return l
}

// MetricsAs puts all of the int metrics with MetricUnit on default context.
func (l *Logger) MetricsAs(m map[string]int, unit MetricUnit) *Logger {
	for name, value := range m {
		l.defaultContext.put(name, value, unit)
	}
	return l
}

// MetricFloatAs puts float metric with MetricUnit on default context.
func (l *Logger) MetricFloatAs(name string, value float64, unit MetricUnit) *Logger {
	l.defaultContext.put(name, value, unit)
	return l
}

// MetricsFloatAs puts all of the float metrics with MetricUnit on default context.
func (l *Logger) MetricsFloatAs(m map[string]float64, unit MetricUnit) *Logger {
	for name, value := range m {
		l.defaultContext.put(name, value, unit)
	}
	return l
}

// Log prints all Contexts and metric values to chosen output in Embedded Metric Format.
func (l *Logger) Log() {
	var metrics []MetricDirective
	if len(l.defaultContext.metricDirective.Metrics) > 0 {
		metrics = append(metrics, l.defaultContext.metricDirective)
	}
	for _, v := range l.contexts {
		if len(v.metricDirective.Metrics) > 0 {
			metrics = append(metrics, v.metricDirective)
		}
	}

	if len(metrics) == 0 {
		return
	}

	l.values["_aws"] = Metadata{
		Timestamp: l.timestamp,
		Metrics:   metrics,
	}
	buf, _ := json.Marshal(l.values)
	_, _ = fmt.Fprintln(l.out, string(buf))
}

// NewContext creates new context for given logger.
func (l *Logger) NewContext() *Context {
	c := newContext(l.values, l.withoutDimensions)
	l.contexts = append(l.contexts, &c)
	return &c
}

// Namespace sets namespace on given context.
func (c *Context) Namespace(namespace string) *Context {
	c.metricDirective.Namespace = namespace
	return c
}

// Dimension adds single dimension on given context.
func (c *Context) Dimension(key, value string) *Context {
	c.metricDirective.Dimensions = append(c.metricDirective.Dimensions, DimensionSet{key})
	c.values[key] = value
	return c
}

// DimensionSet adds multiple dimensions on given context.
func (c *Context) DimensionSet(dimensions ...Dimension) *Context {
	var set DimensionSet
	for _, d := range dimensions {
		set = append(set, d.Key)
		c.values[d.Key] = d.Value
	}
	c.metricDirective.Dimensions = append(c.metricDirective.Dimensions, set)
	return c
}

// Metric puts int metric on given context.
func (c *Context) Metric(name string, value int) *Context {
	return c.put(name, value, None)
}

// MetricFloat puts float metric on given context.
func (c *Context) MetricFloat(name string, value float64) *Context {
	return c.put(name, value, None)
}

// MetricAs puts int metric with MetricUnit on given context.
func (c *Context) MetricAs(name string, value int, unit MetricUnit) *Context {
	return c.put(name, value, unit)
}

// MetricFloatAs puts float metric with MetricUnit on given context.
func (c *Context) MetricFloatAs(name string, value float64, unit MetricUnit) *Context {
	return c.put(name, value, unit)
}

func newContext(values map[string]interface{}, withoutDimensions bool) Context {
	var defaultDimensions []DimensionSet
	if !withoutDimensions {
		// set default dimensions for lambda function
		fnName := os.Getenv("AWS_LAMBDA_FUNCTION_NAME")
		if fnName != "" {
			defaultDimensions = []DimensionSet{{"ServiceName", "ServiceType"}}
			values["ServiceType"] = "AWS::Lambda::Function"
			values["ServiceName"] = fnName
		}
	}

	return Context{
		metricDirective: MetricDirective{
			Namespace:  "aws-embedded-metrics",
			Dimensions: defaultDimensions,
		},
		values: values,
	}
}

func (c *Context) put(name string, value interface{}, unit MetricUnit) *Context {
	c.metricDirective.Metrics = append(c.metricDirective.Metrics, MetricDefinition{
		Name: name,
		Unit: unit,
	})
	c.values[name] = value
	return c
}

package emf_test

import (
	"bytes"
	"io/ioutil"
	"os"
	"testing"

	"github.com/kinbiko/jsonassert"
	"github.com/prozz/aws-embedded-metrics-golang/emf"
)

func TestEmf(t *testing.T) {
	tcs := []struct {
		name     string
		env      map[string]string
		given    func(logger *emf.Logger)
		expected string
	}{
		{
			name: "default namespace, int metric",
			given: func(logger *emf.Logger) {
				logger.Metric("foo", 33)
			},
			expected: "testdata/1.json",
		},
		{
			name: "default namespace, float metric",
			given: func(logger *emf.Logger) {
				logger.MetricFloat("foo", 33.66)
			},
			expected: "testdata/2.json",
		},
		{
			name: "custom namespace, int and float metrics",
			given: func(logger *emf.Logger) {
				logger.Namespace("galaxy").MetricFloat("foo", 33.66).Metric("bar", 666)
			},
			expected: "testdata/3.json",
		},
		{
			name: "custom namespace, int and float metrics, custom units",
			given: func(logger *emf.Logger) {
				logger.Namespace("galaxy").
					MetricFloatAs("foo", 33.66, emf.Milliseconds).
					MetricAs("bar", 666, emf.Count)
			},
			expected: "testdata/4.json",
		},
		{
			name: "new context, default namespace, int and float metrics, custom units",
			given: func(logger *emf.Logger) {
				logger.NewContext().
					MetricFloatAs("foo", 33.66, emf.Milliseconds).
					MetricAs("bar", 666, emf.Count)
			},
			expected: "testdata/5.json",
		},
		{
			name: "new context, custom namespace, int and float metrics, custom units",
			given: func(logger *emf.Logger) {
				logger.NewContext().Namespace("galaxy").
					MetricFloatAs("foo", 33.66, emf.Bits).
					MetricAs("bar", 666, emf.BytesSecond)
			},
			expected: "testdata/6.json",
		},
		{
			name: "default and custom contexts, different metrics names",
			given: func(logger *emf.Logger) {
				logger.NewContext().MetricFloatAs("foo", 33.66, emf.Bits)
				logger.MetricAs("bar", 666, emf.BytesSecond)
			},
			expected: "testdata/7.json",
		},
		{
			name: "set property",
			given: func(logger *emf.Logger) {
				logger.Property("aaa", "666").Metric("foo", 33)
			},
			expected: "testdata/8.json",
		},
		{
			name: "set default properties for lambda",
			env: map[string]string{
				"AWS_LAMBDA_FUNCTION_NAME":        "some-func-name",
				"AWS_EXECUTION_ENV":               "golang",
				"AWS_LAMBDA_FUNCTION_MEMORY_SIZE": "128",
				"AWS_LAMBDA_FUNCTION_VERSION":     "1",
				"AWS_LAMBDA_LOG_STREAM_NAME":      "log/stream",
			},
			given: func(logger *emf.Logger) {
				logger.Metric("foo", 33)
			},
			expected: "testdata/9.json",
		},
		{
			name: "not sampled trace",
			env: map[string]string{
				"_X_AMZN_TRACE_ID": "foo",
			},
			given: func(logger *emf.Logger) {
				logger.Metric("foo", 33)
			},
			expected: "testdata/10.json",
		},
		{
			name: "sampled trace",
			env: map[string]string{
				"_X_AMZN_TRACE_ID": "foo,Sampled=1,bar",
			},
			given: func(logger *emf.Logger) {
				logger.Metric("foo", 33)
			},
			expected: "testdata/11.json",
		},
		{
			name: "one dimension",
			given: func(logger *emf.Logger) {
				logger.Dimension("a", "b").Metric("c", 11)
			},
			expected: "testdata/12.json",
		},
		{
			name: "two dimensions",
			given: func(logger *emf.Logger) {
				logger.Dimension("a", "b").Dimension("o", "p").Metric("c", 11)
			},
			expected: "testdata/13.json",
		},
		{
			name: "one dimension set",
			given: func(logger *emf.Logger) {
				logger.
					DimensionSet(
						emf.NewDimension("a", "b"),
						emf.NewDimension("o", "p")).
					Metric("c", 11)
			},
			expected: "testdata/14.json",
		},
		{
			name: "two dimension sets",
			given: func(logger *emf.Logger) {
				logger.
					DimensionSet(
						emf.NewDimension("a", "b"),
						emf.NewDimension("c", "d")).
					DimensionSet(
						emf.NewDimension("1", "2"),
						emf.NewDimension("3", "4")).
					Metric("foo", 22)
			},
			expected: "testdata/15.json",
		},
		{
			name: "default and custom contexts, multiple dimensions/dimension sets",
			given: func(logger *emf.Logger) {
				logger.NewContext().
					Dimension("CC", "DD").
					DimensionSet(
						emf.NewDimension("gg", "hh"),
						emf.NewDimension("kk", "jj")).
					MetricFloatAs("foo", 33.66, emf.Bits)
				logger.
					Dimension("AA", "BB").
					DimensionSet(
						emf.NewDimension("ww", "ee"),
						emf.NewDimension("rr", "tt")).
					MetricAs("bar", 666, emf.BytesSecond)
			},
			expected: "testdata/16.json",
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			if len(tc.env) > 0 {
				defer unsetenv(t, tc.env)
				setenv(t, tc.env)
			}

			var buf bytes.Buffer
			logger := emf.New(emf.WithWriter(&buf))
			tc.given(logger)
			logger.Log()

			f, err := ioutil.ReadFile(tc.expected)
			if err != nil {
				t.Fatal("unable to read file with expected json")
			}

			jsonassert.New(t).Assertf(buf.String(), string(f))
		})
	}

	t.Run("no metrics set", func(t *testing.T) {
		var buf bytes.Buffer
		logger := emf.New(emf.WithWriter(&buf))
		logger.Log()

		if buf.String() != "" {
			t.Error("Buffer not empty")
		}
	})

	t.Run("new context, no metrics set", func(t *testing.T) {
		var buf bytes.Buffer
		logger := emf.New(emf.WithWriter(&buf))
		logger.NewContext().Namespace("galaxy")
		logger.Log()

		if buf.String() != "" {
			t.Error("Buffer not empty")
		}
	})
}

func TestEmfNoDefaultDimensions(t *testing.T) {
	tcs := []struct {
		name     string
		env      map[string]string
		given    func(logger *emf.Logger)
		expected string
	}{
		{
			name: "default namespace, int metric",
			given: func(logger *emf.Logger) {
				logger.Metric("foo", 33)
			},
			expected: "testdata/1.json",
		},
		{
			name: "default namespace, float metric",
			given: func(logger *emf.Logger) {
				logger.MetricFloat("foo", 33.66)
			},
			expected: "testdata/2.json",
		},
		{
			name: "custom namespace, int and float metrics",
			given: func(logger *emf.Logger) {
				logger.Namespace("galaxy").MetricFloat("foo", 33.66).Metric("bar", 666)
			},
			expected: "testdata/3.json",
		},
		{
			name: "custom namespace, int and float metrics, custom units",
			given: func(logger *emf.Logger) {
				logger.Namespace("galaxy").
					MetricFloatAs("foo", 33.66, emf.Milliseconds).
					MetricAs("bar", 666, emf.Count)
			},
			expected: "testdata/4.json",
		},
		{
			name: "new context, default namespace, int and float metrics, custom units",
			given: func(logger *emf.Logger) {
				logger.NewContext().
					MetricFloatAs("foo", 33.66, emf.Milliseconds).
					MetricAs("bar", 666, emf.Count)
			},
			expected: "testdata/5.json",
		},
		{
			name: "new context, custom namespace, int and float metrics, custom units",
			given: func(logger *emf.Logger) {
				logger.NewContext().Namespace("galaxy").
					MetricFloatAs("foo", 33.66, emf.Bits).
					MetricAs("bar", 666, emf.BytesSecond)
			},
			expected: "testdata/6.json",
		},
		{
			name: "default and custom contexts, different metrics names",
			given: func(logger *emf.Logger) {
				logger.NewContext().MetricFloatAs("foo", 33.66, emf.Bits)
				logger.MetricAs("bar", 666, emf.BytesSecond)
			},
			expected: "testdata/7.json",
		},
		{
			name: "set property",
			given: func(logger *emf.Logger) {
				logger.Property("aaa", "666").Metric("foo", 33)
			},
			expected: "testdata/8.json",
		},

		{
			name: "not sampled trace",
			env: map[string]string{
				"_X_AMZN_TRACE_ID": "foo",
			},
			given: func(logger *emf.Logger) {
				logger.Metric("foo", 33)
			},
			expected: "testdata/10.json",
		},
		{
			name: "one dimension",
			given: func(logger *emf.Logger) {
				logger.Dimension("a", "b").Metric("c", 11)
			},
			expected: "testdata/12.json",
		},
		{
			name: "two dimensions",
			given: func(logger *emf.Logger) {
				logger.Dimension("a", "b").Dimension("o", "p").Metric("c", 11)
			},
			expected: "testdata/13.json",
		},
		{
			name: "one dimension set",
			given: func(logger *emf.Logger) {
				logger.
					DimensionSet(
						emf.NewDimension("a", "b"),
						emf.NewDimension("o", "p")).
					Metric("c", 11)
			},
			expected: "testdata/14.json",
		},
		{
			name: "two dimension sets",
			given: func(logger *emf.Logger) {
				logger.
					DimensionSet(
						emf.NewDimension("a", "b"),
						emf.NewDimension("c", "d")).
					DimensionSet(
						emf.NewDimension("1", "2"),
						emf.NewDimension("3", "4")).
					Metric("foo", 22)
			},
			expected: "testdata/15.json",
		},
		{
			name: "default and custom contexts, multiple dimensions/dimension sets",
			given: func(logger *emf.Logger) {
				logger.NewContext().
					Dimension("CC", "DD").
					DimensionSet(
						emf.NewDimension("gg", "hh"),
						emf.NewDimension("kk", "jj")).
					MetricFloatAs("foo", 33.66, emf.Bits)
				logger.
					Dimension("AA", "BB").
					DimensionSet(
						emf.NewDimension("ww", "ee"),
						emf.NewDimension("rr", "tt")).
					MetricAs("bar", 666, emf.BytesSecond)
			},
			expected: "testdata/16.json",
		},
		{
			name: "default properties for lambda are ignored",
			env: map[string]string{
				"AWS_LAMBDA_FUNCTION_NAME":        "some-func-name",
				"AWS_EXECUTION_ENV":               "golang",
				"AWS_LAMBDA_FUNCTION_MEMORY_SIZE": "128",
				"AWS_LAMBDA_FUNCTION_VERSION":     "1",
				"AWS_LAMBDA_LOG_STREAM_NAME":      "log/stream",
			},
			given: func(logger *emf.Logger) {
				logger.Metric("foo", 33)
			},
			expected: "testdata/17.json",
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			if len(tc.env) > 0 {
				defer unsetenv(t, tc.env)
				setenv(t, tc.env)
			}

			var buf bytes.Buffer
			logger := emf.NewWithoutDefaultDimensions(emf.WithWriter(&buf))
			tc.given(logger)
			logger.Log()

			f, err := ioutil.ReadFile(tc.expected)
			if err != nil {
				t.Fatal("unable to read file with expected json")
			}

			jsonassert.New(t).Assertf(buf.String(), string(f))
		})
	}

	t.Run("no metrics set", func(t *testing.T) {
		var buf bytes.Buffer
		logger := emf.New(emf.WithWriter(&buf))
		logger.Log()

		if buf.String() != "" {
			t.Error("Buffer not empty")
		}
	})

	t.Run("new context, no metrics set", func(t *testing.T) {
		var buf bytes.Buffer
		logger := emf.New(emf.WithWriter(&buf))
		logger.NewContext().Namespace("galaxy")
		logger.Log()

		if buf.String() != "" {
			t.Error("Buffer not empty")
		}
	})
}

func setenv(t *testing.T, env map[string]string) {
	for k, v := range env {
		err := os.Setenv(k, v)
		if err != nil {
			t.Fatalf("unable to set env variable: %s", k)
		}
	}
}

func unsetenv(t *testing.T, env map[string]string) {
	for k := range env {
		err := os.Unsetenv(k)
		if err != nil {
			t.Fatalf("unable to unset env variable: %s", k)
		}
	}
}

package emf

import (
	"fmt"
	"os"
	"testing"
	"time"
)

func TestNew(t *testing.T) {
	tcs := []struct {
		name     string
		opts     []LoggerOption
		expected *Logger
	}{
		{
			name: "default",
			expected: &Logger{
				out:       os.Stdout,
				timestamp: time.Now().UnixNano() / int64(time.Millisecond),
			},
		},
		{
			name: "with options",
			opts: []LoggerOption{
				WithWriter(os.Stderr),
				WithTimestamp(time.Now().Add(time.Hour)),
			},
			expected: &Logger{
				out:       os.Stderr,
				timestamp: time.Now().Add(time.Hour).UnixNano() / int64(time.Millisecond),
			},
		},
		{
			name: "without dimensions",
			opts: []LoggerOption{
				WithoutDimensions(),
			},
			expected: &Logger{
				out:       os.Stdout,
				timestamp: time.Now().UnixNano() / int64(time.Millisecond),
				withoutDimensions: false,
			},
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			actual := New(tc.opts...)
			if err := loggersEqual(actual, tc.expected); err != nil {
				t.Errorf("logger does not match: %v", err)
			}
		})
	}

}

// loggersEqual returns a non-nil error if the loggers do not match.
// Currently it only checks that the loggers' output writer and timestamp match.
func loggersEqual(actual, expected *Logger) error {
	if actual.out != expected.out {
		return fmt.Errorf("output does not match")
	}

	if err := approxInt64(actual.timestamp, expected.timestamp, 100 /* ms */); err != nil {
		return fmt.Errorf("timestamp %v", err)
	}

	return nil
}

func approxInt64(actual, expected, tolerance int64) error {
	diff := expected - actual
	if diff < 0 {
		diff = -diff
	}
	if diff > tolerance {
		return fmt.Errorf("value %v is out of tolerance %v±%v", actual, expected, tolerance)
	}
	return nil
}

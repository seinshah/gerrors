package gerrors_test

import (
	"errors"
	"fmt"
	"strings"
	"sync"
	"testing"

	"github.com/seinshah/gerrors"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	systemKeys = 4
)

type glogger struct {
	history map[string]map[gerrors.LogLevel]int
	mu      sync.Mutex
}

type gcoreErr struct{}

type testData struct {
	name           string
	inputErr       error
	code           gerrors.Code
	metadata       []any
	expectedKeys   int
	expectedOutput string
	supportGrpc    bool
}

func TestNew(t *testing.T) {
	t.Parallel()

	core := gerrors.NewMapper(gerrors.Unknown, gerrors.GetDefaultMapping()).Lookup(gerrors.Internal)

	errOutput := func(err error) string {
		out := core.GetDefaultMessage()
		if err != nil {
			out = err.Error()
		}

		return fmt.Sprintf(
			"error: %s(%d) - %s",
			core.GetIdentifier(),
			core.GetInternalCode(),
			out,
		)
	}

	testCases := []testData{
		{
			name:           "nil input error",
			inputErr:       nil,
			code:           gerrors.Internal,
			metadata:       nil,
			expectedKeys:   0,
			supportGrpc:    true,
			expectedOutput: errOutput(nil),
		},
		{
			name:           "with input error",
			inputErr:       errors.New("example error"),
			code:           gerrors.Internal,
			metadata:       nil,
			expectedKeys:   0,
			supportGrpc:    true,
			expectedOutput: errOutput(errors.New("example error")),
		},
		{
			name:           "with metadata",
			inputErr:       nil,
			code:           gerrors.Internal,
			metadata:       []any{"key1", "val1", "key2", true, "key3", struct{}{}},
			expectedKeys:   3,
			supportGrpc:    true,
			expectedOutput: errOutput(nil),
		},
		{
			name:           "with invalid keys",
			inputErr:       nil,
			code:           gerrors.Internal,
			metadata:       []any{false, "val1", strings.Repeat("k", 70), "val2", "$key%3", "val3"},
			expectedKeys:   2,
			supportGrpc:    true,
			expectedOutput: errOutput(nil),
		},
	}

	cl := &glogger{
		history: make(map[string]map[gerrors.LogLevel]int),
		mu:      sync.Mutex{},
	}

	f := gerrors.NewFormatter(gerrors.WithLogger(cl))

	logLevels := []gerrors.LogLevel{
		gerrors.LogLevelError,
		gerrors.LogLevelWarn,
		gerrors.LogLevelInfo,
		gerrors.LogLevelDebug,
		gerrors.LogLevelTrace,
		gerrors.LogLevelOff,
	}

	for i := range testCases {
		tc := testCases[i]
		index := fmt.Sprintf("t%d", i)

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			nf := f.Clone().AddLabels("testcase", index)

			var err *gerrors.GeneralError

			if tc.metadata == nil || len(tc.metadata) == 0 {
				err = nf.New(tc.inputErr, tc.code)
			} else {
				err = nf.New(tc.inputErr, tc.code, tc.metadata...)
			}
			checkError(t, err, tc)

			for _, level := range logLevels {
				if tc.metadata == nil || len(tc.metadata) == 0 {
					err = nf.NewWithLogLevel(tc.inputErr, tc.code, level)
				} else {
					err = nf.NewWithLogLevel(tc.inputErr, tc.code, level, tc.metadata...)
				}
				checkError(t, err, tc)
			}

			cl.mu.Lock()
			logData := cl.history[index]
			cl.mu.Unlock()

			for _, level := range logLevels {
				expectedVal := 1

				if level == gerrors.LogLevelError {
					expectedVal = 2
				} else if level == gerrors.LogLevelOff {
					continue
				}

				if logData[level] != expectedVal {
					t.Errorf("expected log level %d for %s to be called %d times, called %d time",
						level, index, expectedVal, logData[level])
				}
			}
		})
	}
}

func TestGrpcError(t *testing.T) {
	t.Parallel()

	err := gerrors.GrpcError(errors.New("example error"))
	st, ok := status.FromError(err)

	if !ok {
		t.Fatalf("expected gRPC error, got %T", err)
	}

	if len(st.Details()) != 0 {
		t.Fatalf("non gerrors error should not have details, got %v", st.Details())
	}

	f := gerrors.NewFormatter(
		gerrors.WithLookuper(gerrors.NewMapper(
			gerrors.Unknown,
			map[gerrors.Code]gerrors.CoreError{gerrors.Unknown: gcoreErr{}},
		)),
	)

	err = f.New(errors.New("example error"), gerrors.Code(100)).Grpc()
	st, ok = status.FromError(err)

	if !ok {
		t.Fatalf("expected gRPC error, got %T", err)
	}

	if len(st.Details()) != 0 {
		t.Fatalf("non gerrors error should not have details, got %v", st.Details())
	}
}

func checkError(t *testing.T, err *gerrors.GeneralError, expected testData) {
	t.Helper()

	if err.Error() != expected.expectedOutput {
		t.Errorf("expected %s, got %s", expected.expectedOutput, err.Error())
	}

	if len(err.Metadata())-systemKeys-1 != expected.expectedKeys {
		t.Errorf("expected %d keys in metadata, got %d", expected.expectedKeys, len(err.Metadata())-systemKeys-1)
	}

	ge := err.Grpc()

	st, ok := status.FromError(ge)

	if !ok {
		t.Errorf("expected gRPC error, got %T", ge)

		return
	}

	if expected.supportGrpc {
		if len(st.Details()) != 1 {
			t.Errorf("expected 1 detail in grpc error, got %d", len(st.Details()))

			return
		}

		detail, ok := st.Details()[0].(*errdetails.ErrorInfo)

		if !ok {
			t.Errorf("expected ErrorInfo detail, got %T", st.Details()[0])

			return
		}

		errorMDs := len(detail.GetMetadata()) - systemKeys - 1
		if errorMDs != expected.expectedKeys {
			t.Errorf("expected %d keys in grpc metadata, got %d (actual: %d)",
				expected.expectedKeys, errorMDs, len(detail.GetMetadata()))
		}
	} else if st.Code() != codes.Unknown {
		t.Errorf("expected gRPC error with code %s, got %s", codes.Unknown, st.Code())
	}
}

func (g *glogger) Error(_ error, _ string, keyValues ...any) {
	g.setHistory(gerrors.LogLevelError, keyValues)
}

func (g *glogger) Warn(_ string, keyValues ...any) {
	g.setHistory(gerrors.LogLevelWarn, keyValues)
}

func (g *glogger) Info(_ string, keyValues ...any) {
	g.setHistory(gerrors.LogLevelInfo, keyValues)
}

func (g *glogger) Debug(_ string, keyValues ...any) {
	g.setHistory(gerrors.LogLevelDebug, keyValues)
}

func (g *glogger) Trace(_ string, keyValues ...any) {
	g.setHistory(gerrors.LogLevelTrace, keyValues)
}

func (g *glogger) setHistory(level gerrors.LogLevel, keyValues []any) {
	var (
		testCase string
		ok       bool
	)

	for i := 0; i < len(keyValues); i += 2 {
		if keyValues[i] == "testcase" {
			testCase, ok = keyValues[i+1].(string)

			break
		}
	}

	if !ok || testCase == "" {
		return
	}

	g.mu.Lock()
	defer g.mu.Unlock()

	if g.history[testCase] == nil {
		g.history[testCase] = make(map[gerrors.LogLevel]int)
	}

	g.history[testCase][level]++
}

func (gcoreErr) GetInternalCode() gerrors.Code {
	return gerrors.Code(100)
}

func (gcoreErr) GetDefaultMessage() string {
	return "custom core error"
}

func (gcoreErr) GetIdentifier() string {
	return "custom"
}

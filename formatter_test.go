package gerrors_test

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"testing"

	"github.com/seinshah/gerrors"
	"google.golang.org/grpc/codes"
)

const defaultErrText = "example error"

type customCoreError struct{}

type logger struct{}

func TestFormatter(t *testing.T) {
	t.Parallel()

	core := gerrors.NewMapper(gerrors.Unknown, gerrors.GetDefaultMapping()).Lookup(gerrors.Internal)
	coreg, ok := core.(gerrors.CoreGRPCError)

	var grpcErrCode string
	if ok {
		grpcErrCode = strconv.Itoa(int(coreg.GetGRPCCode()))
	}

	defaultOutput := fmt.Sprintf(
		"error: %s(%d) - %s",
		core.GetIdentifier(),
		core.GetInternalCode(),
		defaultErrText,
	)

	testCases := []struct {
		name                  string
		options               []gerrors.FormatterOption
		additionalLabels      []any
		expectedKeys          int
		expectedMissingValues int
		expectedOutput        string
	}{
		{
			name:           "default formatter",
			options:        []gerrors.FormatterOption{},
			expectedOutput: defaultOutput,
		},
		{
			name:                  "default formatter with missing value",
			options:               []gerrors.FormatterOption{},
			additionalLabels:      []any{"key"},
			expectedKeys:          1,
			expectedMissingValues: 1,
			expectedOutput:        defaultOutput,
		},
		{
			name: "formatter with custom template",
			options: []gerrors.FormatterOption{
				gerrors.WithTemplate(
					"custom: {{.Identifier}}({{.ErrorCode}})({{.GrpcErrorCode}}) - {{.Message}} - {{.DefaultMessage}} - {{.Labels.key}}",
				),
			},
			additionalLabels: []any{"key", "value"},
			expectedKeys:     1,
			expectedOutput: fmt.Sprintf(
				"custom: %s(%d)(%s) - %s - %s - value",
				core.GetIdentifier(),
				core.GetInternalCode(),
				grpcErrCode,
				defaultErrText,
				core.GetDefaultMessage(),
			),
		},
		{
			name: "formatter with custom missing value replacement",
			options: []gerrors.FormatterOption{
				gerrors.WithMissingValueReplacement("missing"),
			},
			additionalLabels:      []any{"key1", 1, "key2", true, "key3", struct{}{}, "key4"},
			expectedKeys:          4,
			expectedMissingValues: 1,
			expectedOutput:        defaultOutput,
		},
		{
			name: "formatter with disabled missing value replacement",
			options: []gerrors.FormatterOption{
				gerrors.WithDisabledMissingValueReplacement(),
			},
			additionalLabels:      []any{"key1", 1, "key2", true, "key3", struct{}{}, "key4"},
			expectedKeys:          3,
			expectedMissingValues: 0,
			expectedOutput:        defaultOutput,
		},
		{
			name: "formatter with default set of labels",
			options: []gerrors.FormatterOption{
				gerrors.WithLabels("dk1", "val1", "dk2", "val2", "dk3", "val3"),
			},
			additionalLabels:      []any{"key1", 1, "key2", true, "key3", struct{}{}},
			expectedKeys:          6,
			expectedMissingValues: 0,
			expectedOutput:        defaultOutput,
		},
		{
			name: "formatter with default set of labels with invalid keys",
			options: []gerrors.FormatterOption{
				gerrors.WithLabels(strings.Repeat("a", 70), "dk1", "dk2", "val2", struct{}{}, "val3"),
			},
			additionalLabels:      []any{"key1", 1, "key2", true, struct{}{}, struct{}{}},
			expectedKeys:          4,
			expectedMissingValues: 0,
			expectedOutput:        defaultOutput,
		},
		{
			name: "formatter with custom core call back",
			options: []gerrors.FormatterOption{
				gerrors.WithLookuper(
					gerrors.NewMapper(gerrors.Code(0), map[gerrors.Code]gerrors.CoreError{
						gerrors.Code(0): CustomCoreError{},
					}),
				),
			},
			expectedKeys:          0,
			expectedMissingValues: 0,
			expectedOutput: fmt.Sprintf(
				"error: %s(%d) - %s",
				customCoreError{}.GetIdentifier(),
				customCoreError{}.GetInternalCode(),
				defaultErrText,
			),
		},
		{
			name: "formatter with logger",
			options: []gerrors.FormatterOption{
				gerrors.WithLogger(logger{}),
			},
			expectedKeys:          0,
			expectedMissingValues: 0,
			expectedOutput:        defaultOutput,
		},
	}

	for i := range testCases {
		tc := testCases[i]

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			f := gerrors.NewFormatter(tc.options...)
			nf := f.Clone()

			if tc.additionalLabels != nil {
				nf.AddLabels(tc.additionalLabels...)
			}

			if len(nf.LabelsSlice())/2 != tc.expectedKeys {
				t.Errorf("expected %d keys, got %d", tc.expectedKeys, len(nf.LabelsSlice())/2)
			}

			labels := nf.LabelsMap()
			missingValues := 0

			if len(labels) != tc.expectedKeys {
				t.Errorf("expected %d keys, got %d", tc.expectedKeys, len(labels))
			}

			if token, ok := nf.MissingValueReplacement(); ok {
				for _, v := range labels {
					if v == token {
						missingValues++
					}
				}
			}

			if missingValues != tc.expectedMissingValues {
				t.Errorf("expected %d missing values, got %d", tc.expectedMissingValues, missingValues)
			}

			err := nf.New(errors.New(defaultErrText), gerrors.Internal)

			if err.Error() != tc.expectedOutput {
				t.Errorf("expected %s, got %s", tc.expectedOutput, err.Error())
			}
		})
	}
}

func TestBadTemplate(t *testing.T) {
	t.Parallel()

	defer func() {
		if r := recover(); r == nil {
			t.Errorf("The code did not panic")
		}
	}()

	gerrors.NewFormatter(gerrors.WithTemplate("{{.$VAR}}"))
}

func (customCoreError) GetGRPCCode() codes.Code {
	return codes.Internal
}

func (customCoreError) GetInternalCode() gerrors.Code {
	return gerrors.Code(100)
}

func (customCoreError) GetDefaultMessage() string {
	return "custom core error"
}

func (customCoreError) GetIdentifier() string {
	return "custom"
}

func (logger) Error(_ error, _ string, _ ...interface{}) {
}

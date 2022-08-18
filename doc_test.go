package gerrors_test

import (
	"errors"
	"fmt"

	"github.com/seinshah/gerrors"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type CustomCoreError struct{}

func Example() {
	// Use default formatter and print error message using the default template.
	err := gerrors.DefaultFormatter.New(errors.New("error"), gerrors.Unknown, "key", "value")
	fmt.Println(err.Error())

	// Use default formatter and get gRPC status error from gerrors.
	gerr := gerrors.DefaultFormatter.New(errors.New("error"), gerrors.Unknown, "key", "value").Grpc()

	st, ok := status.FromError(gerr)
	if !ok {
		fmt.Println("error converting to gRPC status error")

		return
	}

	fmt.Println(st.Message())

	if len(st.Details()) != 1 {
		fmt.Println("converted error to gRPC status error was not of type gerrors.GeneralError")

		return
	}

	details, ok := st.Details()[0].(*errdetails.ErrorInfo)
	if !ok {
		fmt.Println("converted error to gRPC status error was not of type gerrors.GeneralError")

		return
	}

	fmt.Println(details.GetMetadata())
}

func ExampleNewFormatter() {
	f := gerrors.NewFormatter()

	err := f.New(errors.New("error"), gerrors.Unknown, "key", "value")
	fmt.Println(err.Error())
}

func ExampleNewFormatter_withOptions() {
	f := gerrors.NewFormatter(
		// Error method now returns an output populated based on this template.
		gerrors.WithTemplate(
			"custom: {{.Identifier}}({{.ErrorCode}}) - {{.DefaultMessage}} - {{.GrpcErrorCode}} - {{.Labels}}",
		),
		// Any error we create on this formatter uses this lookup function
		// Due to our implementation here, whatever error code you pass to
		// the new error, the same error info will be returned by this function.
		gerrors.WithCoreLookup(
			gerrors.Code(100),
			func(code gerrors.Code) gerrors.CoreError {
				return CustomCoreError{}
			},
		),
		// If we pass key values to the formatter or during error creation,
		// and the value is missing (uneven number of parameters), that key
		// will be completely ignored and won't be added to the error's labels.
		gerrors.WithDisabledMissingValueReplacement(),
		// Any error created using this formatter or any clone from this formatter
		// will have this key value in its labels.
		// Because of previous option, ignored key will be ignored from error's labels.
		gerrors.WithLabels("always", true, "ignored"),
	)

	err := f.New(errors.New("error"), gerrors.Unknown, "key", "value")
	fmt.Println(err.Error())

	// Output: custom: custom(100) - custom core error - 13 - map[_default_message:custom core error _error_code:100 _identifier:custom _original_error:error always:true key:value]
}

func ExampleFormatter_Clone() {
	f := gerrors.NewFormatter(
		gerrors.WithTemplate("{{.Labels}}"),
		gerrors.WithLabels("override", "main", "remains", "yes"),
		gerrors.WithCoreLookup(
			gerrors.Code(100),
			func(code gerrors.Code) gerrors.CoreError {
				return CustomCoreError{}
			},
		),
	)

	f2 := f.Clone()
	f2.AddLabels("override", "cloned", "new", true)

	err := f2.New(errors.New("error"), gerrors.Unknown, "key", "value")
	fmt.Println(err.Error())

	// Output: map[_default_message:custom core error _error_code:100 _identifier:custom _original_error:error key:value new:true override:cloned remains:yes]
}

func (CustomCoreError) GetGRPCCode() codes.Code {
	return codes.Internal
}

func (CustomCoreError) GetInternalCode() gerrors.Code {
	return gerrors.Code(100)
}

func (CustomCoreError) GetDefaultMessage() string {
	return "custom core error"
}

func (CustomCoreError) GetIdentifier() string {
	return "custom"
}

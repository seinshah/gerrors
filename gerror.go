package gerrors

import (
	"bytes"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	// MetadataIdentifier is the ket for accessing the error identifier.
	MetadataIdentifier = "_identifier"

	// MetadataErrorCode is the key for accessing the gerrors internal
	// or customized error code.
	MetadataErrorCode = "_error_code"

	// MetadataDefaultMessage is the key for accessing the default gerrors core
	// error message.
	MetadataDefaultMessage = "_default_message"

	// MetadataOriginalError is the key for accessing the original error
	// which was used during initializing GeneralError.
	MetadataOriginalError = "_original_error"
)

// GeneralError is the error type defined, controlled, and handled by gerrors package.
type GeneralError struct {
	originalError error
	coreError     CoreError
	formatter     *Formatter
	details       *errdetails.ErrorInfo
}

// tplData is used to populate each error's information and then parse the template.
type tplData struct {
	Identifier     string
	ErrorCode      string
	GrpcErrorCode  string
	Message        string
	DefaultMessage string
	Labels         map[string]string
}

// errNoOriginalError is set as the input error whenever there is no original error.
var errNoOriginalError = errors.New("no original error")

// GrpcError accepts an error of gerrors.GeneralError type and returns a gRPC error
// by translating the error to gRPC error and attach all labels as the metadata.
// It supports [Google's AIP 193].
// If the input is not of [GeneralError] type, it smply returns a gRPC error
// with [google.golang.org/grpc/codes.Unknown] error code with the input error
// message as the message.
// If the receiver is receiving the error in gRPC error format, you can check
// [this blog post] on how to parse the error and extract the information from it.
//
// [Google's AIP 193]: https://google.aip.dev/193
// [this blog post]: https://jbrandhorst.com/post/grpc-errors
func GrpcError(err error) error {
	var finalErr *GeneralError

	if !errors.As(err, &finalErr) {
		return status.Error(codes.Unknown, err.Error())
	}

	grpcErr, ok := finalErr.coreError.(CoreGRPCError)

	if !ok {
		return status.Error(codes.Unknown, finalErr.Error())
	}

	st := status.New(grpcErr.GetGRPCCode(), finalErr.Error())

	finalStatus, err := st.WithDetails(finalErr.details)
	if err != nil {
		finalStatus = st
	}

	return finalStatus.Err()
}

// New creates a new [GeneralError] instance using the provided formatter.
// inputErr is the error that is triggered prior to the creation of the error
// and it can be nil. If it's nil, the final error message will be the code's
// default message.
// Any new error can have a list of key values as the metadata. These key values
// will be appended to the formatter's default labels.
// If the formatter has a logger, it will also log the error at Error level.
func (f *Formatter) New(inputErr error, code Code, metadataKeyValues ...any) *GeneralError {
	err := f.createError(inputErr, code, metadataKeyValues...)

	err.log(f.logger, LogLevelError, err.MetadataSlice())

	return err
}

// NewWithLogLevel is the same as New, but it allows the caller to control the log level.
// If the logger is not provided to the formatter, this method is exactly the same as New
// where logging will be ignored.
func (f *Formatter) NewWithLogLevel(
	inputErr error,
	code Code,
	level LogLevel,
	metadataKeyValues ...any,
) *GeneralError {
	err := f.createError(inputErr, code, metadataKeyValues...)

	err.log(f.logger, level, err.MetadataSlice())

	return err
}

func (f *Formatter) createError(inputErr error, code Code, metadataKeyValues ...any) *GeneralError {
	if inputErr == nil {
		inputErr = errNoOriginalError
	}

	err := &GeneralError{
		originalError: inputErr,
		coreError:     f.coreDataLookup.Lookup(code),
		formatter:     f,
		details:       nil,
	}

	err.generateDetails(metadataKeyValues, f.labels)

	return err
}

// Error allows GeneralError to implement the error interface.
// It uses the formatter template and different information of the GeneralError
// to generate the error message.
func (ge *GeneralError) Error() string {
	var buf bytes.Buffer

	data := ge.getTemplateData()

	if err := ge.formatter.template.Execute(&buf, data); err != nil {
		return fmt.Sprintf("failed to execute template: %s (original error: %s)", err.Error(), data.Message)
	}

	return buf.String()
}

// Metadata returns all the combined labels of the given GeneralError.
func (ge *GeneralError) Metadata() map[string]string {
	return ge.details.GetMetadata()
}

// Grpc is the method defined on GeneralError which returns the gRPC error.
// See Grpc function for more details.
func (ge *GeneralError) Grpc() error {
	return GrpcError(ge)
}

func (ge *GeneralError) generateDetails(metadataKeyValues []any, defaultLabels map[string]string) {
	metadata := make(map[string]string)

	for k, v := range defaultLabels {
		metadata[k] = v
	}

	metadata[MetadataIdentifier] = ge.coreError.GetIdentifier()
	metadata[MetadataErrorCode] = strconv.Itoa(int(ge.coreError.GetInternalCode()))
	metadata[MetadataDefaultMessage] = ge.coreError.GetDefaultMessage()
	metadata[MetadataOriginalError] = ge.originalError.Error()

	for index := 0; index < len(metadataKeyValues); index += 2 {
		key, val, ok := ge.formatter.getStringifiedKeyValue(metadataKeyValues, index)
		if !ok {
			continue
		}

		metadata[key] = val
	}

	ge.details = &errdetails.ErrorInfo{
		Reason:   strings.ReplaceAll(strings.ToUpper(ge.coreError.GetIdentifier()), " ", "_"),
		Metadata: metadata,
	}
}

func (ge *GeneralError) getTemplateData() tplData {
	msg := ge.coreError.GetDefaultMessage()
	if !errors.Is(ge.originalError, errNoOriginalError) {
		msg = ge.originalError.Error()
	}

	var grpcCode string

	coreg, ok := ge.coreError.(CoreGRPCError)
	if ok {
		grpcCode = strconv.Itoa(int(coreg.GetGRPCCode()))
	}

	return tplData{
		Identifier:     ge.coreError.GetIdentifier(),
		ErrorCode:      strconv.Itoa(int(ge.coreError.GetInternalCode())),
		GrpcErrorCode:  grpcCode,
		Message:        msg,
		DefaultMessage: ge.coreError.GetDefaultMessage(),
		Labels:         ge.details.GetMetadata(),
	}
}

func (ge *GeneralError) log(logger logger, level LogLevel, metadata []any) {
	if logger == nil || level == LogLevelOff {
		return
	}

	// nolint: exhaustive
	switch level {
	case LogLevelTrace:
		l, ok := logger.(traceLogger)
		if !ok {
			return
		}

		l.Trace(ge.Error(), metadata...)
	case LogLevelDebug:
		l, ok := logger.(debugLogger)
		if !ok {
			return
		}

		l.Debug(ge.Error(), metadata...)
	case LogLevelInfo:
		l, ok := logger.(infoLogger)
		if !ok {
			return
		}

		l.Info(ge.Error(), metadata...)
	case LogLevelWarn:
		l, ok := logger.(warnLogger)
		if !ok {
			return
		}

		l.Warn(ge.Error(), metadata...)
	default:
		logger.Error(
			ge.originalError,
			ge.Error(),
			metadata...,
		)
	}
}

func (ge *GeneralError) MetadataSlice() []any {
	s := make([]any, len(ge.details.GetMetadata())*2)
	index := 0

	for k, v := range ge.details.GetMetadata() {
		s[index], s[index+1] = k, v
		index += 2
	}

	return s
}

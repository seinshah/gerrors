package gerrors

import "google.golang.org/grpc/codes"

// Code is gerrors internal error type.
// If a customized core call back function is used, customized error codes
// should be of this type as well.
type Code int

// CoreError is an interface that every error mapper should implement.
// Each mapper for an error maps that error code to the rest of error's
// information. That information should implement this interface to make
// error's data accessible.
type CoreError interface {
	// GetInternalCode returns the gerrors internal code of type Code.
	GetInternalCode() Code

	// GetIdentifier returns a human-readable that explains the code
	// in words. (one or two words)
	GetIdentifier() string

	// GetDefaultMessage returns a default longer message that explains
	// the error code. If a new error is being created with a nil initial
	// error, this default message will be used to explain the error.
	GetDefaultMessage() string
}

// CoreGRPCError can provide support for gRPC error messages.
// If the provided error mapper implements this interface, the error
// can be converted to a gRPC error.
type CoreGRPCError interface {
	// GetGRPCCode returns a gRPC error code that can be matched to
	// an internal gerrors error code.
	GetGRPCCode() codes.Code
}

// Lookuper is an interface that shows how a mapper should be implemented.
// Every mapper should have a lookup method to translate [Code] to [CoreError].
type Lookuper interface {
	Lookup(code Code) CoreError
}

// gerrorCore is the default implementation of CoreError and CoreGRPCError.
type gerrorCore struct {
	internalCode   Code
	identifier     string
	defaultMessage string
	grpcCode       codes.Code
}

// Mapper is the data type that maps gerrors error code to the core error
// which holds more information about the error code.
// This type will be used by different methods to translate the error code
// to the underlying error details.
type Mapper struct {
	mapping      map[Code]CoreError
	unknownError Code
}

const (
	// Unknown generates an unhandled error and is useful whenever the error
	// we want to generate is unknown to us.
	// It translates to GRPC Unknown codes.Code.
	// If Grpc method is called on an error that i not of gerrors type,
	// this error code will be used to convert the error to gerrors,
	// regardless of whether a custom core callback is provided or not.
	Unknown Code = iota + 1

	// NotFound generate a not found error and is useful whenever some database
	// or similar lookup fails because of no record.
	// It translates to GRPC NotFound codes.Code.
	NotFound

	// InvalidArgument generates an invalid argument error and is useful whenever
	// the provided argument by caller is not of a valid type and therefore cannot
	// be processed.
	// It translates to GRPC InvalidArgument codes.code.
	InvalidArgument

	// Marshal generates a marshaling error and is useful whenever there is an error
	// related to marshaling or unmarshaling of some data to some other data.
	// It translates to GRPC Internal codes.Code.
	Marshal

	// Storage generates a storage error and is useful for whenever there is an error
	// because of some storage operation which could be a file-system, database, redis,
	// etc.
	// It translates to GRPC Internal codes.Code.
	Storage

	// Threshold generates an out of range error and is useful for whenever some
	// received argument is out of your expected range. The value is still a valid
	// type and format, but out of your expected range.
	// It translates to GRPC OutOfRange codes.code.
	Threshold

	// Unimplemented generate an unimplemented error and is useful for whenever
	// there is a request from user that has not been implemented yet.
	// This error type is used by GRPC internal packages in some special cases as well.
	// It translates to GRPC Unimplemented codes.code.
	Unimplemented

	// Unauthorized generate an unauthorized error and is useful for whenever
	// the request is not permitted for the requester. It might be that there
	// is no authorized user in the context or header at all or that user is not
	// allowed to perform the operation in question.
	// It translates to grpc Unauthenticated codes.Code.
	Unauthorized

	// Internal generates an internal error and is useful for whenever there is
	// an error that is related to non user-facing error. System is solely responsible
	// for causing these kind of errors.
	// It translates to GRPC Internal codes.Code.
	Internal

	// Unavailable generates an unavailable error and is useful for whenever
	// the requested action is not available to the requester.
	// It translates to GRPC Unavailable codes.Code.
	Unavailable

	// ExternalRequest generates an internal error and is useful for whenever
	// system fails when calling a third-party service. The third-party service
	// can be within organization or outside of organization.
	// It translates to GRPC Internal codes.Code.
	ExternalRequest
)

// NewMapper initiates the Mapper with all available one-to-one mapping
// information from an error code to error details.
// mapping is a map that maps the [Code] to [CoreError]. This can be customized
// based on your needs.
// unknownErrorCode will be used whenever mapper cannot translate a [Code]
// to [CoreError] and this unknown code is used as a fallback.
func NewMapper(unknownErrorCode Code, mapping map[Code]CoreError) *Mapper {
	return &Mapper{
		mapping:      mapping,
		unknownError: unknownErrorCode,
	}
}

// Lookup helps [Mapper] to implement [Lookuper] interface that can be passed to
// [Formatter] and acts as the translator for translating [Code] to [CoreError].
func (m *Mapper) Lookup(code Code) CoreError {
	var selectedInfo CoreError

	if rec, ok := m.mapping[code]; ok {
		selectedInfo = rec
	} else {
		selectedInfo = m.mapping[m.unknownError]
	}

	return selectedInfo
}

// GetInternalCode is part of CoreError interface implementation.
// It returns the internal code of the error. e.g. Unknown, NotFound, etc.
func (g *gerrorCore) GetInternalCode() Code {
	return g.internalCode
}

// GetIdentifier is part of CoreError interface implementation.
// It returns the identifier of the error. e.g. "unhandled", "no-record", etc.
func (g *gerrorCore) GetIdentifier() string {
	return g.identifier
}

// GetDefaultMessage is part of CoreError interface implementation.
// It returns the default message of the error.
func (g *gerrorCore) GetDefaultMessage() string {
	return g.defaultMessage
}

// GetGRPCCode is part of CoreGRPCError interface implementation.
// It returns the GRPC code of the error. e.g. codes.Unknown, codes.NotFound, etc.
func (g *gerrorCore) GetGRPCCode() codes.Code {
	return g.grpcCode
}

// GetDefaultMapping returns a map that contains translation between package's
// default error codes to detailed information. These information can be customized.
// Check [Formatter] and [WithLookuper] for more information.
func GetDefaultMapping() map[Code]CoreError {
	return map[Code]CoreError{
		Unknown: &gerrorCore{
			internalCode:   Unknown,
			identifier:     "unknown",
			defaultMessage: "no information is available for this type of error",
			grpcCode:       codes.Unknown,
		},

		NotFound: &gerrorCore{
			internalCode:   NotFound,
			identifier:     "not-found",
			defaultMessage: "no record was found with given information",
			grpcCode:       codes.NotFound,
		},

		InvalidArgument: &gerrorCore{
			internalCode:   InvalidArgument,
			identifier:     "invalid-argument",
			defaultMessage: "some of the arguments in the request are invalid",
			grpcCode:       codes.InvalidArgument,
		},

		Marshal: &gerrorCore{
			internalCode:   Marshal,
			identifier:     "marshal",
			defaultMessage: "unable to marshal/unmarshal provided data",
			grpcCode:       codes.Internal,
		},

		Storage: &gerrorCore{
			internalCode:   Storage,
			identifier:     "storage",
			defaultMessage: "unable to perform storage-related operation",
			grpcCode:       codes.Internal,
		},

		Threshold: &gerrorCore{
			internalCode:   Threshold,
			identifier:     "out-of-range",
			defaultMessage: "provided argument is out of valid range",
			grpcCode:       codes.OutOfRange,
		},

		Unimplemented: &gerrorCore{
			internalCode:   Unimplemented,
			identifier:     "unimplemented",
			defaultMessage: "provided argument led to an unimplemented operation",
			grpcCode:       codes.Unimplemented,
		},

		Unauthorized: &gerrorCore{
			internalCode:   Unauthorized,
			identifier:     "unauthorized",
			defaultMessage: "requester is not authorized to perform the requested operation",
			grpcCode:       codes.Unauthenticated,
		},

		Internal: &gerrorCore{
			internalCode:   Internal,
			identifier:     "internal",
			defaultMessage: "there is an internal error in the system",
			grpcCode:       codes.Internal,
		},

		Unavailable: &gerrorCore{
			internalCode:   Unavailable,
			identifier:     "unavailable",
			defaultMessage: "requested action is not available to the requester",
			grpcCode:       codes.Unavailable,
		},

		ExternalRequest: &gerrorCore{
			internalCode:   ExternalRequest,
			identifier:     "external-request",
			defaultMessage: "system failed during the request to external service",
			grpcCode:       codes.Internal,
		},
	}
}

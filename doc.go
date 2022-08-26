// Package gerrors helps to have rich and cohesive error handling throughout the code base.
// This package helps with formatting the final error message and all the error's metadata
// that can be used by different parties to understand the error.
//
// gerrors can be used by any package for simply handling and logging errors with detailed
// information. Furthermore, it can be used whenever a client is involved where a general message
// need to be shown to the end user, but more information about the error need to be shared with
// the client. Error types of gerror have metadata attached to them that can help with this.
//
// Every gerror type can be easily converted to a fully fledged gRPC error type which follows
// [Google's AIP 193] that includes gRPC error message and status, alongside more information regarding
// the error represented by [error details] protocol buffer message.
// In other words, this package can be used to send comprehensive structured error information to the
// receiver through gRPC, REST, or any other protocol.
//
// # Formatter
//
// This package has a default formatter, but there is possibility to customize every aspect of it.
// NewFormatter accepts a variadic number of FormatterOption which have some helper functions to customize
// the formatter and has been explained in their section.
//
// Formatter uses text/template to generate the final error message. It uses a default template, which can
// be customized using WithTemplate helper function. More information on the available variables have been
// explained in helper's documentation.
//
// # gRPC
//
// gerrors defines a set of default error codes that can translate to different error messages
// and different gRPC error codes. These default error codes have been defined to help using the
// package without much customization. However, they can be easily customized using WithCustomCoreCallback
// helper function.
//
// [Google's AIP 193]: https://google.aip.dev/193
// [error details]: https://github.com/googleapis/googleapis/blob/master/google/rpc/error_details.proto#L111
package gerrors

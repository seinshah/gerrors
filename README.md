![gerrors CI Flow](https://github.com/seinshah/gerrors/actions/workflows/ci.yaml/badge.svg) [![Maintainability](https://api.codeclimate.com/v1/badges/2df8a9a23ca8e274b360/maintainability)](https://codeclimate.com/github/seinshah/gerrors/maintainability)

# gerrors
An extensive general-purpose error handling for Go applications.
This package wraps errors with extensive information that can help debugging or pinpointing the actual reason behind the error while hiding non-sense from the end user.

# What?
A mechanism to standardize error messages with tailord information for different parties, including end-users, client developers, and server developer themselves.

All variables, constants, types, methods, and functions have already been throughly
documented. Please check [gerrors documentation on pkg.go.dev](https://pkg.go.dev/github.com/seinshah/gerrors#pkg-constants).

# Why?
When we return error messages to our clients (not users), we need to follow a
standard that has been decided among the team. We need to implement a wrapper
on top of our decided standard and use it throughout the source code to avoid
any confusion or any kind of misrepresentings.

This package comes handy in these situation. You decide on the standard you
want to follow and create a Formatter that satisfy it and start creating your
universally parsable error messages.

On the other hand, in some occasions, more detailed information need to be
shared with the client while user-facing error stays clean. This kind of metadata
helps server and client to talk the same language and share useful information
to pinpoint the cause of the error without going back and force through the source
code, logs, errors and so on.

# gerrors
An extensive general-purpose error handling for Go applications.
This package wraps errors with extensive information that can help debugging or
pinpointing the actual reason behind the error while hiding non-sense from the 
end user.

# Why This Module?
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

# Documentation
All variables, constants, types, methods, and functions have already been throughly
documented. Please check [gerrors documentation on pkg.go.dev](https://pkg.go.dev/github.com/seinshah/gerrors#pkg-constants).

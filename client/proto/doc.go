/*
Package proto is a separate module that implements a GRPC Cadence client, for sending requests.

Or it will be when it's complete.
For now, it's just reserving space, for when we fully remove Proto from the main module's dependencies.

Keeping this a separate package ensures both we and our users are not bound to a specific GRPC/Protobuf code generator,
and that we do not restrict their versions by our choices.
*/
package proto

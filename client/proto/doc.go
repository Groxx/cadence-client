/*
Package proto is a separate module that implements a GRPC Cadence client, for sending requests.

Keeping this a separate package ensures both we and our users are not bound to a specific GRPC/Protobuf code generator,
and that we do not restrict their versions by our choices.
*/
package proto

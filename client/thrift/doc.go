/*
Package thrift is a separate module that implements a Thrift Cadence client, for sending requests.

Or it will be when it's complete.
For now, it's just reserving space, for when we fully remove Thrift from the main module's dependencies.

Keeping this a separate package ensures both we and our users are not bound to a specific Thrift code version,
and that we do not restrict their versions by our choices.
*/
package thrift

/*
Package thrift is a separate module that implements a Thrift Cadence client, for sending requests.

Keeping this a separate package ensures both we and our users are not bound to a specific Thrift code version,
and that we do not restrict their versions by our choices.
*/
package thrift

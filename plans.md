High-level v2 plans:
- get rid of .gen, idls, lean entirely on cadence-idl
- consider getting rid of top-level entirely
  - maybe keep for global things, like custom error types that both activities and workflows need access to?
- keep internal/tools setup
- move to `go test ./...` including for coverage
- rewrite all interfaces by hand, no internal aliases at all
  - ... rpc agnostic?  probably required as part of this.
- typesafe preview folder, e.g. in unstable/*?
- "external" folder for interfaces of other types?
  - should we wrap zap and tally?  and provide zap/tally packages with adapters?
  - no, it'll be more pain to use than it's worth, just make a v3 when that happens.
- interceptors: only allow a single one?  is it reasonable to have internal ones if this is the case?
- error wrapping!  everywhere!
- no global registries period

more detailed?
- maybe I should target un-exposing internal things like search attrs, and then just copy/paste to make a non-aliased version?
  - feels like probably, as it keeps steps smaller and more verifiable.
- no zero-valued iotas, all should have a validate func
- definitely get rid of idls, use idl module
- get rid of mocks, users need to make their own.  or at least we should provide it in a different library.
- separate "encoded value" into "json-encoded", "unknown-encoded (dataconverter)", and maybe "from-context unknown-encoded"
  - these have different calling requirements, they should be different types
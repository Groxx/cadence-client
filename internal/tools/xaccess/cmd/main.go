package main

import (
	"flag"
	"strings"

	"go.uber.org/cadence/internal/tools/xaccess"
	"golang.org/x/tools/go/analysis/singlechecker"
)

var usage = `limit access to the given package or type name, applied in the order that is passed on the command line.
the value is a comma-separated list of package or type names, which will be split and passed to LimitAccess verbatim, see Config docs for details:
	"some.package/path" is equivalent to 'LimitAccess("some.package/path")'.
	"some.package/path.Func,other.package/etc" is equvalent to 'LimitAccess("some.package/path.Func", "other.package/etc")'.
	multiple allowed packages can be specified in a single flag, just separate with commas.`

func main() {
	flag.Func(
		"limit-access", usage,
		func(s string) error {
			parts := strings.Split(s, ",")
			xaccess.LimitAccess(parts[0], parts[1:]...)
			return nil
		})
	// do not explicitly parse, or it'll be executed twice due to singlechecker
	singlechecker.Main(xaccess.Analyzer)
}

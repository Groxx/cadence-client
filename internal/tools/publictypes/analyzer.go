package publictypes

import (
	"fmt"
	"go/types"
	"strings"

	"golang.org/x/tools/go/analysis"
)

var Analyzer = &analysis.Analyzer{
	Name:      "threepex",
	Doc:       "reports third-party library exposure via public code paths",
	Run:       run,
	FactTypes: []analysis.Fact{&exposure{}},
}

type exposure struct {
	IsExternal  bool                    // for all external types
	HasExternal bool                    // for internal, if contains external
	Exposes     map[string]types.Object // X exposes Y.  follow transitively to find path.
}

func (e *exposure) AFact() {}

func run(pass *analysis.Pass) (interface{}, error) {
	/*
		a type either *is* an external type, or it *contains* an external type (via its own type, a field, or a method).
		or neither.

		so just categorize everything that way.
	*/

	// things to exclude:
	domain := strings.Split(pass.Pkg.Path(), "/")[0]
	if len(strings.Split(domain, ".")) == 1 {
		// no domain name -> stdlib
		return nil, nil
	}
	if strings.HasSuffix(pass.Pkg.Path(), ".test") {
		// test package, cannot be exposed
		return nil, nil
	}

	// mark all non-internal as external
	if !strings.HasPrefix(pass.Pkg.Path(), "go.uber.org/cadence") {
		for _, obj := range pass.TypesInfo.Defs {
			if obj == nil {
				continue
			}
			pass.ExportObjectFact(obj, &exposure{
				IsExternal: true,
			})
		}

		return nil, nil
	}

	// check all internal for external references
	for id, obj := range pass.TypesInfo.Defs {
		_ = id
		if obj == nil {
			continue // package def, etc
		}
		switch tobj := obj.Type().(type) {
		case *types.Struct:
			for i := tobj.NumFields() - 1; i >= 0; i-- {
				f := tobj.Field(i)
				nf, ok := f.Type().(*types.Named)
				if !ok {
					continue // anonymous or basic type... yea?  verify
				}
				upsert(pass, obj, f.Id(), nf.Obj())
			}
		case *types.Signature:
			p := tobj.Params()
			for i := p.Len() - 1; i >= 0; i-- {
				pi := p.At(i)
				pt := pi.Type()
				for ppt, ok := pt.(*types.Pointer); ok; {
					if pt == ppt || ppt == ppt.Elem() {
						break
					}
				}
				npi, ok := pt.(*types.Named)
				if !ok {
					// fmt.Println(pi.String(), "is not a named type")
					continue
				}
				upsert(pass, obj, fmt.Sprintf("arg %v", i), npi.Obj())
			}
		}
	}
	return nil, nil
}

func upsert(pass *analysis.Pass, host types.Object, thing string, references types.Object) {
	var ex exposure
	pass.ImportObjectFact(references, &ex)
	if ex.IsExternal {
		if anyPrefix(
			references.Pkg().Path(),
			"go.uber.org/zap",   // known and okay
			"go.uber.org/yarpc", // known and okay

			"github.com/stretchr/testify",           // known and mostly NOT okay... ish?  rather limiting, particularly when embedded.
			"github.com/uber-go/tally",              // known and NOT okay due to v4
			"github.com/opentracing/opentracing-go", // known and NOT okay, migrating to otel
		) {
			// skip whitelisted
			return
		}
	}
	if ex.IsExternal || ex.HasExternal {
		var add exposure
		pass.ImportObjectFact(host, &add)
		if add.Exposes == nil {
			add.Exposes = make(map[string]types.Object)
		}
		add.HasExternal = true
		add.Exposes[host.Id()] = references
		fmt.Println(host.Id(), "has{", host.Name(), host.String(), "}")
		pass.ExportObjectFact(host, &add)
	}
}

func anyPrefix(s string, p ...string) bool {
	for _, pp := range p {
		if strings.HasPrefix(s, pp) {
			return true
		}
	}
	return false
}

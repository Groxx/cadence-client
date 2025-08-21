package xaccess

import (
	"go/ast"
	"go/types"

	"golang.org/x/tools/go/analysis"
)

// fact marks that the target requires accessibility checks.
// because the analyzer has the accessibility config in memory, no actual data needs to be stored in the fact itself.
type fact struct{}

func (f *fact) AFact() {}

// Analyzer is a simple analyzer that finds a hardcoded target and exports a fact
var Analyzer = &analysis.Analyzer{
	Name:      "xaccess",
	Doc:       "controls access to objects",
	Run:       run,
	FactTypes: []analysis.Fact{(*fact)(nil)},
}

type Rules struct {
	Allowed map[string]bool
}

type Config map[string]Rules

// ControlledAccess is a shared global to allow code to override,
// because that's kinda the only option in bazel,
// and there's no great way to pass it in for plain analyzers either.
//
// I hate it, but oh well.
var ControlledAccess = Config{}

func run(pass *analysis.Pass) (interface{}, error) {
	markObjects(pass, ControlledAccess)
	checkUsage(pass, ControlledAccess)
	return nil, nil
}

func markObjects(pass *analysis.Pass, controlled Config) {
	for _, file := range pass.Files {
		// loop over top-level decls, simpler to check than TypesInfo.Defs
		for _, decl := range file.Decls {
			// Check if this is a function or type declaration
			if obj := declaredObject(pass, decl); obj != nil {
				// look for package-wide restrictions, mark everything
				if _, ok := controlled[pass.Pkg.Path()]; ok {
					pass.ExportObjectFact(obj, &fact{})
					continue
				}
				// else look for exact matches
				fullName := getFullName(obj)
				if _, ok := controlled[fullName]; ok {
					pass.ExportObjectFact(obj, &fact{})
				}
			}
		}
	}
}

func checkUsage(pass *analysis.Pass, controlled Config) {
	for id, usedThing := range pass.TypesInfo.Uses {
		if usedThing.Pkg() == pass.Pkg {
			continue // anything defined can be used in its own package
			// TODO: should this also allow subpackages?  probably?
		}

		var out fact
		if !pass.ImportObjectFact(usedThing, &out) {
			continue // not a tracked object, so it's not interesting
		}

		name := getFullName(usedThing)
		if controlled[name].Allowed[pass.Pkg.Path()] {
			continue // allowed to use on a specific level
		}

		if controlled[usedThing.Pkg().Path()].Allowed[pass.Pkg.Path()] {
			continue // allowed to use on a package-wide level
		}

		pass.Reportf(id.Pos(), "using controlled object %s, marked to require access checks", name)
	}
}

// declaredObject extracts the types.Object from a declaration, if any
func declaredObject(pass *analysis.Pass, decl interface{}) types.Object {
	switch d := decl.(type) {
	case *ast.FuncDecl:
		if d.Name != nil && d.Name.IsExported() {
			return pass.TypesInfo.Defs[d.Name]
		}
	case *ast.GenDecl:
		for _, spec := range d.Specs {
			switch s := spec.(type) {
			case *ast.TypeSpec:
				if s.Name != nil && s.Name.IsExported() {
					return pass.TypesInfo.Defs[s.Name]
				}
			case *ast.ValueSpec:
				for _, name := range s.Names {
					if name != nil && name.IsExported() {
						return pass.TypesInfo.Defs[name]
					}
				}
			}
		}
	}
	return nil
}

// matchesTarget checks if the given object matches our hardcoded target
func getFullName(obj types.Object) string {
	if obj == nil {
		// irrelevant, can't be accessed
		return "<nil obj>"
	}

	// Get the full qualified name
	pkg := obj.Pkg()
	if pkg == nil {
		// probably not possible, seems like it'd be a builtin?
		return "<nil pkg>"
	}

	return pkg.Path() + "." + obj.Name()
}

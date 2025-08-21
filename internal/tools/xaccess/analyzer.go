package xaccess

import (
	"fmt"
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

// LimitAccess limits the access to a "thing" to the list of allowed packages.
//
// If no allowed paths are given, the path/item is completely blocked.
// If some paths are given, they will be *added to* the allowed list.
//
// All values must be specific, i.e. you cannot use a wildcard or glob.
// If you really need to do that, build a helper in the allowed package,
// and use the blocked thing through that helper.
//
// No duplicates or conflicting rules are allowed, i.e. you cannot
// allow something on some paths and then block it completely:
//
//	LimitAccess("example.org/x", "your.org")
//	LimitAccess("example.org/x") // panics
//
// The reverse order is allowed though, i.e. you can block something
// and then allow it on some paths:
//
//	LimitAccess("example.org/x")
//	LimitAccess("example.org/x", "your.org")      // only your.org is allowed
//	LimitAccess("example.org/x", "someother.org") // both orgs are allowed
//	LimitAccess("example.org/x", "your.org")      // panics, this is a duplicate
//
// You can target entire packages and/or specific (accessible) things within a package,
// which can be specified in any order:
//
//	LimitAccess("example.org/x")       // everything in the package is limited
//	LimitAccess("example.org/x.Thing") // only this one `Thing` is limited
//
// Limiting a package does not currently prevent it from being underscore-imported,
// as this does not actually "use" anything, so it can still be used to trigger
// `init` behavior.  This may be restricted later, at which point there should be a
// `your/package.init` path to allow this behavior.
//
// To clear all access rules, e.g. for testing, pass only an empty string:
//
//	LimitAccess("") // clears all access rules
func LimitAccess(pkgOrTypeName string, allowed ...string) {
	if pkgOrTypeName == "" && len(allowed) == 0 {
		ControlledAccess = Config{}
		return
	}

	rules, ok := ControlledAccess[pkgOrTypeName]
	if ok && len(allowed) == 0 {
		if len(rules.Allowed) == 0 {
			panic(fmt.Sprintf("limited-access %q is already limited", pkgOrTypeName))
		} else {
			panic(fmt.Sprintf("limited-access %q is already allowing some packages, cannot be fully limited: %v", pkgOrTypeName, rules.Allowed))
		}
	}

	if len(rules.Allowed) == 0 {
		rules.Allowed = map[string]bool{}
	}
	for _, a := range allowed {
		if _, ok := rules.Allowed[a]; ok {
			panic(fmt.Sprintf("limited-access %q is already allowing %q", pkgOrTypeName, a))
		}
		rules.Allowed[a] = true
	}
	ControlledAccess[pkgOrTypeName] = rules
}

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
			for _, obj := range declaredObjects(pass, decl) {
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

func declaredObjects(pass *analysis.Pass, decl ast.Decl) []types.Object {
	switch d := decl.(type) {
	case *ast.FuncDecl:
		// top-level functions and methods on types, i.e.:
		//   func (r receiver) Method() {}
		// ^ because that's still "a function".
		//
		// both of these can be accessed and have a "type name", e.g. "import/path.(*Thing).Method"
		if d.Name != nil && d.Name.IsExported() {
			return []types.Object{pass.TypesInfo.Defs[d.Name]}
		}
	case *ast.GenDecl:
		var all []types.Object
		for _, spec := range d.Specs {
			switch s := spec.(type) {
			case *ast.TypeSpec:
				// the types themselves are the only things that have "names".
				// this covers everything starting with `type`, e.g. structs, interfaces, etc.
				if s.Name != nil && s.Name.IsExported() {
					all = append(all, pass.TypesInfo.Defs[s.Name])
				}
			case *ast.ValueSpec:
				for _, name := range s.Names {
					if name != nil && name.IsExported() {
						all = append(all, pass.TypesInfo.Defs[name])
					}
				}
			}
		}
		return all
	}
	return nil
}

func getFullName(obj types.Object) string {
	if obj == nil {
		// irrelevant, can't be accessed
		return "<nil obj>"
	}

	// Get the full qualified name
	pkg := obj.Pkg()
	if pkg == nil {
		// probably not possible, seems like it'd be a builtin?
		// that wouldn't be "exported" though, they're all lowercase.
		return "<nil pkg>"
	}

	return pkg.Path() + "." + obj.Name()
}

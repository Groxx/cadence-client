package structfill

import (
	"go/ast"
	"go/token"
	"go/types"
	"strings"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
)

const (
	// Magic comment strings
	mustFillComment = "lint:must-fill"
	canSkipComment  = "lint:can-skip"
)

// MustFillFact marks types that require all fields to be explicitly filled
// It also stores which fields can be skipped via // lint:can-skip comments
type MustFillFact struct {
	SkippableFields []string
}

// AFact implements analysis.Fact interface
func (*MustFillFact) AFact() {}

var Analyzer = &analysis.Analyzer{
	Name:      "structfill",
	Doc:       "Ensures structs with // lint:must-fill comments have all fields explicitly filled",
	Run:       run,
	FactTypes: []analysis.Fact{(*MustFillFact)(nil)},
	Requires:  []*analysis.Analyzer{inspect.Analyzer},
}

func run(pass *analysis.Pass) (interface{}, error) {
	markStructs(pass)
	enforceRules(pass)
	return nil, nil
}

func markStructs(pass *analysis.Pass) {
	// Get the inspector from the inspect analyzer
	inspect := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)

	// Only look at GenDecl nodes for efficiency
	nodeFilter := []ast.Node{(*ast.GenDecl)(nil)}

	inspect.Preorder(nodeFilter, func(node ast.Node) {
		n := node.(*ast.GenDecl)

		// Check if this is a type declaration
		if n.Tok != token.TYPE {
			return
		}

		// Check each type spec in the declaration
		for _, spec := range n.Specs {
			if typeSpec, ok := spec.(*ast.TypeSpec); ok {
				if structType, ok := typeSpec.Type.(*ast.StructType); ok {
					// Found a struct declaration, check for the magic comment
					if hasMustFillComment(n.Doc) {
						// Get the type object for this struct
						obj := pass.TypesInfo.Defs[typeSpec.Name]
						if obj != nil {
							// Find which fields can be skipped, and export the fact
							skippableFields := parseSkippableFields(structType)
							pass.ExportObjectFact(obj, &MustFillFact{
								SkippableFields: skippableFields,
							})
						}
					}
				}
			}
		}
	})
}

func enforceRules(pass *analysis.Pass) {
	// Get the inspector from the inspect analyzer
	inspect := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)

	// Only look at CompositeLit nodes for efficiency
	nodeFilter := []ast.Node{(*ast.CompositeLit)(nil)}

	inspect.Preorder(nodeFilter, func(node ast.Node) {
		n := node.(*ast.CompositeLit)
		// Found a struct literal, check if it needs complete filling
		checkStructLiteral(pass, n)
	})
}

func hasMustFillComment(commentGroup *ast.CommentGroup) bool {
	if commentGroup == nil {
		return false
	}

	// commentGroup.Text() removes // markers and leading/trailing whitespace
	text := commentGroup.Text()
	lines := strings.Split(text, "\n")

	for _, line := range lines {
		if strings.TrimSpace(line) == mustFillComment {
			return true
		}
	}
	return false
}

func parseSkippableFields(structType *ast.StructType) []string {
	var skippable []string

	for _, field := range structType.Fields.List {
		if hasCanSkipComment(field.Doc) || hasCanSkipComment(field.Comment) {
			// Add all names in this field as skippable
			for _, name := range field.Names {
				skippable = append(skippable, name.Name)
			}
		}
	}

	return skippable
}

func hasCanSkipComment(commentGroup *ast.CommentGroup) bool {
	if commentGroup == nil {
		return false
	}

	// commentGroup.Text() removes // markers and leading/trailing whitespace
	text := commentGroup.Text()
	lines := strings.Split(text, "\n")

	for _, line := range lines {
		if strings.TrimSpace(line) == canSkipComment {
			return true
		}
	}
	return false
}

// checkStructLiteral verifies that struct literals for MustFill types have all fields explicitly filled
func checkStructLiteral(pass *analysis.Pass, lit *ast.CompositeLit) {
	// Get the type of this composite literal
	tv, ok := pass.TypesInfo.Types[lit]
	if !ok {
		return
	}

	// Check if it's a struct type
	structType, ok := tv.Type.Underlying().(*types.Struct)
	if !ok {
		return
	}

	// Get the named type (if any) to check for facts
	var namedType *types.Named
	if named, ok := tv.Type.(*types.Named); ok {
		namedType = named
	} else {
		// ignore anonymous structs.  arguably we could, but I don't think that's worth encouraging.
		return
	}

	// Check if this type has the MustFillFact
	var fact MustFillFact
	if !pass.ImportObjectFact(namedType.Obj(), &fact) {
		return // This type doesn't require complete filling
	}

	// Convert skippable fields slice to map for efficient lookup
	skippableFields := make(map[string]bool)
	for _, field := range fact.SkippableFields {
		skippableFields[field] = true
	}

	// Build a map of filled fields from the literal
	filledFields := make(map[string]bool)
	for _, elt := range lit.Elts {
		if kv, ok := elt.(*ast.KeyValueExpr); ok {
			if ident, ok := kv.Key.(*ast.Ident); ok {
				filledFields[ident.Name] = true
			}
		} else {
			// positional initialization is always safe in this respect, as it cannot be missing fields
			return
		}
	}

	// Check all struct fields are present (unless they can be skipped)
	for i := 0; i < structType.NumFields(); i++ {
		field := structType.Field(i)
		if !field.Exported() {
			continue // Skip unexported fields
		}

		if !filledFields[field.Name()] && !skippableFields[field.Name()] {
			pass.Reportf(lit.Pos(), "missing %q", strings.ToLower(field.Name()))
		}
	}
}

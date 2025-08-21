package structfill

import (
	"go/ast"
	"go/token"
	"go/types"
	"path/filepath"
	"strings"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
)

const (
	// magic comment strings
	mustFillComment = "lint:must-fill"
	canSkipComment  = "lint:can-skip"
)

var (
	// enforcePaths specifies package paths where ALL structs should be treated as must-fill
	enforcePaths string
	// ignoreTests controls whether to skip analysis in test files (_test.go)
	ignoreTests bool
	// ignoreGenerated controls whether to skip analysis in generated files
	ignoreGenerated bool
)

// MustFillFact marks types that require all fields to be explicitly filled
// It also stores which fields can be skipped via // lint:can-skip comments
type MustFillFact struct {
	SkippableFields []string
}

// AFact implements analysis.Fact interface
func (*MustFillFact) AFact() {}

var Analyzer = &analysis.Analyzer{
	Name: "structfill",
	Doc: `Ensures structs with // lint:must-fill comments have all fields explicitly filled.

Use -enforce flag to treat ALL structs in specified packages as must-fill:
  -enforce="github.com/example/api,github.com/example/types"
  -enforce="github.com/example/..."  (matches all subpackages)

Use -ignore-tests flag to skip analysis of test files:
  -ignore-tests=true  (skips all *_test.go files)

Use -ignore-generated flag to skip analysis of generated files:
  -ignore-generated=true  (skips files with 'Code generated' and 'DO NOT EDIT' comments)

Comments:
  // lint:must-fill   - struct requires all fields to be filled
  // lint:can-skip    - at struct level: exempts struct from -enforce mode
                     - at field level: allows field to be omitted`,
	Run:       run,
	FactTypes: []analysis.Fact{(*MustFillFact)(nil)},
	Requires:  []*analysis.Analyzer{inspect.Analyzer},
}

func init() {
	Analyzer.Flags.StringVar(&enforcePaths, "enforce", "", "comma-separated list of import paths where all structs are treated as must-fill")
	Analyzer.Flags.BoolVar(&ignoreTests, "ignore-tests", false, "skip analysis of test files (_test.go)")
	Analyzer.Flags.BoolVar(&ignoreGenerated, "ignore-generated", false, "skip analysis of generated files (containing 'Code generated' and 'DO NOT EDIT' comments)")
}

func run(pass *analysis.Pass) (interface{}, error) {
	markStructs(pass)
	enforceRules(pass)
	return nil, nil
}

// shouldIgnoreFile returns true if the file should be ignored based on ignore flags
func shouldIgnoreFile(pass *analysis.Pass, filename string) bool {
	if ignoreTests && strings.HasSuffix(filepath.Base(filename), "_test.go") {
		return true
	}

	if ignoreGenerated {
		// find the file in pass.Files to check for generated file comments
		for _, file := range pass.Files {
			filePos := pass.Fset.Position(file.Pos())
			if filePos.Filename == filename {
				return isGeneratedFile(file)
			}
		}
	}

	return false
}

// isGeneratedFile checks if a file contains the standard generated file comment
func isGeneratedFile(file *ast.File) bool {
	if len(file.Comments) == 0 {
		return false
	}

	// check the first few comment groups for the generated file pattern
	for _, commentGroup := range file.Comments[:min(len(file.Comments), 3)] {
		text := commentGroup.Text()
		// look for the standard pattern: "Code generated" and "DO NOT EDIT"
		if strings.Contains(text, "Code generated") && strings.Contains(text, "DO NOT EDIT") {
			return true
		}
	}

	return false
}

// min returns the smaller of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// isPackageEnforced checks if the current package should have all structs treated as must-fill
func isPackageEnforced(pass *analysis.Pass) bool {
	if enforcePaths == "" {
		return false
	}

	currentPkg := pass.Pkg.Path()
	enforcedPaths := strings.Split(enforcePaths, ",")

	for _, path := range enforcedPaths {
		path = strings.TrimSpace(path)
		if path == "" {
			continue
		}
		// check for exact match or prefix match with "/..." suffix
		if currentPkg == path {
			return true
		}
		if strings.HasSuffix(path, "/...") {
			prefix := strings.TrimSuffix(path, "/...")
			if strings.HasPrefix(currentPkg, prefix+"/") || currentPkg == prefix {
				return true
			}
		}
	}
	return false
}

func markStructs(pass *analysis.Pass) {
	// get the inspector from the inspect analyzer
	inspect := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)

	// only look at GenDecl nodes for efficiency
	nodeFilter := []ast.Node{(*ast.GenDecl)(nil)}

	// check if this package is in enforcer mode
	packageEnforced := isPackageEnforced(pass)

	inspect.Preorder(nodeFilter, func(node ast.Node) {
		n := node.(*ast.GenDecl)

		// skip if this node is in a test file and we're ignoring tests
		pos := pass.Fset.Position(n.Pos())
		if shouldIgnoreFile(pass, pos.Filename) {
			return
		}

		// check if this is a type declaration
		if n.Tok != token.TYPE {
			return
		}

		// check each type spec in the declaration
		for _, spec := range n.Specs {
			if typeSpec, ok := spec.(*ast.TypeSpec); ok {
				if structType, ok := typeSpec.Type.(*ast.StructType); ok {
					// found a struct declaration
					// check if it should be treated as must-fill (either has comment or package is enforced)
					hasMustFill := hasCommentWithText(n.Doc, mustFillComment)
					hasCanSkip := hasCommentWithText(n.Doc, canSkipComment)

					// determine if we should enforce: must have explicit must-fill OR be in enforced package AND not have can-skip
					shouldEnforce := hasMustFill || (packageEnforced && !hasCanSkip)

					if shouldEnforce {
						// get the type object for this struct
						obj := pass.TypesInfo.Defs[typeSpec.Name]
						if obj != nil {
							// find which fields can be skipped, and export the fact
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
	inspect := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)

	// only look at CompositeLit nodes for efficiency
	nodeFilter := []ast.Node{(*ast.CompositeLit)(nil)}

	inspect.Preorder(nodeFilter, func(node ast.Node) {
		n := node.(*ast.CompositeLit)

		// skip if this node is in a test file and we're ignoring tests
		pos := pass.Fset.Position(n.Pos())
		if shouldIgnoreFile(pass, pos.Filename) {
			return
		}

		// found a struct literal, check if it needs complete filling
		checkStructLiteral(pass, n)
	})
}

// hasCommentWithText checks if the comment group contains a line that starts with the target string
// the target must be at the beginning of the line, optionally followed by a space and additional text
func hasCommentWithText(commentGroup *ast.CommentGroup, target string) bool {
	if commentGroup == nil {
		return false
	}

	// commentGroup.Text() removes // markers and leading/trailing whitespace
	text := commentGroup.Text()
	lines := strings.Split(text, "\n")

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == target {
			return true // exact match
		}
		if strings.HasPrefix(trimmed, target+" ") {
			return true // can have additional text after the target
		}
	}
	return false
}

func parseSkippableFields(structType *ast.StructType) []string {
	var skippable []string

	for _, field := range structType.Fields.List {
		if hasCommentWithText(field.Doc, canSkipComment) || hasCommentWithText(field.Comment, canSkipComment) {
			// add all names in this field as skippable
			for _, name := range field.Names {
				skippable = append(skippable, name.Name)
			}
		}
	}

	return skippable
}

// checkStructLiteral verifies that struct literals for MustFill types have all fields explicitly filled
func checkStructLiteral(pass *analysis.Pass, lit *ast.CompositeLit) {
	// get the type of this composite literal
	tv, ok := pass.TypesInfo.Types[lit]
	if !ok {
		return
	}

	// must be a struct type
	structType, ok := tv.Type.Underlying().(*types.Struct)
	if !ok {
		return
	}

	// named types have an Obj() so we can get the fact.
	// TODO: unsure if there are other possibilities here
	var namedType *types.Named
	if named, ok := tv.Type.(*types.Named); ok {
		namedType = named
	} else {
		// ignore anonymous structs.  arguably we could, but I don't think that's worth encouraging.
		return
	}

	var fact MustFillFact
	if !pass.ImportObjectFact(namedType.Obj(), &fact) {
		return // un-checked type
	}

	// convert skippable fields slice to map for efficient lookup
	skippableFields := make(map[string]bool)
	for _, field := range fact.SkippableFields {
		skippableFields[field] = true
	}

	// build a map of filled fields from the literal
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

	// check all struct fields are present (unless they can be skipped)
	for i := 0; i < structType.NumFields(); i++ {
		field := structType.Field(i)
		if !field.Exported() {
			continue // skip unexported fields
		}

		if !filledFields[field.Name()] && !skippableFields[field.Name()] {
			pass.Reportf(lit.Pos(), "missing %q", field.Name())
		}
	}
}

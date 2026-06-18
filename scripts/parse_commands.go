package main

import (
	"encoding/json"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
)

type CommandInfo struct {
	GoFile     string `json:"go_file"`
	CmdName    string `json:"cmd_name"` // from godoc comment first word
	VarName    string `json:"var_name"`
	Doc        string `json:"doc"`
	Use        string `json:"use"`
	Short      string `json:"short"`
	HasCommand bool   `json:"has_command"`
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run parse_commands.go <dir>")
		os.Exit(1)
	}
	dir := os.Args[1]
	files, err := os.ReadDir(dir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading dir: %v\n", err)
		os.Exit(1)
	}

	commands := []CommandInfo{}

	for _, f := range files {
		if f.IsDir() {
			continue
		}
		name := f.Name()
		if !strings.HasSuffix(name, ".go") || strings.HasSuffix(name, "_test.go") || name == "main.go" || name == "root.go" {
			continue
		}

		filePath := filepath.Join(dir, name)
		info, err := parseGoFile(filePath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error parsing file %s: %v\n", filePath, err)
			continue
		}
		if info.HasCommand {
			commands = append(commands, info)
		}
	}

	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	if err := enc.Encode(commands); err != nil {
		fmt.Fprintf(os.Stderr, "Error encoding JSON: %v\n", err)
		os.Exit(1)
	}
}

func parseGoFile(filePath string) (CommandInfo, error) {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, filePath, nil, parser.ParseComments)
	if err != nil {
		return CommandInfo{}, err
	}

	info := CommandInfo{
		GoFile: filePath,
	}

	// Helper to check if a type is *cobra.Command
	isCobraCmdType := func(expr ast.Expr) bool {
		star, ok := expr.(*ast.StarExpr)
		if !ok {
			return false
		}
		sel, ok := star.X.(*ast.SelectorExpr)
		if !ok {
			return false
		}
		xIdent, ok := sel.X.(*ast.Ident)
		if !ok {
			return false
		}
		return xIdent.Name == "cobra" && sel.Sel.Name == "Command"
	}

	// Helper to extract fields from &cobra.Command composite literal
	extractCobraCmdFields := func(expr ast.Expr) (use, short string, found bool) {
		unary, ok := expr.(*ast.UnaryExpr)
		if !ok || unary.Op != token.AND {
			return
		}
		comp, ok := unary.X.(*ast.CompositeLit)
		if !ok {
			return
		}
		sel, ok := comp.Type.(*ast.SelectorExpr)
		if !ok {
			return
		}
		xIdent, ok := sel.X.(*ast.Ident)
		if !ok || xIdent.Name != "cobra" || sel.Sel.Name != "Command" {
			return
		}

		found = true
		for _, el := range comp.Elts {
			kv, ok := el.(*ast.KeyValueExpr)
			if !ok {
				continue
			}
			keyIdent, ok := kv.Key.(*ast.Ident)
			if !ok {
				continue
			}
			lit, ok := kv.Value.(*ast.BasicLit)
			if !ok || lit.Kind != token.STRING {
				continue
			}
			val := strings.Trim(lit.Value, "`\"")
			if keyIdent.Name == "Use" {
				use = val
			} else if keyIdent.Name == "Short" {
				short = val
			}
		}
		return
	}

	// Inspect top-level declarations
	for _, decl := range file.Decls {
		switch d := decl.(type) {
		case *ast.GenDecl:
			if d.Tok == token.VAR {
				for _, spec := range d.Specs {
					vspec, ok := spec.(*ast.ValueSpec)
					if !ok {
						continue
					}
					// Check type
					isCmd := false
					if vspec.Type != nil && isCobraCmdType(vspec.Type) {
						isCmd = true
					}
					var use, short string
					for _, val := range vspec.Values {
						if u, s, ok := extractCobraCmdFields(val); ok {
							isCmd = true
							if use == "" {
								use = u
							}
							if short == "" {
								short = s
							}
						}
					}
					if isCmd {
						info.HasCommand = true
						if len(vspec.Names) > 0 {
							info.VarName = vspec.Names[0].Name
						}
						info.Use = use
						info.Short = short
						// Get comment
						var commentGroup *ast.CommentGroup
						if vspec.Doc != nil {
							commentGroup = vspec.Doc
						} else if d.Doc != nil {
							commentGroup = d.Doc
						}
						if commentGroup != nil {
							info.Doc = strings.TrimSpace(commentGroup.Text())
							lines := strings.Split(info.Doc, "\n")
							if len(lines) > 0 && len(lines[0]) > 0 {
								// First word of comment text
								words := strings.Fields(lines[0])
								if len(words) > 0 {
									info.CmdName = words[0]
								}
							}
						}
						break
					}
				}
			}
		case *ast.FuncDecl:
			// Check if return type is *cobra.Command
			isCmd := false
			if d.Type.Results != nil {
				for _, field := range d.Type.Results.List {
					if isCobraCmdType(field.Type) {
						isCmd = true
						break
					}
				}
			}
			if isCmd {
				info.HasCommand = true
				info.VarName = d.Name.Name
				if d.Doc != nil {
					info.Doc = strings.TrimSpace(d.Doc.Text())
					lines := strings.Split(info.Doc, "\n")
					if len(lines) > 0 && len(lines[0]) > 0 {
						words := strings.Fields(lines[0])
						if len(words) > 0 {
							info.CmdName = words[0]
						}
					}
				}
				break
			}
		}
		if info.HasCommand {
			break
		}
	}

	return info, nil
}

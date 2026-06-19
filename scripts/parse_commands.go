package main

import (
	"encoding/json"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type FlagInfo struct {
	Name        string `json:"name"`
	Short       string `json:"short"`
	Type        string `json:"type"`
	Description string `json:"description"`
}

type FlagGroupInfo struct {
	Name        string     `json:"name"`
	Description string     `json:"description"`
	Example     string     `json:"example"`
	Flags       []FlagInfo `json:"flags"`
}

type CommandInfo struct {
	GoFile     string          `json:"go_file"`
	CmdName    string          `json:"cmd_name"` // from godoc comment first word
	VarName    string          `json:"var_name"`
	Doc        string          `json:"doc"`
	Use        string          `json:"use"`
	Short      string          `json:"short"`
	HasCommand bool            `json:"has_command"`
	FlagGroups []FlagGroupInfo `json:"flag_groups"`
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
		if !strings.HasSuffix(name, ".go") || strings.HasSuffix(name, "_test.go") || name == "main.go" {
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

func parseExprString(expr ast.Expr) string {
	switch e := expr.(type) {
	case *ast.BasicLit:
		if e.Kind == token.STRING {
			return strings.Trim(e.Value, "`\"")
		}
	case *ast.Ident:
		if e.Name == "name" {
			return "l"
		}
		if e.Name == "short" {
			return "an ls replacement"
		}
		return e.Name
	case *ast.BinaryExpr:
		if e.Op == token.ADD {
			return parseExprString(e.X) + parseExprString(e.Y)
		}
	}
	return ""
}

func parseCommentGroup(text string) (groupName, desc, example string, isGroup bool) {
	lines := strings.Split(text, "\n")
	var parsingDesc, parsingExample bool
	var descLines, exampleLines []string

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "@group:") {
			groupName = strings.TrimSpace(strings.TrimPrefix(trimmed, "@group:"))
			isGroup = true
			parsingDesc = false
			parsingExample = false
			continue
		}
		if strings.HasPrefix(trimmed, "@description:") {
			parsingDesc = true
			parsingExample = false
			rest := strings.TrimSpace(strings.TrimPrefix(trimmed, "@description:"))
			if rest != "" {
				descLines = append(descLines, rest)
			}
			continue
		}
		if strings.HasPrefix(trimmed, "@example:") {
			parsingExample = true
			parsingDesc = false
			rest := strings.TrimSpace(strings.TrimPrefix(trimmed, "@example:"))
			if rest != "" {
				exampleLines = append(exampleLines, rest)
			}
			continue
		}

		if parsingDesc {
			descLines = append(descLines, line)
		} else if parsingExample {
			exampleLines = append(exampleLines, line)
		}
	}

	if isGroup {
		desc = cleanBlockText(descLines)
		example = cleanBlockText(exampleLines)
	}
	return
}

func cleanBlockText(lines []string) string {
	if len(lines) == 0 {
		return ""
	}
	minChars := -1
	for _, l := range lines {
		if strings.TrimSpace(l) == "" {
			continue
		}
		chars := 0
		for _, r := range l {
			if r == ' ' || r == '\t' {
				chars++
			} else {
				break
			}
		}
		if minChars == -1 || chars < minChars {
			minChars = chars
		}
	}
	if minChars == -1 {
		minChars = 0
	}

	var cleaned []string
	for _, l := range lines {
		if len(l) <= minChars {
			cleaned = append(cleaned, "")
		} else {
			cleaned = append(cleaned, strings.TrimRight(l[minChars:], " \t"))
		}
	}
	start := 0
	for start < len(cleaned) && cleaned[start] == "" {
		start++
	}
	end := len(cleaned)
	for end > start && cleaned[end-1] == "" {
		end--
	}
	return strings.Join(cleaned[start:end], "\n")
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

	// Helper to extract flag info from a CallExpr
	extractFlagInfo := func(call *ast.CallExpr) (name, short, flagType, description string, ok bool) {
		sel, okSel := call.Fun.(*ast.SelectorExpr)
		if !okSel {
			return "", "", "", "", false
		}
		funName := sel.Sel.Name
		recvCall, okRecvCall := sel.X.(*ast.CallExpr)
		if !okRecvCall {
			return "", "", "", "", false
		}
		recvSel, okRecvSel := recvCall.Fun.(*ast.SelectorExpr)
		if !okRecvSel || (recvSel.Sel.Name != "Flags" && recvSel.Sel.Name != "PersistentFlags") {
			return "", "", "", "", false
		}

		isFlagMethod := false
		for _, prefix := range []string{"Bool", "String", "Int", "Duration", "Float", "Var"} {
			if strings.HasPrefix(funName, prefix) {
				isFlagMethod = true
				break
			}
		}
		if !isFlagMethod {
			return "", "", "", "", false
		}

		args := call.Args
		if strings.HasSuffix(funName, "VarP") {
			if len(args) >= 5 {
				name = parseExprString(args[1])
				short = parseExprString(args[2])
				description = parseExprString(args[4])
				flagType = strings.TrimSuffix(funName, "VarP")
				ok = true
			}
		} else if strings.HasSuffix(funName, "Var") {
			if len(args) >= 4 {
				name = parseExprString(args[1])
				short = ""
				description = parseExprString(args[3])
				flagType = strings.TrimSuffix(funName, "Var")
				ok = true
			}
		} else {
			if strings.HasSuffix(funName, "P") {
				if len(args) >= 4 {
					name = parseExprString(args[0])
					short = parseExprString(args[1])
					description = parseExprString(args[3])
					flagType = strings.TrimSuffix(funName, "P")
					ok = true
				}
			} else {
				if len(args) >= 3 {
					name = parseExprString(args[0])
					short = ""
					description = parseExprString(args[2])
					flagType = funName
					ok = true
				}
			}
		}
		return
	}

	// Helper to find &cobra.Command inside a node (e.g. FuncDecl or ValueSpec)
	findCobraCommandInNode := func(node ast.Node) (use, short string, found bool) {
		ast.Inspect(node, func(n ast.Node) bool {
			unary, ok := n.(*ast.UnaryExpr)
			if !ok || unary.Op != token.AND {
				return true
			}
			comp, ok := unary.X.(*ast.CompositeLit)
			if !ok {
				return true
			}
			sel, ok := comp.Type.(*ast.SelectorExpr)
			if !ok {
				return true
			}
			xIdent, ok := sel.X.(*ast.Ident)
			if !ok || xIdent.Name != "cobra" || sel.Sel.Name != "Command" {
				return true
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
				if keyIdent.Name == "Use" {
					use = parseExprString(kv.Value)
				} else if keyIdent.Name == "Short" {
					short = parseExprString(kv.Value)
				}
			}
			return false
		})
		return
	}

	// Inspect top-level declarations to find command definition and set basic info
	for _, decl := range file.Decls {
		switch d := decl.(type) {
		case *ast.GenDecl:
			if d.Tok == token.VAR {
				for _, spec := range d.Specs {
					vspec, ok := spec.(*ast.ValueSpec)
					if !ok {
						continue
					}
					isCmd := false
					if vspec.Type != nil && isCobraCmdType(vspec.Type) {
						isCmd = true
					}
					var use, short string
					for _, val := range vspec.Values {
						if u, s, ok := findCobraCommandInNode(val); ok {
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
								words := strings.Fields(lines[0])
								if len(words) > 0 {
									info.CmdName = strings.Trim(words[0], "`\"'")
								}
							}
						}
						break
					}
				}
			}
		case *ast.FuncDecl:
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
				if u, s, ok := findCobraCommandInNode(d.Body); ok {
					info.Use = u
					info.Short = s
				}
				if d.Doc != nil {
					info.Doc = strings.TrimSpace(d.Doc.Text())
					lines := strings.Split(info.Doc, "\n")
					if len(lines) > 0 && len(lines[0]) > 0 {
						words := strings.Fields(lines[0])
						if len(words) > 0 {
							info.CmdName = strings.Trim(words[0], "`\"'")
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

	if info.HasCommand {
		type ASTNodeWithLine struct {
			Line     int
			Comment  *ast.CommentGroup
			FlagCall *ast.CallExpr
		}
		var items []ASTNodeWithLine

		for _, cg := range file.Comments {
			items = append(items, ASTNodeWithLine{
				Line:    fset.Position(cg.Pos()).Line,
				Comment: cg,
			})
		}

		ast.Inspect(file, func(n ast.Node) bool {
			if call, ok := n.(*ast.CallExpr); ok {
				if _, _, _, _, isFlag := extractFlagInfo(call); isFlag {
					items = append(items, ASTNodeWithLine{
						Line:     fset.Position(call.Pos()).Line,
						FlagCall: call,
					})
				}
			}
			return true
		})

		sort.Slice(items, func(i, j int) bool {
			return items[i].Line < items[j].Line
		})

		var flagGroups []FlagGroupInfo
		var currentGroup *FlagGroupInfo

		for _, item := range items {
			if item.Comment != nil {
				gName, desc, example, isGroup := parseCommentGroup(item.Comment.Text())
				if isGroup {
					if currentGroup != nil {
						flagGroups = append(flagGroups, *currentGroup)
					}
					currentGroup = &FlagGroupInfo{
						Name:        gName,
						Description: desc,
						Example:     example,
					}
				}
			} else if item.FlagCall != nil {
				name, short, flagType, desc, ok := extractFlagInfo(item.FlagCall)
				if ok {
					if currentGroup == nil {
						currentGroup = &FlagGroupInfo{
							Name: "Flags",
						}
					}
					currentGroup.Flags = append(currentGroup.Flags, FlagInfo{
						Name:        name,
						Short:       short,
						Type:        flagType,
						Description: desc,
					})
				}
			}
		}
		if currentGroup != nil {
			flagGroups = append(flagGroups, *currentGroup)
		}

		info.FlagGroups = flagGroups
	}

	return info, nil
}

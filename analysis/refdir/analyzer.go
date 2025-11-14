package refdir

import (
	"errors"
	"flag"
	"fmt"
	"go/ast"
	"go/token"
	"go/types"

	"github.com/ppipada/refdir/analysis/refdir/color"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
)

var Analyzer = &analysis.Analyzer{
	Name:     "refdir",
	Doc:      "Report potential reference-to-declaration ordering issues",
	Run:      run,
	Flags:    flag.FlagSet{},
	Requires: []*analysis.Analyzer{inspect.Analyzer},
}

var (
	verbose  bool
	colorize bool
)

type RefKind string

const (
	Func     RefKind = "func"
	Type     RefKind = "type"
	RecvType RefKind = "recvtype"
	Var      RefKind = "var"
	Const    RefKind = "const"
)

var RefKinds = []RefKind{
	Func,
	Type,
	RecvType,
	Var,
	Const,
}

type Direction string

const (
	Down   Direction = "down"
	Up     Direction = "up"
	Ignore Direction = "ignore"
)

var Directions = []Direction{
	Down,
	Up,
	Ignore,
}

var RefOrder = map[RefKind]Direction{
	Func:     Down,
	Type:     Up,
	RecvType: Up,
	Var:      Up,
	Const:    Up,
}

func init() {
	Analyzer.Flags.BoolVar(&verbose, "verbose", false, `print all details`)
	Analyzer.Flags.BoolVar(&colorize, "color", true, `colorize terminal`)
	addDirectionFlag := func(kind RefKind, desc string) {
		Analyzer.Flags.Func(
			string(kind)+"-dir",
			fmt.Sprintf("%s (default %s)", desc, RefOrder[kind]),
			func(s string) error {
				switch dir := Direction(s); dir {
				case Down, Up, Ignore:
					RefOrder[kind] = dir
					return nil
				default:
					return fmt.Errorf("must be %s, %s, or %s", Up, Down, Ignore)
				}
			},
		)
	}
	addDirectionFlag(Func, "direction of references to functions and methods")
	addDirectionFlag(Type, "direction of type references, excluding references to the receiver type")
	addDirectionFlag(RecvType, "direction of references to the receiver type")
	addDirectionFlag(Var, "direction of references to var declarations")
	addDirectionFlag(Const, "direction of references to const declarations")
}

func run(pass *analysis.Pass) (any, error) {
	var printer Printer = SimplePrinter{Pass: pass}
	if colorize {
		printer = ColorPrinter{
			Pass:       pass,
			ColorError: color.Red,
			ColorInfo:  color.Gray,
			ColorOk:    color.Green,
		}
	}
	printer = VerbosePrinter{Verbose: verbose, Printer: printer}
	printer = &SortedPrinter{Pass: pass, Printer: printer}
	defer printer.Flush()

	analysisInspector, ok := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)
	if !ok {
		return nil, errors.New("could not get analyzer")
	}

	check := func(ref *ast.Ident, def token.Pos, kind RefKind) {
		if !def.IsValid() {
			// So far only seen on calls to Error method of error interface.
			printer.Info(ref.Pos(), fmt.Sprintf("got invalid definition position for %q", ref.Name))
			return
		}

		if RefOrder[kind] == Ignore {
			printer.Info(ref.Pos(), fmt.Sprintf("%s reference %s ignored by options", kind, ref.Name))
			return
		}

		if pass.Fset.File(ref.Pos()).Name() != pass.Fset.File(def).Name() {
			printer.Info(
				ref.Pos(),
				fmt.Sprintf(
					`%s reference %s is to definition in separate file (%s)`,
					kind,
					ref.Name,
					pass.Fset.Position(def),
				),
			)
			return
		}

		refLine, defLine := pass.Fset.Position(ref.Pos()).Line, pass.Fset.Position(def).Line
		if refLine == defLine {
			printer.Ok(
				ref.Pos(),
				fmt.Sprintf(
					`%s reference %s is on same line as definition (%s)`,
					kind,
					ref.Name,
					pass.Fset.Position(def),
				),
			)
			return
		}

		refBeforeDef := refLine < defLine
		order := "before"
		if !refBeforeDef {
			order = "after"
		}
		var message string
		if verbose {
			message = fmt.Sprintf(
				`%s reference %s is %s definition (%s)`,
				kind,
				ref.Name,
				order,
				pass.Fset.Position(def),
			)
		} else {
			message = fmt.Sprintf(`%s reference %s is %s definition`, kind, ref.Name, order)
		}

		if orderOk := refBeforeDef == (RefOrder[kind] == Down); orderOk {
			printer.Ok(ref.Pos(), message)
		} else {
			printer.Error(ref.Pos(), message)
		}
	}

	// Map selector identifiers (the "Sel" in x.Sel) to their selections so we can
	// distinguish interface method selections from concrete ones.
	selOfIdent := make(map[*ast.Ident]*types.Selection)

	// State for keeping track of the receiver type.
	// No need for a stack as method declarations can only be at file scope.
	var (
		funcDecl       *ast.FuncDecl
		recvType       *types.TypeName
		beforeFuncType bool
	)

	analysisInspector.Nodes(nil, func(n ast.Node, push bool) (proceed bool) {
		if !push {
			if funcDecl == n {
				funcDecl = nil
				recvType = nil
			}
			return true
		}

		switch node := n.(type) {
		case *ast.File:
			if ast.IsGenerated(node) {
				printer.Info(node.Pos(), "skipping generated file")
				return false
			}

		case *ast.SelectorExpr:
			if sel := pass.TypesInfo.Selections[node]; sel != nil {
				selOfIdent[node.Sel] = sel
			}

		case *ast.FuncDecl:
			if funcDecl == nil {
				funcDecl = node
				beforeFuncType = true
			}

		case *ast.FuncType:
			beforeFuncType = false

		case *ast.Ident:
			// If this ident is a definition or otherwise has no associated use,
			// skip it to avoid noisy "unexpected ident" messages.
			obj := pass.TypesInfo.Uses[node]
			if obj == nil {
				break
			}

			switch def := obj.(type) {
			case *types.Var:
				def = def.Origin()
				switch {
				case def.IsField():
					printer.Info(node.Pos(), fmt.Sprintf("skipping var ident %s for field %s", node.Name,
						pass.Fset.Position(def.Pos())))
				case def.Parent() != def.Pkg().Scope():
					printer.Info(node.Pos(), fmt.Sprintf("skipping var ident %s with inner parent scope %s", node.Name,
						pass.Fset.Position(def.Parent().Pos())))
				default:
					check(node, def.Pos(), Var)
				}
			case *types.Const:
				if def.Parent() != def.Pkg().Scope() {
					printer.Info(node.Pos(), fmt.Sprintf("skipping var ident %s with inner parent scope %s", node.Name, pass.Fset.Position(def.Parent().Pos())))
				} else {
					check(node, def.Pos(), Const)
				}

			case *types.Func:
				def = def.Origin()
				// Allow direct self-recursion (call to the function we're inside).
				if funcDecl != nil {
					curr, ok := pass.TypesInfo.Defs[funcDecl.Name].(*types.Func)
					if ok && curr != nil && curr.Origin() == def {
						// For a recursive call, pass.TypesInfo.Uses[node] returns the current function’s object;
						// comparing its Origin() to the current func’s Origin() lets us detect direct recursion even
						// with generics instantiation.
						break
					}
				}

				// Handle interface method selections as type references.
				// If this is a method selection, and the receiver is an interface type,
				// treat it as a reference to the interface type (not a function).
				if sel := selOfIdent[node]; sel != nil {
					recv := sel.Recv()
					// Unwrap pointers.
					for {
						if p, ok := recv.(*types.Pointer); ok {
							recv = p.Elem()
							continue
						}
						break
					}
					handled := false
					switch rt := recv.(type) {
					case *types.Named:
						if _, ok := rt.Underlying().(*types.Interface); ok {
							// Count this as a type reference to the named interface.
							check(node, rt.Obj().Pos(), Type)
							handled = true
						}
					case *types.Interface:
						// Unnamed interface type; nothing to order against at package scope.
						printer.Info(node.Pos(), fmt.Sprintf("skipping interface method reference %s on unnamed interface type", node.Name))
						handled = true
					case *types.TypeParam:
						// Method selected via a type parameter's interface constraint.
						printer.Info(node.Pos(), fmt.Sprintf("skipping method reference %s on type parameter %s", node.Name, rt.Obj().Name()))
						handled = true
					}
					if handled {
						break
					}
				}

				if def.Parent() != nil && def.Parent() != def.Pkg().Scope() {
					printer.Info(node.Pos(), fmt.Sprintf("skipping func ident %s with inner parent scope %s", node.Name, pass.Fset.Position(def.Parent().Pos())))
				} else {
					check(node, def.Pos(), Func)
				}

			case *types.TypeName:
				if def.Pkg() == nil {
					printer.Info(node.Pos(), "skipping predeclared type "+node.Name)
					break
				}
				if def.Parent() != def.Pkg().Scope() {
					printer.Info(node.Pos(), fmt.Sprintf("skipping type ident %s with inner parent scope %s", node.Name, pass.Fset.Position(def.Parent().Pos())))
					break
				}

				if funcDecl != nil && beforeFuncType {
					check(node, def.Pos(), RecvType)
					recvType = def
					break
				}
				if funcDecl != nil && recvType == def {
					// Reference to the receiver type within a method type or body.
					break
				}
				check(node, def.Pos(), Type)

			case *types.Builtin:
				// Built-in functions like len, make, panic, etc.
				printer.Info(node.Pos(), "skipping builtin "+node.Name)
			case *types.PkgName:
				// Package qualifier in selectors like fmt.Println.
				printer.Info(node.Pos(), "skipping package name "+node.Name)
			case *types.Label:
				printer.Info(node.Pos(), "skipping label "+node.Name)
			default:
				printer.Info(node.Pos(), fmt.Sprintf("unexpected ident def type %T for %q", pass.TypesInfo.Uses[node], node.Name))
			}
		}

		return true
	})

	//nolint:nilnil // Done.
	return nil, nil
}

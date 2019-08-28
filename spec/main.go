package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"go/ast"
	"go/importer"
	"go/parser"
	"go/token"
	"go/types"
	"io/ioutil"
	"os"
	"reflect"
	"strings"
	"unicode"
)

// filter is used to filter out go test files.
func filter(fi os.FileInfo) bool {
	return !strings.Contains(fi.Name(), "test")
}

func isExported(name string) bool {
	return name[0] == byte(unicode.ToUpper(rune(name[0])))
}

type Spec struct {
	Definitions map[string]interface{}
}
type ClassSpec struct {
	Name        string
	Description string
	Kind        string
	Entries     map[string]interface{}
}
type FunctionSpec struct {
	Name        string
	Description string
	Kind        string
	Params      []*ParamSpec
}

type ParamSpec struct {
	Name string
	Type string
}
type FieldSpec struct {
	Kind        string
	Name        string
	Description string
	Type        string
}

func receiver(f *ast.FuncDecl) string {
	if f.Recv != nil {
		switch f.Recv.List[0].Type.(type) {
		case *ast.StarExpr:
			t := f.Recv.List[0].Type.(*ast.StarExpr)
			return "" + t.X.(*ast.Ident).Name
		case *ast.Ident:
			t := f.Recv.List[0].Type.(*ast.Ident)
			return "" + t.Name
		default:
			return "unknown"
		}
	}
	return "enigma"
}
func main() {

	spec := Spec{
		Definitions: make(map[string]interface{}),
	}

	f, err := os.Create("./output.txt")
	defer f.Close()

	w := bufio.NewWriter(f)

	path := "."
	fset := token.NewFileSet()
	pkgs, err := parser.ParseDir(fset, ".", filter, parser.ParseComments)

	var pkg *ast.Package
	for _, v := range pkgs {
		pkg = v
	}

	conf := &types.Config{Importer: importer.Default(), Error: func(err error) {

	}}

	files := make([]*ast.File, len(pkg.Files))
	i := 0
	for _, file := range pkg.Files {
		files[i] = file
		for _, d := range file.Decls {

			switch d.(type) {

			case *ast.FuncDecl:
				f := d.(*ast.FuncDecl)
				name := f.Name.Name
				unicode.ToUpper(rune(name[0]))
				if isExported(name) {
					owner := receiver(f)
					fmt.Fprintln(w, owner, f.Name)
					if owner != "" {
						definition := spec.Definitions[owner]
						if definition != nil {

						} else {
							fmt.Println("Non matching doc")
						}
					}
					fmt.Fprintln(w, f.Doc.Text())
					if f.Name.Name == "GetProperties" {
					}

				}
			case *ast.GenDecl:
				//fmt.Println(d.(*ast.GenDecl).Specs)
				d.End()
			default:
				d.End()
			}
		}
		i++
	}
	conf.Error = func(err error) {

	}
	p, err := conf.Check(path, fset, files, nil)
	if err != nil {
		fmt.Println(err)
	}

	scope := p.Scope()

	fmt.Fprintln(w, "--------------------------Second level-------------------------")
	for _, name := range scope.Names() {
		o := scope.Lookup(name)
		if o.Exported() {

			//if (reflect.TypeOf(o.Type()).Name())

			fmt.Println(reflect.TypeOf(o.Type()))
			if o.Name() == "RemoteObject" {
				fmt.Println("dialer")
			}

			if o.Name() == "NxPage" {
				otype := o.Type()
				fmt.Fprintln(w, o.Pkg().Name()+"."+o.Name(), otype)
			}

			switch o.Type().(type) {

			case *types.Named:

				clazz := getOrCreateClassSpec(spec, o)

				namedType := o.Type().(*types.Named)
				methodCount := namedType.NumMethods()
				namedType.Underlying()
				underlying := namedType.Underlying()
				for i = 0; i < methodCount; i++ {
					m := namedType.Method(i)
					if isExported(m.Name()) {
						signature, _ := m.Type().(*types.Signature)
						params := toSpecParams(signature.Params())
						methodSpec := &FunctionSpec{Kind: "Function", Name: m.Name(), Params: params}
						clazz.Entries[m.Name()] = methodSpec

						fmt.Fprintln(w, "   "+m.Name()+"()")
					}
				}
				switch underlying.(type) {
				case *types.Struct:
					strukt := underlying.(*types.Struct)
					fieldCount := strukt.NumFields()
					for i = 0; i < fieldCount; i++ {
						m := strukt.Field(i)
						if isExported(m.Name()) {
							mt := toSpecType(m.Type())
							if m.Embedded() {
								fieldSpec := &FieldSpec{Kind: "Field", Name: m.Name(), Type: mt}
								clazz.Entries[m.Name()] = fieldSpec
							} else {
								fieldSpec := &FieldSpec{Kind: "Embedded", Name: m.Name(), Type: mt}
								clazz.Entries[m.Name()] = fieldSpec
							}

							fmt.Fprintln(w, "   "+m.Name()+"()")
						}
					}

				}
				//			m := o.(*types.Named).Method(0)

			case *types.Signature:
				signature := o.Type().(*types.Signature)

				functionn := FunctionSpec{Kind: "function", Name: o.Name()}
				spec.Definitions[o.Name()] = functionn
				functionn.Params = toSpecParams(signature.Params())
				fmt.Fprintln(w, "   signature: ", signature.String())

			default:
				werwrwerwer := reflect.TypeOf(o.Type())
				fmt.Fprintln(w, "UNKOWN TYPE:", werwrwerwer)

			}
		}
	}

	i = 0
	for _, file := range pkg.Files {
		files[i] = file
		for _, d := range file.Decls {

			switch d.(type) {
			case *ast.FuncDecl:
				f := d.(*ast.FuncDecl)
				name := f.Name.Name
				if isExported(name) {
					fnReceiver := receiver(f)
					fnName := f.Name.Name
					if fnReceiver != "" {
						var methodSpec interface{}
						classSpec := spec.Definitions[fnReceiver]
						if classSpec != nil {
							methodSpec = classSpec.(*ClassSpec).Entries[fnName]
						}
						if methodSpec != nil {
							switch methodSpec.(type) {
							//case ClassSpec:
							//	classSpec := methodSpec.(ClassSpec)
							//	classSpec.Description = f.Doc.Text()
							case *FunctionSpec:
								functionSpec := methodSpec.(*FunctionSpec)
								text := f.Doc.Text()
								functionSpec.Description = text

								if name == "AbortModal" {
									fmt.Println("here")
								}
							}
						} else {
							fmt.Println("Non matching doc")
						}
					}
					fmt.Fprintln(w, f.Doc.Text())
					if f.Name.Name == "GetProperties" {
					}

				}
			case *ast.GenDecl:
				specs := d.(*ast.GenDecl).Specs
				for _, astSpec := range specs {
					fmt.Println("---------", reflect.TypeOf(astSpec).String())
					switch astSpec.(type) {
					case *ast.TypeSpec:
						typeSpec := astSpec.(*ast.TypeSpec)
						fmt.Println("TYPE SPEC:", typeSpec.Name)
						definition := spec.Definitions[typeSpec.Name.Name]
						if definition != nil {
							if definition != nil {
								switch definition.(type) {
								case *ClassSpec:
									classSpec := definition.(*ClassSpec)
									text := typeSpec.Doc.Text()
									classSpec.Description = text
									//case FunctionSpec:
									//	functionSpec := definition.(FunctionSpec)
									//	functionSpec.Description = f.Doc.Text()
								}
							} else {
								fmt.Println("Non matching doc")
							}
						}
					default:
						fmt.Println("gendecl")
					}
				}
				fmt.Println(d)
			default:
			}
		}
		i++
	}

	fmt.Println(spec)

	specFile, _ := json.MarshalIndent(spec, "", " ")

	_ = ioutil.WriteFile("output.json", specFile, 0644)
}

func getOrCreateClassSpec(spec Spec, o types.Object) *ClassSpec {

	clazz := spec.Definitions[o.Name()]
	if clazz == nil {
		clazz = &ClassSpec{Kind: "class", Name: o.Name(), Entries: make(map[string]interface{})}
		spec.Definitions[o.Name()] = clazz
	}
	return clazz.(*ClassSpec)
}

func toSpecParams(tuple *types.Tuple) []*ParamSpec {
	result := make([]*ParamSpec, tuple.Len())
	for i := 0; i < tuple.Len(); i++ {
		param := tuple.At(i)
		result[i] = &ParamSpec{Name: param.Name(), Type: toSpecType(param.Type())}
	}
	return result
}

func toSpecType(typ types.Type) string {
	switch typ.(type) {
	case *types.Named:
		namedType := typ.(*types.Named)
		if namedType.Obj().Pkg() == nil {
			return namedType.Obj().Name()
		} else if namedType.Obj().Pkg().Path() == "." {
			return "#/definitions/" + namedType.Obj().Name()
		} else {
			return "GOLANG/" + namedType.Obj().Pkg().Path() + "/" + namedType.Obj().Name()
		}

	case *types.Basic:
		basicType := typ.(*types.Basic)
		return "BASICTYPE:" + basicType.Name()
	case *types.Slice:
		sliceType := typ.(*types.Slice)
		return "Array of " + toSpecType(sliceType.Elem())
	case *types.Pointer:
		pointerType := typ.(*types.Pointer)
		return "Pointer to " + toSpecType(pointerType.Elem())
	case *types.Interface:
		return "interface{}"
	case *types.Chan:
		chanType := typ.(*types.Chan)
		return "Chan " + chanType.String()
	case *types.Signature:
		return "Function"
	case *types.Map:
		mapType := typ.(*types.Map)
		return "Map of " + toSpecType(mapType.Elem())
	}
	os.Exit(1)
	return typ.String()
}

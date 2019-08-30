package main

import (
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

type Info struct {
	Name        string `json:"name,omitempty"`
	Version     string `json:"version,omitempty"`
	Description string `json:"description,omitempty"`
	License     string `json:"license,omitempty"`
	Stability   string `json:"stability,omitempty"`
	Visibility  string `json:"x-qlik-visibility,omitempty"`
}
type Spec struct {
	OAppy       string                 `json:"oappy,omitempty"`
	Info        Info                   `json:"info,omitempty"`
	Definitions map[string]Describable `json:"definitions,omitempty"`
}
type SpecNode struct {
	Kind        string `json:"kind,omitempty"`
	name        string
	Description string                 `json:"description,omitempty"`
	Type        string                 `json:"type,omitempty"`
	Embedded    bool                   `json:"embedded,omitempty"`
	Entries     map[string]Describable `json:"entries,omitempty"`
	Items       *SpecNode              `json:"items,omitempty"`
	Generics    []*SpecNode            `json:"generics,omitempty"`
	RefKind     string                 `json:"refkind,omitempty"`
	Params      []*SpecNode            `json:"params,omitempty"`
	Returns     []*SpecNode            `json:"returns,omitempty"`
}

type Describable interface {
	setDescription(description string)
	getEntry(string) Describable
}

func (f *SpecNode) setDescription(description string) {
	f.Description = description
}

func (f *SpecNode) getEntry(name string) Describable {
	return f.Entries[name]
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
		OAppy: "1.0.0",
		Info: Info{
			Name:       "enigma-go",
			Version:    "0.0.1",
			Stability:  "experimental",
			Visibility: "public",
			License:    "MIT",

			Description: "enigma-go is a library that helps you communicate with a Qlik Associative Engine.",
		},
		Definitions: make(map[string]Describable),
	}

	astPackage, scope := compilePackage()
	for _, name := range scope.Names() {
		o := scope.Lookup(name)
		if o.Exported() {
			switch o.Type().(type) {
			case *types.Named:
				namedType := o.Type().(*types.Named)
				underlying := namedType.Underlying()
				specNode := translateDeclaration(underlying)
				fillInMethods(namedType, specNode)
				spec.Definitions[name] = specNode
			case *types.Signature:
				signature := o.Type().(*types.Signature)
				specNode := translateDeclaration(signature)
				specNode.Kind = "function"
				spec.Definitions[o.Name()] = specNode
			default:
				panic("Unknown otype")
			}
		}
	}

	fmt.Println("--------------- Appending comments ---------------")
	for _, file := range astPackage.Files {
		for _, astDecl := range file.Decls {
			switch astDecl.(type) {
			case *ast.FuncDecl:
				f := astDecl.(*ast.FuncDecl)
				name := f.Name.Name
				if isExported(name) {
					fnReceiver := receiver(f)
					fnName := f.Name.Name
					if fnReceiver != "" {
						var methodSpec interface{}
						classSpec := spec.Definitions[fnReceiver]
						if classSpec != nil {
							methodSpec = classSpec.(*SpecNode).Entries[fnName]
						}
						if methodSpec != nil {
							switch methodSpec.(type) {
							case *SpecNode:
								SpecNode := methodSpec.(*SpecNode)
								text := f.Doc.Text()
								SpecNode.Description = text
							}
						} else {
							fmt.Println("Non matching doc for method: ", fnName)
						}
					}
				}
			case *ast.GenDecl:
				specs := astDecl.(*ast.GenDecl).Specs
				for _, astSpec := range specs {
					switch astSpec.(type) {
					case *ast.TypeSpec:
						typeSpec := astSpec.(*ast.TypeSpec)
						if isExported(typeSpec.Name.Name) {
							definition := spec.Definitions[typeSpec.Name.Name]
							if definition != nil {
								definition.setDescription(typeSpec.Doc.Text())
								switch typeSpec.Type.(type) {
								case *ast.StructType:
									astStruct := typeSpec.Type.(*ast.StructType)
									for _, y := range astStruct.Fields.List {
										if len(y.Names) > 0 {
											fieldName := y.Names[0].Name

											fieldSpec := definition.getEntry(fieldName)
											if fieldSpec != nil {
												fieldSpec.setDescription(y.Doc.Text())
											} else if isExported(y.Names[0].Name) {
												fmt.Println("Non matching field: ", y.Names[0].Name)
											}
										} else {
											if len(y.Doc.Text()) > 0 {
												fmt.Println("Embedded with doc: ", y)
											}
										}

									}
								case *ast.InterfaceType:
									astInterface := typeSpec.Type.(*ast.InterfaceType)
									for _, y := range astInterface.Methods.List {
										if len(y.Names) > 0 {
											fieldName := y.Names[0].Name

											fieldSpec := definition.getEntry(fieldName)
											if fieldSpec != nil {
												fieldSpec.setDescription(y.Doc.Text())
											} else if isExported(y.Names[0].Name) {
												fmt.Println("Non matching method: ", y.Names[0].Name)
											}
										} else {
											if len(y.Doc.Text()) > 0 {
												fmt.Println("Embedded with doc: ", y)
											}
										}

									}
								case *ast.ArrayType:
									//No extra doc needed
								case *ast.Ident:
									//No extra doc needed
								case *ast.FuncType:
									//No extra doc needed
								default:
									panic("Unknown doc decl: ")
								}
							} else {
								fmt.Println("Non matching doc for ", typeSpec.Name)
							}
						}
					case *ast.ImportSpec:
					default:
						panic("MISSED ON LEVEL 2:" + reflect.TypeOf(astSpec).String())

					}
				}
			default:
				panic("MISSED ON LEVEL 1:" + reflect.TypeOf(astDecl).String())
			}
		}
	}

	specFile, _ := json.MarshalIndent(spec, "", "   ")
	_ = ioutil.WriteFile("api-spec.json", specFile, 0644)
}

func fillInStructMembers(strukt *types.Struct, clazz *SpecNode) {
	fieldCount := strukt.NumFields()
	for i := 0; i < fieldCount; i++ {
		m := strukt.Field(i)
		if isExported(m.Name()) {
			mt := translateTypeReference(m.Type())
			if m.Embedded() {
				mt.Embedded = true
			}
			clazz.Entries[m.Name()] = mt
		}
	}
}

func fillInMethods(namedType MethodContainer, clazz *SpecNode) {
	methodCount := namedType.NumMethods()
	for i := 0; i < methodCount; i++ {
		m := namedType.Method(i)
		if isExported(m.Name()) {
			methodSpec := translateDeclaration(m.Type())
			methodSpec.Kind = "method"
			if clazz.Entries == nil {
				clazz.Entries = make(map[string]Describable)
			}
			clazz.Entries[m.Name()] = methodSpec
		}
	}
}

//
//func translateFuncToSpec(m *types.Func) *SpecNode {
//	signature, _ := m.Type().(*types.Signature)
//	params := translateTupleToSpec(signature.Params())
//	results := translateTupleToSpec(signature.Results())
//	methodSpec := &SpecNode{Kind: "function", name: m.Name(), Params: params, Returns: results}
//	return methodSpec
//}

type MethodContainer interface {
	NumMethods() int
	Method(i int) *types.Func
}

func fillInEmbeddedMethods(namedType types.Named, clazz *SpecNode) {
	methodCount := namedType.NumMethods()
	for i := 0; i < methodCount; i++ {
		m := namedType.Method(i)
		if isExported(m.Name()) {
			signature, _ := m.Type().(*types.Signature)
			params := translateTupleToSpec(signature.Params())
			results := translateTupleToSpec(signature.Results())
			methodSpec := &SpecNode{Kind: "function", name: m.Name(), Params: params, Returns: results}
			if clazz.Entries == nil {
				clazz.Entries = make(map[string]Describable)
			}
			clazz.Entries[m.Name()] = methodSpec
		}
	}
	//var x types.Interface
	//var y types.Struct

}

func compilePackage() (*ast.Package, *types.Scope) {
	path := "."
	fset := token.NewFileSet()
	pkgs, err := parser.ParseDir(fset, ".", filter, parser.ParseComments)
	var pkg *ast.Package
	for _, v := range pkgs {
		pkg = v
	}
	files := make([]*ast.File, 0)
	for x, file := range pkg.Files {
		fmt.Println(file.Name.Name, x)
		files = append(files, file)
	}
	conf := &types.Config{Importer: importer.Default(), Error: func(err error) {

	}}
	conf.Error = func(err error) {

	}
	p, err := conf.Check(path, fset, files, nil)
	if err != nil {
		fmt.Println(err)
	}
	return pkg, p.Scope()
}

func translateTupleToSpec(tuple *types.Tuple) []*SpecNode {
	result := make([]*SpecNode, tuple.Len())
	for i := 0; i < tuple.Len(); i++ {
		param := tuple.At(i)
		result[i] = translateTypeReference(param.Type())
	}
	return result
}

func translateDeclaration(underlying types.Type) *SpecNode {
	switch underlying.(type) {
	case *types.Struct:
		strukt := underlying.(*types.Struct)
		clazz := &SpecNode{Kind: "struct", Entries: make(map[string]Describable)}
		fillInStructMembers(strukt, clazz)
		return clazz
	case *types.Signature:
		return translateSignature(underlying)
	case *types.Slice:
		sliceType := underlying.(*types.Slice)
		arraySpec := &SpecNode{
			Kind:  "slice",
			Items: translateTypeReference(sliceType.Elem()),
		}
		return arraySpec
	case *types.Interface:
		interfac := underlying.(*types.Interface)
		interfaceSpec := &SpecNode{Kind: "interface", name: interfac.String(), Entries: make(map[string]Describable)}
		fillInMethods(interfac, interfaceSpec)
		return interfaceSpec
	case *types.Basic:
		basic := underlying.(*types.Basic)
		basicSpec := &SpecNode{Kind: "basic", Type: basic.String()}
		return basicSpec
	default:
		panic("Unknown type")
	}
}

func translateSignature(underlying types.Type) *SpecNode {
	signature := underlying.(*types.Signature)
	functionn := &SpecNode{Kind: "signature"}
	functionn.Params = translateTupleToSpec(signature.Params())
	functionn.Returns = translateTupleToSpec(signature.Results())
	return functionn
}

func defaultIsPointer(typ types.Type) bool {

	switch typ.(type) {
	case *types.Named:
		return defaultIsPointer(typ.Underlying())
	case *types.Struct:
		return true
	default:
		return false
	}
}
func translateTypeReference(typ types.Type) *SpecNode {
	switch typ.(type) {
	case *types.Named:
		namedType := typ.(*types.Named)
		actualName := getNamedName(namedType)
		if defaultIsPointer(namedType.Underlying()) {
			return &SpecNode{
				Type:    actualName,
				RefKind: "value",
			}
		} else {
			return &SpecNode{
				Type: actualName,
			}
		}

	case *types.Basic:
		basicType := typ.(*types.Basic)
		return &SpecNode{
			Type: basicType.Name(),
		}
	case *types.Slice:
		sliceType := typ.(*types.Slice)
		result := &SpecNode{
			Kind:  "slice",
			Items: translateTypeReference(sliceType.Elem()),
		}
		return result
	case *types.Pointer:
		pointerType := typ.(*types.Pointer)
		result := translateTypeReference(pointerType.Elem())
		if defaultIsPointer(pointerType.Elem().Underlying()) {
			result.RefKind = ""
		} else {
			result.RefKind = "pointer"
		}
		return result
	case *types.Interface:
		interfaceType := typ.(*types.Interface)
		if interfaceType.NumMethods() == 0 {
			return &SpecNode{
				Type: "interface {}",
			}
		} else {
			panic("Inline interface not supported")
		}
	case *types.Chan:
		chanType := typ.(*types.Chan)
		result := &SpecNode{
			Kind:  "chan",
			Items: translateTypeReference(chanType.Elem()),
		}
		return result
	case *types.Signature:
		signatureType := typ.(*types.Signature)
		result := translateSignature(signatureType)
		result.Kind = "function"
		return result
	case *types.Struct:
		structType := typ.(*types.Struct)
		if structType.NumFields() == 0 {
			result := &SpecNode{
				Type: "struct {}",
			}
			return result
		} else {
			panic("Inline struct not supported")
		}
	default:
		panic("Unknown type")
	}
}

func getNamedName(namedType *types.Named) string {
	if namedType.Obj().Pkg() == nil {
		return namedType.Obj().Name()
	} else if namedType.Obj().Pkg().Path() == "." {
		return "#/definitions/" + namedType.Obj().Name()
	} else {
		return "" + namedType.Obj().Pkg().Path() + "." + namedType.Obj().Name()
	}
	return ""
}

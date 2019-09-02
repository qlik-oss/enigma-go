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
	OAppy       string               `json:"oappy,omitempty"`
	Info        Info                 `json:"info,omitempty"`
	Definitions map[string]*SpecNode `json:"definitions,omitempty"`
}
type SpecNode struct {
	//Kind        string `json:"kind,omitempty"`
	name        string
	Description string               `json:"description,omitempty"`
	Type        string               `json:"type,omitempty"`
	Embedded    bool                 `json:"embedded,omitempty"`
	Entries     map[string]*SpecNode `json:"entries,omitempty"`
	Items       *SpecNode            `json:"items,omitempty"`
	Generics    []*SpecNode          `json:"generics,omitempty"`
	RefKind     string               `json:"refkind,omitempty"`
	Params      []*SpecNode          `json:"params,omitempty"`
	Returns     []*SpecNode          `json:"returns,omitempty"`
}
type MethodContainer interface {
	NumMethods() int
	Method(i int) *types.Func
}

var descriptions map[string]string

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
	astPackage, scope := compilePackage()
	descriptions = grabComments(astPackage)
	spec := buildSpec(scope)
	specFile, _ := json.MarshalIndent(spec, "", "   ")
	_ = ioutil.WriteFile("api-spec.json", specFile, 0644)
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
	for _, file := range pkg.Files {
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

func grabComments(astPackage *ast.Package) map[string]string {
	docz := make(map[string]string)
	for _, file := range astPackage.Files {
		for _, astDecl := range file.Decls {
			switch astDecl.(type) {
			case *ast.FuncDecl:
				f := astDecl.(*ast.FuncDecl)
				name := f.Name.Name
				if isExported(name) {
					fnReceiver := receiver(f)
					fnName := f.Name.Name
					docz[fnReceiver+"."+fnName] = f.Doc.Text()
				}
			case *ast.GenDecl:
				specs := astDecl.(*ast.GenDecl).Specs
				for _, astSpec := range specs {
					switch astSpec.(type) {
					case *ast.TypeSpec:
						typeSpec := astSpec.(*ast.TypeSpec)
						docz[typeSpec.Name.Name] = typeSpec.Doc.Text()
						switch typeSpec.Type.(type) {
						case *ast.StructType:
							astStruct := typeSpec.Type.(*ast.StructType)
							for _, y := range astStruct.Fields.List {
								if len(y.Names) > 0 {
									fieldName := y.Names[0].Name
									docz[typeSpec.Name.Name+"."+fieldName] = y.Doc.Text()
								}

							}
						case *ast.InterfaceType:
							astInterface := typeSpec.Type.(*ast.InterfaceType)
							for _, y := range astInterface.Methods.List {
								if len(y.Names) > 0 {
									fieldName := y.Names[0].Name
									docz[typeSpec.Name.Name+"."+fieldName] = y.Doc.Text()
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
	return docz
}

func buildSpec(scope *types.Scope) Spec {
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
		Definitions: make(map[string]*SpecNode),
	}
	for _, name := range scope.Names() {
		o := scope.Lookup(name)
		if o.Exported() {
			switch o.Type().(type) {
			case *types.Named:
				namedType := o.Type().(*types.Named)
				underlying := namedType.Underlying()
				specNode := translateTypeUnified(name, underlying)
				fillInMethods(name, namedType, specNode)
				specNode.Description = descriptions[name]
				spec.Definitions[name] = specNode
			case *types.Signature:
				signature := o.Type().(*types.Signature)
				specNode := translateTypeUnified("", signature)
				specNode.Type = "function"
				specNode.Description = descriptions[name]
				spec.Definitions[o.Name()] = specNode
			default:
				panic("Unknown otype")
			}
		}
	}
	return spec
}

func translateTupleToSpec(tuple *types.Tuple) []*SpecNode {
	result := make([]*SpecNode, tuple.Len())
	for i := 0; i < tuple.Len(); i++ {
		param := tuple.At(i)
		result[i] = translateTypeUnified("", param.Type())
	}
	return result
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

func translateTypeUnified(docNamespace string, typ types.Type) *SpecNode {
	switch typ.(type) {
	case *types.Named:
		namedType := typ.(*types.Named)
		actualName := getNamedName(namedType)
		result := &SpecNode{
			Type: actualName,
		}
		if defaultIsPointer(namedType.Underlying()) {
			result.RefKind = "value"
		}
		return result
	case *types.Basic:
		basicType := typ.(*types.Basic)
		return &SpecNode{
			Type: basicType.Name(),
		}
	case *types.Slice:
		sliceType := typ.(*types.Slice)
		result := &SpecNode{
			Type:  "slice",
			Items: translateTypeUnified(docNamespace, sliceType.Elem()),
		}
		return result
	case *types.Pointer:
		pointerType := typ.(*types.Pointer)
		result := translateTypeUnified(docNamespace, pointerType.Elem())
		if defaultIsPointer(pointerType.Elem().Underlying()) {
			result.RefKind = ""
		} else {
			result.RefKind = "pointer"
		}
		return result
	case *types.Chan:
		chanType := typ.(*types.Chan)
		result := &SpecNode{
			Type:  "chan",
			Items: translateTypeUnified(docNamespace, chanType.Elem()),
		}
		return result
	case *types.Signature:
		signatureType := typ.(*types.Signature)
		result := &SpecNode{
			Type:    "function-signature",
			Params:  translateTupleToSpec(signatureType.Params()),
			Returns: translateTupleToSpec(signatureType.Results()),
		}
		return result
	case *types.Struct:
		structType := typ.(*types.Struct)
		result := &SpecNode{
			Type:    "struct",
			Entries: make(map[string]*SpecNode),
		}
		fillInStructFields(docNamespace, structType, result)
		fillInEmbeddedMethods(structType, result)
		return result
	case *types.Interface:
		interfaceType := typ.(*types.Interface)
		interfaceSpec := &SpecNode{
			Type:    "interface",
			Entries: make(map[string]*SpecNode),
		}
		fillInMethods(docNamespace, interfaceType, interfaceSpec)
		return interfaceSpec
	default:
		panic("Unknown type")
	}
}

func fillInStructFields(docNamespace string, struktType *types.Struct, clazz *SpecNode) {
	fieldCount := struktType.NumFields()
	for i := 0; i < fieldCount; i++ {
		m := struktType.Field(i)
		if isExported(m.Name()) {
			mt := translateTypeUnified(docNamespace, m.Type())
			if m.Embedded() {
				mt.Embedded = true
			}
			if mt.Type == "function-signature" {
				mt.Type = "function"
			}

			mt.Description = descriptions[docNamespace+"."+m.Name()]
			clazz.Entries[m.Name()] = mt
		}
	}
}

func fillInMethods(docNamespace string, namedType MethodContainer, clazz *SpecNode) {
	methodCount := namedType.NumMethods()
	for i := 0; i < methodCount; i++ {
		m := namedType.Method(i)
		if isExported(m.Name()) {
			methodSpec := translateTypeUnified(docNamespace, m.Type())
			methodSpec.Type = "method"
			methodSpec.Description = descriptions[docNamespace+"."+m.Name()]

			if clazz.Entries == nil {
				clazz.Entries = make(map[string]*SpecNode)
			}
			clazz.Entries[m.Name()] = methodSpec
		}
	}
}
func fillInEmbeddedMethods(typ types.Type, clazz *SpecNode) {
	switch typ.(type) {
	case *types.Struct:
		struktType := typ.(*types.Struct)
		fieldCount := struktType.NumFields()
		for i := 0; i < fieldCount; i++ {
			field := struktType.Field(i)
			if field.Embedded() && !field.Exported() {
				embeddedFieldType := field.Type()
				switch embeddedFieldType.(type) {
				case *types.Pointer:
					embeddedFieldType = embeddedFieldType.(*types.Pointer).Elem()
				}
				switch embeddedFieldType.(type) {
				case *types.Named:
					embeddedNamedType := embeddedFieldType.(*types.Named)
					fillInMethods(field.Name(), embeddedNamedType, clazz)
					embeddedFieldType = embeddedFieldType.(*types.Named).Underlying()
				}
				switch embeddedFieldType.(type) {
				case *types.Struct:
					fillInEmbeddedMethods(embeddedFieldType.(*types.Struct), clazz)
				}
			}
		}
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

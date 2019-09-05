package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"go/ast"
	"go/importer"
	"go/parser"
	"go/token"
	"go/types"
	"io/ioutil"
	"os"
	"reflect"
	"regexp"
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

type DescriptionAndTags struct {
	Descr             string `json:"description,omitempty"`
	Stability         string `json:"x-qlik-stability,omitempty"`
	Deprecated        bool   `json:"deprecated,omitempty"`
	DeprecatedComment string `json:"x-qlik-deprecation-description,omitempty"`
}
type Info struct {
	Name                string `json:"name,omitempty"`
	GoPackageName       string `json:"go-package-name,omitempty"`
	GoPackageImportPath string `json:"go-package-import-path,omitempty"`
	Version             string `json:"version,omitempty"`
	Description         string `json:"description,omitempty"`
	License             string `json:"license,omitempty"`
	Visibility          string `json:"x-qlik-visibility,omitempty"`
	Stability           string `json:"x-qlik-stability,omitempty"`
	Deprecated          bool   `json:"x-qlik-deprecated,omitempty"`
}
type Spec struct {
	OAppy       string               `json:"oappy,omitempty"`
	Info        *Info                `json:"info,omitempty"`
	Definitions map[string]*SpecNode `json:"definitions,omitempty"`
}
type SpecNode struct {
	name string
	*DescriptionAndTags
	Type     string               `json:"type,omitempty"`
	Embedded bool                 `json:"embedded,omitempty"`
	Entries  map[string]*SpecNode `json:"entries,omitempty"`
	Items    *SpecNode            `json:"items,omitempty"`
	Generics []*SpecNode          `json:"generics,omitempty"`
	RefType  string               `json:"ref-type,omitempty"`
	Params   []*SpecNode          `json:"params,omitempty"`
	Returns  []*SpecNode          `json:"returns,omitempty"`
}
type MethodContainer interface {
	NumMethods() int
	Method(i int) *types.Func
}

var descriptions map[string]*DescriptionAndTags

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

var version = flag.String("version", "devbuild", "Specification version")

func main() {
	info := &Info{
		Name:                "enigma",
		GoPackageImportPath: "github.com/qlik-oss/enigma-go",
		GoPackageName:       "enigma",
		Version:             *version,
		Stability:           "locked",
		Visibility:          "public",
		License:             "MIT",
		Description:         "enigma-go is a library that helps you communicate with a Qlik Associative Engine.",
	}

	workingdir, _ := os.Getwd()
	if strings.HasSuffix(workingdir, "/spec") {
		specFile := generateSpec("..", info)
		_ = ioutil.WriteFile("../api-spec.json", specFile, 0644)
	} else {
		specFile := generateSpec(".", info)
		_ = ioutil.WriteFile("api-spec.json", specFile, 0644)
	}
}

func generateSpec(packagePath string, info *Info) []byte {
	astPackage, scope := compilePackage(packagePath, info.GoPackageName)
	descriptions = grabComments(astPackage)
	spec := buildSpec(scope, info)
	specFile, _ := json.MarshalIndent(spec, "", "   ")
	return specFile
}

func compilePackage(path string, packageName string) (*ast.Package, *types.Scope) {
	fset := token.NewFileSet()
	pkgs, err := parser.ParseDir(fset, path, filter, parser.ParseComments)
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
	p, err := conf.Check(packageName, fset, files, nil)
	if err != nil {
		fmt.Println(err)
	}
	return pkg, p.Scope()
}

var DeprecatedRE1 = regexp.MustCompile("(^|\\n)Deprecated: ([^\\n]*)")
var StabilityRE1 = regexp.MustCompile("(^|\\n)Stability: ([^\\n]*)")
var TrailingNewlinesRE = regexp.MustCompile("\\n*$")

func splitDoc(doc string) *DescriptionAndTags {
	node := &DescriptionAndTags{}
	if strings.Index(doc, "Deprecated") >= 0 {
		fmt.Println(doc)
	}
	if DeprecatedRE1.MatchString(doc) {
		node.Deprecated = true
		node.DeprecatedComment = DeprecatedRE1.ReplaceAllString(DeprecatedRE1.FindString(doc), "$2")
	}

	if StabilityRE1.MatchString(doc) {
		node.Stability = StabilityRE1.ReplaceAllString(StabilityRE1.FindString(doc), "$2")
	}

	// Remove tags from comment
	node.Descr = TrailingNewlinesRE.ReplaceAllString(DeprecatedRE1.ReplaceAllString(StabilityRE1.ReplaceAllString(doc, ""), ""), "")

	return node
}

func grabComments(astPackage *ast.Package) map[string]*DescriptionAndTags {
	docz := make(map[string]*DescriptionAndTags)
	for _, file := range astPackage.Files {
		for _, astDecl := range file.Decls {
			switch astDecl.(type) {
			case *ast.FuncDecl:
				f := astDecl.(*ast.FuncDecl)
				name := f.Name.Name
				if isExported(name) {
					fnReceiver := receiver(f)
					fnName := f.Name.Name
					docz[fnReceiver+"."+fnName] = splitDoc(f.Doc.Text())
				}
			case *ast.GenDecl:
				genDecl := astDecl.(*ast.GenDecl)
				specs := genDecl.Specs
				for _, astSpec := range specs {
					switch astSpec.(type) {
					case *ast.TypeSpec:
						typeSpec := astSpec.(*ast.TypeSpec)
						if typeSpec.Doc.Text() != "" {
							docz[typeSpec.Name.Name] = splitDoc(typeSpec.Doc.Text())
						} else {
							docz[typeSpec.Name.Name] = splitDoc(genDecl.Doc.Text())
						}
						switch typeSpec.Type.(type) {
						case *ast.StructType:
							astStruct := typeSpec.Type.(*ast.StructType)
							for _, y := range astStruct.Fields.List {
								if len(y.Names) > 0 {
									fieldName := y.Names[0].Name
									docz[typeSpec.Name.Name+"."+fieldName] = splitDoc(y.Doc.Text())
								}
							}
						case *ast.InterfaceType:
							astInterface := typeSpec.Type.(*ast.InterfaceType)
							for _, y := range astInterface.Methods.List {
								if len(y.Names) > 0 {
									fieldName := y.Names[0].Name
									docz[typeSpec.Name.Name+"."+fieldName] = splitDoc(y.Doc.Text())
								}
							}
						case *ast.ArrayType:
							//No extra doc needed
						case *ast.Ident:
							//No extra doc needed
						case *ast.FuncType:
							//No extra doc needed
						default:
							panic("Unknown typeSpec: " + reflect.TypeOf(typeSpec.Type).String())
						}
					case *ast.ImportSpec:
					default:
						panic("Unknown astSpec: " + reflect.TypeOf(astSpec).String())
					}
				}
			default:
				panic("Unknown astDecl:" + reflect.TypeOf(astDecl).String())
			}
		}
	}
	return docz
}

func buildSpec(scope *types.Scope, info *Info) Spec {
	spec := Spec{
		OAppy:       "0.0.1",
		Info:        info,
		Definitions: make(map[string]*SpecNode),
	}
	for _, name := range scope.Names() {
		namedLangEntity := scope.Lookup(name)
		if namedLangEntity.Exported() {
			switch namedLangEntity.Type().(type) {
			case *types.Named:
				namedType := namedLangEntity.Type().(*types.Named)
				underlying := namedType.Underlying()
				specNode := translateTypeUnified(name, underlying)
				if defaultIsPointer(underlying) && specNode.RefType == "value" {
					specNode.RefType = "" //Reset the value RefType for types where value is not default behaviour
				}
				fillInMethods(name, namedType, specNode)
				specNode.DescriptionAndTags = descriptions[name]
				spec.Definitions[name] = specNode
			case *types.Signature:
				signature := namedLangEntity.Type().(*types.Signature)
				specNode := translateTypeUnified("", signature)
				specNode.Type = "function"
				specNode.DescriptionAndTags = descriptions[name]
				spec.Definitions[namedLangEntity.Name()] = specNode
			default:
				panic("Unknown namedLangEntity: " + reflect.TypeOf(namedLangEntity.Type()).String())
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
			// For types where default ref type is pointer set the ref type to "value". This allows it to be
			// reset to "" by the *types.Pointer case below. Leaving only the non-pointer uses to "value"
			// It is also reset to "" in all named language entities.
			result.RefType = "value"
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
			result.RefType = ""
		} else {
			result.RefType = "pointer"
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
			RefType: "value",
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
		panic("Unknown type: " + reflect.TypeOf(typ).String())
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

			mt.DescriptionAndTags = descriptions[docNamespace+"."+m.Name()]
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
			methodSpec.DescriptionAndTags = descriptions[docNamespace+"."+m.Name()]

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
	} else {
		return "" + namedType.Obj().Pkg().Path() + "." + namedType.Obj().Name()
	}
}

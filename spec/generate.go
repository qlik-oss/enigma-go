package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"go/ast"
	"go/doc"
	"go/importer"
	"go/parser"
	"go/token"
	"go/types"
	"io/ioutil"
	"os"
	"reflect"
	"regexp"
	"strings"
)

// filter is used to filter out go test files.
func filter(fi os.FileInfo) bool {
	return !strings.Contains(fi.Name(), "test")
}

type descriptionAndTags struct {
	Descr             string `json:"description,omitempty"`
	Constant          bool   `json:"constant,omitempty"`
	Stability         string `json:"x-qlik-stability,omitempty"`
	Deprecated        bool   `json:"deprecated,omitempty"`
	DeprecatedComment string `json:"x-qlik-deprecation-description,omitempty"`
}
type info struct {
	Name                string `json:"name,omitempty"`
	GoPackageName       string `json:"go-package-name,omitempty"`
	GoPackageImportPath string `json:"go-package-import-path,omitempty"`
	Version             string `json:"version,omitempty"`
	Description         string `json:"description,omitempty"`
	License             string `json:"license,omitempty"`
	Deprecated          bool   `json:"x-qlik-deprecated,omitempty"`
}
type spec struct {
	OAppy       string               `json:"oappy,omitempty"`
	Info        *info                `json:"info,omitempty"`
	Visibility  string               `json:"x-qlik-visibility,omitempty"`
	Stability   string               `json:"x-qlik-stability,omitempty"`
	Definitions map[string]*specNode `json:"definitions,omitempty"`
}
type specNode struct {
	*descriptionAndTags
	Type     string               `json:"type,omitempty"`
	Embedded bool                 `json:"embedded,omitempty"`
	Entries  map[string]*specNode `json:"entries,omitempty"`
	Items    *specNode            `json:"items,omitempty"`
	Generics []*specNode          `json:"generics,omitempty"`
	RefType  string               `json:"ref-type,omitempty"`
	Params   []*specNode          `json:"params,omitempty"`
	Returns  []*specNode          `json:"returns,omitempty"`
}
type methodContainer interface {
	NumMethods() int
	Method(i int) *types.Func
}

var descriptions map[string]*descriptionAndTags

func receiver(f *ast.FuncDecl) string {
	var recv string
	if f.Recv != nil {
		ast.Inspect(f.Recv, func(n ast.Node) bool {
			if id, ok := n.(*ast.Ident); ok {
				recv = id.Name
				return false
			}
			return true
		})
	}
	return recv
}

var version = flag.String("version", "devbuild", "Specification version")
var currentPackage string

func main() {
	info := &info{
		Name:                "enigma",
		GoPackageImportPath: "github.com/qlik-oss/enigma-go",
		GoPackageName:       "enigma",
		Version:             *version,
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

func generateSpec(packagePath string, info *info) []byte {
	astPackage, scope := compilePackage(packagePath, info.GoPackageName)
	currentPackage = info.GoPackageName
	buildDescriptions(astPackage)
	spec := buildSpec(scope, info, astPackage)
	specFile, _ := json.MarshalIndent(spec, "", "  ")
	return specFile
}

func compilePackage(path string, packageName string) (*ast.Package, *types.Scope) {
	fset := token.NewFileSet()
	pkgs, err := parser.ParseDir(fset, path, filter, parser.ParseComments)
	var pkg *ast.Package
	// There should only be one package in the path. Panic otherwise.
	if len(pkgs) > 1 {
		panic("Too many packages")
	}
	// Extract the only present package from the map[string]*ast.Package.
	// Might be a better way of doing this?
	for _, v := range pkgs {
		pkg = v
	}
	// Convert map[string]*ast.File to slice.
	files := make([]*ast.File, len(pkg.Files))
	i := 0
	for _, file := range pkg.Files {
		files[i] = file
		i++
	}
	conf := &types.Config{
		Importer: importer.ForCompiler(fset, "source", nil),
	}
	p, err := conf.Check(packageName, fset, files, nil)
	if err != nil {
		fmt.Println(err)
	}
	return pkg, p.Scope()
}

var deprecatedRE1 = regexp.MustCompile("(^|\\n)Deprecated: ([^\\n]*)")
var stabilityRE1 = regexp.MustCompile("(^|\\n)Stability: ([^\\n]*)")
var trailingNewlinesRE = regexp.MustCompile("\\n*$")

func splitDoc(doc string) *descriptionAndTags {
	node := &descriptionAndTags{}
	if deprecatedRE1.MatchString(doc) {
		node.Deprecated = true
		node.DeprecatedComment = deprecatedRE1.ReplaceAllString(deprecatedRE1.FindString(doc), "$2")
	}

	if stabilityRE1.MatchString(doc) {
		node.Stability = stabilityRE1.ReplaceAllString(stabilityRE1.FindString(doc), "$2")
	}

	// Remove tags from comment
	descr := deprecatedRE1.ReplaceAllString(doc, "")
	descr = stabilityRE1.ReplaceAllString(descr, "")
	node.Descr = strings.Trim(descr, " \n")

	return node
}

func buildDescriptions(astPackage *ast.Package) {
	descriptions = map[string]*descriptionAndTags{}
	packageDoc := doc.New(astPackage, "", doc.AllDecls)
	for _, c := range packageDoc.Consts {
		describeValue(c, true)
	}
	for _, v := range packageDoc.Vars {
		describeValue(v, false)
	}
	for _, t := range packageDoc.Types {
		describeType(t)
	}
	for _, f := range packageDoc.Funcs {
		descriptions[f.Name] = splitDoc(f.Doc)
	}
}

func describeValue(v *doc.Value, constant bool) {
	desc := splitDoc(v.Doc)
	desc.Constant = constant
	for _, name := range v.Names {
		descriptions[name] = desc
	}
}

func describeType(t *doc.Type) {
	// For the type 't'.
	descriptions[t.Name] = splitDoc(t.Doc)
	// For functions returning type 't'.
	for _, f := range t.Funcs {
		descriptions[f.Name] = splitDoc(f.Doc)
	}
	// For methods (with 't' as receiver).
	for _, m := range t.Methods {
		descriptions[t.Name+"."+m.Name] = splitDoc(m.Doc)
	}
	// For embedded fields of 't'.
	ast.Inspect(t.Decl, func(n ast.Node) bool {
		switch n.(type) {
		case *ast.GenDecl:
			if n.(*ast.GenDecl).Tok == token.TYPE {
				return true
			}
		case *ast.TypeSpec:
			return true
		case *ast.StructType:
			return true
		case *ast.FieldList:
			return true
		case *ast.Field:
			field := n.(*ast.Field)
			desc := splitDoc(field.Doc.Text())
			for _, id := range field.Names {
				descriptions[t.Name+"."+id.Name] = desc
			}
			return false
		}
		return false
	})
}

func buildSpec(scope *types.Scope, info *info, astPkg *ast.Package) *spec {
	spec := &spec{
		OAppy:       "0.0.1",
		Info:        info,
		Stability:   "locked",
		Visibility:  "public",
		Definitions: make(map[string]*specNode),
	}
	for _, name := range scope.Names() {
		namedLangEntity := scope.Lookup(name)
		if namedLangEntity.Exported() {
			switch namedLangEntity.Type().(type) {
			// Type definition.
			case *types.Named:
				namedType := namedLangEntity.Type().(*types.Named)
				underlying := namedType.Underlying()
				specNode := translateTypeUnified(name, underlying)
				if defaultIsPointer(underlying) && specNode.RefType == "value" {
					specNode.RefType = "" //Reset the value RefType for types where value is not default behaviour
				}
				fillInMethods(name, namedType, specNode)
				specNode.descriptionAndTags = descriptions[name]
				spec.Definitions[name] = specNode
				// Function definition.
			case *types.Signature:
				signature := namedLangEntity.Type().(*types.Signature)
				specNode := translateTypeUnified("", signature)
				specNode.Type = "function"
				specNode.descriptionAndTags = descriptions[name]
				spec.Definitions[namedLangEntity.Name()] = specNode
			case *types.Basic:
				basic := namedLangEntity.Type().(*types.Basic)
				specNode := translateTypeUnified("", basic)
				specNode.descriptionAndTags = descriptions[name]
				spec.Definitions[namedLangEntity.Name()] = specNode
			default:
				panic("Unknown namedLangEntity: " + reflect.TypeOf(namedLangEntity.Type()).String())
			}
		}
	}
	return spec
}

func translateTupleToSpec(tuple *types.Tuple) []*specNode {
	result := make([]*specNode, tuple.Len())
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
func removeUntyped(typ string) string {
	if strings.HasPrefix(typ, "untyped ") {
		return typ[8:]
	}
	return typ
}
func translateTypeUnified(docNamespace string, typ types.Type) *specNode {
	switch typ.(type) {
	case *types.Named:
		namedType := typ.(*types.Named)
		actualName := getNamedName(namedType)
		result := &specNode{
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
		return &specNode{
			Type: removeUntyped(basicType.Name()),
		}
	case *types.Slice:
		sliceType := typ.(*types.Slice)
		result := &specNode{
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
		result := &specNode{
			Type:  "chan",
			Items: translateTypeUnified(docNamespace, chanType.Elem()),
		}
		return result
	case *types.Signature:
		signatureType := typ.(*types.Signature)
		result := &specNode{
			Type:    "function-signature",
			Params:  translateTupleToSpec(signatureType.Params()),
			Returns: translateTupleToSpec(signatureType.Results()),
		}
		return result
	case *types.Struct:
		structType := typ.(*types.Struct)
		result := &specNode{
			Type:    "struct",
			Entries: make(map[string]*specNode),
			RefType: "value",
		}
		fillInStructFields(docNamespace, structType, result)
		fillInEmbeddedMethods(structType, result)
		return result
	case *types.Interface:
		interfaceType := typ.(*types.Interface)
		result := &specNode{
			Type:    "interface",
			Entries: make(map[string]*specNode),
		}
		fillInMethods(docNamespace, interfaceType, result)
		return result
	default:
		panic("Unknown type: " + reflect.TypeOf(typ).String())
	}
}

func fillInStructFields(docNamespace string, struktType *types.Struct, clazz *specNode) {
	fieldCount := struktType.NumFields()
	for i := 0; i < fieldCount; i++ {
		m := struktType.Field(i)
		if m.Exported() {
			mt := translateTypeUnified(docNamespace, m.Type())
			if m.Embedded() {
				mt.Embedded = true
			}
			if mt.Type == "function-signature" {
				mt.Type = "function"
			}
			mt.descriptionAndTags = descriptions[docNamespace+"."+m.Name()]
			clazz.Entries[m.Name()] = mt
		}
	}
}

func fillInMethods(docNamespace string, namedType methodContainer, clazz *specNode) {
	methodCount := namedType.NumMethods()
	for i := 0; i < methodCount; i++ {
		m := namedType.Method(i)
		if m.Exported() {
			methodSpec := translateTypeUnified(docNamespace, m.Type())
			methodSpec.Type = "method"
			methodSpec.descriptionAndTags = descriptions[docNamespace+"."+m.Name()]

			if clazz.Entries == nil {
				clazz.Entries = make(map[string]*specNode)
			}
			clazz.Entries[m.Name()] = methodSpec
		}
	}
}

func fillInEmbeddedMethods(typ types.Type, clazz *specNode) {
	switch typ.(type) {
	case *types.Struct:
		struktType := typ.(*types.Struct)
		fieldCount := struktType.NumFields()
		for i := 0; i < fieldCount; i++ {
			field := struktType.Field(i)
			if field.Embedded() && !field.Exported() {
				embeddedFieldType := field.Type()
				if ptr, ok := embeddedFieldType.(*types.Pointer); ok {
					embeddedFieldType = ptr.Elem()
				}
				if embeddedNamedType, ok := embeddedFieldType.(*types.Named); ok {
					fillInMethods(field.Name(), embeddedNamedType, clazz)
					embeddedFieldType = embeddedNamedType.Underlying()
				}
				if strct, ok := embeddedFieldType.(*types.Struct); ok {
					fillInEmbeddedMethods(strct, clazz)
				}
			}
		}
	}
}

func getNamedName(namedType *types.Named) string {
	if namedType.Obj().Pkg() == nil {
		return namedType.Obj().Name()
	}
	pkg := namedType.Obj().Pkg().Path()
	if pkg == currentPackage {
		return "#/definitions/" + namedType.Obj().Name()
	}
	return "https://golang.org/pkg/" + namedType.Obj().Pkg().Path() + "/" + namedType.Obj().Name()
}

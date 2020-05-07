package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"regexp"
	"sort"
	"strings"
	"sync/atomic"

	"github.com/jaytaylor/html2text"
)

// QlikExtensions represents Qlik JSON Schema extensinos
type QlikExtensions struct {
	QlikStability              string `json:"x-qlik-stability,omitempty"`
	QlikVisibility             string `json:"x-qlik-visibility,omitempty"`
	QlikDeprecated1            bool   `json:"deprecated,omitempty"`
	QlikDeprecated2            bool   `json:"x-qlik-deprecated,omitempty"`
	QlikDeprecationDescription string `json:"x-qlik-deprecation-description,omitempty"`
}

// Info represents engine information
type Info struct {
	Version string `json:"version,omitempty"` //Engine version
	Title   string `json:"title,omitempty"`
}

type OpenRpcFile struct {
	QlikExtensions
	Info       Info               `json:"info,omitempty"`
	OpenRPC    string             `json:"openrpc,omitempty"`
	Methods    []*OpenRpcMethod   `json:"methods,omitempty"`
	Components *OpenRpcComponents `json:"components,omitempty"`
}

type OpenRpcComponents struct {
	Schemas map[string]*Type `json:"schemas,omitempty"`
}

type OpenRpcMethod struct {
	QlikExtensions
	Name        string         `json:"name,omitempty"`
	Description string         `json:"description,omitempty"`
	Parameters  []*Type        `json:"params,omitempty"`
	Responses   *OpenRpcResult `json:"result,omitempty"`
}

type OpenRpcResult struct {
	QlikExtensions
	Name        string `json:"description,omitempty"`
	Description string `json:"description,omitempty"`
	Schema      *Type  `json:"schema,omitempty"`
}

type AdditionalProperties struct {
	AnyOf []*Type `json:"anyOf,omitempty"`
}

// Type represents a JSON Schema object type.
type Type struct {
	QlikExtensions
	AdditionalProperties *AdditionalProperties   `json:"additionalProperties,omitempty"`
	Default              interface{}             `json:"default,omitempty"`
	Description          string                  `json:"description,omitempty"`
	Enum                 []interface{}           `json:"enum,omitempty"`
	Format               string                  `json:"format,omitempty"`
	Items                *Type                   `json:"items,omitempty"`
	Name                 string                  `json:"name,omitempty"`
	OneOf                []*Option               `json:"oneOf,omitempty"`
	Properties           map[OrderAwareKey]*Type `json:"properties,omitempty"` // Special trick with the OrderAwareKey to retain the order of the properties
	Ref                  string                  `json:"$ref,omitempty"`
	Required             bool                    `json:"required,omitempty"`
	Type                 string                  `json:"type,omitempty"`
	Schema               *Type                   `json:"schema,omitempty"`
}

// Option represents a possible value in case of an "enum".
// Title and ConstValue are, and always should be, present.
type Option struct {
	Description string `json:"description,omitempty"`
	Title       string `json:"title"`
	ConstValue  int    `json:"x-qlik-const"`
}

// OrderAwareKey is a special key construct to retain order from json spec for properties
type OrderAwareKey struct {
	Key   string
	Order uint64
}

// OrderAwareKeySlice is an alias of []OrderAwareKey so we can add functions to it. This is used retain the original order of the items in the Properties map
type OrderAwareKeySlice []OrderAwareKey

func (p OrderAwareKeySlice) Len() int           { return len(p) }
func (p OrderAwareKeySlice) Less(i, j int) bool { return p[i].Order < p[j].Order }
func (p OrderAwareKeySlice) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }

var keyOrderCounter uint64

var typesMap map[string]string

// Holds the enigma package to be included before some types like Float64 and RemoteObject if an external spec is used
var enigmaStandardTypesPrefix string

// UnmarshalText allows OrderAwareKey to be used as a json map key
func (k *OrderAwareKey) UnmarshalText(text []byte) error {
	i := atomic.AddUint64(&keyOrderCounter, 1)
	k.Order = i
	k.Key = string(text)
	return nil
}

func getOriginalOrderSortedKeys(contents map[OrderAwareKey]*Type) []OrderAwareKey {
	sortedKeys := make(OrderAwareKeySlice, 0, len(contents))
	for k := range contents {
		sortedKeys = append(sortedKeys, k)
	}
	sort.Sort(sortedKeys)
	return sortedKeys
}
func getAlphabeticSortedKeys(contents map[string]*Type) []string {
	sortedKeys := make([]string, 0, len(contents))
	for k := range contents {
		sortedKeys = append(sortedKeys, k)
	}
	sort.Strings(sortedKeys)
	return sortedKeys
}
func getSortedMethodKeys(contents map[string]*OpenRpcMethod) []string {
	sortedKeys := make([]string, 0, len(contents))
	for k := range contents {
		sortedKeys = append(sortedKeys, k)
	}
	sort.Strings(sortedKeys)
	return sortedKeys
}
func getSortedServiceKeys(contents map[string]map[string]*OpenRpcMethod) []string {
	sortedKeys := make([]string, 0, len(contents))
	for k := range contents {
		sortedKeys = append(sortedKeys, k)
	}
	sort.Strings(sortedKeys)
	return sortedKeys
}

func restructureByRemoteObject(contents []*OpenRpcMethod) map[string]map[string]*OpenRpcMethod {
	mapmap := make(map[string]map[string]*OpenRpcMethod)
	for _, method := range contents {

		separated := strings.Split(method.Name, ".")
		serviceName := separated[0]
		methodName := separated[1]

		if mapmap[serviceName] == nil {
			mapmap[serviceName] = make(map[string]*OpenRpcMethod)
		}
		mapmap[serviceName][methodName] = method
	}
	return mapmap
}

func refToName(refName string) string {
	if refName == "#/components/schemas/JsonObject" {
		return "json.RawMessage"
	}
	name := strings.Replace(refName, "#/components/schemas/", "", 1)
	if typesMap[name] == "string" {
		return "string"
	}
	if typesMap[name] == "array" {
		return name
	}
	return "*" + name
}

func getTypeName(t *Type) string {
	if t.Type == "" && t.Schema != nil {
		t = t.Schema
	}

	switch t.Type {
	case "":
		if t.Ref != "" {
			return refToName(t.Ref)
		}
		return "string"
	case "object":
		if t.Ref != "" {
			return refToName(t.Ref)
		}
		return "json.RawMessage"
	case "array":
		if t.Items != nil {
			return "[]" + getTypeName(t.Items)
		}
		if t.Ref != "" {
			return "[]" + refToName(t.Ref)
		}
		return "[]int"
	case "string":
		return "string"
	case "boolean":
		return "bool"
	case "int":
		return "int"
	case "integer":
		return "int"
	case "number":
		if t.Format != "" {
			switch t.Format {
			case "double":
				return enigmaStandardTypesPrefix + "Float64"
			}
			panic("Unknown format:" + t.Format)
		}
		return enigmaStandardTypesPrefix + "Float64"
	default:
		panic("Unknown type:" + t.Type)
	}
}

func isBuiltInType(t *Type) bool {
	var builtIn = true
	typeName := getTypeName(t)
	if strings.Contains(typeName, "*") && typeName != "*ObjectInterface" {
		builtIn = false
	}
	return builtIn
}

func allTypesBuiltInMap(types map[OrderAwareKey]*Type) bool {
	var result = true
	for _, t := range types {
		result = result && isBuiltInType(t)
	}
	return result
}
func allTypesBuiltIn(types []*Type) bool {
	var result = true
	for _, t := range types {
		result = result && isBuiltInType(t)
	}
	return result
}

func createTypesMap(schema *OpenRpcFile) map[string]string {
	result := map[string]string{}
	for id, t := range schema.Components.Schemas {
		result[id] = t.Type
	}
	return result
}

func loadSchemaFile(schemaFilePath string) (*OpenRpcFile, error) {
	rawSchemaFile, err := ioutil.ReadFile(schemaFilePath)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
	var schema = &OpenRpcFile{}
	err = json.Unmarshal(rawSchemaFile, schema)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	typesMap = createTypesMap(schema)
	return schema, err
}

// Read the file that describes what remote object type is created by each method
func createObjectFunctionToObjectTypeMapping(schemaCompanionFilePath string) map[string]string {
	var result struct {
		RemoteObjectReturnTypes map[string]string `json:"remote-object-return-types"`
	}
	file, err := ioutil.ReadFile(schemaCompanionFilePath)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	err = json.Unmarshal(file, &result)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	return result.RemoteObjectReturnTypes
}

func patchMissingTypeInfo(param *Type, name, methodName string) {
	patchToStringArray := func() {
		if param.Items == nil {
			param.Items = &Type{
				Type: "string",
			}
		}
	}
	patchToIntArray := func() {
		if param.Items == nil {
			param.Items = &Type{
				Type: "int",
			}
		}
	}

	patchToBeArray := func() {
		if param.Type == "" {
			param.Type = "array"
		}
	}

	switch name {
	case "qLabels", "qObjectIdsToPatch", "qNames", "qv", "qTags", "qTerms", "qParameters", "qPaths", "qIds":
		patchToStringArray()
	case "qWarnings", "qColIndices", "qColumnsToSelect", "qRowIndices", "qValues":
		patchToIntArray()
	case "qBreakpoints", "qFieldValues", "qPatches", "qPages", "qRanges", "qFieldsOrColumnsWithWildcards", "qSelections", "qDataRanges":
		patchToBeArray()
	}
}

func getRawInputTypeName(t *Type) string {
	tn := getTypeName(t)
	if strings.Contains(tn, "*") {
		return "interface{}"
	}
	return tn
}

func getRawOutputTypeName(t *Type) string {
	tn := getTypeName(t)
	if strings.Contains(tn, "*") && tn != "*ObjectInterface" {
		return "json.RawMessage"
	}
	return tn
}

func cleanCommentString(comment string, indent string) (string, error) {

	// Fix headings so they end up on a new line
	comment = regexp.MustCompile("(?m)####([^\\n]*)$").ReplaceAllString(comment, "<br><br>$1:<br><br>")
	comment = regexp.MustCompile("(?m)###([^\\n]*)$").ReplaceAllString(comment, "<br><br>$1:<br><br>")
	comment = regexp.MustCompile("(?m)##([^\\n]*)$").ReplaceAllString(comment, "<br><br>$1:<br><br>")
	comment = regexp.MustCompile("(?m)#([^\\n]*)$").ReplaceAllString(comment, "<br><br>$1:<br><br>")
	comment = regexp.MustCompile("(?m)\\*\\*([^\\*]*)\\*\\*$").ReplaceAllString(comment, "<br><br>$1:<br><br>")

	// Remove markdown
	comment = regexp.MustCompile("(?m) \\*\\*([^\\*]*)\\*\\*").ReplaceAllString(comment, " $1")
	comment = regexp.MustCompile("(?m) _([^_]*)_").ReplaceAllString(comment, " $1")

	comment = strings.Replace(comment, "\n", "<br>\n", -1)
	comment, err := html2text.FromString(comment, html2text.Options{PrettyTables: true})
	comment = strings.Replace(comment, "\n*", "\n\n•", -1)
	comment = strings.Replace(comment, "\n+", "\n  +", -1)
	comment = strings.Replace(comment, "\n|", "\n  |", -1)
	comment = strings.Replace(comment, "\n", "\n"+indent+"// ", -1)
	comment = strings.Replace(strings.Replace(comment, "<i>", "   ", -1), "</i>", "    ", -1)
	return comment, err
}

func formatComment(indent string, description string, parameters []*Type) string {
	description, err := cleanCommentString(description, indent)
	if err != nil {
		fmt.Println(err)
	}
	comment := "" + indent + "// " + description
	if parameters != nil && len(parameters) > 0 && parameters[0].Description != "" {
		comment = comment + "\n" + indent + "//" + "\n" + indent + "// Parameters:" + "\n//"
		longestParamNameLen := 0
		for _, param := range parameters {
			if len(param.Name) > longestParamNameLen {
				longestParamNameLen = len(toParamName(param.Name))
			}
		}
		for _, param := range parameters {
			paramDescription, err := cleanCommentString(param.Description, indent)
			if err != nil {
				fmt.Println(err)
			}
			comment = comment + "\n" + indent + "// ◾ " + toParamName(param.Name) + strings.Repeat(" ", longestParamNameLen-len(toParamName(param.Name))) + "   -   " + paramDescription + "\n//"
		}
	}
	return comment
}

func toPublicMemberName(name string) string {
	if name[0:1] == "q" {
		name = name[1:]
	}
	return strings.ToUpper(name[0:1]) + name[1:]
}

func toParamName(name string) string {
	if name[0:1] == "q" {
		name = name[1:]
	}
	return strings.ToLower(name[0:1]) + name[1:]
}

func nilNameInEarlyReturnAfterError(typeName string) string {
	switch typeName {
	case "string":
		return "\"\""
	case "bool":
		return "false"
	default:
		return "nil"
	}
}

func atLeastOneObjectInterface(method *OpenRpcMethod) bool {
	for _, response := range method.Responses.Schema.Properties {
		if getTypeName(response) == "*ObjectInterface" {
			return true
		}
	}
	return false
}

func getPropertiesWithoutFilteredNxInfo(responses map[OrderAwareKey]*Type, methodName string) map[OrderAwareKey]*Type {

	if len(responses) > 1 {
		var qReturnKey OrderAwareKey
		var qReturn *Type
		var qInfo *Type
		for key, value := range responses {
			if key.Key == "qReturn" {
				qReturnKey = key
				qReturn = value
			}
			if key.Key == "qInfo" {
				qInfo = value
			}
		}
		if qReturn != nil && qInfo != nil {
			if len(responses) > 2 {
				panic("Unexpected case that is not handled")
			}
			responses = map[OrderAwareKey]*Type{qReturnKey: qReturn}
		}
	}
	return responses
}

func getPropertiesWithoutRedundantResponses(responses map[OrderAwareKey]*Type, methodName string) map[OrderAwareKey]*Type {
	if len(responses) > 1 && methodName != "GetHyperCubeContinuousData" {
		for key, value := range responses {
			if value.Type == "object" {
				responses = map[OrderAwareKey]*Type{key: value}
				break
			}
		}
	}
	return responses
}

func wrapWithEnigmaStandardTypesPrefix(str string) string {
	switch str {
	case "*ObjectInterface":
		return "*" + enigmaStandardTypesPrefix + "ObjectInterface"
	}
	return str
}

// Generate an ordinary fully typed method
func printMethod(method *OpenRpcMethod, out *os.File, serviceName string, methodName string, objectFuncToObject map[string]string) {
	responseMap := getPropertiesWithoutFilteredNxInfo(method.Responses.Schema.Properties, methodName)
	sortedResponseKeys := getOriginalOrderSortedKeys(responseMap)
	actualResponseMap := getPropertiesWithoutRedundantResponses(responseMap, methodName)
	sortedActualResponseKeys := getOriginalOrderSortedKeys(actualResponseMap)
	// Generate Description
	if method.Description != "" {
		fmt.Fprintln(out, formatComment("", method.Description, method.Parameters))
	}
	printExtensionTags(out, "", method.QlikExtensions)
	fmt.Fprint(out, "func (obj *", serviceName, ") ", methodName, "(ctx context.Context")

	// Generate Parameters
	for _, param := range method.Parameters {
		fmt.Fprint(out, ", ", "", toParamName(param.Name), " ", getTypeName(param))
	}
	fmt.Fprint(out, ") ")

	// Generate Return Types
	if len(actualResponseMap) > 0 {
		fmt.Fprint(out, "(")
	}

	for _, responseKey := range sortedActualResponseKeys {
		responseType := actualResponseMap[responseKey]
		typeName := getTypeName(responseType)
		if typeName == "*ObjectInterface" {
			// Replace the generic ObjectInterface pointer with the right Remote Object API struct
			objectTypeName := objectFuncToObject[serviceName+"."+methodName]
			if objectTypeName == "" {
				fmt.Println("method with unknown return type:" + method.Name)
				fmt.Fprint(out, "*"+enigmaStandardTypesPrefix+"RemoteObject, ")
			} else {
				fmt.Fprint(out, "*"+objectTypeName, ", ")
			}
		} else {
			fmt.Fprint(out, typeName, ", ")
		}

	}
	fmt.Fprint(out, "error")
	if len(actualResponseMap) > 0 {
		fmt.Fprint(out, ")")
	}

	// Generate Start of Function body
	fmt.Fprintln(out, " {")

	// Generate an anonymous result container struct
	var resultParamName string
	if len(sortedResponseKeys) > 0 {
		fmt.Fprintln(out, "\tresult := &struct {")
		for _, responseKey := range sortedResponseKeys {
			responseType := responseMap[responseKey]
			fmt.Fprintln(out, "\t\t"+toPublicMemberName(responseKey.Key), wrapWithEnigmaStandardTypesPrefix(getTypeName(responseType))+" `json:\""+responseKey.Key+"\"`")
		}
		fmt.Fprintln(out, "\t} {}")
		resultParamName = "result"
	} else {
		resultParamName = "nil"
	}

	// Generate the actual call down to the RPC machinery
	fmt.Fprint(out, "\terr := obj.Rpc(ctx, \"", methodName, "\", ", resultParamName)
	for _, param := range method.Parameters {
		// Fill in the parameters in the parameter array
		fmt.Fprint(out, ", ", toParamName(param.Name))
	}
	fmt.Fprintln(out, ")")

	// If there is at least one Remote Object that will be created then check for nil and return early
	if atLeastOneObjectInterface(method) {
		fmt.Fprintln(out, "\tif err != nil {")
		fmt.Fprint(out, "\t\treturn ")
		for _, responseKey := range sortedActualResponseKeys {
			responseType := actualResponseMap[responseKey]
			fmt.Fprint(out, nilNameInEarlyReturnAfterError(getTypeName(responseType))+", ")
		}
		fmt.Fprintln(out, "err")
		fmt.Fprintln(out, "\t}")
	}

	fmt.Fprint(out, getExtraCrossAssignmentLine(methodName))

	// Return the result including creating a new Remote Object if needed
	fmt.Fprint(out, "\treturn ")
	for _, responseKey := range sortedActualResponseKeys {
		responseType := actualResponseMap[responseKey]
		if getTypeName(responseType) == "*ObjectInterface" {
			objectAPITypeName := objectFuncToObject[serviceName+"."+methodName]
			if objectAPITypeName == "" {
				fmt.Fprint(out, "obj.GetRemoteObject(result."+responseKey.Key[1:]+"), ")
			} else {
				fmt.Fprint(out, "&"+objectAPITypeName+"{obj.GetRemoteObject(result."+responseKey.Key[1:]+")}, ")
			}
		} else {
			fmt.Fprint(out, "result."+toPublicMemberName(responseKey.Key)+", ")
		}
	}
	fmt.Fprintln(out, "err ")
	fmt.Fprintln(out, "}")
	fmt.Fprintln(out, "")
}

// Generate a raw method that allows raw json input and output
func printRawMethod(method *OpenRpcMethod, out *os.File, serviceName string, methodName string, objectFuncToObject map[string]string) {
	responseMap := getPropertiesWithoutFilteredNxInfo(method.Responses.Schema.Properties, methodName)
	sortedResponseKeys := getOriginalOrderSortedKeys(responseMap)
	actualResponseMap := getPropertiesWithoutRedundantResponses(responseMap, methodName)
	sortedActualResponseKeys := getOriginalOrderSortedKeys(actualResponseMap)

	// Generate Description
	if method.Description != "" {
		fmt.Fprintln(out, formatComment("", method.Description, method.Parameters))
	}
	printExtensionTags(out, "", method.QlikExtensions)
	fmt.Fprint(out, "func (obj *", serviceName, ") ", methodName, "Raw(ctx context.Context")

	// Generate Parameters
	for _, param := range method.Parameters {
		fmt.Fprint(out, ", ", "", toParamName(param.Name), " ", getRawInputTypeName(param))
	}
	fmt.Fprint(out, ")")

	// Generate Return Types
	if len(actualResponseMap) > 0 {
		fmt.Fprint(out, " (")
	}
	for _, responseKey := range sortedActualResponseKeys {
		response := actualResponseMap[responseKey]
		typeName := getRawOutputTypeName(response)
		if typeName == "*ObjectInterface" {
			objectTypeName := objectFuncToObject[serviceName+"."+methodName]
			if objectTypeName == "" {
				fmt.Fprint(out, "*"+enigmaStandardTypesPrefix+"RemoteObject, ")
			} else {
				fmt.Fprint(out, "*"+objectTypeName, ", ")
			}
		} else {
			fmt.Fprint(out, getRawOutputTypeName(response), ", ")
		}

	}
	fmt.Fprint(out, "error")
	if len(actualResponseMap) > 0 {
		fmt.Fprint(out, ")")
	}

	// Generate Start of Function body
	fmt.Fprintln(out, " {")

	// Generate an anonymous result container struct
	var resultParamName string
	if len(actualResponseMap) > 0 {
		fmt.Fprintln(out, "\tresult := &struct {")
		for _, responseKey := range sortedResponseKeys {
			responseType := responseMap[responseKey]
			fmt.Fprintln(out, "\t\t"+toPublicMemberName(responseKey.Key), wrapWithEnigmaStandardTypesPrefix(getRawOutputTypeName(responseType))+" `json:\""+responseKey.Key+"\"`")
		}
		fmt.Fprintln(out, "\t} {}")
		resultParamName = "result"
	} else {
		resultParamName = "nil"
	}

	// Generate the actual call down to the RPC machinery
	fmt.Fprint(out, "\terr := obj.Rpc(ctx, \"", methodName, "\", ", resultParamName)
	for _, param := range method.Parameters {
		fmt.Fprint(out, ", ", toParamName(param.Name))
	}
	fmt.Fprintln(out, ")")

	// If there is at least one Remote Object that will be created then check for nil and return early
	if atLeastOneObjectInterface(method) {
		fmt.Fprintln(out, "\tif err != nil {")
		fmt.Fprint(out, "\t\treturn ")
		for _, response := range actualResponseMap {
			// Fill in the parameters in the parameter array
			fmt.Fprint(out, nilNameInEarlyReturnAfterError(response.Type)+", ")
		}
		fmt.Fprintln(out, "err")
		fmt.Fprintln(out, "\t}")
	}

	fmt.Fprint(out, getExtraCrossAssignmentLine(methodName))

	// Return the result including creating a new Remote Object if needed
	fmt.Fprint(out, "\treturn ")
	for _, responseKey := range sortedActualResponseKeys {
		responseType := actualResponseMap[responseKey]
		if getTypeName(responseType) == "*ObjectInterface" {
			objectAPITypeName := objectFuncToObject[serviceName+"."+methodName]
			if objectAPITypeName == "" {
				fmt.Fprint(out, "obj.GetRemoteObject(result."+responseKey.Key[1:]+"), ")
			} else {
				fmt.Fprint(out, "&"+objectAPITypeName+"{obj.GetRemoteObject(result."+responseKey.Key[1:]+")}, ")
			}
		} else {
			fmt.Fprint(out, "result."+toPublicMemberName(responseKey.Key)+", ")
		}
	}
	fmt.Fprintln(out, "err ")
	fmt.Fprintln(out, "}")
	fmt.Fprintln(out, "")
}

func printExtensionTags(out *os.File, indent string, extensions QlikExtensions) {

	if extensions.QlikDeprecated1 || extensions.QlikDeprecated2 {
		if extensions.QlikDeprecationDescription != "" {
			fmt.Fprintln(out, indent+"// Deprecated: "+extensions.QlikDeprecationDescription)
		} else {
			fmt.Fprintln(out, indent+"// Deprecated: This will be removed in a future version")
		}
	}
	if extensions.QlikStability != "" {
		fmt.Fprintln(out, indent+"// Stability: "+extensions.QlikStability)
	}

}

func printErrorCodeLookup(out *os.File, def *Type) {
	fmt.Fprintln(out, "func errorCodeLookup(c int) string {")
	fmt.Fprintln(out, "switch c {")
	for _, opt := range def.OneOf {
		fmt.Fprintln(out, "case", opt.ConstValue, ":")
		fmt.Fprintln(out, "return \""+opt.Title+"\"")
	}
	fmt.Fprint(out, "}\nreturn \"\"\n}\n\n")
}

func isNonZero(value interface{}) bool {
	return !(value == nil || value == "" || value == float64(0) || value == 0 || value == false)
}

func hasEnumRef(property *Type) bool {
	if property.Ref != "" {
		name := refToName(property.Ref)
		// "Enums" have type string but are sent as "objects" with a list of valid options for said "enum".
		return name == "string"
	}
	return false
}

func getExtraCrossAssignmentLine(methodName string) string {
	switch methodName {
	case "GetMediaList":
		return ""
	case "GetInteract":
		return ""
	case "CreateSessionApp":
		return "\tresult.Return.GenericId = result.SessionAppId\n"
	case "CreateDocEx":
		return "\tresult.Return.GenericId = result.DocId\n"
	case "CreateSessionAppFromApp":
		return "\tresult.Return.GenericId = result.SessionAppId\n"
	default:
		return ""
	}
}

func main() {

	if len(os.Args) < 5 {
		fmt.Println("Usage: go run schema/generate.go <file path to spec.json> <file path to generated file.go> <generated package name>")
		os.Exit(1)
	}
	schemaFilePath := os.Args[1]
	schemaCompanionFilePath := os.Args[2]
	generatedFilePath := os.Args[3]
	generatedFilePackage := os.Args[4]
	enigmaImports := ""
	enigmaStandardTypesPrefix = ""
	// The fourth argument disable-enigma-import is used to indicate that we are genering the standard
	// schema file inside of enigma-go and not an external one.
	if !(len(os.Args) > 5 && os.Args[5] == "disable-enigma-import") {
		enigmaImports = "\t\"github.com/qlik-oss/enigma-go\"\n"
		enigmaStandardTypesPrefix = "enigma."
	}

	objectFuncToObject := createObjectFunctionToObjectTypeMapping(schemaCompanionFilePath)
	schemaFile, err := loadSchemaFile(schemaFilePath)

	// Start generating the go file
	out, err := os.Create(generatedFilePath)
	defer out.Close()
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	fmt.Fprintln(out, "// Code generated by QIX generator (./schema/generate.go) for Qlik Associative Engine version", schemaFile.Info.Version, ". DO NOT EDIT.")
	fmt.Fprintln(out)
	fmt.Fprintln(out, "package "+generatedFilePackage)
	fmt.Fprintln(out, "import (")
	fmt.Fprintln(out, "\t\"context\"")
	fmt.Fprintln(out, "\t\"encoding/json\"")
	fmt.Fprint(out, enigmaImports)
	fmt.Fprint(out, ")\n\n")
	fmt.Fprintln(out, "// Version of the schema used to generate the enigma.go QIX API")
	fmt.Fprintf(out, "const QIX_SCHEMA_VERSION = \"%s\"\n\n", schemaFile.Info.Version)

	// Generate definition data type structs
	definitionKeys := getAlphabeticSortedKeys(schemaFile.Components.Schemas)
	for _, defName := range definitionKeys {
		if defName == "JsonObject" {
			continue
		}
		def := schemaFile.Components.Schemas[defName]
		if def.Description != "" {
			fmt.Fprintln(out, formatComment("", def.Description, nil))
		}
		printExtensionTags(out, "", def.QlikExtensions)
		switch def.Type {
		case "object":
			fmt.Fprintln(out, "type", defName, "struct {")
			// types
			propertiesKeys := getOriginalOrderSortedKeys(def.Properties)
			for _, key := range propertiesKeys {
				propertyName := key.Key
				property := def.Properties[key]
				printStructMember(propertyName, property, out)
			}
			if def.AdditionalProperties != nil {
				for _, property := range def.AdditionalProperties.AnyOf {
					printStructMember(property.Name, property.Schema, out)
				}
			}
			fmt.Fprintln(out, "}")
			fmt.Fprintln(out, "")
		case "array":
			fmt.Fprintln(out, "type", defName, getTypeName(def))
			fmt.Fprintln(out, "")
		case "string":
			// Enums are strings now and only one of them is of particular interest.
			if defName == "NxLocalizedErrorCode" {
				printErrorCodeLookup(out, def)
			}
		default:
			fmt.Fprintln(out, "<<<other>>>", defName, def.Type)
			fmt.Fprintln(out, "")
		}
	}

	mapmap := restructureByRemoteObject(schemaFile.Methods)
	// Generate structs for the remote objects (service APIs)
	serviceNames := getSortedServiceKeys(mapmap)
	for _, serviceName := range serviceNames {

		var serviceImplName = serviceName
		fmt.Fprintln(out, "type ", serviceImplName, "struct {")
		fmt.Fprintln(out, "\t*"+enigmaStandardTypesPrefix+"RemoteObject")
		fmt.Fprintln(out, "}")

		methodKeys := getSortedMethodKeys(mapmap[serviceName])
		for _, methodName := range methodKeys {
			method := mapmap[serviceName][methodName]

			for _, par := range method.Parameters {
				patchMissingTypeInfo(par.Schema, par.Name, methodName)
			}

			for key, par := range method.Responses.Schema.Properties {
				patchMissingTypeInfo(par, key.Key, methodName)
			}

			// Generate typed methods
			printMethod(method, out, serviceName, methodName, objectFuncToObject)

			// Generate untyped (raw) methods for those with complex parameters and/or return values
			actualResponses := getPropertiesWithoutFilteredNxInfo(method.Responses.Schema.Properties, methodName)
			if !(allTypesBuiltIn(method.Parameters) && allTypesBuiltInMap(actualResponses)) {
				printRawMethod(method, out, serviceName, methodName, objectFuncToObject)
			}
		}
	}
}

func printStructMember(propertyName string, property *Type, out *os.File) {
	if propertyName[0:1] == "q" {
		propertyName = propertyName[1:]
	}
	if property.Description != "" {
		fmt.Fprintln(out, formatComment("\t", property.Description, nil))
	}
	printExtensionTags(out, "\t", property.QlikExtensions)
	if isNonZero(property.Default) && !hasEnumRef(property) {
		fmt.Fprintln(out, "\t// When set to nil the default value is used, when set to point at a value that value is used (including golang zero values)")
		fmt.Fprint(out, "\t", toPublicMemberName(propertyName), " *", getTypeName(property), " `json:\"q", propertyName, ",omitempty\"`")
	} else {
		fmt.Fprint(out, "\t", toPublicMemberName(propertyName), " ", getTypeName(property), " `json:\"q", propertyName, ",omitempty\"`")
	}

	fmt.Fprintln(out, "")
}

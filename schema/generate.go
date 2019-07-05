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
	QlikStability  string `json:"x-qlik-stability,omitempty"`
	QlikVisibility string `json:"x-qlik-visibility,omitempty"`
}

// Info represents engine information
type Info struct {
	Version string `json:"version,omitempty"` //Engine version
	Title   string `json:"title,omitempty"`
}

// Layout represents GenericObject Layout definition
type Layout struct {
	Prop   *Type `json:"prop,omitempty"`
	Layout *Type `json:"layout,omitempty"`
}

// Schema represents a JSON Schema
type Schema struct {
	QlikExtensions
	Info       Info                        `json:"info,omitempty"`
	OpenAPI    string                      `json:"openapi,omitempty"`
	Components map[string]map[string]*Type `json:"components,omitempy"`
	Services   map[string]*Service         `json:"x-qlik-services,omitempy"`
}

// Service represents a JSON Schema service
type Service struct {
	QlikExtensions
	Description string
	Methods     map[string]*Methodx `json:"methods,omitempty"`
	Layouts     []*Layout           `json:"layouts,omitempty"`
}

// Methodx represents a JSON Schema method
type Methodx struct {
	QlikExtensions
	Description string  `json:"description,omitempty"`
	Parameters  []*Type `json:"parameters,omitempty"`
	Responses   []*Type `json:"responses,omitempty"`
}

// Type represents a JSON Schema object type.
type Type struct {
	QlikExtensions
	AdditionalItems      *Type                   `json:"additionalItems,omitempty"`
	AdditionalProperties json.RawMessage         `json:"additionalProperties,omitempty"`
	AllOf                []*Type                 `json:"allOf,omitempty"`
	AnyOf                []*Type                 `json:"anyOf,omitempty"`
	BinaryEncoding       string                  `json:"binaryEncoding,omitempty"`
	Default              interface{}             `json:"default,omitempty"`
	Definitions          map[string]*Type        `json:"definitions,omitempty"`
	Dependencies         map[string]*Type        `json:"dependencies,omitempty"`
	Description          string                  `json:"description,omitempty"`
	Enum                 []interface{}           `json:"enum,omitempty"`
	ExclusiveMaximum     bool                    `json:"exclusiveMaximum,omitempty"`
	ExclusiveMinimum     bool                    `json:"exclusiveMinimum,omitempty"`
	Format               string                  `json:"format,omitempty"`
	Items                *Type                   `json:"items,omitempty"`
	Maximum              int                     `json:"maximum,omitempty"`
	MaxItems             int                     `json:"maxItems,omitempty"`
	MaxLength            int                     `json:"maxLength,omitempty"`
	MaxProperties        int                     `json:"maxProperties,omitempty"`
	Media                *Type                   `json:"media,omitempty"`
	Minimum              int                     `json:"minimum,omitempty"`
	MinItems             int                     `json:"minItems,omitempty"`
	MinLength            int                     `json:"minLength,omitempty"`
	MinProperties        int                     `json:"minProperties,omitempty"`
	MultipleOf           int                     `json:"multipleOf,omitempty"`
	Not                  *Type                   `json:"not,omitempty"`
	Name                 string                  `json:"name,omitempty"`
	OneOf                []*Option               `json:"oneOf,omitempty"`
	Pattern              string                  `json:"pattern,omitempty"`
	PatternProperties    map[string]*Type        `json:"patternProperties,omitempty"`
	Properties           map[OrderAwareKey]*Type `json:"properties,omitempty"` // Special trick with the OrderAwareKey to retain the order of the properties
	Ref                  string                  `json:"$ref,omitempty"`
	Required             bool                    `json:"required,omitempty"`
	Title                string                  `json:"title,omitempty"`
	Type                 string                  `json:"type,omitempty"`
	UniqueItems          bool                    `json:"uniqueItems,omitempty"`
	Schema               *Type                   `json:"schema,omitempty"`
}

// Option represents a possible value in case of an "enum"
type Option struct {
	Description string `json:"description,omitempty"`
	Title       string `json:"title,omitempty"`
	ConstValue  int    `json:"x-qlik-const,omitempty"`
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

// UnmarshalText allows OrderAwareKey to be used as a json map key
func (k *OrderAwareKey) UnmarshalText(text []byte) error {
	i := atomic.AddUint64(&keyOrderCounter, 1)
	k.Order = i
	k.Key = string(text)
	return nil
}

func newOrderAwareKey(key string) OrderAwareKey {
	i := atomic.AddUint64(&keyOrderCounter, 1)
	return OrderAwareKey{Key: key, Order: i}
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
func getSortedMethodKeys(contents map[string]*Methodx) []string {
	sortedKeys := make([]string, 0, len(contents))
	for k := range contents {
		sortedKeys = append(sortedKeys, k)
	}
	sort.Strings(sortedKeys)
	return sortedKeys
}
func getSortedServiceKeys(contents map[string]*Service) []string {
	sortedKeys := make([]string, 0, len(contents))
	for k := range contents {
		sortedKeys = append(sortedKeys, k)
	}
	sort.Strings(sortedKeys)
	return sortedKeys
}

func refToName(refName string) string {
	if refName == "#/components/schemas/JsonObject" {
		return "json.RawMessage"
	}
	name := strings.Replace(refName, "#/components/schemas/", "", 1)
	if typesMap[name] == "string" {
		return name
	}
	if typesMap[name] == "array" {
		return name
	}
	return "*" + name
}

func getInnerType(t *Type) string {
	var innerType *Type
	if t.Items != nil {
		innerType = t.Items
	} else if t.Schema != nil {
		innerType = t.Schema
	} else {
		innerType = t
	}
	var s string
	if innerType.Ref != "" {
		s = refToName(innerType.Ref)
	} else {
		s = getTypeName(innerType)
	}
	return s
}

func getTypeName(t *Type) string {
	switch t.Type {
	case "object":
		return getInnerType(t)
	case "array":
		return "[]" + getInnerType(t)
	case "boolean":
		return "bool"
	case "integer":
		return "int"
	case "number":
		if t.Format != "" {
			switch t.Format {
			case "double":
				return "Float64"
			}
			return t.Format
		}
		return "Float64"
	default:
		// Handle cases where type is not defined
		if t.Type == "" {
			if t.Items != nil {
				return "[]" + getInnerType(t)
			}
		}
		return t.Type
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

func allTypesBuiltIn(types []*Type) bool {
	var result = true
	for _, t := range types {
		result = result && isBuiltInType(t)
	}
	return result
}

func createTypesMap(schema *Schema) map[string]string {
	result := map[string]string{}
	for id, t := range schema.Components["schemas"] {
		result[id] = t.Type
	}
	return result
}

func loadSchemaFile() (*Schema, error) {
	pwd, err := os.Getwd()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	rawSchemaFile, err := ioutil.ReadFile(pwd + "/schema.json")
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
	var schema = &Schema{}
	err = json.Unmarshal(rawSchemaFile, schema)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	// Fix some issues in the spec file
	schema.Components["schemas"]["ObjectInterface"].Type = "object"
	delete(schema.Components["schemas"], "JsonObject")
	typesMap = createTypesMap(schema)
	return schema, err
}

func expandGenericObjectWithLayouts(schema *Schema) {
	// Expand the generic object properties schema with layouts
	genericObjectProperties := schema.Components["schemas"]["GenericObjectProperties"]
	genericObjectLayout := schema.Components["schemas"]["GenericObjectLayout"]
	layouts := schema.Services["GenericObject"].Layouts
	for _, layout := range layouts {
		if layout.Prop != nil {
			propName := layout.Prop.Name
			if propName[0:1] == "q" {
				propName = propName[1:]
			}
			genericObjectProperties.Properties[newOrderAwareKey(propName)] = layout.Prop
		}
		if layout.Layout != nil {
			layoutName := layout.Layout.Name
			if layoutName[0:1] == "q" {
				layoutName = layoutName[1:]
			}
			genericObjectLayout.Properties[newOrderAwareKey(layoutName)] = layout.Layout
		}
	}
}

func createObjectFunctionToObjectTypeMapping() map[string]string {
	// Register mappings that describe what remote object type is created by each method
	objectFuncToObject := make(map[string]string)
	objectFuncToObject["Global.GetActiveDoc"] = "Doc"
	objectFuncToObject["Global.OpenDoc"] = "Doc"
	objectFuncToObject["Global.CreateDocEx"] = "Doc"
	objectFuncToObject["Global.CreateSessionAppFromApp"] = "Doc"
	objectFuncToObject["Global.CreateSessionApp"] = "Doc"
	objectFuncToObject["GenericObject.CreateChild"] = "GenericObject"
	objectFuncToObject["GenericObject.GetChild"] = "GenericObject"
	objectFuncToObject["GenericObject.GetSnapshotObject"] = "GenericObject"
	objectFuncToObject["GenericObject.GetParent"] = "GenericObject"
	objectFuncToObject["Doc.CreateSessionObject"] = "GenericObject"
	objectFuncToObject["Doc.CreateBookmark"] = "GenericBookmark"
	objectFuncToObject["Doc.GetDimension"] = "GenericDimension"
	objectFuncToObject["Doc.CreateMeasure"] = "GenericMeasure"
	objectFuncToObject["Doc.GetField"] = "Field"
	objectFuncToObject["Doc.CreateSessionVariable"] = "Variable"
	objectFuncToObject["Doc.GetVariable"] = "Variable"
	objectFuncToObject["Doc.GetObject"] = "GenericObject"
	objectFuncToObject["Doc.GetVariableById"] = "Variable"
	objectFuncToObject["Doc.CreateObject"] = "GenericObject"
	objectFuncToObject["Doc.CreateVariableEx"] = "Variable"
	objectFuncToObject["Doc.CreateDimension"] = "GenericDimension"
	objectFuncToObject["Doc.GetBookmark"] = "GenericBookmark"
	objectFuncToObject["Doc.GetVariableByName"] = "GenericVariable"
	objectFuncToObject["Doc.GetMeasure"] = "GenericMeasure"
	return objectFuncToObject
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

func atLeastOneObjectInterface(method *Methodx) bool {
	for _, response := range method.Responses {
		if getTypeName(response) == "*ObjectInterface" {
			return true
		}
	}
	return false
}

func filterResponses(responses []*Type, methodName string) []*Type {
	if len(responses) > 1 && methodName != "GetHyperCubeContinuousData" {
		for _, response := range responses {
			if response.Type == "object" {
				return []*Type{response}
			}
		}
	}
	return responses
}

func removeRedundantNxInfo(schema *Schema) {
	for _, service := range schema.Services {
		for _, method := range service.Methods {
			if len(method.Responses) > 1 {
				if method.Responses[0].Name == "qInfo" && method.Responses[1].Name == "qReturn" {
					method.Responses = []*Type{method.Responses[1]}
				} else if method.Responses[0].Name == "qReturn" && method.Responses[1].Name == "qInfo" {
					method.Responses = []*Type{method.Responses[0]}
				}
			}
		}
	}
}

func printEnumMethods(out *os.File, typeName string) {
  // Create String() method for type
  // Using Fprint instead of Fprintln due to lintchecks on trailing newline otherwise
  fmt.Fprintf(out, "func (t %s) String() string {\n", typeName)
  fmt.Fprint(out, "\treturn string(t)\n")
  fmt.Fprint(out, "}\n\n")
  // Create MarshalText() method for type
  fmt.Fprintf(out, "func (t %s) MarshalText() ([]byte, error) {\n", typeName)
  fmt.Fprint(out, "\terr := validateArg(t)\n")
  fmt.Fprint(out, "\treturn []byte(t), err\n")
  fmt.Fprint(out, "}\n\n")
}

func generateArgumentInitForType(typeName string, t *Type) string {
  validArgs := []string{}
  for _, opt := range t.OneOf {
    if opt.Description != "" {
      validArgs = append(validArgs, "\"" + opt.Description + "\"")
    }
    if opt.Title != "" {
      validArgs = append(validArgs, "\"" + opt.Title + "\"")
    }
  }
  validArgsString := "[]string{" + strings.Join(validArgs, ", ") + "}"
  // Add valid arguments for the type
  return fmt.Sprintf("AddArgumentsForType(%s(\"\"), %s)", typeName, validArgsString)
}

func generateArgumentInitFunc(out *os.File, funcCalls []string) {
  fmt.Fprint(out, "var argInitCalled = false\n\n")
  fmt.Fprint(out, "//argInit initializes all the valid arguments for the generated \"enum\" (string) types.\n")
  fmt.Fprint(out, "//This method should be called once for the argument validation to work.\n")
  fmt.Fprint(out, "func argInit() {\n")
  fmt.Fprint(out, "\targInitCalled = true\n")
  fmt.Fprintf(out, "\t%s\n", strings.Join(funcCalls, "\n\t"))
  fmt.Fprint(out, "}\n\n")
}

// Generate an ordinary fully typed method
func printMethod(method *Methodx, out *os.File, serviceName string, methodName string, objectFuncToObject map[string]string) {

	actualResponses := filterResponses(method.Responses, methodName)
	// Generate Description
	if method.Description != "" {
		fmt.Fprintln(out, formatComment("", method.Description, method.Parameters))
	}
	fmt.Fprint(out, "func (obj *", serviceName, ") ", methodName, "(ctx context.Context")

	// Generate Parameters
	for _, param := range method.Parameters {
		fmt.Fprint(out, ", ", "", toParamName(param.Name), " ", getTypeName(param))
	}
	fmt.Fprint(out, ") ")

	// Generate Return Types
	if len(actualResponses) > 0 {
		fmt.Fprint(out, "(")
	}
	for _, response := range actualResponses {
		typeName := getTypeName(response)
		if typeName == "*ObjectInterface" {
			// Replace the generic ObjectInterface pointer with the right Remote Object API struct
			objectTypeName := objectFuncToObject[serviceName+"."+methodName]
			if objectTypeName == "" {
				panic("Unknown remote object type for " + serviceName + "." + methodName)
			}
			fmt.Fprint(out, "*"+objectTypeName, ", ")
		} else {
			fmt.Fprint(out, getTypeName(response), ", ")
		}

	}
	fmt.Fprint(out, "error")
	if len(actualResponses) > 0 {
		fmt.Fprint(out, ")")
	}

	// Generate Start of Function body
	fmt.Fprintln(out, " {")

	// Generate an anonymous result container struct
	var resultParamName string
	if len(method.Responses) > 0 {
		fmt.Fprintln(out, "\tresult := &struct {")
		for _, response := range method.Responses {
			fmt.Fprintln(out, "\t\t"+toPublicMemberName(response.Name), getTypeName(response)+"\t`json:\""+response.Name+"\"`")
		}
		fmt.Fprintln(out, "\t} {}")
		resultParamName = "result"
	} else {
		resultParamName = "nil"
	}

	// Generate the actual call down to the RPC machinery
	fmt.Fprint(out, "\terr := obj.rpc(ctx, \"", methodName, "\", ", resultParamName)
	for _, param := range method.Parameters {
		// Fill in the parameters in the parameter array
		fmt.Fprint(out, ", ", toParamName(param.Name))
	}
	fmt.Fprintln(out, ")")

	// If there is at least one Remote Object that will be created then check for nil and return early
	if atLeastOneObjectInterface(method) {
		fmt.Fprintln(out, "\tif err != nil {")
		fmt.Fprint(out, "\t\treturn ")
		for _, response := range actualResponses {
			fmt.Fprint(out, nilNameInEarlyReturnAfterError(response.Type)+", ")
		}
		fmt.Fprintln(out, "err")
		fmt.Fprintln(out, "\t}")
	}

	fmt.Fprint(out, getExtraCrossAssignmentLine(methodName))

	// Return the result including creating a new Remote Object if needed
	fmt.Fprint(out, "\treturn ")
	for _, response := range actualResponses {
		if getTypeName(response) == "*ObjectInterface" {
			objectAPITypeName := objectFuncToObject[serviceName+"."+methodName]
			fmt.Fprint(out, "&"+objectAPITypeName+"{obj.session.getRemoteObject(result."+response.Name[1:]+")}, ")
		} else {
			fmt.Fprint(out, "result."+toPublicMemberName(response.Name)+", ")
		}
	}
	fmt.Fprintln(out, "err ")
	fmt.Fprintln(out, "}")
	fmt.Fprintln(out, "")
}

// Generate a raw method that allows raw json input and output
func printRawMethod(method *Methodx, out *os.File, serviceName string, methodName string, objectFuncToObject map[string]string) {

	actualResponses := filterResponses(method.Responses, methodName)

	// Generate Description
	if method.Description != "" {
		fmt.Fprintln(out, formatComment("", method.Description, method.Parameters))
	}
	fmt.Fprint(out, "func (obj *", serviceName, ") ", methodName, "Raw(ctx context.Context")

	// Generate Parameters
	for _, param := range method.Parameters {
		fmt.Fprint(out, ", ", "", toParamName(param.Name), " ", getRawInputTypeName(param))
	}
	fmt.Fprint(out, ")")

	// Generate Return Types
	if len(actualResponses) > 0 {
		fmt.Fprint(out, " (")
	}
	for _, response := range actualResponses {
		typeName := getRawOutputTypeName(response)
		if typeName == "*ObjectInterface" {
			objectTypeName := objectFuncToObject[serviceName+"."+methodName]
			if objectTypeName == "" {
				panic("Unknown remote object type for" + serviceName + "." + methodName)
			}
			fmt.Fprint(out, "*"+objectTypeName, ", ")
		} else {
			fmt.Fprint(out, getRawOutputTypeName(response), ", ")
		}

	}
	fmt.Fprint(out, "error")
	if len(actualResponses) > 0 {
		fmt.Fprint(out, ")")
	}

	// Generate Start of Function body
	fmt.Fprintln(out, " {")

	// Generate an anonymous result container struct
	var resultParamName string
	if len(method.Responses) > 0 {
		fmt.Fprintln(out, "\tresult := &struct {")
		for _, response := range method.Responses {
			fmt.Fprintln(out, "\t\t"+toPublicMemberName(response.Name), getRawOutputTypeName(response)+"\t`json:\""+response.Name+"\"`")
		}
		fmt.Fprintln(out, "\t} {}")
		resultParamName = "result"
	} else {
		resultParamName = "nil"
	}

	// Generate the actual call down to the RPC machinery
	fmt.Fprint(out, "\terr := obj.rpc(ctx, \"", methodName, "\", ", resultParamName)
	for _, param := range method.Parameters {
		fmt.Fprint(out, ", ensureEncodable(", toParamName(param.Name), ")")
	}
	fmt.Fprintln(out, ")")

	// If there is at least one Remote Object that will be created then check for nil and return early
	if atLeastOneObjectInterface(method) {
		fmt.Fprintln(out, "\tif err != nil {")
		fmt.Fprint(out, "\t\treturn ")
		for _, response := range actualResponses {
			// Fill in the parameters in the parameter array
			fmt.Fprint(out, nilNameInEarlyReturnAfterError(response.Type)+", ")
		}
		fmt.Fprintln(out, "err")
		fmt.Fprintln(out, "\t}")
	}

	fmt.Fprint(out, getExtraCrossAssignmentLine(methodName))

	// Return the result including creating a new Remote Object if needed
	fmt.Fprint(out, "\treturn ")
	for _, response := range actualResponses {
		if getTypeName(response) == "*ObjectInterface" {
			objectAPITypeName := objectFuncToObject[serviceName+"."+methodName]
			fmt.Fprint(out, "&"+objectAPITypeName+"{obj.session.getRemoteObject(result."+response.Name[1:]+")}, ")
		} else {
			fmt.Fprint(out, "result."+toPublicMemberName(response.Name)+", ")
		}
	}
	fmt.Fprintln(out, "err ")
	fmt.Fprintln(out, "}")
	fmt.Fprintln(out, "")
}

func isNonZero(value interface{}) bool {
	return !(value == nil || value == "" || value == float64(0) || value == 0 || value == false)
}

func hasEnumRef(property *Type) bool {
	if property.Ref != "" {
		name := refToName(property.Ref)
		// "Enums" have type string but are sent as "objects" with a list of valid options for said "enum".
		return typesMap[name] == "string"
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
	objectFuncToObject := createObjectFunctionToObjectTypeMapping()

	schema, err := loadSchemaFile()

	pwd, err := os.Getwd()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	expandGenericObjectWithLayouts(schema)
	removeRedundantNxInfo(schema)
	// Start generating the go file
	out, err := os.Create(pwd + "/../qix_generated.go")
	defer out.Close()
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	fmt.Fprintln(out, "// Code generated by QIX generator (./schema/generate.go) for Qlik Associative Engine version", schema.Info.Version, ". DO NOT EDIT.")
	fmt.Fprintln(out)
	fmt.Fprintln(out, "package enigma")
	fmt.Fprintln(out, "import (")
	fmt.Fprintln(out, "\t\"context\"")
	fmt.Fprintln(out, "\t\"encoding/json\"")
	fmt.Fprint(out, ")\n\n")

	// Generate definition data type structs
	definitionKeys := getAlphabeticSortedKeys(schema.Components["schemas"])
  // To be added into the argument validation init function
  argumentInits := []string{}
	for _, defName := range definitionKeys {
		def := schema.Components["schemas"][defName]
		if def.Description != "" {
			fmt.Fprintln(out, formatComment("", def.Description, nil))
		}
		switch def.Type {
		case "object":
			fmt.Fprintln(out, "type", defName, "struct {")
			// types
			propertiesKeys := getOriginalOrderSortedKeys(def.Properties)
			for _, key := range propertiesKeys {
				propertyName := key.Key
				property := def.Properties[key]
				if propertyName[0:1] == "q" {
					propertyName = propertyName[1:]
				}
				if property.Description != "" {
					fmt.Fprintln(out, formatComment("\t", property.Description, nil))
				}

				if isNonZero(property.Default) && !hasEnumRef(property) {
					fmt.Fprintln(out, "\t// When set to nil the default value is used, when set to point at a value that value is used (including golang zero values)")
					fmt.Fprint(out, "\t", toPublicMemberName(propertyName), " *", getTypeName(property), " `json:\"q", propertyName, ",omitempty\"`")
        } else if hasEnumRef(property) {
					fmt.Fprint(out, "\t", toPublicMemberName(propertyName), " ", refToName(property.Ref), " `json:\"q", propertyName, ",omitempty\"`")
				} else {
					fmt.Fprint(out, "\t", toPublicMemberName(propertyName), " ", getTypeName(property), " `json:\"q", propertyName, ",omitempty\"`")
				}

				fmt.Fprintln(out, "")
			}
			fmt.Fprintln(out, "}")
			fmt.Fprintln(out, "")
		case "array":
			fmt.Fprintln(out, "type", defName, getTypeName(def))
			fmt.Fprintln(out, "")
		case "string":
      fmt.Fprintln(out, "type", defName, "string")
      fmt.Fprintln(out, "")
      printEnumMethods(out, defName)
      argumentInits = append(argumentInits, generateArgumentInitForType(defName, def))
			// Enums are strings now
		default:
			fmt.Fprintln(out, "<<<other>>>", defName, def.Type)
			fmt.Fprintln(out, "")
		}
	}

	// Generate structs for the remote objects (service APIs)
	serviceKeys := getSortedServiceKeys(schema.Services)
	for _, serviceName := range serviceKeys {
		service := schema.Services[serviceName]
		if service.Description != "" {
			fmt.Fprintln(out, formatComment("", service.Description, nil))
		}
		var serviceImplName = serviceName
		fmt.Fprintln(out, "type ", serviceImplName, "struct {")
		fmt.Fprintln(out, "\t*RemoteObject")
		fmt.Fprintln(out, "}")

		methodKeys := getSortedMethodKeys(service.Methods)
		for _, methodName := range methodKeys {
			method := service.Methods[methodName]
			// Generate typed methods
			printMethod(method, out, serviceName, methodName, objectFuncToObject)

			// Generate untyped (raw) methods for those with complex parameters and/or return values
			actualResponses := filterResponses(method.Responses, methodName)
			if !(allTypesBuiltIn(method.Parameters) && allTypesBuiltIn(actualResponses)) {
				printRawMethod(method, out, serviceName, methodName, objectFuncToObject)
			}
		}
	}

  // Generate argument validation init function
  generateArgumentInitFunc(out, argumentInits)
}

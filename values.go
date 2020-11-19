package gqlcost

// Functions in this file are copied from
//   github.com/graphql-go/graphql@v0.7.9/values.go
// Keep in sync those.

import (
	"math"
	"reflect"

	"github.com/graphql-go/graphql"
	"github.com/graphql-go/graphql/language/ast"
)

func isNullish(src interface{}) bool {
	if src == nil {
		return true
	}
	value := reflect.ValueOf(src)
	if value.Kind() == reflect.Ptr {
		if value.IsNil() {
			return true
		}
		value = value.Elem()
	}
	switch value.Kind() {
	case reflect.String:
		// if src is ptr type and len(string)=0, it returns false
		if !value.IsValid() {
			return true
		}
	case reflect.Int:
		return value.Int() == 0
	case reflect.Float32, reflect.Float64:
		return math.IsNaN(value.Float())
	}
	return false
}

func valueFromAST(valueAST ast.Value, ttype graphql.Input, variables map[string]interface{}) interface{} {
	if valueAST == nil {
		return nil
	}
	// precedence: value > type
	if valueAST, ok := valueAST.(*ast.Variable); ok {
		if valueAST.Name == nil || variables == nil {
			return nil
		}
		// Note: we're not doing any checking that this variable is correct. We're
		// assuming that this query has been validated and the variable usage here
		// is of the correct type.
		return variables[valueAST.Name.Value]
	}
	switch ttype := ttype.(type) {
	case *graphql.NonNull:
		return valueFromAST(valueAST, ttype.OfType, variables)
	case *graphql.List:
		values := []interface{}{}
		if valueAST, ok := valueAST.(*ast.ListValue); ok {
			for _, itemAST := range valueAST.Values {
				values = append(values, valueFromAST(itemAST, ttype.OfType, variables))
			}
			return values
		}
		return append(values, valueFromAST(valueAST, ttype.OfType, variables))
	case *graphql.InputObject:
		var (
			ok bool
			ov *ast.ObjectValue
			of *ast.ObjectField
		)
		if ov, ok = valueAST.(*ast.ObjectValue); !ok {
			return nil
		}
		fieldASTs := map[string]*ast.ObjectField{}
		for _, of = range ov.Fields {
			if of == nil || of.Name == nil {
				continue
			}
			fieldASTs[of.Name.Value] = of
		}
		obj := map[string]interface{}{}
		for name, field := range ttype.Fields() {
			var value interface{}
			if of, ok = fieldASTs[name]; ok {
				value = valueFromAST(of.Value, field.Type, variables)
			} else {
				value = field.DefaultValue
			}
			if !isNullish(value) {
				obj[name] = value
			}
		}
		return obj
	case *graphql.Scalar:
		return ttype.ParseLiteral(valueAST)
	case *graphql.Enum:
		return ttype.ParseLiteral(valueAST)
	}

	return nil
}

// Prepares an object map of argument values given a list of argument
// definitions and list of argument AST nodes.
func getArgumentValues(
	argDefs []*graphql.Argument, argASTs []*ast.Argument,
	variableValues map[string]interface{}) map[string]interface{} {

	argASTMap := map[string]*ast.Argument{}
	for _, argAST := range argASTs {
		if argAST.Name != nil {
			argASTMap[argAST.Name.Value] = argAST
		}
	}
	results := map[string]interface{}{}
	for _, argDef := range argDefs {
		var (
			tmp   interface{}
			value ast.Value
		)
		if tmpValue, ok := argASTMap[argDef.PrivateName]; ok {
			value = tmpValue.Value
		}
		if tmp = valueFromAST(value, argDef.Type, variableValues); isNullish(tmp) {
			tmp = argDef.DefaultValue
		}
		if !isNullish(tmp) {
			results[argDef.PrivateName] = tmp
		}
	}
	//log.Printf("getArgumentValues()=%+v", results)
	return results
}

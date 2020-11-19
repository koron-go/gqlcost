package gqlcost

import (
	"github.com/graphql-go/graphql/language/ast"
)

// Cost provides each cost value for type.field
type Cost struct {
	UseMultipliers bool     `json:"useMultipliers,omitempty"`
	Complexity     int      `json:"complexity,omitempty"`
	Multipliers    []string `json:"multipliers,omitempty"`
}

// FieldsCost provides costs for each fields.
type FieldsCost map[string]Cost

// TypeCost provides costs for a type and its fields.
type TypeCost struct {
	// Cost is cost of type itself
	Cost *Cost
	// Fields is costs for each fields.
	Fields FieldsCost
}

// CostMap provides costs for type and fields.
type CostMap map[string]TypeCost

func (m CostMap) getCost(contextTypeName string, fieldNode *ast.Field, fieldTypeName string) *Cost {
	if fieldNode == nil || fieldNode.Name == nil {
		return nil
	}
	if typeCost, ok := m[contextTypeName]; ok {
		if fieldCost, ok := typeCost.Fields[fieldNode.Name.Value]; ok {
			return &fieldCost
		}
	}
	if typeCost, ok := m[fieldTypeName]; ok {
		return typeCost.Cost
	}
	return nil
}

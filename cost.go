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

func (m CostMap) getCost(parentTyp string, node *ast.Field) *Cost {
	if node == nil || node.Name == nil {
		return nil
	}
	fieldName := node.Name.Value
	typeCost, ok := m[parentTyp]
	if !ok {
		return nil
	}
	c, ok := typeCost.Fields[fieldName]
	if !ok {
		return typeCost.Cost
	}
	return &c
}

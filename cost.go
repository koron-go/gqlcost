package gqlcost

import (
	"github.com/graphql-go/graphql/language/ast"
)

// Cost provides each cost value for type.field
type Cost struct {
	// UseMultipliers is flag to use multiplier.
	// Multipliers and MultiplierFunc are refered only when this is true.
	UseMultipliers bool `json:"useMultipliers,omitempty"`

	// Complexity define default complexity of field or type.
	Complexity int `json:"complexity,omitempty"`

	// Multipliers enumerates name of arguments to be used to calculate
	// multiplier.
	Multipliers []string `json:"multipliers,omitempty"`

	// MultiplierFunc is for customizing multiplier calculation.
	// When it available Multipliers is ignored.
	MultiplierFunc func(map[string]interface{}) int
}

func (c Cost) getMultiplier(args map[string]interface{}) int {
	if c.MultiplierFunc != nil {
		return c.MultiplierFunc(args)
	}
	var mul int
	for _, n := range c.Multipliers {
		v, ok := args[n]
		if !ok {
			continue
		}
		if n, ok := toNumber(v); ok {
			mul += n
		}
	}
	return mul
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

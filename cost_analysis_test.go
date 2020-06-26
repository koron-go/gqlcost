package gqlcost

import (
	"testing"

	"github.com/graphql-go/graphql"
	"github.com/graphql-go/graphql/language/ast"
	"github.com/graphql-go/graphql/language/parser"
	"github.com/graphql-go/graphql/language/visitor"
)

var (
	schema   *graphql.Schema
	typeInfo *graphql.TypeInfo
)

type valueTyp struct {
	String string
	Int    int
}

func init() {
	limitArgs := graphql.FieldConfigArgument{
		"limit": &graphql.ArgumentConfig{
			Type: graphql.Int,
		},
	}

	basicInterfaceType := graphql.NewInterface(graphql.InterfaceConfig{
		Name: "BasicInterface",
		Fields: graphql.Fields{
			"string": &graphql.Field{Type: graphql.String},
			"int":    &graphql.Field{Type: graphql.Int},
		},
		ResolveType: func(_ graphql.ResolveTypeParams) *graphql.Object {
			return nil
		},
	})

	secondType := graphql.NewObject(graphql.ObjectConfig{
		Name:       "Second",
		Interfaces: []*graphql.Interface{basicInterfaceType},
		Fields: graphql.Fields{
			"string": &graphql.Field{
				Type: graphql.String,
			},
			"int": &graphql.Field{
				Type: graphql.Int,
			},
			"third": &graphql.Field{
				Type: graphql.String,
				Args: limitArgs,
			},
		},
	})

	firstType := graphql.NewObject(graphql.ObjectConfig{
		Name:       "First",
		Interfaces: []*graphql.Interface{basicInterfaceType},
		Fields: graphql.Fields{
			"string": &graphql.Field{
				Type: graphql.String,
			},
			"int": &graphql.Field{
				Type: graphql.Int,
			},
			"second": &graphql.Field{
				Type: secondType,
				Args: limitArgs,
			},
			"anotherSecond": &graphql.Field{
				Type: secondType,
				Args: limitArgs,
			},
			"basicInterface": &graphql.Field{
				Type: basicInterfaceType,
				Args: limitArgs,
			},
		},
	})

	firstOrSecondType := graphql.NewUnion(graphql.UnionConfig{
		Name:  "FirstOrSecond",
		Types: []*graphql.Object{firstType, secondType},
		ResolveType: func(_ graphql.ResolveTypeParams) *graphql.Object {
			return firstType
		},
	})

	firstType.AddFieldConfig("firstOrSecond", &graphql.Field{
		Type: firstOrSecondType,
		Args: limitArgs,
	})

	typeCostType := graphql.NewObject(graphql.ObjectConfig{
		Name: "TypeCost",
		Fields: graphql.Fields{
			"string": &graphql.Field{
				Type: graphql.String,
			},
			"int": &graphql.Field{
				Type: graphql.Int,
			},
		},
	})

	queryType := graphql.NewObject(graphql.ObjectConfig{
		Name: "Query",
		Fields: graphql.Fields{
			"defaultCost": &graphql.Field{
				Type: graphql.Int,
				Resolve: func(_ graphql.ResolveParams) (interface{}, error) {
					return 1, nil
				},
			},
			"customCost": &graphql.Field{
				Type: graphql.Int,
				Resolve: func(_ graphql.ResolveParams) (interface{}, error) {
					return 2, nil
				},
			},
			"first": &graphql.Field{
				Type: firstType,
				Args: limitArgs,
				Resolve: func(_ graphql.ResolveParams) (interface{}, error) {
					return &valueTyp{String: "first", Int: 1}, nil
				},
			},
			"customCostWithResolver": &graphql.Field{
				Type: graphql.Int,
				Args: limitArgs,
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					n := p.Args["limit"].(int)
					return n, nil
				},
			},
			"overrideTypeCost": &graphql.Field{
				Type: typeCostType,
				Resolve: func(_ graphql.ResolveParams) (interface{}, error) {
					return nil, nil
				},
			},
			"getCostByType": &graphql.Field{
				Type: typeCostType,
				Resolve: func(_ graphql.ResolveParams) (interface{}, error) {
					return nil, nil
				},
			},
			"badComplexityArgument": &graphql.Field{
				Type: graphql.Int,
			},
			"severalMultipliers": &graphql.Field{
				Type: graphql.Int,
				Args: graphql.FieldConfigArgument{
					"first": &graphql.ArgumentConfig{
						Type: graphql.Int,
					},
					"last": &graphql.ArgumentConfig{
						Type: graphql.Int,
					},
					"list": &graphql.ArgumentConfig{
						Type: graphql.NewList(graphql.String),
					},
				},
				Resolve: func(_ graphql.ResolveParams) (interface{}, error) {
					return nil, nil
				},
			},
		},
	})

	sch, err := graphql.NewSchema(graphql.SchemaConfig{Query: queryType})
	if err != nil {
		panic(err)
	}
	schema = &sch
	typeInfo = graphql.NewTypeInfo(&graphql.TypeInfoConfig{Schema: schema})
}

func parseQuery(t *testing.T, q string) *ast.Document {
	t.Helper()
	astDoc, err := parser.Parse(parser.ParseParams{Source: q})
	if err != nil {
		t.Fatalf("parse failed: %s", err)
	}
	return astDoc
}

func testCost(t *testing.T, query string, opts AnalysisOptions, expectedCost int) *costAnalysis {
	t.Helper()
	astDoc := parseQuery(t, query)
	ctx := graphql.NewValidationContext(schema, astDoc, typeInfo)
	ca := newCostAnalysis(ctx, opts)
	visitor.Visit(astDoc, ca.visitorOptions(), nil)
	if errs := ca.ctx.Errors(); len(errs) > 0 {
		t.Errorf("detect errors #1: %s", errs[0])
	}
	if ca.cost != expectedCost {
		t.Fatalf("wrong expectedCost: want=%d got=%d", expectedCost, ca.cost)
	}
	return ca
}

func testErrs(t *testing.T, query string, opts AnalysisOptions, expErrs ...string) *costAnalysis {
	t.Helper()
	astDoc := parseQuery(t, query)
	ctx := graphql.NewValidationContext(schema, astDoc, typeInfo)
	ca := newCostAnalysis(ctx, opts)
	visitor.Visit(astDoc, ca.visitorOptions(), nil)
	errs := ca.ctx.Errors()
	if len(errs) != len(expErrs) {
		t.Fatalf("no errors expected: want=%d got=%d", len(expErrs), len(errs))
	}
	for i, err := range errs {
		act, exp := err.Error(), expErrs[i]
		if act != exp {
			t.Fatalf("%d error mismatch:\nwant=%s\ngot=%s", i, exp, act)
		}
	}
	return ca
}

func TestSimple(t *testing.T) {
	testCost(t, `query { defaultCost }`, AnalysisOptions{
		MaximumCost: 100,
	}, 0)
}

func TestDefaultCost(t *testing.T) {
	testCost(t, `query { defaultCost }`, AnalysisOptions{
		MaximumCost: 100,
		DefaultCost: 12,
	}, 12)
}

func TestCustomCost(t *testing.T) {
	testCost(t, `query { customCost }`, AnalysisOptions{
		MaximumCost: 100,
		CostMap: CostMap{
			"Query": {Fields: FieldsCost{"customCost": {Complexity: 8}}},
		},
	}, 8)
}

func limitCost(complexity int) Cost {
	return Cost{
		UseMultipliers: true,
		Complexity:     complexity,
		Multipliers:    []string{"limit"},
	}
}

func TestRecursive(t *testing.T) {
	testCost(t, `
		query{
			first(limit: 10) {
				second(limit: 10) {
					third(limit: 10)
				}
			}
		}`,
		AnalysisOptions{
			MaximumCost: 10000,
			CostMap: CostMap{
				"Query":  {Fields: FieldsCost{"first": limitCost(2)}},
				"First":  {Fields: FieldsCost{"second": limitCost(5)}},
				"Second": {Fields: FieldsCost{"third": limitCost(6)}},
			},
		}, 6520)
}

func TestRecursive_WithEmptyMultipliers(t *testing.T) {
	testCost(t, `
		query{
			first(limit: 10) {
				second(limit: 10) {
					third(limit: 10)
				}
			}
			customCost
		}`,
		AnalysisOptions{
			MaximumCost: 10000,
			CostMap: CostMap{
				"Query": {Fields: FieldsCost{
					"first":      limitCost(2),
					"customCost": {Complexity: 8},
				}},
				"First":  {Fields: FieldsCost{"second": limitCost(5)}},
				"Second": {Fields: FieldsCost{"third": limitCost(6)}},
			},
		}, 6528)
}

func TestFragment_OnInterface(t *testing.T) {
	testCost(t, `
		query{
			first(limit: 10) {
				basicInterface(limit: 10) {
					string
					...firstFields
					...secondFields
				}
			}
		}
		fragment firstFields on First {
			second(limit: 10)
		}
		fragment secondFields on Second {
			third(limit: 10)
		}`,
		AnalysisOptions{
			MaximumCost: 10000,
			CostMap: CostMap{
				"BasicInterface": {Fields: FieldsCost{"string": {Complexity: 8}}},
				"Query":          {Fields: FieldsCost{"first": limitCost(2)}},
				"First": {Fields: FieldsCost{
					"second":         limitCost(5),
					"basicInterface": limitCost(3),
				}},
				"Second": {Fields: FieldsCost{"third": limitCost(6)}},
			},
		}, 6328)
}

func TestFragment_OnUnion(t *testing.T) {
	testCost(t, `
		query{
			first(limit: 10) {
				firstOrSecond(limit: 10) {
					...firstFields
					...secondFields
				}
			}
		}
		fragment firstFields on First {
			second(limit: 10)
		}
		fragment secondFields on Second {
			third(limit: 10)
		}`,
		AnalysisOptions{
			MaximumCost: 10000,
			CostMap: CostMap{
				"Query": {Fields: FieldsCost{"first": limitCost(2)}},
				"First": {Fields: FieldsCost{
					"firstOrSecond": limitCost(3),
					"second":        limitCost(5),
				}},
				"Second": {Fields: FieldsCost{"third": limitCost(6)}},
			},
		}, 6320)
}

func TestInlineFragment(t *testing.T) {
	testCost(t, `
		query{
			first(limit: 10) {
				firstOrSecond(limit: 10) {
					... on First {
						second(limit: 10)
					}
					... on Second {
						third(limit: 10)
					}
				}
			}
		}`,
		AnalysisOptions{
			MaximumCost: 10000,
			CostMap: CostMap{
				"Query": {Fields: FieldsCost{"first": limitCost(2)}},
				"First": {Fields: FieldsCost{
					"firstOrSecond": limitCost(3),
					"second":        limitCost(5),
				}},
				"Second": {Fields: FieldsCost{"third": limitCost(6)}},
			},
		}, 6320)
}

func TestCostExceed(t *testing.T) {
	testErrs(t, `query { customCost }`, AnalysisOptions{
		MaximumCost: 1,
		CostMap: CostMap{
			"Query": {Fields: FieldsCost{"customCost": {Complexity: 8}}},
		},
	}, `The query exceeds the maximum cost of 1. Actual cost is 8`)
}

func TestNotNegative(t *testing.T) {
	testCost(t, `query { customCostWithResolver(limit: -10) }`,
		AnalysisOptions{
			MaximumCost: 100,
			CostMap: CostMap{
				"Query": {Fields: FieldsCost{
					"customCostWithResolver": limitCost(4),
				}},
			},
		}, 0)
}

func TestFieldOverrideTypeCost(t *testing.T) {
	testCost(t, `query { overrideTypeCost }`, AnalysisOptions{
		MaximumCost: 100,
		CostMap: CostMap{
			"Query": {
				Cost:   &Cost{Complexity: 3},
				Fields: FieldsCost{"overrideTypeCost": {Complexity: 2}},
			},
		},
	}, 2)
}

func TestCostType(t *testing.T) {
	testCost(t, `query { getCostByType }`, AnalysisOptions{
		MaximumCost: 100,
		CostMap: CostMap{
			"Query": {Cost: &Cost{Complexity: 3}},
		},
	}, 3)
}

func TestDefaultCost_EmptyCostMap(t *testing.T) {
	testCost(t, `query { first(limit: 10) }`, AnalysisOptions{
		MaximumCost: 100,
		DefaultCost: 34,
		CostMap:     CostMap{},
	}, 34)
}

func TestComplexityRange(t *testing.T) {
	ca := testErrs(t, `query { badComplexityArgument }`, AnalysisOptions{
		MaximumCost: 1000,
		DefaultCost: 2,
		CostMap: CostMap{
			"Query": {Fields: FieldsCost{
				"badComplexityArgument": {Complexity: 12},
			}},
		},
		ComplexityRange: ComplexityRange{Min: 1, Max: 3},
	}, "The complexity argument must be between 1 and 3")
	if ca.cost != 2 {
		t.Fatalf("wrong expectedCost: want=%d got=%d", 2, ca.cost)
	}
}

func TestComplexityRange_Invaild(t *testing.T) {
	testErrs(t, `query { badComplexityArgument }`, AnalysisOptions{
		MaximumCost: 1000,
		CostMap: CostMap{
			"Query": {Fields: FieldsCost{
				"badComplexityArgument": {Complexity: 12},
			}},
		},
		ComplexityRange: ComplexityRange{Min: 100, Max: 1},
	},
		"Invalid minimum and maximum complexity",
		"The complexity argument must be between 100 and 1")
}

func TestSeveralMultipliers(t *testing.T) {
	testCost(t, `query { severalMultipliers(first: 10, last: 4) }`,
		AnalysisOptions{
			MaximumCost: 1000,
			CostMap: CostMap{"Query": {Fields: FieldsCost{
				"severalMultipliers": {
					Multipliers:    []string{"coucou", "first", "last", "list"},
					UseMultipliers: true,
					Complexity:     4,
				},
			}}},
		}, 56)
}

func TestSeveralMultipliers_WithList(t *testing.T) {
	testCost(t, `
		query {
			severalMultipliers(first: 10, last: 4, list: ["this", "is", "a", "test"])
		}`,
		AnalysisOptions{
			MaximumCost: 1000,
			CostMap: CostMap{"Query": {Fields: FieldsCost{
				"severalMultipliers": {
					Multipliers:    []string{"coucou", "first", "last", "list"},
					UseMultipliers: true,
					Complexity:     4,
				},
			}}},
		}, 72)
}

func TestMultipleRecusive(t *testing.T) {
	testCost(t, `
		query{
			first(limit: 10) {
				second(limit: 10) {
					int
				}
				anotherSecond(limit: 10) {
					int
				}
			}
		}`,
		AnalysisOptions{
			MaximumCost: 10000,
			CostMap: CostMap{
				"Query": {Fields: FieldsCost{"first": limitCost(2)}},
				"First": {Fields: FieldsCost{
					"second":        limitCost(5),
					"anotherSecond": limitCost(5),
				}},
			},
		}, 1020)
}

// TODO: add tests

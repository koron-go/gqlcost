# koron-go/gqlcost

[![GoDoc](https://godoc.org/github.com/koron-go/gqlcost?status.svg)](https://godoc.org/github.com/koron-go/gqlcost)
[![Actions/Go](https://github.com/koron-go/gqlcost/workflows/Go/badge.svg)](https://github.com/koron-go/gqlcost/actions?query=workflow%3AGo)
[![Go Report Card](https://goreportcard.com/badge/github.com/koron-go/gqlcost)](https://goreportcard.com/report/github.com/koron-go/gqlcost)

<strong>gqlcost</strong> provides cost analysis validation rule for
[graphql-go/graphql][graphql-go]. This is a port of
[pa-bru/graphql-cost-analysis][graphql-cost-analysis].

<strong>gqlcost</strong> supports only cost map, and not support `@cost`
directive for now. Because graphql-go/graphql's `graphql.Schema`, `Object` and
`Field` don't have spaces to store values for directives.

## Getting started

Put this codes to your project.

It adds a cost analysis validation rule to `graphql.SpecifiedRules` Currently,
graphql-go/graphql doesn't have any methods to customize validation rules
exclude modifing `graphql.SpecifiedRules`.

```go
import "github.com/koron-go/gqlcost"

func init() {
    gqlcost.AddCostAnalysisRule(gqlcost.AnalysisOptions{
        // FIXME: modify maximum cost as you need.
        MaximumCost: 1000,
        // CostMap defines cost for types and fields.
        CostMap: gqlcost.CostMap{
            // FIXME: modify CostMap as you need.
            "Query": gqlcost.TypeCost{
                // cost for "Query" type.
                Cost: &gqlcost.Cost{Complexity: 1},
                // cost for fields of "Query" type.
                Fields: gqlcost.FieldsCost{
                    "todoList": gqlcost.Cost{
                        UseMultipliers: true,
                        Complexity:     2,
                        Multipliers:    []string{"limit"},
                    },
                    // TODO: add other fields.
                },
            },
            // TODO: add other costs for types by using `TypeCost`
        },
    })
}
```

[graphql-go]:https://github.com/graphql-go/graphql
[graphql-cost-analysis]:https://github.com/pa-bru/graphql-cost-analysis

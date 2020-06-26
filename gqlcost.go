package gqlcost

import (
	"sync"

	"github.com/graphql-go/graphql"
)

// AnalysisOptions provides options for cost analysis.
type AnalysisOptions struct {
	MaximumCost int
	DefaultCost int
	Valiables   map[string]interface{}

	CostMap         CostMap
	ComplexityRange ComplexityRange
}

var addRule sync.Once

// AddCostAnalysisRule adds a rule of cost analysis to
// graphql.SpecifiedRules.
func AddCostAnalysisRule(opts AnalysisOptions) {
	addRule.Do(func() {
		r := AnalysisRule(opts)
		graphql.SpecifiedRules = append(graphql.SpecifiedRules, r)
	})
}

// AnalysisRule provides cost analysis rule (function)
func AnalysisRule(opts AnalysisOptions) graphql.ValidationRuleFn {
	r := &costAnalysisRule{
		opts: opts,
	}
	return r.validationRule
}

type costAnalysisRule struct {
	opts AnalysisOptions
}

func (r *costAnalysisRule) validationRule(context *graphql.ValidationContext) *graphql.ValidationRuleInstance {
	ca := newCostAnalysis(context, r.opts)
	return &graphql.ValidationRuleInstance{VisitorOpts: ca.visitorOptions()}
}

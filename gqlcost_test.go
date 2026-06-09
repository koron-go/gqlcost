package gqlcost

import (
	"testing"

	"github.com/graphql-go/graphql"
)

func TestAnalysisRule(t *testing.T) {
	opts := AnalysisOptions{MaximumCost: 100}
	ruleFn := AnalysisRule(opts)
	if ruleFn == nil {
		t.Fatal("AnalysisRule returned nil")
	}
	astDoc := parseQuery(t, `query { defaultCost }`)
	ctx := graphql.NewValidationContext(schema, astDoc, typeInfo)
	inst := ruleFn(ctx)
	if inst == nil {
		t.Fatal("validationRule returned nil")
	}
	if inst.VisitorOpts == nil {
		t.Fatal("VisitorOpts is nil")
	}
}

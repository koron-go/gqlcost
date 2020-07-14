package gqlcost

import (
	"fmt"

	"github.com/graphql-go/graphql"
	"github.com/graphql-go/graphql/gqlerrors"
	"github.com/graphql-go/graphql/language/ast"
	"github.com/graphql-go/graphql/language/kinds"
	"github.com/graphql-go/graphql/language/visitor"
)

type costAnalysis struct {
	opts AnalysisOptions
	ctx  *graphql.ValidationContext
	cost int

	defaultComplexity int
}

func newCostAnalysis(ctx *graphql.ValidationContext, opts AnalysisOptions) *costAnalysis {
	ca := &costAnalysis{
		opts:              opts,
		ctx:               ctx,
		defaultComplexity: 1,
	}
	cr := ca.opts.ComplexityRange
	if cr.Min != 0 && cr.Max != 0 && cr.Min > cr.Max {
		ca.reportError("Invalid minimum and maximum complexity", nil)
	}
	if ca.opts.ComplexityRange.Min != 0 {
		ca.defaultComplexity = ca.opts.ComplexityRange.Min
	}
	return ca
}

func (ca *costAnalysis) visitorOptions() *visitor.VisitorOptions {
	return &visitor.VisitorOptions{
		KindFuncMap: map[string]visitor.NamedVisitFuncs{
			kinds.OperationDefinition: {
				Enter: ca.opDefEnter,
				Leave: ca.opDefLeave,
			},
		},
	}
}

func (ca *costAnalysis) opDefEnter(p visitor.VisitFuncParams) (string, interface{}) {
	od, ok := p.Node.(*ast.OperationDefinition)
	if !ok {
		return visitor.ActionSkip, nil
	}
	switch od.GetOperation() {
	case "query":
		if op := ca.ctx.Schema().QueryType(); op != nil {
			ca.cost += ca.computeNodeCost(od, op, nil)
		}
		return visitor.ActionNoChange, nil
	case "mutation":
		if op := ca.ctx.Schema().MutationType(); op != nil {
			ca.cost += ca.computeNodeCost(od, op, nil)
		}
		return visitor.ActionNoChange, nil
	case "subscription":
		if op := ca.ctx.Schema().SubscriptionType(); op != nil {
			ca.cost += ca.computeNodeCost(od, op, nil)
		}
		return visitor.ActionNoChange, nil
	default:
		return visitor.ActionSkip, nil
	}
}

func (ca *costAnalysis) opDefLeave(p visitor.VisitFuncParams) (string, interface{}) {
	od, ok := p.Node.(*ast.OperationDefinition)
	if !ok {
		return visitor.ActionSkip, nil
	}
	if ca.cost > ca.opts.MaximumCost {
		ca.reportError(fmt.Sprintf("The query exceeds the maximum cost of %d. Actual cost is %d", ca.opts.MaximumCost, ca.cost), []ast.Node{od})
	}
	return visitor.ActionNoChange, nil
}

func (ca *costAnalysis) reportError(msg string, nodes []ast.Node) {
	ca.ctx.ReportError(gqlerrors.NewError(msg, nodes, "", nil, []int{}, nil))
}

func (ca *costAnalysis) getSectionSet(node ast.Node) (*ast.SelectionSet, bool) {
	sel, ok := node.(ast.Selection)
	if !ok {
		return nil, false
	}
	selectionSet := sel.GetSelectionSet()
	if selectionSet == nil {
		return nil, false
	}
	return selectionSet, true
}

func (ca *costAnalysis) getFieldDefinitionMap(typDef interface{}) graphql.FieldDefinitionMap {
	if x, ok := typDef.(interface {
		Fields() graphql.FieldDefinitionMap
	}); ok {
		return x.Fields()
	}
	return graphql.FieldDefinitionMap{}
}

func (ca *costAnalysis) computeNodeCost(node ast.Node, typDef interface{}, parentMultipliers []int) int {
	selectionSet, ok := ca.getSectionSet(node)
	if !ok {
		return 0
	}

	fm := ca.getFieldDefinitionMap(typDef)

	var (
		total         int
		fragmentCosts []int
	)

	for _, iSelection := range selectionSet.Selections {
		nodeCost := ca.opts.DefaultCost
		switch childNode := iSelection.(type) {

		case *ast.Field:
			if childNode.Name == nil {
				break
			}
			field, ok := fm[childNode.Name.Value]
			if !ok {
				break
			}

			// NOTE: graphql-go/graphql doesn't support directives in
			// schema. So this package supports only used defined CostMap.
			if len(ca.opts.CostMap) == 0 {
				nodeCost += ca.computeNodeCost(childNode, field.Type, parentMultipliers)
				break
			}

			var costMapArgs nodeCostConfig
			if n, ok := typName(typDef); ok {
				fieldArgs := getArgumentValues(field.Args, childNode.Arguments, ca.opts.Valiables)
				costMapArgs = ca.getArgsFromCostMap(childNode, n, fieldArgs)
			}
			multipliers := copyInts(parentMultipliers)
			nodeCost, multipliers = ca.computeCost(costMapArgs, multipliers)
			nodeCost += ca.computeNodeCost(childNode, field.Type, multipliers)

		case *ast.FragmentSpread:
			fragName := ""
			if childNode.Name != nil {
				fragName = childNode.Name.Value
			}
			fr := ca.ctx.Fragment(fragName)
			if fr == nil || fr.TypeCondition == nil || fr.TypeCondition.Name == nil {
				fragmentCosts = append(fragmentCosts, ca.opts.DefaultCost)
				nodeCost = 0
				break
			}
			fragType := ca.ctx.Schema().Type(fr.TypeCondition.Name.Value)
			fragCost := ca.computeNodeCost(fr, fragType, parentMultipliers)
			fragmentCosts = append(fragmentCosts, fragCost)
			nodeCost = 0

		case *ast.InlineFragment:
			if childNode == nil {
				fragmentCosts = append(fragmentCosts, ca.opts.DefaultCost)
				nodeCost = 0
				break
			}
			if childNode.TypeCondition == nil || childNode.TypeCondition.Name == nil {
				fragCost := ca.computeNodeCost(childNode, typDef, parentMultipliers)
				fragmentCosts = append(fragmentCosts, fragCost)
				nodeCost = 0
				break
			}
			fragType := ca.ctx.Schema().Type(childNode.TypeCondition.Name.Value)
			fragCost := ca.computeNodeCost(childNode, fragType, parentMultipliers)
			fragmentCosts = append(fragmentCosts, fragCost)
			nodeCost = 0

		default:
			if n, ok := childNode.(ast.Node); ok {
				nodeCost = ca.computeNodeCost(n, typDef, nil)
			}
		}
		if nodeCost > 0 {
			total += nodeCost
		}
	}

	return total + maxCost(fragmentCosts)
}

type nodeCostConfig struct {
	useMultipliers bool
	complexity     int
	multipliers    []int
}

func (ca *costAnalysis) getArgsFromCostMap(node *ast.Field, parentTyp string, fieldArgs map[string]interface{}) (ncc nodeCostConfig) {
	cost := ca.opts.CostMap.getCost(parentTyp, node)
	if cost == nil {
		return nodeCostConfig{}
	}
	return nodeCostConfig{
		useMultipliers: cost.UseMultipliers,
		complexity:     cost.Complexity,
		multipliers:    ca.getMultipliersFromString(cost.Multipliers, fieldArgs),
	}
}

func (ca *costAnalysis) computeCost(ncc nodeCostConfig, parentMultipliers []int) (int, []int) {
	if ca.opts.ComplexityRange.outside(ncc.complexity) {
		ca.reportError(fmt.Sprintf("The complexity argument must be between %d and %d", ca.opts.ComplexityRange.Min, ca.opts.ComplexityRange.Max), nil)
		return ca.opts.DefaultCost, parentMultipliers
	}

	if !ncc.useMultipliers {
		return ncc.complexity, parentMultipliers
	}

	if len(ncc.multipliers) > 0 {
		mul := 0
		for _, v := range ncc.multipliers {
			mul += v
		}
		parentMultipliers = append(parentMultipliers, mul)
	}

	acc := ncc.complexity
	for _, v := range parentMultipliers {
		acc *= v
	}

	return acc, parentMultipliers
}

func (ca *costAnalysis) getMultipliersFromString(multipliers []string, fieldArgs map[string]interface{}) []int {
	muls := make([]int, 0, len(multipliers))
	for _, n := range multipliers {
		v, ok := fieldArgs[n]
		if !ok {
			continue
		}
		if n, ok := toNumber(v); ok && n != 0 {
			muls = append(muls, n)
			continue
		}
	}
	return muls
}

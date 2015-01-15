package main

import (
	"fmt"
	"math/rand"
	"reflect"
	"time"
)

func indent(n int) string {
	str := ""
	for i := int(0); i < n; i++ {
		str += "  "
	}

	return str
}

// ===== FLOAT VALUE ===============================================================================

func NewRandomFloatValue(min, max float64) FloatValue {
	return FloatValue{min: min, max: max, random: true}
}

func NewFixedFloatValue(val float64) FloatValue {
	return FloatValue{min: val, max: val, random: false}
}

type FloatValue struct {
	min    float64
	max    float64
	random bool
}

func (fv FloatValue) Value() float64 {
	if fv.random {
		return 1.23
	} else {
		return fv.min
	}
}


// ===== GENETIC ALGORITHM =========================================================================

func NewGA() *GeneticAlgorithm {
	return &GeneticAlgorithm{}
}

type GeneticAlgorithm struct {
	values []interface{}
	tree Node
	treeDepth int
}

func (ga *GeneticAlgorithm) GenerateTree(perFunc, perParam float64) {
	// TODO: validate inputs between 0.0 and 1.0
	ga.tree = makeRandomTree(len(ga.values), ga.treeDepth, perFunc, perParam, nil)
	ga.tree.Display(0)
}

func (ga *GeneticAlgorithm) Run() bool {
	return ga.tree.Evaluate(ga.values).(bool)
}

func (ga *GeneticAlgorithm) SetData(values... interface{}) {
	data := []interface{}{}

	for _, val := range values {
		var newVal interface{}

		if ival, ok := val.(int); ok {
			newVal = float64(ival)
		} else {
			newVal = val
		}

		data = append(data, newVal)
	}

	ga.values = data
}

func (ga *GeneticAlgorithm) SetTreeDepth(n int) {
	ga.treeDepth = n
}

// ===== NODE TYPES ================================================================================

type Node interface {
	Display(int)
	Evaluate([]interface{}) interface{}
	NodeCount() int
}

type FWrapper struct {
	function   func([]interface{}) interface{}
	childCount int
	name       string
	accepts    string
}

// ===== BASIC NODE ================================================================================

type BasicNode struct {
	function func([]interface{}) interface{}
	name string
	children []Node
	nodeType string
}

// func (bn BasicNode) IsBool

func (bn BasicNode) Display(indentLevel int) {
	fmt.Printf("%s%s\n", indent(indentLevel), bn.name)

	for _, child := range bn.children {
		child.Display(indentLevel + 1)
	}
}

func (bn BasicNode) Evaluate(inputs []interface{}) interface{} {
	results := []interface{}{}

	for _, child := range bn.children {
		results = append(results, child.Evaluate(inputs))
	}

	return bn.function(results)
}

func (bn BasicNode) NodeCount() int {
	count := int(1)

	for _, child := range bn.children {
		count += child.NodeCount()
	}

	return count
}

// ===== PARAM NODE ================================================================================

type ParamNode struct {
	index int
}

func (pn ParamNode) Display(indentLevel int) {
	fmt.Printf("%sp%d\n", indent(indentLevel), pn.index)
}

func (pn ParamNode) Evaluate(inputs []interface{}) interface{} {
	return inputs[pn.index]
}

func (pn ParamNode) NodeCount() int {
	return 1
}

// ===== CONST NODE ================================================================================ 

type ConstNode struct {
	value float64
}

func (cn ConstNode) Display(indentLevel int) {
	fmt.Printf("%s%f\n", indent(indentLevel), cn.value)
}

func (cn ConstNode) Evaluate(inputs []interface{}) interface{} {
	return cn.value
}

func (cn ConstNode) NodeCount() int {
	return 1
}

// ===== MATH FUNCTIONS ============================================================================

func fAdder(o []interface{}) interface{} {
	if _, ok := o[0].(float64); ok {
		return o[0].(float64) + o[1].(float64)
	} else {
		panic("unsupported type: " + reflect.TypeOf(o[0]).Name())
	}
}

func fSubtractor(o []interface{}) interface{} {
	if _, ok := o[0].(float64); ok {
		return o[0].(float64) - o[1].(float64)
	} else {
		panic("unsupported type: " + reflect.TypeOf(o[0]).Name())
	}
}

func fMultiplier(o []interface{}) interface{} {
	if _, ok := o[0].(float64); ok {
		return o[0].(float64) * o[1].(float64)
	} else {
		panic("unsupported type: " + reflect.TypeOf(o[0]).Name())
	}
}

func fDivider(o []interface{}) interface{} {
	if _, ok := o[0].(float64); ok {
		// TODO: handle o[1] = 0 here
		return o[0].(float64) / o[1].(float64)
	} else {
		panic("unsupported type: " + reflect.TypeOf(o[0]).Name())
	}
}

// ===== BOOLEAN FUNCTIONS =========================================================================

func ifFunc(o []interface{}) interface{} {
	if _, ok := o[0].(bool); ok {
		if o[0].(bool) {
			return o[1]
		} else {
			return o[2]
		}
	} else {
		panic("unsupported type: " + reflect.TypeOf(o[0]).Name())
	}
}

func isGreater(o []interface{}) interface{} {
	if _, ok := o[0].(float64); ok {
		return o[0].(float64) > o[1].(float64)
	} else {
		panic("unsupported type: " + reflect.TypeOf(o[0]).Name())
	}
}

func isGreaterOrEqual(o []interface{}) interface{} {
	if _, ok := o[0].(float64); ok {
		return o[0].(float64) >= o[1].(float64)
	} else {
		panic("unsupported type: " + reflect.TypeOf(o[0]).Name())
	}
}

func isLess(o []interface{}) interface{} {
	if _, ok := o[0].(float64); ok {
		return o[0].(float64) < o[1].(float64)
	} else {
		panic("unsupported type: " + reflect.TypeOf(o[0]).Name())
	}
}

func isLessOrEqual(o []interface{}) interface{} {
	if _, ok := o[0].(float64); ok {
		return o[0].(float64) <= o[1].(float64)
	} else {
		panic("unsupported type: " + reflect.TypeOf(o[0]).Name())
	}
}

// ===== FUNCTION WRAPPERS =========================================================================

func randomFunc(f map[string]) (FWrapper, string) {
	index := rand.Intn(len(f))
	count := 0

	for key, value := range f {
		if count == index {
			return value, key
		} else {
			count++
		}
	}
}

var mathfunctions = map[string]FWrapper{
	"add":  FWrapper{function: fAdder,      childCount: 2, name: "add"},
	"subw": FWrapper{function: fSubtractor, childCount: 2, name: "subtract"},
	"mulw": FWrapper{function: fMultiplier, childCount: 2, name: "multiply"},
	"divw": FWrapper{function: fDivider,    childCount: 2, name: "divide"},
}

var boolfunctions = map[string]FWrapper{
	"isgreater":        FWrapper{function: isGreater,        childCount: 2, name: "isgreater"},
	"isgreaterorequal": FWrapper{function: isGreaterOrEqual, childCount: 2, name: "isgreaterorequal"},
	"isless":           FWrapper{function: isLess,           childCount: 2, name: "isless"},
	"islessorequal":    FWrapper{function: isLessOrEqual,    childCount: 2, name: "islessorequal"},
}

var addw = FWrapper{function: fAdder,      childCount: 2, name: "add"}
var subw = FWrapper{function: fSubtractor, childCount: 2, name: "subtract"}
var mulw = FWrapper{function: fMultiplier, childCount: 2, name: "multiply"}
var divw = FWrapper{function: fDivider,    childCount: 2, name: "divide"}

var iffunc           = FWrapper{function: ifFunc,           childCount: 3, name: "if"}
var isgreater        = FWrapper{function: isGreater,        childCount: 2, name: "isgreater"}
var isgreaterorequal = FWrapper{function: isGreaterOrEqual, childCount: 2, name: "isgreaterorequal"}
var isless           = FWrapper{function: isLess,           childCount: 2, name: "isless"}
var islessorequal    = FWrapper{function: isLessOrEqual,    childCount: 2, name: "islessorequal"}

var flist     = []FWrapper{addw, subw, mulw, divw, iffunc, isgreater, isgreaterorequal, isless, islessorequal}
var boolflist = []FWrapper{iffunc, isgreater, isgreaterorequal, isless, islessorequal}
var mathflist = []FWrapper{addw, subw, mulw, divw}

// func randomFunction() FWrapper {
// 	return flist[rand.Intn(len(flist)-1)]
// }

// func randomBoolFunction() FWrapper {
// 	return boolflist[rand.Intn(len(boolflist)-1)]
// }

// func randomMathFunction() FWrapper {
// 	return mathflist[rand.Intn(len(mathflist)-1)]
// }

func makeRandomTree(numInputs, maxDepth int, funcPer, paramPer float64, parent Node) Node {
	// if there's no parent, create an "if" node which only allows bool children
	if nil == parent {
		f := iffunc

		bn := BasicNode{
			name:     f.name,
			function: f.function,
		}

		// this is just so that the child nodes are also bools
		boolNode := BasicNode{
			nodeType: "bool",
		}

		children := []Node{}

		for i := int(0); i < f.childCount; i++ {
			tree := makeRandomTree(numInputs, maxDepth - 1, funcPer, paramPer, boolNode)
			children = append(children, tree)
		}

		bn.children = children

		return bn
	}

	if maxDepth > 0 {
		if "bool" == parent.nodeType {
			// if the parent node is a bool, that means the parent node is an if bool
			f, 
		} else if "math" == parent.nodeType {
			// if the parent node is a math, valid children are either math nodes or
			// float value nodes (random or param)
		} else {
			panic("unknown node type: " + parent.nodeType)
		}
	}
}

// func makeRandomTree(numInputs, maxDepth int, funcPer, paramPer float64, parent Node) Node {
// 	if rand.Float64() < funcPer && maxDepth > 0 {
// 		f := randomFunction()

// 		bn := BasicNode{
// 			name:     f.name,
// 			function: f.function,
// 		}

// 		children := []Node{}

// 		for i := int(0); i < f.childCount; i++ {
// 			tree := makeRandomTree(numInputs, maxDepth - 1, funcPer, paramPer, bn)
// 			children = append(children, tree)
// 		}

// 		bn.children = children

// 		return bn
// 	} else if rand.Float64() < paramPer {
// 		return ParamNode{index: rand.Intn(numInputs - 1)}
// 	} else {
// 		return ConstNode{value: float64(rand.Intn(10))}
// 	}
// }

// ===== PROGRAM ENTRYPOINT ========================================================================

func main() {
	rand.Seed(time.Now().UTC().UnixNano())

	ga := NewGA()
	ga.SetData(16, false, 1.0, true)
	ga.SetTreeDepth(4)
	ga.GenerateTree(0.5, 0.6)

	fmt.Printf("Nodes: %d\n", ga.tree.NodeCount())
	fmt.Printf("Result: %.2f\n", ga.Run())
}

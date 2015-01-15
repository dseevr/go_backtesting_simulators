package variables

import (
	"fmt"
	"math/rand"
)

// ===== BOOL ======================================================================================

func NewBoolVar() *BoolVar {
	return &BoolVar{}
}

type BoolVar struct {
	value bool
}

func (b *BoolVar) Randomize() {
	b.value = 0 == rand.Intn(1)
}

func (b *BoolVar) Value() bool {
	return b.value
}

// ===== FLOAT =====================================================================================

func NewFloatVar(lower, upper float64) *FloatVar {
	if lower >= upper {
		panic(fmt.Sprintf(
			"lower must be > upper (got lower: %f, upper: %f)\n",
			lower,
			upper,
		))
	}

	return &FloatVar{lower: lower, upper: upper}
}

type FloatVar struct {
	lower float64
	upper float64
	value float64
}

func (f *FloatVar) Randomize() {
	f.value = (rand.Float64() * (f.upper - f.lower)) + f.lower
	fmt.Printf("f: %f\n", f.value)
}

func (f *FloatVar) Value() float64 {
	return f.value
}

// ===== INT =======================================================================================

func NewIntVar(lower, upper int) *IntVar {
	if lower >= upper {
		panic(fmt.Sprintf(
			"lower must be > upper (got lower: %d, upper: %d)\n",
			lower,
			upper,
		))
	}

	return &IntVar{lower: lower, upper: upper}
}

type IntVar struct {
	lower int
	upper int
	value int
}

func (i *IntVar) Randomize() {
	i.value = rand.Intn(i.upper) + i.lower
}

func (i *IntVar) Value() int {
	return i.value
}

// ===== VARIABLES =================================================================================

type Variables struct {
	floats map[string]*FloatVar
	ints   map[string]*IntVar
	bools  map[string]*BoolVar
}

func (v *Variables) Init() {
	fmt.Println("variables.Init()")
	v.floats = make(map[string]*FloatVar)
	v.ints   = make(map[string]*IntVar)
	v.bools  = make(map[string]*BoolVar)
	fmt.Println("end of variables.Init()")
}

func (v *Variables) CreateBool(name string) {
	v.bools[name] = NewBoolVar()
}

func (v *Variables) CreateFloat(name string, lower, upper float64) {
	v.floats[name] = NewFloatVar(lower, upper)
}

func (v *Variables) CreateInt(name string, lower, upper int) {
	v.ints[name] = NewIntVar(lower, upper)
}

func (v *Variables) GetBool(name string) bool {
	b, ok := v.bools[name]
	if !ok {
		panic("unknown bool var: " + name)
	}

	return b.Value()
}

func (v *Variables) GetFloat(name string) float64 {
	f, ok := v.floats[name]
	if !ok {
		panic("unknown float var: " + name)
	}

	return f.Value()
}

func (v *Variables) GetInt(name string) int {
	i, ok := v.ints[name]
	if !ok {
		panic("unknown int var: " + name)
	}

	return i.Value()
}

func (vars *Variables) Randomize() {
	fmt.Println("Randomize()")

	for _, v := range vars.bools {
		v.Randomize()
	}

	for _, v := range vars.ints {
		v.Randomize()
	}

	for _, v := range vars.floats {
		v.Randomize()
	}

	fmt.Println("end of Randomize()")
}

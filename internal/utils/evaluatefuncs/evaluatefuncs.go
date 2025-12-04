package evaluatefuncs

import(
	"maps"
	"github.com/Knetic/govaluate"
)

func GovalFuncs() map[string]govaluate.ExpressionFunction {
	functions := make(map[string]govaluate.ExpressionFunction)

	// Merge all function categories
	mergeFunctions(functions, MathFunctions())
	mergeFunctions(functions, StatisticalFunctions())
	mergeFunctions(functions, StringFunctions())
	mergeFunctions(functions, DateTimeFunctions())
	mergeFunctions(functions, LogicalFunctions())

	return functions
}

// mergeFunctions merges source functions into target map
func mergeFunctions(target, source map[string]govaluate.ExpressionFunction) {
	if target == nil || source == nil {
		return
	}
	maps.Copy(target, source)
}

package restructure

import "regexp/syntax"

type transformer func(expr *syntax.Regexp) ([]*syntax.Regexp, error)

// transform replaces each node in a regex AST with the return value of the given function
// it processes the children of a node before the node itself
func transform(expr *syntax.Regexp, f transformer) (*syntax.Regexp, error) {
	var newchildren []*syntax.Regexp
	for _, child := range expr.Sub {
		newchild, err := transform(child, f)
		if err != nil {
			return nil, err
		}
		replacements, err := f(newchild)
		if err != nil {
			return nil, err
		}
		newchildren = append(newchildren, replacements...)
	}
	expr.Sub = newchildren
	return expr, nil
}

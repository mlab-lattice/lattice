package tree

// The Node interface represents a node in a tree.
type Node interface {
	Path() Path
	Value() interface{}
	Parent() Node
	Children() map[string]Node
}

func Lookup(n Node, p Path) (Node, bool, error) {
	if p.IsRoot() {
		return n, true, nil
	}

	remainder, c, err := p.Shift(1)
	if err != nil {
		return nil, false, err
	}

	name := c[0]
	child, ok := n.Children()[name]
	if !ok {
		return nil, false, nil
	}

	return Lookup(child, remainder)
}

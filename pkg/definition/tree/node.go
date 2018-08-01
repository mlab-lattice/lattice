package tree

// The Node interface represents a node in a tree.
type Node interface {
	Path() Path
	Value() interface{}
	Parent() Node
	Children() map[string]Node
}

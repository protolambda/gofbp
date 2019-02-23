package fbp

type Scope []NodeID

// Simple way of defining a new scope inside this scope.
// return a copy of the scope, with id appended.
func (s Scope) Inner(id NodeID) Scope {
	inner := make([]NodeID, len(s)+1, len(s)+1)
	copy(inner, s)
	inner[len(s)] = id
	return inner
}

// A node has a scope, but this may be temporary (if the node changes its parent).
// Hence the choice to define scope as a stand-alone reference, instead of embedding it fully in the child.
// Also, most child don't need a full reference of their scope anyway,
//  this would only make it easy to create dependency problems.
type ScopedNode struct {
	Scope
	Node
}

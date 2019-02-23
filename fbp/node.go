package fbp

// ID of a node, simply embed it into a struct to give it ID functionality as part of implementing the Node interface.
type NodeID string

func (id NodeID) ID() NodeID {
	return id
}

// The base is a child: every node should be able to be embedded into a parent.
type Child interface {
	// returns the parent node. May be nil if no parent exists.
	Parent() Parent
	// Sets the parent node (a child can only have one parent)
	SetParent(p Parent)
}

// The default child implementation: just keep track of the parent.
type BasicChildImpl struct {
	parent Parent
}

func (n *BasicChildImpl) Parent() Parent {
	return n.parent
}

func (n *BasicChildImpl) SetParent(p Parent) {
	n.parent = p
}

// A node is really just an item with an identity (ID), and possibly a parent.
type Node interface {
	Child
	// returns the ID of the node
	ID() NodeID
}

// The most common node implementation: ID, parent relationship, propagate errors to parent.
type BasicNodeImpl struct {
	NodeID
	BasicChildImpl
}

func (n *BasicNodeImpl) OnError(err error) {
	// no parent? Then it's un-catched, all we can do is panic.
	if n.parent == nil {
		panic(err)
	}
	// propagate error to parent by default
	if ep, ok := n.parent.(ErrorDealer); ok {
		ep.OnError(err)
	} else {
		// parent does deal with errors, no way to propagate. Panic.
		panic(err)
	}
}

// Parents are a special type of node: they can have 1 or more children nodes.
type Parent interface {
	Node
	// Retrieve a child from a parent by its ID. Return nil if the child ID is unknown to the parent.
	GetChild(id NodeID) Node
	// Add child to the parent.
	// Check parent implementation if it auto-removes child from old parent.
	// Note: a node may only have 1 parent node.
	// Adding it to multiple different parents results in problematic iteration and double closing.
	// Adding it twice to the same parent results in double error messages.
	AddChild(node Node)
	// Remove child from the parent
	RemoveChild(id NodeID)
}

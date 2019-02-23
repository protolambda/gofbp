package fbp

// A basic collection node: can have any amount of children. Children are indexed by their ID.
type Graph struct {
	BasicNodeImpl

	nodes map[NodeID]Node
}

func NewGraph(id NodeID) *Graph {
	g := new(Graph)
	g.NodeID = id
	g.nodes = make(map[NodeID]Node)
	return g
}

func (g *Graph) CollectNodes(scope Scope, out *[]ScopedNode) {
	newScope := scope.Inner(g.ID())
	for _, n := range g.nodes {
		*out = append(*out, ScopedNode{newScope, n})
		if ng, ok := n.(*Graph); ok {
			ng.CollectNodes(newScope, out)
		}
	}
}

func (g *Graph) AddChild(node Node) {
	// Remove the node from its previous parent, if it has one
	if p := node.Parent(); p != nil {
		p.RemoveChild(node.ID())
	}
	// add it to this node
	g.nodes[node.ID()] = node
	// make this the parent of the node
	node.SetParent(g)
}

func (g *Graph) RemoveChild(id NodeID) {
	if n, ok := g.nodes[id]; ok {
		n.SetParent(nil)
		delete(g.nodes, id)
	}
}

func (g *Graph) Close() {
	for _, n := range g.nodes {
		if c, ok := n.(Closeable); ok {
			c.Close()
		}
	}
}

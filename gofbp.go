package gofbp

type Process interface {
	Run()
}
type Initialisable interface {
	Init()
}
type Closeable interface {
	Close()
}

type ErrorDealer interface {
	OnError(err error)
}

type Msg interface{}

type Source struct {
	Out chan<- Msg
}

func (s *Source) Close() {
	close(s.Out)
}

func (s *Source) MsgOut() *chan<- Msg {
	return &s.Out
}

type MsgIn interface {
	MsgIn() *<-chan Msg
}

type MsgOut interface {
	MsgOut() *chan<- Msg
}

type Sink struct {
	In <-chan Msg
}

func (s *Sink) MsgIn() *<-chan Msg {
	return &s.In
}

func Bind(src MsgOut, dst MsgIn, cap uint) {
	BindRaw(src.MsgOut(), dst.MsgIn(), cap)
}

func BindRaw(src *chan<- Msg, dst *<-chan Msg, cap uint) {
	c := make(chan Msg, cap)
	*src = c
	*dst = c
}

type NodeID string

type Parent interface {
	Node
	// Add child to the parent.
	// Check parent implementation if it auto-removes child from old parent.
	// Note: a node may only have 1 parent node.
	// Adding it to multiple different parents results in problematic iteration and double closing.
	// Adding it twice to the same parent results in double error messages.
	AddChild(node Node)
	// Remove child from the parent
	RemoveChild(id NodeID)
}

type Child interface {
	// returns the parent node. May be nil if no parent exists.
	Parent() Parent
	// Sets the parent node (a child can only have one parent)
	SetParent(p Parent)
}

type Node interface {
	Child
	// returns the ID of the node
	ID() NodeID
}

type WithID struct {
	id NodeID
}

func (w *WithID) ID() NodeID {
	return w.id
}

type BasicChildImpl struct {
	parent Parent
}

func (n *BasicChildImpl) Parent() Parent {
	return n.parent
}

func (n *BasicChildImpl) SetParent(p Parent) {
	n.parent = p
}

// The most common child implementation: ID, parent relationship, propagate errors to parent.
type ChildImpl struct {
	WithID
	BasicChildImpl
}

func (n *ChildImpl) OnError(err error) {
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

// A simple wrapper component that can deal with errors by directing them in a special channel,
//  ErrOut, to be consumed somewhere else.
// Use this as a wrapper to prevent errors from propagating upwards (as would happen with ChildImpl)
type ErrorCatcher struct {
	WithID
	BasicChildImpl
	Child  Node
	errCh  chan<- error
	ErrOut <-chan error
}

func NewErrorCatcher(id NodeID) *ErrorCatcher {
	ec := new(ErrorCatcher)
	ec.id = id
	ec.Init()
	return ec
}

func (ec *ErrorCatcher) Init() {
	ch := make(chan error)
	ec.errCh = ch
	ec.ErrOut = ch
}

func (ec *ErrorCatcher) AddChild(node Node) {
	if ec.Child != nil {
		ec.Child.SetParent(nil)
		ec.Child = nil
	}
	ec.Child = node
	node.SetParent(ec)
}

func (ec *ErrorCatcher) RemoveChild(id NodeID) {
	if ec.Child != nil && ec.Child.ID() == id {
		ec.Child.SetParent(nil)
		ec.Child = nil
	}
}

func (ec *ErrorCatcher) OnError(err error) {
	ec.errCh <- err
}

type Graph struct {
	ChildImpl

	nodes map[NodeID]Node
}

func NewGraph(id NodeID) *Graph {
	g := new(Graph)
	g.id = id
	g.Init()
	return g
}

type Scope []NodeID

// return a copy of the scope, with id appended.
func (s Scope) Inner(id NodeID) Scope {
	inner := make([]NodeID, len(s)+1, len(s)+1)
	copy(inner, s)
	inner[len(s)] = id
	return inner
}

type ScopedNode struct {
	Scope
	Node
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

func (g *Graph) Init() {
	g.nodes = make(map[NodeID]Node)
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

// TODO structure everything into files, unit tests, etc.



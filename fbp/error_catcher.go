package fbp

// A simple wrapper component that can deal with errors by directing them in a special channel,
//  ErrOut, to be consumed somewhere else.
// Use this as a wrapper to prevent errors from propagating upwards (as would happen with ChildImpl)
type ErrorCatcher struct {
	NodeID
	BasicChildImpl
	Child  Node
	errCh  chan<- error
	ErrOut <-chan error
}

func NewErrorCatcher(id NodeID) *ErrorCatcher {
	ec := new(ErrorCatcher)
	ec.NodeID = id
	ch := make(chan error)
	ec.errCh = ch
	ec.ErrOut = ch
	return ec
}

func (ec *ErrorCatcher) GetChild(id NodeID) Node {
	if id == ec.Child.ID() {
		return ec.Child
	} else {
		return nil
	}
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

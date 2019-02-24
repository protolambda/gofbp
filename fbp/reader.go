package fbp

// A way of retrieving a pointer the input channel (i.e. reading for this node),
//  to set it to a channel that is written to by another node.
type MsgReader interface {
	MsgReadPort() NodePort
	MsgReadCh() *<-chan Msg
}

// A simple implementation of MsgIn, to be embedded/added to your node structs.
type NodeInput struct {
	Owner NodeID
	PortID
	In <-chan Msg
}

func Input(owner NodeID, id PortID) *NodeInput {
	return &NodeInput{Owner: owner, PortID: id}
}

func (ni *NodeInput) OwnerID() NodeID {
	return ni.Owner
}

func (ni *NodeInput) MsgReadPort() NodePort {
	return ni
}

func (ni *NodeInput) MsgReadCh() *<-chan Msg {
	return &ni.In
}

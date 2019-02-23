package fbp

// A way of retrieving a pointer the input channel (i.e. reading for this node),
//  to set it to a channel that is written to by another node.
type MsgReader interface {
	MsgReadPort() *<-chan Msg
}

// A simple implementation of MsgIn, to be embedded/added to your node structs.
type NodeInput struct {
	In <-chan Msg
}

func (ni *NodeInput) MsgReadPort() *<-chan Msg {
	return &ni.In
}

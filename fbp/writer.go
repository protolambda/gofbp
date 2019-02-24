package fbp

// A way of retrieving a pointer the output channel (i.e. writing for this node),
//  to set it to a channel that is read from by another node.
type MsgWriter interface {
	MsgWritePort() NodePort
	MsgWriteCh() *chan<- Msg
}

// A simple implementation of MsgOut, to be embedded/added to your node structs.
type NodeOutput struct {
	Owner NodeID
	PortID
	Out chan<- Msg
}

func Output(owner NodeID, id PortID) *NodeOutput {
	return &NodeOutput{Owner: owner, PortID: id}
}

func (no *NodeOutput) OwnerID() NodeID {
	return no.Owner
}

// Sources need to be closed to clean up resources (i.e. the channel used for the communication)
func (no *NodeOutput) Close() {
	close(no.Out)
}

func (no *NodeOutput) MsgWritePort() NodePort {
	return no
}

func (no *NodeOutput) MsgWriteCh() *chan<- Msg {
	return &no.Out
}


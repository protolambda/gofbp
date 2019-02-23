package fbp

// A way of retrieving a pointer the input channel (i.e. reading for this node),
//  to set it to a channel that is written to by another node.
type MsgIn interface {
	MsgIn() *<-chan Msg
}

// A simple implementation of MsgIn, to be embedded/added to your node structs.
type Sink struct {
	In <-chan Msg
}

func (s *Sink) MsgIn() *<-chan Msg {
	return &s.In
}

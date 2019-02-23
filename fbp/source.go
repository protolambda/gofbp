package fbp

// A way of retrieving a pointer the output channel (i.e. writing for this node),
//  to set it to a channel that is read from by another node.
type MsgWriter interface {
	MsgWritePort() *chan<- Msg
}

// A simple implementation of MsgOut, to be embedded/added to your node structs.
type Source struct {
	Out chan<- Msg
}

// Sources need to be closed to clean up resources (i.e. the channel used for the communication)
func (s *Source) Close() {
	close(s.Out)
}

func (s *Source) MsgWritePort() *chan<- Msg {
	return &s.Out
}

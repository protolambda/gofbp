package fbp

// Binding is very simple: an edge between two nodes is created by overwriting the:
//  - read-channel (in) of the destination of the edge
//  - write-channel (out) of the source the edge
// With a newly created channel. The channel should be closed by the writing source.
func Bind(src MsgOut, dst MsgIn, cap uint) {
	BindRaw(src.MsgOut(), dst.MsgIn(), cap)
}

// Bind two channel ports directly. I.e. init a new channel, and overwrite the in and out ports.
func BindRaw(src *chan<- Msg, dst *<-chan Msg, cap uint) {
	c := make(chan Msg, cap)
	*src = c
	*dst = c
}

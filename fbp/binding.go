package fbp

type Bond struct {
	Chan chan Msg
	MsgWriter
	MsgReader
}


// Binding is very simple: an edge between two nodes is created by overwriting the:
//  - read-channel (in) of the destination of the edge
//  - write-channel (out) of the source the edge
// With a newly created channel. The channel should be closed by the writing source.
func Bind(src MsgWriter, dst MsgReader, cap uint) Bond {
	srcCh := src.MsgWriteCh()
	dstCh := dst.MsgReadCh()
	c := make(chan Msg, cap)
	*srcCh = c
	*dstCh = c
	return Bond{c, src, dst}
}

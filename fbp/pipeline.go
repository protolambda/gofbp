package fbp

type MsgForward interface {
	MsgReader
	MsgWriter
}

// Creates a new pipeline, with the same given capacity for each hop,
//  and will setup nodes to have process messages in given order. (through their default read and write ports)
func Pipeline(cap uint, nodes ...MsgForward) (MsgReader, MsgWriter, []Bond) {
	if len(nodes) == 0 {
		return nil, nil, []Bond{}
	}
	if len(nodes) == 1 {
		return nodes[0], nodes[0], []Bond{}
	}

	bonds := make([]Bond, len(nodes)-1, len(nodes)-1)

	for i := 0; i < len(nodes)-1; i++ {
		// bind to next input
		bonds[i] = Bind(nodes[i], nodes[i+1], cap)
	}

	// no errors, pipeline is complete.
	return nodes[0], nodes[len(bonds)], bonds
}

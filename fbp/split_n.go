package fbp

type SplitN struct {
	BasicNodeImpl

	// single input
	*NodeInput

	Outputs []*NodeOutput
}

func NewSplitN(id NodeID) *SplitN {
	sn := new(SplitN)
	sn.NodeID = id
	sn.NodeInput = Input(id, "in")
	sn.Outputs = make([]*NodeOutput, 0)
	return sn
}

// Add output
func (sn *SplitN) AddOutput(id PortID) *NodeOutput {
	output := Output(sn.NodeID, id)
	sn.Outputs = append(sn.Outputs, output)
	return output
}

func (sn *SplitN) Run() {
	for item := range sn.In {
		for _, out := range sn.Outputs {
			out.Out <- item
		}
	}
}

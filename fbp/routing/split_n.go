package routing

import "protolambda.com/gofbp/fbp"

type SplitN struct {
	fbp.BasicNodeImpl

	// single input
	*fbp.NodeInput

	Outputs []*fbp.NodeOutput
}

func NewSplitN(id fbp.NodeID) *SplitN {
	sn := new(SplitN)
	sn.NodeID = id
	sn.NodeInput = fbp.Input(id, "in")
	sn.Outputs = make([]*fbp.NodeOutput, 0)
	return sn
}

// Add output
func (sn *SplitN) AddOutput(id fbp.PortID) *fbp.NodeOutput {
	output := fbp.Output(sn.NodeID, id)
	sn.Outputs = append(sn.Outputs, output)
	return output
}

func (sn *SplitN) Run() {
	for item := range sn.In {
		for _, out := range sn.Outputs {
			go func(m fbp.Msg) {
				out.Out <- m
			}(item)
		}
	}
}

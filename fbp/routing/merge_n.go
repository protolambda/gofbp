package routing

import (
	"protolambda.com/gofbp/fbp"
	"sync"
)

type MergeN struct {
	fbp.BasicNodeImpl

	// single output
	*fbp.NodeOutput

	Inputs []*fbp.NodeInput
}

func NewMergeN(id fbp.NodeID) *MergeN {
	mn := new(MergeN)
	mn.NodeID = id
	mn.Inputs = make([]*fbp.NodeInput, 0)
	mn.NodeOutput = fbp.Output(id, "out")
	return mn
}

// Add input (ignored if added after running)
func (mn *MergeN) AddInput(id fbp.PortID) *fbp.NodeInput {
	input := fbp.Input(mn.NodeID, id)
	mn.Inputs = append(mn.Inputs, input)
	return input
}

func (mn *MergeN) Run() {
	var wg sync.WaitGroup
	wg.Add(len(mn.Inputs))
	for _, input := range mn.Inputs {
		go func(c *fbp.NodeInput) {
			for v := range c.In {
				mn.Out <- v
			}
			wg.Done()
		}(input)
	}
	wg.Wait()
}


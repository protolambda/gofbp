package fbp

import "sync"

type MergeN struct {
	BasicNodeImpl

	// single output
	*NodeOutput

	Inputs []*NodeInput
}

func NewMergeN(id NodeID) *MergeN {
	mn := new(MergeN)
	mn.NodeID = id
	mn.Inputs = make([]*NodeInput, 0)
	mn.NodeOutput = Output(id, "out")
	return mn
}

// Add input (ignored if added after running)
func (mn *MergeN) AddInput(id PortID) *NodeInput {
	input := Input(mn.NodeID, id)
	mn.Inputs = append(mn.Inputs, input)
	return input
}

func (mn *MergeN) Run() {
	var wg sync.WaitGroup
	wg.Add(len(mn.Inputs))
	for _, input := range mn.Inputs {
		go func(c *NodeInput) {
			for v := range c.In {
				mn.Out <- v
			}
			wg.Done()
		}(input)
	}
	wg.Wait()
}


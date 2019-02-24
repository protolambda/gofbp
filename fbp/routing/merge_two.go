package routing

import "protolambda.com/gofbp/fbp"

type MergeTwo struct {
	fbp.BasicNodeImpl

	// single output
	*fbp.NodeOutput

	InA *fbp.NodeInput
	InB *fbp.NodeInput
}

func NewMergeTwo(id fbp.NodeID) *MergeTwo {
	mt := new(MergeTwo)
	mt.NodeID = id
	mt.InA = fbp.Input(id, "in_a")
	mt.InB = fbp.Input(id, "in_b")
	mt.NodeOutput = fbp.Output(id, "out")
	return mt
}

func (mt *MergeTwo) Run() {
	a := mt.InA.In
	b := mt.InB.In
	for a != nil || b != nil {
		select {
		case v, ok := <-a:
			if !ok {
				a = nil
				continue
			}
			mt.Out <- v
		case v, ok := <-b:
			if !ok {
				b = nil
				continue
			}
			mt.Out <- v
		}
	}
}

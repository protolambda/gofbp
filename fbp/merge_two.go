package fbp

type MergeTwo struct {
	BasicNodeImpl

	// single output
	*NodeOutput

	InA *NodeInput
	InB *NodeInput
}

func NewMergeTwo(id NodeID) *MergeTwo {
	mt := new(MergeTwo)
	mt.NodeID = id
	mt.InA = Input(id, "in_a")
	mt.InB = Input(id, "in_b")
	mt.NodeOutput = Output(id, "out")
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

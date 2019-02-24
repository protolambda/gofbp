package fbp

type SplitTwo struct {
	BasicNodeImpl

	// single input
	*NodeInput

	OutA *NodeOutput
	OutB *NodeOutput

}

func NewSplitTwo(id NodeID) *SplitTwo {
	st := new(SplitTwo)
	st.NodeID = id
	st.NodeInput = Input(id, "in")
	st.OutA = Output(id, "out_a")
	st.OutB = Output(id, "out_b")
	return st
}

func (st *SplitTwo) Run() {
	for item := range st.In {
		st.OutA.Out <- item
		st.OutB.Out <- item
	}
}

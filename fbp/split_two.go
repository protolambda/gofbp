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
	sendA := func(item Msg) {
		st.OutA.Out <- item
	}
	sendB := func(item Msg) {
		st.OutB.Out <- item
	}
	for item := range st.In {
		go sendA(item)
		go sendB(item)
	}
}

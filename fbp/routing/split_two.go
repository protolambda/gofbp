package routing

import "protolambda.com/gofbp/fbp"

type SplitTwo struct {
	fbp.BasicNodeImpl

	// single input
	*fbp.NodeInput

	OutA *fbp.NodeOutput
	OutB *fbp.NodeOutput

}

func NewSplitTwo(id fbp.NodeID) *SplitTwo {
	st := new(SplitTwo)
	st.NodeID = id
	st.NodeInput = fbp.Input(id, "in")
	st.OutA = fbp.Output(id, "out_a")
	st.OutB = fbp.Output(id, "out_b")
	return st
}

func (st *SplitTwo) Run() {
	sendA := func(item fbp.Msg) {
		st.OutA.Out <- item
	}
	sendB := func(item fbp.Msg) {
		st.OutB.Out <- item
	}
	for item := range st.In {
		go sendA(item)
		go sendB(item)
	}
}

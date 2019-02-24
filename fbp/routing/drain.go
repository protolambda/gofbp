package routing

import "protolambda.com/gofbp/fbp"

type Drain struct {
	 fbp.BasicNodeImpl

	// single input
	*fbp.NodeInput

}

func NewDrain(id fbp.NodeID) *Drain {
	st := new(Drain)
	st.NodeID = id
	st.NodeInput = fbp.Input(id, "in")
	return st
}

func (st *Drain) Run() {
	for range st.In {
		// simply drain input
	}
}

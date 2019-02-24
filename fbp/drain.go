package fbp

type Drain struct {
	BasicNodeImpl

	// single input
	*NodeInput

}

func NewDrain(id NodeID) *Drain {
	st := new(Drain)
	st.NodeID = id
	st.NodeInput = Input(id, "in")
	return st
}

func (st *Drain) Run() {
	for range st.In {
		// simply drain input
	}
}

package fbp

import "errors"

// A pipeline process is a graph with an input and an output.
// Nodes in the graph are processed through the order.
type Pipeline struct {
	Graph
	NodeOutput
	NodeInput
}

// Creates a new pipeline, with the same given capacity for each hop,
//  and will setup nodes to have process messages in given order. (through their default read and write ports)
func NewPipeLine(id NodeID, cap uint, nodes ...Node) (*Pipeline, error) {
	p := new(Pipeline)
	p.NodeID = id
	p.Init()

	var w MsgWriter = &p.NodeOutput
	for _, item := range nodes {
		p.AddChild(item)
		itemR, ok := item.(MsgReader)
		if !ok {
			return nil, errors.New("Item: "+string(item.ID())+" in pipeline"+string(p.NodeID)+" has no default output")
		}
		// bind to previous output
		Bind(w, itemR, cap)
		itemW, ok := item.(MsgWriter)
		if !ok {
			return nil, errors.New("Item: "+string(item.ID())+" in pipeline"+string(p.NodeID)+" has no default output")
		}
		w = itemW
	}
	// Bind last output of pipeline items to pipeline output.
	Bind(w, &p.NodeInput, cap)

	// no errors, pipeline is complete.
	return p, nil
}

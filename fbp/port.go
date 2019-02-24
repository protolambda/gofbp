package fbp

// The ID of a port
type PortID string

func (id PortID) Port() PortID {
	return id
}

func (id PortID) String() string {
	return string(id)
}

type WithPort interface {
	Port() PortID
}

// A node-port is a port, owned by a node
type NodePort interface {
	OwnerID() NodeID
	WithPort
}

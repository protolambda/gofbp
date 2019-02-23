package fbp

// Nodes that implement a Process can be started to run (commonly used at graph-execution-time)
type Process interface {
	Run()
}

// Nodes that implement a Closeable can be stopped and cleaned-up (commonly at end of graph-execution-time)
type Closeable interface {
	Close()
}

// Error-dealers are a very simple error-forwarding mechanic:
//  simply call OnError(...) to publish your error. And Implement OnError(...) to handle it.
// The default implementation (BasicNodeImpl) tries to forward errors to the parent node.
// Use an ErrorCatcher (single-child parent) to wrap some node that you want to channel errors away from.
type ErrorDealer interface {
	OnError(err error)
}

// TODO structure everything into files, unit tests, etc.

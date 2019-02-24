# GoFBP

Go Flow Based Programming (FBP) framework, by @protolambda.

**This is in a prototype phase: it's functional, but experimental.**

See `example/` for a very verbose but interesting FizzBuzz implementation.
It not just maps to "fizz", "buzz", "fizzbuzz", but routes the different cases to different nodes,
 maps them, merges them back, and takes the first 100.
After this it stops the input number generator node, and exits. It also integrates OS signals in a FBP way.

## Usage

The idea of Flow-Based-Programming is that every mutation (or group of mutations) is represented by a node in a graph.
Nodes are connected, grouped, and the graph is executed as a whole.
This is similar to the graph-model implemented in frameworks like Tensorflow.

Since Go is already very capable of running lots of lightweight concurrent processes, a FBP approach works quite well.
Channels, embedding and go-routines are core to this framework.

### Setting up a graph

The core concepts are **binding** and **process nodes**.

#### Binding

It can be as simple as calling:
```
Bind(numberProducer, numberConsumer, 1)
```

This connects the default output of `numberProducer` to the default input of `numberConsumer`.
The connection (a newly created channel) will have a capacity of 1.

Sometimes you want access to the created bond, e.g. to keep track of edges (visualization), or to close channels 
(the GC can get rid of open channels, but maybe your nodes have longer lifespans).
The `Bind(...)` function returns a `Bond` which provides the information you need:
 source and destination identities (node id + port id for sides of bond) and the used channel.


#### Process nodes

When you instantiate a node, you just create a part of the model: it's not activated.
You have to call `go Run()` (from `fbp.Process` interface) to make them do their work (some don't represent processes).

I recommend to just create a local function that collects the processes you create, to later start running them all.
You're free to implement your own run-time stages for your model.
 
See fizzbuzz example.


### Implementing a node

Nodes can be anything, all they need is an ID. Most nodes have inputs and outputs however.

The most common node formats are:

- generators/sources: only have a default output
- middleware: have a default input and a default output
- sinks: only have a default input
- graphs: collections of nodes, to encapsulate errors etc.

#### Examples

A very basic node example is one that forwards messages after processing.
It has a default input and output, and is simple to implement using embedding:

```go
type PrintMiddleware struct {
	fbp.BasicNodeImpl
	*fbp.NodeInput
	*fbp.NodeOutput
}

func NewPrintMiddleware(id fbp.NodeID) *PrintMiddleware {
	m := new(PrintMiddleware)
	m.NodeID = id
	m.NodeInput = fbp.Input(id, "in")
	m.NodeOutput = fbp.Output(id, "out")
	return m
}
```

To make a node do something, it has to implement the `Process` interface:

```go
func (m *PrintMiddleware) Run() {
	for item := range m.In {
		fmt.Println("[", m.ID(), "]: ", item)
		// pass it to next component
		m.Out <- item
	}
}
```

Nodes can also just have an output:

```go
type OsSignal struct {
	fbp.BasicNodeImpl
	*fbp.NodeOutput
	Sig os.Signal
}

func NewOsSignal(id fbp.NodeID, sig os.Signal) *OsSignal {
	sl := new(OsSignal)
	sl.NodeID = id
	sl.NodeOutput = fbp.Output(id, "out")
	sl.Sig = sig
	return sl
}

func (sl *OsSignal) Run() {
	exit := make(chan os.Signal)
	signal.Notify(exit, sl.Sig)
	// wait for exit
	<-exit
	// make output aware
	sl.Out <- "interrupt"
}
```

Or only an input. (e.g. Storage-writer)

More advanced nodes are also possible, with named in/outputs, and error handling:

This node forwards only messages that are divisible by the given divisor,
 other messages are passed to the named `Filtered` output, to consume elsewhere.

```go
type FilterByDivisorMiddleware struct {
	fbp.BasicNodeImpl
	*fbp.NodeInput
	*fbp.NodeOutput
	Filtered *fbp.NodeOutput
	div      uint64
}

func NewFilterByDivisorMiddleware(id fbp.NodeID, div uint64) *FilterByDivisorMiddleware {
	m := new(FilterByDivisorMiddleware)
	m.NodeID = id
	m.NodeInput = fbp.Input(id, "in")
	m.NodeOutput = fbp.Output(id, "out")
	m.Filtered = fbp.Output(id, "filtered")
	m.div = div
	return m
}

func (m *FilterByDivisorMiddleware) Run() {
	for item := range m.In {
		v, ok := item.(uint64)
		if !ok {
			m.OnError(errors.New("div filter" + string(m.NodeID) + " cannot process non-indexed msg"))
			continue
		}
		// filter
		if v%m.div == 0 {
			// pass it to default next component
			m.Out <- item
		} else {
			// pass it to filtered queue
			m.Filtered.Out <- item
		}
	}
}
```

##### Advanced

Alternatively, you can implement `MsgReader` and `MsgWriter` interfaces yourself,
 for rare use-cases where you need more dynamic access to requests for the in/out-puts of a node.


## License

See `LICENSE` file.


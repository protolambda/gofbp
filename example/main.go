package main

import (
	"errors"
	"fmt"
	"os"
	"os/signal"
	"protolambda.com/gofbp/fbp"
)

// --------------- testing things out

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

func (m *PrintMiddleware) Run() {
	for item := range m.In {
		v, ok := item.(fmt.Stringer)
		if !ok {
			m.OnError(errors.New("printer "+string(m.NodeID)+" cannot process non-stringer msg"))
			continue
		}
		fmt.Println("[", m.ID(), "]: ", v)
		// pass it to next component
		m.Out <- item
	}
}

type FilterByDivisorMiddleware struct {
	fbp.BasicNodeImpl
	*fbp.NodeInput
	*fbp.NodeOutput
	Filtered *fbp.NodeOutput
	div uint64
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
			m.OnError(errors.New("div filter" + string(m.NodeID)+" cannot process non-indexed msg"))
			continue
		}
		// filter
		if v % m.div == 0 {
			// pass it to default next component
			m.Out <- item
		} else {
			// pass it to filtered queue
			m.Filtered.Out <- item
		}
	}
}


type EventMiddleware struct {
	fbp.BasicNodeImpl
	*fbp.NodeInput
	*fbp.NodeOutput
	event fbp.Msg
}

func NewEventMiddleware(id fbp.NodeID, event fbp.Msg) *EventMiddleware {
	m := new(EventMiddleware)
	m.NodeID = id
	m.NodeInput = fbp.Input(id, "in")
	m.NodeOutput = fbp.Output(id, "out")
	m.event = event
	return m
}

func (m *EventMiddleware) Run() {
	for range m.In {
		// simply replace it with event.
		m.Out <- m.event
	}
}

type NumberGenerator struct {
	fbp.BasicNodeImpl
	*fbp.NodeOutput
}

func NewNumberGenerator(id fbp.NodeID) *NumberGenerator {
	m := new(NumberGenerator)
	m.NodeID = id
	m.NodeOutput = fbp.Output(id, "out")
	return m
}

func (m *NumberGenerator) Run() {
	// TODO implement generator
}


func main() {

	exit := make(chan os.Signal)
	signal.Notify(exit, os.Interrupt)

	bonds := make([]fbp.Bond, 0)

	bind := func(src fbp.MsgWriter, dst fbp.MsgReader) {
		bonds = append(bonds, fbp.Bind(src, dst, 1))
	}

	numbers := NewNumberGenerator("numbers")

	inputPrinter := NewPrintMiddleware("input_printer")
	bind(numbers, inputPrinter)

	div3 := NewFilterByDivisorMiddleware("fizz?", 3)
	bind(inputPrinter, div3)
	div5 := NewFilterByDivisorMiddleware("buzz?", 5)
	bind(div3.Filtered, div5)
	div3and5 := NewFilterByDivisorMiddleware("fizzbuzz?", 5)
	bind(div3, div3and5)


	fizz := NewEventMiddleware("to_fizz", "fizz")
	buzz := NewEventMiddleware("to_fizz", "buzz")
	fizzbuzz := NewEventMiddleware("to_fizzbuzz", "fizzbuzz")

	// TODO hook up event translation


	outputPrinter := NewPrintMiddleware("output_printer")

	// TODO merge filtered back together

	// TODO create sink node that stops process after N numbers

}
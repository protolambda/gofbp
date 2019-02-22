package main

import (
	"errors"
	"fmt"
	"os"
	"os/signal"
	"protolambda.com/gofbp"
)

// --------------- testing things out


type AbcComponent struct {
	id gofbp.NodeID
	gofbp.Sink
	gofbp.Source
}

func NewAbcComponent(id gofbp.NodeID) *AbcComponent {
	abc := &AbcComponent{id: id}
	return abc
}

func (abc *AbcComponent) ID() gofbp.NodeID {
	return abc.id
}

func (abc *AbcComponent) OnError(err error) {

}

func (abc *AbcComponent) Run() {
	for item := range abc.In {
		bl := item.(uint64)
		fmt.Println(abc.ID(), " - ", bl)
		abc.Out <- item
		abc.OnError(errors.New("test error"))
	}
}



func main() {

	exit := make(chan os.Signal)
	signal.Notify(exit, os.Interrupt)

	foo := NewAbcComponent("FOO")
	bar := NewAbcComponent("BAR")
	var in chan<- gofbp.Msg
	var out <-chan gofbp.Msg
	gofbp.BindRaw(&in, &foo.In, 1)
	gofbp.Bind(foo, bar, 1)
	gofbp.BindRaw(&bar.Out, &out, 1)
	g := gofbp.NewGraph("foobar")

	// TODO bigger example graph + small ascii viz

	ec := gofbp.NewErrorCatcher("error_zone_1")

	// TODO generate test input, pipe into input channel
	for {
		select {
		case o := <-out:
			fmt.Println("Outer out: ", o)
		case err := <-ec.ErrOut:
			fmt.Println("Err zone 1 err: ", err)
		// TODO more zones in example
		//case err := <- ec2.ErrOut:
		//	fmt.Println("Err zone 2 err: ", err)
		case sig := <-exit:
			fmt.Println("Exiting, sig: ", sig)
			// close graph
			g.Close()
			fmt.Println("Closed FBP graph")
			return
		}
	}
}
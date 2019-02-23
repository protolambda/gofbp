package main

import (
	"errors"
	"fmt"
	"os"
	"os/signal"
	"protolambda.com/gofbp/fbp"
)

// --------------- testing things out


type AbcComponent struct {
	id fbp.NodeID
	fbp.NodeInput
	fbp.NodeOutput
}

func NewAbcComponent(id fbp.NodeID) *AbcComponent {
	abc := &AbcComponent{id: id}
	return abc
}

func (abc *AbcComponent) ID() fbp.NodeID {
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
	var in chan<- fbp.Msg
	var out <-chan fbp.Msg
	fbp.BindRaw(&in, &foo.In, 1)
	fbp.Bind(foo, bar, 1)
	fbp.BindRaw(&bar.Out, &out, 1)
	g := fbp.NewGraph("foobar")

	// TODO bigger example graph + small ascii viz

	ec := fbp.NewErrorCatcher("error_zone_1")

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
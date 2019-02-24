package main

import (
	"errors"
	"fmt"
	"os"
	"os/signal"
	"protolambda.com/gofbp/fbp"
	"time"
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
		fmt.Println("[", m.ID(), "]: ", item)
		// pass it to next component
		m.Out <- item
	}
}

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

type MapperFn func(m fbp.Msg) fbp.Msg

type Mapper struct {
	fbp.BasicNodeImpl
	*fbp.NodeInput
	*fbp.NodeOutput
	Fn MapperFn
}

// Simple function-mapping node
func NewMapper(id fbp.NodeID, f MapperFn) *Mapper {
	ma := new(Mapper)
	ma.NodeID = id
	ma.NodeInput = fbp.Input(id, "in")
	ma.NodeOutput = fbp.Output(id, "out")
	ma.Fn = f
	return ma
}

func (ma *Mapper) Run() {
	for item := range ma.In {
		// simply embed it in an event.
		ma.Out <- ma.Fn(item)
	}
}

type NumberGenerator struct {
	fbp.BasicNodeImpl
	*fbp.NodeOutput
	Stop *fbp.NodeInput
	Done *fbp.NodeOutput
}

func NewNumberGenerator(id fbp.NodeID) *NumberGenerator {
	ng := new(NumberGenerator)
	ng.NodeID = id
	ng.NodeOutput = fbp.Output(id, "out")
	ng.Stop = fbp.Input(id, "stop")
	ng.Done = fbp.Output(id, "done")
	return ng
}

func (ng *NumberGenerator) Run() {
	i := uint64(0)
	for {
		if len(ng.Stop.In) == 0 {
			ng.Out <- i
			i++
		} else {
			// remove signal
			<-ng.Stop.In
			// To test out exit timeout / force exit:
			//time.Sleep(100 * time.Second)
			// send done signal
			ng.Done.Out <- "stopped"
			// stop generating
			return
		}
	}
}

type TakeN struct {
	fbp.BasicNodeImpl
	*fbp.NodeInput
	*fbp.NodeOutput
	Done *fbp.NodeOutput
	Amount uint64
}

func NewTakeN(id fbp.NodeID, amount uint64) *TakeN {
	tn := new(TakeN)
	tn.NodeID = id
	tn.NodeInput = fbp.Input(id, "in")
	tn.NodeOutput = fbp.Output(id, "out")
	tn.Done = fbp.Output(id, "done")
	tn.Amount = amount
	return tn
}

func (tn *TakeN) Run() {
	i := uint64(0)
	for item := range tn.In {
		// count current in
		i++
		// stop running when we already took N messages.
		if i >= tn.Amount {
			tn.Done.Out <- "complete"
			return
		}
		tn.Out <- item
		// Check if we can stop before waiting for next item.
		if i + 1 >= tn.Amount {
			tn.Done.Out <- "complete"
			return
		}
	}
}

type Sleeper struct {
	fbp.BasicNodeImpl
	*fbp.NodeInput
	*fbp.NodeOutput
	T     time.Duration
	Async bool
}

func NewSleeper(id fbp.NodeID, t time.Duration, async bool) *Sleeper {
	sl := new(Sleeper)
	sl.NodeID = id
	sl.NodeInput = fbp.Input(id, "in")
	sl.NodeOutput = fbp.Output(id, "out")
	sl.T = t
	sl.Async = async
	return sl
}

func (sl *Sleeper) Run() {
	// delay each message by the given duration
	f := func(m fbp.Msg) {
		time.Sleep(sl.T)
		sl.Out <- m
	}
	for item := range sl.In {
		if sl.Async {
			go f(item)
		} else {
			f(item)
		}
	}
}

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

func main() {

	processes := make([]fbp.Process, 0)
	bonds := make([]fbp.Bond, 0)

	p := func(pro fbp.Process) {
		processes = append(processes, pro)
	}

	// Try changing the capacity from 1 to 10:
	//  nodes will not have to weight for their outputs to be processed, and can buffer some work.
	bind := func(src fbp.MsgWriter, dst fbp.MsgReader) {
		bonds = append(bonds, fbp.Bind(src, dst, 1))
	}

	/*
		INPUT
	 */
	numbers := NewNumberGenerator("numbers")
	p(numbers)

	// Try changing this to a nanoseconds, and you will see the concurrency in effect.
	// (Also changes bond capacity for bigger effect)
	numbersSlow := NewSleeper("numbers_slow", 10 * time.Millisecond, false)
	p(numbersSlow)
	bind(numbers, numbersSlow)

	inputPrinter := NewPrintMiddleware("input_printer")
	p(inputPrinter)
	bind(numbersSlow, inputPrinter)

	/*
		FIZZ BUZZ
	 */

	// First choice: mod 3?
	div3 := NewFilterByDivisorMiddleware("div3", 3)
	p(div3)
	bind(inputPrinter, div3)

	// No? -> Second choice: mod 5?
	div5 := NewFilterByDivisorMiddleware("div5", 5)
	p(div5)
	bind(div3.Filtered, div5)

	// Yes? -> Second choice: mod 5?
	div3div5 := NewFilterByDivisorMiddleware("div3_div5", 5)
	p(div3div5)
	bind(div3, div3div5)

	// map to events
	// Yes3, No5: fizz
	fizz := NewMapper("to_fizz", func(m fbp.Msg) fbp.Msg { return fmt.Sprintf("fizz (%d)", m)})
	p(fizz)
	bind(div3div5.Filtered, fizz)
	// No3, Yes5: buzz
	buzz := NewMapper("to_buzz", func(m fbp.Msg) fbp.Msg { return fmt.Sprintf("buzz (%d)", m)})
	p(buzz)
	bind(div5, buzz)
	// Yes3, Yes5: fizzbuzz
	fizzbuzz := NewMapper("to_fizzbuzz", func(m fbp.Msg) fbp.Msg { return fmt.Sprintf("fizzbuzz (%d)", m)})
	p(fizzbuzz)
	bind(div3div5, fizzbuzz)
	// No3, No5: #number
	// (nothing to do) Numbers output to div5.Filtered

	// Merge channels
	merged := fbp.NewMergeN("merged")
	p(merged)
	bind(fizz, merged.AddInput("fizz"))
	bind(buzz, merged.AddInput("buzz"))
	bind(fizzbuzz, merged.AddInput("fizzbuzz"))
	bind(div5.Filtered, merged.AddInput("normal"))

	/*
		END FIZZBUZZ
	 */
	 // We only want the first 100, i.e. 0...99
	take100 := NewTakeN("take100", 100)
	p(take100)
	bind(merged, take100)

	// Print output
	outputPrinter := NewPrintMiddleware("output_printer")
	p(outputPrinter)
	bind(take100, outputPrinter)

	// After printing we are not interested in the values anymore, drain them.
	drain := fbp.NewDrain("drain")
	p(drain)
	bind(outputPrinter, drain)

	/*
		STOP/CLOSE HANDLING
	 */

	gracefulStopSignal := NewOsSignal("graceful_stop", os.Interrupt)
	p(gracefulStopSignal)

	// Try to graciously stop:
	gracefulExitOptions := fbp.NewMergeTwo("end_cases")
	p(gracefulExitOptions)
	bind(take100.Done, gracefulExitOptions.InA)
	bind(gracefulStopSignal, gracefulExitOptions.InB)
	// Use a MergeN to add more graceful-exit options

	exitHandlers := fbp.NewSplitTwo("exit_handling")
	p(exitHandlers)
	bind(gracefulExitOptions, exitHandlers)

	// Timeout the exit handling
	timeoutDelay := NewSleeper("timeout_delay", 5*time.Second, true)
	p(timeoutDelay)
	bind(exitHandlers.OutA, timeoutDelay)

	// Try adding a delay to the "stopped" event. If it's not soon enough, the program will exit because of this exit.
	timeout := NewMapper("timeout", func(m fbp.Msg) fbp.Msg { return "graceful_exit_timeout"})
	p(timeout)
	bind(timeoutDelay, timeout)

	// wait for end case, close input
	bind(exitHandlers.OutB, numbers.Stop)

	exitCases := fbp.NewMergeN("exit_cases")
	p(exitCases)
	bind(timeout, exitCases.AddInput("timeout"))
	bind(numbers.Done, exitCases.AddInput("graceful"))
	// More force-exit cases can be added like this:
	// bind(otherSignal, exitCases.AddInput("other_stop"))

	exit := fbp.Input("exit", "exit")
	bind(exitCases, exit)

	/*
		RUN GRAPH
	 */

	// Start running all nodes in the graph.
	for _, p := range processes {
		go p.Run()
	}

	/*
		WAIT FOR EXIT/COMPLETION
	 */

	// wait for program to exit
	exitReason := <-exit.In
	fmt.Println("Completed graph execution, exit reason:", exitReason)
	os.Exit(0)

}

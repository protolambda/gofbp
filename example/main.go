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
		// simply embed it in an event.
		m.Out <- m.event
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
	Amount uint64
}

func NewTakeN(id fbp.NodeID, amount uint64) *TakeN {
	tn := new(TakeN)
	tn.NodeID = id
	tn.NodeInput = fbp.Input(id, "in")
	tn.NodeOutput = fbp.Output(id, "in")
	tn.Amount = amount
	return tn
}

func (tn *TakeN) Run() {
	i := uint64(0)
	for range tn.In {
		i++
		if i >= tn.Amount {
			// stop running when we took N messages.
			tn.NodeOutput.Out <- "complete"
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
	s := <-exit
	if s == os.Interrupt {
		sl.Out <- "interrupt"
	} else {
		sl.Out <- "kill"
	}
}

func main() {

	processes := make([]fbp.Process, 0)
	bonds := make([]fbp.Bond, 0)

	p := func(pro fbp.Process) {
		processes = append(processes, pro)
	}

	bind := func(src fbp.MsgWriter, dst fbp.MsgReader) {
		bonds = append(bonds, fbp.Bind(src, dst, 1))
	}

	/*
		INPUT
	 */
	numbers := NewNumberGenerator("numbers")
	p(numbers)
	numbersSlow := NewSleeper("numbers_slow", 100 * time.Millisecond, false)
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
	fizz := NewEventMiddleware("to_fizz", "fizz")
	p(fizz)
	bind(div3div5.Filtered, fizz)
	// No3, Yes5: buzz
	buzz := NewEventMiddleware("to_buzz", "buzz")
	p(buzz)
	bind(div5, buzz)
	// Yes3, Yes5: fizzbuzz
	fizzbuzz := NewEventMiddleware("to_fizzbuzz", "fizzbuzz")
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

	// Print merged output
	outputPrinter := NewPrintMiddleware("output_printer")
	p(outputPrinter)
	bind(merged, outputPrinter)

	/*
		STOP/CLOSE HANDLING
	 */

	endCondition := NewTakeN("end_condition", 100)
	p(endCondition)
	bind(outputPrinter, endCondition)

	gracefulStopSignal := NewOsSignal("graceful_stop", os.Interrupt)
	p(gracefulStopSignal)

	// Try to graciously stop:
	gracefulExitOptions := fbp.NewMergeTwo("end_cases")
	p(gracefulExitOptions)
	bind(endCondition, gracefulExitOptions.InA)
	bind(gracefulStopSignal, gracefulExitOptions.InB)

	exitHandlers := fbp.NewSplitTwo("exit_handling")
	p(exitHandlers)
	bind(gracefulExitOptions, exitHandlers)

	// Timeout the exit handling
	timeout := NewSleeper("timeout", 5*time.Second, true)
	p(timeout)
	bind(exitHandlers.OutA, timeout)

	// wait for end case, close input
	bind(exitHandlers.OutB, numbers.Stop)

	forceStopSignal := NewOsSignal("force_stop", os.Kill)
	p(forceStopSignal)

	exitCases := fbp.NewMergeN("exit_cases")
	p(exitCases)
	bind(timeout, exitCases.AddInput("timeout"))
	bind(forceStopSignal, exitCases.AddInput("force_stop"))
	bind(numbers.Done, exitCases.AddInput("graceful"))

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

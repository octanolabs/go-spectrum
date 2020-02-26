package syncronizer

import (
	"math/big"
)

func (s *Synchronizer) startRoutineManager() {
	// As we create routines with AddLink, and make them block at r.Link()
	// this goroutine will receive tasks from the channel and continue each one
	// finish for it to terminate and proceed to the next task

	go func() {
		//var i int
	loop:
		for {
			select {
			case task := <-s.routines:
				task.wait()

				var b = s.didAbort(task)

				//fmt.Println(i, " did abort ", b)

				if b {
					s.Aborted = true

					//fmt.Println("aborting")

					task.stop()

					//fmt.Println("Aborted")

					break loop
				}

				//fmt.Println("task")

				task.release()
				task.finish()

				var c = s.didAbort(task)

				//fmt.Println(i, " did abort after ", c)

				if c {
					s.Aborted = true

					//fmt.Println("aborting after")

					task.closeNext()

					//fmt.Println("Aborted after")

					break loop
				}

				if len(s.routines) == 0 {
					s.quit()
					break loop
				}
				//i++
			}
		}
		//fmt.Println("finish sync")
		return
	}()
}

type Synchronizer struct {
	routines    chan *Task
	logChan     chan *logObject
	quitChan    chan int
	abortChan   chan *Task
	nextChannel chan int
	Aborted     bool
}

func (s *Synchronizer) Log(blockNo uint64, txns, transfers, uncles int, minted *big.Int, supply *big.Int) {
	s.logChan <- &logObject{
		blockNo:        blockNo,
		txns:           txns,
		tokentransfers: transfers,
		uncleNo:        uncles,
		minted:         minted,
		supply:         supply,
	}
}

func (s *Synchronizer) AddLink(body func(*Task)) {

	if s.Aborted {
		return
	}

	nr := newTask(s, body, s.nextChannel)

	c := make(chan int)
	s.nextChannel = c

	nr.abortFunc = func() {
		close(c)
	}

	go nr.run()

loop:
	for {
		select {
		case s.routines <- nr:
			break loop
		default:
			if s.Aborted {
				break loop
			}
		}
	}

}

func (s *Synchronizer) Finish() (closed bool) {
	for {
		select {
		case _, more := <-s.nextChannel:
			if more {
				return false
			}
			return true
		}
	}
}

func (s *Synchronizer) quit() {
	s.nextChannel <- 0
}

func (s *Synchronizer) didAbort(t *Task) bool {
	select {
	case closedTask := <-s.abortChan:
		if closedTask == t {
			return true
		} else {
			s.abortChan <- closedTask
			return false
		}
	default:
		return false
	}
}

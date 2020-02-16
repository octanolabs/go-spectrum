package syncronizer

import (
	"math/big"
)

func (s *Synchronizer) startRoutineManager() {
	// As we create routines with AddLink, and make them block at r.Link()
	// this goroutine will receive tasks from the channel and continue each one
	// wait for it to terminate and proceed to the next task

	go func() {
		for {
			select {
			case _, open := <-s.abort:
				if !open {
					s.aborted = true
				}
				return
			}
		}
	}()

	go func() {
		for {
			select {
			case task := <-s.routines:

				if s.aborted {
					task.close()
				} else {
					task.release()
					task.wait()
					if s.aborted {
						task.close()
					}
				}

				if len(s.routines) == 0 {
					s.quit <- 0
					s.abort <- 0
					return
				}
			}
		}
	}()
}

type Synchronizer struct {
	routines    chan *Task
	logChan     chan *logObject
	quit, abort chan int
	aborted     bool
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

func (s *Synchronizer) Abort() {
	close(s.abort)
}

func (s *Synchronizer) AddLink(body func(*Task)) {

	nr := newTask(body)

	go nr.run()

	s.routines <- nr

}

func (s *Synchronizer) Finish() (closed bool) {
	for {
		select {
		case <-s.quit:
			return s.aborted
		}
	}
}

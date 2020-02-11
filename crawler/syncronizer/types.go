package syncronizer

import "math/big"

type Synchronizer struct {
	routines    []*Routine
	maxRoutines int
	logChan     chan *logObject
	quit        chan int
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

func (s *Synchronizer) AddLink(f func(t *Routine)) {

	r := s.newLink()

	go func() {

		f(r)

		if r.isClosed() {
			return
		}

		r.send(0)
	}()

}

func (s *Synchronizer) Abort() {
	// An effect of how we add and clean routines is that the first one in the slice is always the one that's running
	s.routines[0].close()
}

func (s *Synchronizer) Finish() (closed bool) {
	for {
		select {
		case _, open := <-s.quit:
			if !open {
				return false
			}
			return true
		}
	}
}

type Routine struct {
	in           chan int
	out          chan int
	done, closed bool
}

func (r *Routine) Link() (closed bool) {
	closed = r.receive()
	return
}

// Returns a new sync object with no routines
// Routines should be linked inside the routine function body via Routines.Link()
// All routines should be linked together with syncBlock.AddLink()

func NewSync(maxRoutines int) *Synchronizer {
	s := &Synchronizer{routines: []*Routine{}, maxRoutines: maxRoutines}

	startLogger(s)

	return s
}

package syncronizer

//TODO: scrap everything, try with buffered channel
// if taks are kept inside buffered channel, we loose the ability to
// abort sync, as we can't routine.close() so easily

// Returns a new sync object with no routines
// Routines should be linked inside the routine function body via Routines.Link()
// All routines should be linked together with syncBlock.AddLink()

func NewSync(maxRoutines int) *Synchronizer {
	s := &Synchronizer{routines: make(chan *Task, maxRoutines), abortChan: make(chan *Task, 1), quitChan: make(chan int), nextChannel: make(chan int)}

	s.startRoutineManager()

	return s
}

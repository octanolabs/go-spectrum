package syncronizer

func (r *Routine) isDone() bool {
	return r.done
}
func (r *Routine) isClosed() bool {
	return r.closed
}

func (r *Routine) receive() (closed bool) {
	for {
		select {
		case _, more := <-r.in:
			if more {
				return false
			} else {
				r.close()
				return true
			}
		}
	}
}
func (r *Routine) send(v int) {
	r.done = true
	r.out <- v
}

func (r *Routine) close() {
	r.closed = true
	close(r.out)
}

func newRoutine(in chan int) *Routine {
	return &Routine{in, make(chan int, 1), false, false}
}

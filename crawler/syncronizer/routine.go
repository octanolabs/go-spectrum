package syncronizer

type Task struct {
	hang, done chan int
	fn         func()
}

func (r *Task) Link() (closed bool) {
	closed = r.receive()
	return
}

func (r *Task) release() {
	r.hang <- 0
}

func (r *Task) wait() {
	<-r.done
}

func (r *Task) run() {
	r.fn()
}

func (r *Task) receive() (closed bool) {
	for {
		select {
		case _, more := <-r.hang:
			if more {
				return false
			} else {
				return true
			}
		}
	}
}

func (r *Task) close() {
	close(r.hang)
}

func newTask(fn func(*Task)) *Task {

	r := &Task{make(chan int), make(chan int), nil}

	rFn := func() {
		fn(r)

		r.done <- 0
	}

	r.fn = rFn

	return r
}

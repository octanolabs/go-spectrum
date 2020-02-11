package syncronizer

import (
	log "github.com/sirupsen/logrus"
)

func (s *Synchronizer) isAvailable() bool {
	if len(s.routines) < s.maxRoutines {
		return true
	}
	return false
}

func (s *Synchronizer) hangUntilFree() {
	waitChan := make(chan int, 1)

	go func() {
		for {
			s.clean()
			if s.isAvailable() {
				waitChan <- len(s.routines)
			}
		}
	}()

wait:
	for {
		select {
		case d := <-waitChan:
			log.Debugf("resuming, len routines %v", d)
			break wait
		}
	}
	return
}

func (s *Synchronizer) clean() {
	for i := 0; i < len(s.routines); i++ {
		if s.routines[i].isDone() || s.routines[i].isClosed() {
			// shift left by 1
			s.routines = s.routines[1:]
		} else {
			break
		}
	}
}

func (s *Synchronizer) lastRoutine() *Routine {
	return s.routines[len(s.routines)-1]
}

func (s *Synchronizer) addRoutine(prevOut chan int) *Routine {
	nr := newRoutine(prevOut)
	s.quit = nr.out

	s.routines = append(s.routines, nr)

	return nr
}

// newLink creates a new routine, adds it to the slice and returns it to the caller

func (s *Synchronizer) newLink() *Routine {

	var newLink *Routine

	if len(s.routines) == 0 {
		firstChan := make(chan int, 1)
		firstChan <- 0

		newLink = s.addRoutine(firstChan)

	} else {
		//We save the value of lastRoutine before cleaning the slice
		l := *(s.lastRoutine())

		s.clean()

		if !s.isAvailable() {
			s.hangUntilFree()
		}

		newLink = s.addRoutine(l.out)

	}
	return newLink
}

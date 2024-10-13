package active

import "sync"

// Simple implementation of active object pattern
type ActiveObject struct {
	commandsChan chan func()

	dispatchWaitGroup sync.WaitGroup
	workerWaitGroup   sync.WaitGroup
}

func NewActiveObject(queueSize int) *ActiveObject {
	return &ActiveObject{
		commandsChan: make(chan func(), queueSize),
	}
}

func (active *ActiveObject) Start() {
	active.workerWaitGroup.Add(1)
	go func() {
		defer active.workerWaitGroup.Done()
		for command := range active.commandsChan {
			command()
		}
	}()
}

func (active *ActiveObject) Close() {
	active.dispatchWaitGroup.Wait()
	close(active.commandsChan)
	active.workerWaitGroup.Wait()
}

// Asynchronously adds a command at the end of the commands queue
// Use with caution as this is unbound by back pressure
func (active *ActiveObject) DispatchCommand(command func()) {
	active.dispatchWaitGroup.Add(1)
	go func() {
		defer active.dispatchWaitGroup.Done()
		active.commandsChan <- command
	}()
}

func (active *ActiveObject) EnqueueCommand(command func()) {
	active.commandsChan <- command
}

func RunCommandSync[T any](active *ActiveObject, command func() T) T {
	resChan := make(chan T)

	active.EnqueueCommand(func() {
		resChan <- command()
		close(resChan)
	})

	return <-resChan
}

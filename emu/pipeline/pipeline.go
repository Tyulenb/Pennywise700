package pipeline

import "sync"

// The pipeline emits a conveyor
// Commands to be executed at each stage will be written in pipe
// Mutex is needed to eliminate race conditions
type Pipeline struct {
	Pipe []uint32
    Mtx sync.Mutex
}

func NewPipeline() *Pipeline {
    return &Pipeline{
        Pipe: make([]uint32, 5),
    }
}

//Read new cmd and moves pipeline stages for next step
func (p *Pipeline) Move(cmd uint32) {
    p.Mtx.Lock()
    defer p.Mtx.Unlock()
    for i := len(p.Pipe)-1; i >= 1; i-- {
        p.Pipe[i] = p.Pipe[i-1]
    }
    p.Pipe[0] = cmd
}

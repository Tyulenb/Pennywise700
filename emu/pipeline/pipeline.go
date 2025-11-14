package pipeline

import (
	"fmt"
	"sync"
)

//Imitation of Arithmetic Logic Unit
type ALU struct {
	Op1 uint16
	Op2 uint16
	Res uint16
}

func (a *ALU) ToString() string {
    return fmt.Sprintf("  OP1:%v\n   OP2:%v\n   RES:%v", a.Op1, a.Op2, a.Res)
}

// The pipeline emits a conveyor
// Commands to be executed at each stage will be written in pipe
// Mutex is needed to eliminate race conditions
type Pipeline struct {
	Pipe []uint32
    Alu []ALU
    Mtx sync.Mutex
}

func NewPipeline(stages int) *Pipeline {
    return &Pipeline{
        Pipe: make([]uint32, stages),
        Alu: make([]ALU, stages),
    }
}

//Read new cmd and moves pipeline stages for next step
func (p *Pipeline) Move(cmd uint32) {
    p.Mtx.Lock()
    defer p.Mtx.Unlock()
    for i := len(p.Pipe)-1; i >= 1; i-- {
        p.Pipe[i] = p.Pipe[i-1]
        p.Alu[i] = p.Alu[i-1]
    }
    p.Pipe[0] = cmd
    p.Alu[0] = ALU{}
}

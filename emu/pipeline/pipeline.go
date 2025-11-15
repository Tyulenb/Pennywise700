package pipeline

import (
	"fmt"
	"strconv"
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

func (p *Pipeline) FetchOpcode(stage int) uint8 {
    return uint8(p.Pipe[stage] & 0x00F00000 >> 20)
}

func (p *Pipeline) DecodeAdrR1(stage int) uint16 {
    return uint16(p.Pipe[stage] & 0x000F0000 >> 16)
}

func (p *Pipeline) DecodeAdrR2(stage int) uint16 {
    return uint16(p.Pipe[stage] & 0x0000F000 >> 12)
}

func (p *Pipeline) DecodeAdrR3(stage int) uint16 {
    return uint16(p.Pipe[stage] & 0x00000F00 >> 8)
}

func (p *Pipeline) DecodeAdrM(stage int) uint16 {
    return uint16(p.Pipe[stage] & 0x000003FF)
}

func (p *Pipeline) DecodeLiteral(stage int) uint16 {
    return uint16(p.Pipe[stage] & 0x000FFC00 >> 10) 
}

func (p *Pipeline) DecodeAdrToJump(stage int) uint16 {
    return p.DecodeAdrM(stage) 
}

func (p *Pipeline) DropPipe() {
    for i := range 4 {
        p.Pipe[i] = 0
    }
}

func (p *Pipeline) PipeToString() [5]string {
    result := [5]string{}
    for i := range p.Pipe {
        result[i] = commandToString(p.Pipe[i])
    }
    return result 
}

// Public func to convert commands
func (p *Pipeline) CommandToString(stage int) string {
    return commandToString(p.Pipe[stage])
}

// Private func to convert commands
func commandToString(cmd uint32) string {
    opcode := cmd & 0x00F00000 >> 20
    switch opcode {
    case 0:
        return "NOP"
    case 1:
        lit := cmd & 0x000FFC00 >> 10
        adr_m := cmd & 0x000003FF 
        return fmt.Sprintf("LTM %v %v", strconv.Itoa(int(lit)), strconv.Itoa(int(adr_m)))
    case 2:
        adr_r := cmd & 0x000F0000>>16 
        adr_m := cmd & 0x000003FF 
        return fmt.Sprintf("MTR %v %v", strconv.Itoa(int(adr_r)), strconv.Itoa(int(adr_m)))
    case 3:
        adr_r1 := cmd & 0x000F0000>>16 
        adr_r2 := cmd & 0x0000F000>>12 
        return fmt.Sprintf("RTR %v %v", strconv.Itoa(int(adr_r1)), strconv.Itoa(int(adr_r2)))
    case 4:
        adr_r1 := cmd & 0x000F0000>>16 
        adr_r2 := cmd & 0x0000F000>>12 
        adr_r3 := cmd & 0x00000F00>>8
        return fmt.Sprintf("SUB %v %v %v", strconv.Itoa(int(adr_r1)), strconv.Itoa(int(adr_r2)), strconv.Itoa(int(adr_r3)))
    case 5:
        adr_r1 := cmd & 0x000F0000>>16 
        adr_r2 := cmd & 0x0000F000>>12 
        adr_m := cmd & 0x000003FF
        return fmt.Sprintf("SUB %v %v %v", strconv.Itoa(int(adr_r1)), strconv.Itoa(int(adr_r2)), strconv.Itoa(int(adr_m)))
    case 6:
        adr_r1 := cmd & 0x000F0000>>16 
        adr_r2 := cmd & 0x0000F000>>12 
        return fmt.Sprintf("MTRK %v %v", strconv.Itoa(int(adr_r1)), strconv.Itoa(int(adr_r2)))
    case 7:
        adr_r1 := cmd & 0x000F0000>>16 
        adr_r2 := cmd & 0x0000F000>>12 
        return fmt.Sprintf("RTMK %v %v", strconv.Itoa(int(adr_r1)), strconv.Itoa(int(adr_r2)))
    case 8:
        adr_to_jmp := cmd & 0x000003FF
        return fmt.Sprintf("JMP %v", adr_to_jmp)
    case 9:
        adr_r1 := cmd & 0x000F0000>>16 
        adr_r2 := cmd & 0x0000F000>>12 
        adr_r3 := cmd & 0x00000F00>>8
        return fmt.Sprintf("SUM %v %v %v", strconv.Itoa(int(adr_r1)), strconv.Itoa(int(adr_r2)), strconv.Itoa(int(adr_r3)))
    }
    return ""
}

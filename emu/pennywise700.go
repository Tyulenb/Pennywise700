package main

import (
    "sync"
    "github.com/Tyulenb/Pennywise700/emu/pipeline"
)

const (
    NOP = iota
    LTM
    MTR
    RTR
    SUB
    JUMP_LESS
    MTRK
    RTMK
    JMP
    SUM
)

type ALU struct {
	op1 uint16
	op2 uint16
	res uint16
	mtx sync.Mutex
}

type Pennywise700 struct {
	//memory of commands
	cmd_mem  [1024]uint32
	//main memory
	mem      [1024]uint16
	//registers
	RF       [16]uint16
	//program counter
	pc       uint16

	pipeline pipeline.Pipeline
	alu      ALU
}

func (p *Pennywise700) EmulateCycle() {
    //TO DO ADD MIPS
    for {
        var wg sync.WaitGroup 
        p.pipeline.Move(p.cmd_mem[p.pc]) //Zero stage, FETCH COMMAND
        wg.Go(p.stageOne)
        wg.Go(p.stageTwo)
        wg.Go(p.stageThree)
        wg.Go(p.stageFour)
        wg.Wait()
    }
}

//DECODE OP 1
func (p *Pennywise700) stageOne() {
    //fetching current command on stage one
    p.pipeline.Mtx.Lock()
    cmd := p.pipeline.Pipe[1]&0x00F00000
    p.pipeline.Mtx.Unlock()

    //execute command with alu
    p.alu.mtx.Lock()
    switch cmd {
    case LTM:
        p.alu.op1 = uint16(cmd&0x000FFC00>>10)

    case MTR:
        adr_m := cmd&0x000003FF
        p.alu.op1 = p.mem[adr_m]

    case RTR, MTRK, RTMK:
        adr_r2 := cmd&0x0000F000>>12
        p.alu.op1 = p.RF[adr_r2]

    case SUB, JUMP_LESS, SUM:
        adr_r1 := cmd&0x000F0000>>16
        p.alu.op1 = p.RF[adr_r1]

    case JMP:
        p.alu.op1 = uint16(cmd&0x000003FF) 
    }
    p.alu.mtx.Unlock()
}

//DECODE OP 2
func (p *Pennywise700) stageTwo() {
    //fetching current command on stage two 
    p.pipeline.Mtx.Lock()
    cmd := p.pipeline.Pipe[2]&0x00F00000
    p.pipeline.Mtx.Unlock()

    //execute command with alu
    p.alu.mtx.Lock()
    switch cmd {
    case SUB, SUM, JUMP_LESS:
        adr_r2 := cmd&0x0000F000>>12
        p.alu.op2 = p.RF[adr_r2]
    }
    p.alu.mtx.Unlock()
}

//EXECUTE
func (p *Pennywise700) stageThree() {
    //fetching current command on stage three 
    p.pipeline.Mtx.Lock()
    cmd := p.pipeline.Pipe[3]&0x00F00000
    p.pipeline.Mtx.Unlock()

    //execute command with alu
    p.alu.mtx.Lock()
    switch cmd {
    case LTM, MTR, RTR, MTRK, RTMK, JMP:
        p.alu.res = p.alu.op1

    case SUB:
       p.alu.res = p.alu.op2 - p.alu.op1 

    case SUM:
        p.alu.res = p.alu.op1 + p.alu.op2

    case JUMP_LESS:
        if p.alu.op1 >= p.alu.op2 {
            p.alu.res = 1 
        }else{
            p.alu.res = 0
        }
    }
    p.alu.mtx.Unlock()
}

//WRITEBACK
func (p *Pennywise700) stageFour() {
    //fetching current command on stage four 
    p.pipeline.Mtx.Lock()
    cmd := p.pipeline.Pipe[4]&0x00F00000
    p.pipeline.Mtx.Unlock()

    //execute command with alu
    p.alu.mtx.Lock()
    switch cmd {
    case LTM:
        adr_r1 := cmd&0x000F0000>>16
        p.mem[adr_r1] = p.alu.res
        p.pc += 1

    case MTR, RTR:
        adr_r1 := cmd&0x000F0000>>16
        p.RF[adr_r1] = p.alu.res
        p.pc += 1 

    case SUB, SUM:
        adr_r3 := cmd&0x00000F00>>8
        p.RF[adr_r3] = p.alu.res
        p.pc += 1

    case JUMP_LESS:
        if p.alu.res == 1 {
            p.pc = p.alu.res
        }else {
            p.pc += 1
        }

    case MTRK:
        adr_r1 := cmd&0x000F0000>>16
        p.RF[adr_r1] = p.mem[p.alu.res]

    case RTMK:
        adr_r2 := cmd&0x0000F000>>12
        p.mem[p.alu.res] = p.RF[adr_r2]

    case JMP:
        p.pc = p.alu.res

    default:
        p.pc += 1
    }
    p.alu.mtx.Unlock()
}

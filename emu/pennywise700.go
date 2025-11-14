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
    //necessary for MIPS optimizations
	alu_prev ALU
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
        p.alu_prev.op1, p.alu_prev.op2, p.alu_prev.res = p.alu.op1, p.alu.op2, p.alu.res
    }
}

//DECODE OP 1
func (p *Pennywise700) stageOne() {
    J_L_MIPS := false
    SUB_MIPS := false
    MTRK_MIPS := false

    //fetching current command on stage one
    p.pipeline.Mtx.Lock()
    cmd := (p.pipeline.Pipe[1]&0x00F00000)>>20
    cmd5 := (p.pipeline.Pipe[4]&0x00F00000)>>20
    //MIPS for J_L and RTR case
    if cmd == JUMP_LESS && cmd5 == RTR {
        J_L_adr_r1 := p.pipeline.Pipe[1]&0x000F0000>>16
        RTR_adr_r1 := p.pipeline.Pipe[4]&0x000F0000>>16
        if J_L_adr_r1 == RTR_adr_r1 {
            J_L_MIPS = true
        }
    }
    //MIPS for SUB and RTR case
    if cmd == SUB && cmd5 == RTR {
        SUB_adr_r1 := p.pipeline.Pipe[1]&0x000F0000>>16
        RTR_adr_r1 := p.pipeline.Pipe[4]&0x000F0000>>16
        if SUB_adr_r1 == RTR_adr_r1 {
            SUB_MIPS = true
        }
    }
    //MIPS for MTRK and SUB case
    if cmd == MTRK && cmd5 == SUB {
        MTRK_adr_r2 := p.pipeline.Pipe[1]&0x0000F000>>12
        SUB_adr_r3 := p.pipeline.Pipe[4]&0x00000F00>>8
        if MTRK_adr_r2 == SUB_adr_r3 {
            MTRK_MIPS = true
        }
    }

    p.pipeline.Mtx.Unlock()

    //execute command with alu
    p.alu.mtx.Lock()
    defer p.alu.mtx.Unlock()
    switch cmd {
    case LTM:
        p.alu.op1 = uint16(cmd&0x000FFC00>>10)

    case RTR, MTRK, RTMK:
        if MTRK_MIPS {
            p.alu.op1 = p.alu_prev.res
            return
        }
        adr_r2 := cmd&0x0000F000>>12
        p.alu.op1 = p.RF[adr_r2]

    case SUB, JUMP_LESS, SUM:
        if J_L_MIPS || SUB_MIPS {
            p.alu.op1 = p.alu_prev.res
            return
        }
        adr_r1 := cmd&0x000F0000>>16
        p.alu.op1 = p.RF[adr_r1]

    case JMP:
        p.alu.op1 = uint16(cmd&0x000003FF) 
    }
}

//DECODE OP 2
func (p *Pennywise700) stageTwo() {
    MTR_MIPS := false
    J_L_MIPS := false

    //fetching current command on stage two 
    p.pipeline.Mtx.Lock()
    cmd := (p.pipeline.Pipe[2]&0x00F00000)>>20
    cmd5 := (p.pipeline.Pipe[4]&0x00F00000)>>20
    //MIPS for MTR to LTM case
    if cmd == MTR && cmd5 == LTM {
        MTR_adr_m := p.pipeline.Pipe[2]&0x000003FF
        LTM_adr_m := p.pipeline.Pipe[4]&0x000003FF
        if MTR_adr_m == LTM_adr_m {
            MTR_MIPS = true
        }
    }
    //MIPS for J_L to MTR case
    if cmd == JUMP_LESS && cmd5 == MTR {
        J_L_adr_r2 := p.pipeline.Pipe[2]&0x0000F000>>12
        MTR_adr_r1 := p.pipeline.Pipe[4]&0x000F0000>>16
        if J_L_adr_r2 == MTR_adr_r1 {
            J_L_MIPS = true
        }
    }
    //MIPS for J_L to RTR case
    if cmd == JUMP_LESS && cmd5 == RTR {
        J_L_adr_r2 := p.pipeline.Pipe[2]&0x0000F000>>12
        RTR_adr_r2 := p.pipeline.Pipe[4]&0x0000F000>>12
        if J_L_adr_r2 == RTR_adr_r2 {
            J_L_MIPS = true
        }
    }
    p.pipeline.Mtx.Unlock()

    //execute command with alu
    p.alu.mtx.Lock()
    defer p.alu.mtx.Unlock()
    switch cmd {
    case SUB, SUM, JUMP_LESS:
        if J_L_MIPS {
            p.alu.op2 = p.alu_prev.res
        }
        adr_r2 := cmd&0x0000F000>>12
        p.alu.op2 = p.RF[adr_r2]

    case MTR:
        if MTR_MIPS {
            p.alu.op1 = p.alu_prev.res
            return
        }
        adr_m := cmd&0x000003FF
        p.alu.op1 = p.mem[adr_m]
    }
    p.alu.mtx.Unlock()
}

//EXECUTE
func (p *Pennywise700) stageThree() {
    //fetching current command on stage three 
    p.pipeline.Mtx.Lock()
    cmd := (p.pipeline.Pipe[3]&0x00F00000)>>20
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
    cmd := (p.pipeline.Pipe[4]&0x00F00000)>>20
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

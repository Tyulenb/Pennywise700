package cpu

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strconv"
	"sync"

	"github.com/Tyulenb/Pennywise700/pipeline"
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


type Pennywise700 struct {
	//memory of commands
	cmd_mem  [1024]uint32
	//main memory
	mem      [1024]uint16
	//registers
	RF       [16]uint16
	//program counter
	pc       uint16
    ignoreWR uint8
    DebugMode bool
	pipeline *pipeline.Pipeline
}

func NewPennywise700() *Pennywise700 {
    pipeline := pipeline.NewPipeline(5) //Since we have 5 stages
    p := &Pennywise700{
        pc: 0,
        pipeline: pipeline,
    }
    p.RF[1] = 1
    return p
}

func (p *Pennywise700) EmulateCycle() {
    var wg sync.WaitGroup 
    p.pipeline.Move(p.cmd_mem[p.pc]) //Zero stage, FETCH COMMAND
    wg.Go(p.stageOne)
    wg.Go(p.stageTwo)
    wg.Go(p.stageThree)
    wg.Go(p.stageFour)
    wg.Wait()
}

//DECODE OP 1
func (p *Pennywise700) stageOne() {
    stage := 1
    stage2 := 2
    stage0 := 0
    stage5 := 4
    J_L_MIPS := false
    SUB_MIPS := false
    MTRK_MIPS := false

    //fetching current command on stage one
    p.pipeline.Mtx.Lock()
    defer p.pipeline.Mtx.Unlock()
    cmd := p.pipeline.Pipe[stage]
    opCode := (p.pipeline.Pipe[stage]&0x00F00000)>>20
    opCode0 := (p.pipeline.Pipe[stage0]&0x00F00000)>>20
    opCode2 := p.pipeline.Pipe[stage2]&0x00F00000>>20
    opCode5 := (p.pipeline.Pipe[stage5]&0x00F00000)>>20

    //Prediction logic
    if opCode0 == JUMP_LESS && (opCode == MTR || opCode == RTR) {
        p.pc -= 1
        p.pipeline.Pipe[stage0] = 0 
    }
    
    if opCode0 == MTRK && (opCode == SUB || opCode2 == SUB) {
        p.pc -= 1
        p.pipeline.Pipe[stage0] = 0 
    }

    //MIPS for J_L and RTR case
    if opCode == JUMP_LESS && opCode5 == RTR {
        J_L_adr_r1 := p.pipeline.Pipe[stage]&0x000F0000>>16
        RTR_adr_r1 := p.pipeline.Pipe[stage5]&0x000F0000>>16
        if J_L_adr_r1 == RTR_adr_r1 {
            J_L_MIPS = true
        }
    }
    //MIPS for SUB and RTR case
    if opCode == SUB && opCode5 == RTR {
        SUB_adr_r1 := p.pipeline.Pipe[stage]&0x000F0000>>16
        RTR_adr_r1 := p.pipeline.Pipe[stage5]&0x000F0000>>16
        if SUB_adr_r1 == RTR_adr_r1 {
            SUB_MIPS = true
        }
    }
    //MIPS for MTRK and SUB case
    if opCode == MTRK && opCode5 == SUB {
        MTRK_adr_r2 := p.pipeline.Pipe[stage]&0x0000F000>>12
        SUB_adr_r3 := p.pipeline.Pipe[stage5]&0x00000F00>>8
        if MTRK_adr_r2 == SUB_adr_r3 {
            MTRK_MIPS = true
        }
    }

    //execute command with alu
    switch opCode {
    case LTM:
        p.pipeline.Alu[stage].Op1 = uint16(cmd&0x000FFC00>>10)

    case RTR, MTRK:
        if MTRK_MIPS {
            if p.DebugMode {
                fmt.Println("OP1: MTRK_MIPS EXECUTED")
            }   
            p.pipeline.Alu[stage].Op1 = p.pipeline.Alu[stage5].Res
        }else {
            adr_r2 := cmd&0x0000F000>>12
            p.pipeline.Alu[stage].Op1 = p.RF[adr_r2]
        }

    case SUB, JUMP_LESS, SUM, RTMK:
        if J_L_MIPS || SUB_MIPS {
            if p.DebugMode {
                fmt.Println("OP1: J_L_/SUB_MIPS EXECUTED")
            }
            p.pipeline.Alu[stage].Op1 = p.pipeline.Alu[stage5].Res
        }else {
            adr_r1 := cmd&0x000F0000>>16
            p.pipeline.Alu[stage].Op1 = p.RF[adr_r1]
        }

    case JMP:
        p.pipeline.Alu[stage].Op1 = uint16(cmd&0x000003FF) 
    }
    if p.DebugMode {
         fmt.Printf("\nDECODE OP1\nOpCode: %v\nALU:\n %v\n", numToCommand(opCode), p.pipeline.Alu[stage].ToString())
    }
}

//DECODE OP 2
func (p *Pennywise700) stageTwo() {
    stage := 2
    stage5 := 4
    MTR_MIPS := false
    J_L_MIPS := false

    //fetching current command on stage two 
    p.pipeline.Mtx.Lock()
    defer p.pipeline.Mtx.Unlock()
    cmd := p.pipeline.Pipe[stage]
    opCode := (p.pipeline.Pipe[stage]&0x00F00000)>>20
    opCode5 := (p.pipeline.Pipe[stage5]&0x00F00000)>>20


    //MIPS for MTR to LTM case
    if opCode == MTR && opCode5 == LTM {
        MTR_adr_m := p.pipeline.Pipe[stage]&0x000003FF
        LTM_adr_m := p.pipeline.Pipe[stage5]&0x000003FF
        if MTR_adr_m == LTM_adr_m {
            MTR_MIPS = true
        }
    }
    //MIPS for J_L to MTR case
    if opCode == JUMP_LESS && opCode5 == MTR {
        J_L_adr_r2 := p.pipeline.Pipe[stage]&0x0000F000>>12
        MTR_adr_r1 := p.pipeline.Pipe[stage5]&0x000F0000>>16
        if J_L_adr_r2 == MTR_adr_r1 {
            J_L_MIPS = true
        }
    }
    //MIPS for J_L to RTR case
    if opCode == JUMP_LESS && opCode5 == RTR {
        J_L_adr_r2 := p.pipeline.Pipe[stage]&0x0000F000>>12
        RTR_adr_r2 := p.pipeline.Pipe[stage5]&0x0000F000>>12
        if J_L_adr_r2 == RTR_adr_r2 {
            J_L_MIPS = true
        }
    }

    //execute command with alu
    switch opCode {
    case SUB, SUM, JUMP_LESS:
        if J_L_MIPS {
            if p.DebugMode {
                fmt.Println("OP2 J_L_MIPS was executed!")
            }
            p.pipeline.Alu[stage].Op2 = p.pipeline.Alu[stage5].Res
        }else {
            adr_r2 := cmd&0x0000F000>>12
            p.pipeline.Alu[stage].Op2 = p.RF[adr_r2]
        }

    case MTR:
        if MTR_MIPS {
            if p.DebugMode {
                fmt.Println("OP2 MTR_MIPS was executed!")
            }
            p.pipeline.Alu[stage].Op1 = p.pipeline.Alu[stage5].Res
        }else {
            adr_m := cmd&0x000003FF
            p.pipeline.Alu[stage].Op1 = p.mem[adr_m] //In this command can write to any operand
        }
    }

    if p.DebugMode {
        fmt.Printf("\nDECODE OP2\nOpCode: %v\nALU:\n %v\n", numToCommand(opCode), p.pipeline.Alu[stage].ToString())
    }
}

//EXECUTE
func (p *Pennywise700) stageThree() {
    stage := 3

    //fetching current command on stage three 
    p.pipeline.Mtx.Lock()
    defer p.pipeline.Mtx.Unlock()
    opCode := (p.pipeline.Pipe[stage]&0x00F00000)>>20


    //execute command with alu
    switch opCode {
    case LTM, MTR, RTR, MTRK, RTMK, JMP:
        p.pipeline.Alu[stage].Res = p.pipeline.Alu[stage].Op1

    case SUB:
       p.pipeline.Alu[stage].Res = p.pipeline.Alu[stage].Op1 - p.pipeline.Alu[stage].Op2 

    case SUM:
       p.pipeline.Alu[stage].Res = p.pipeline.Alu[stage].Op2 + p.pipeline.Alu[stage].Op1 

    case JUMP_LESS:
        if p.pipeline.Alu[stage].Op1 >= p.pipeline.Alu[stage].Op2 {
            p.pipeline.Alu[stage].Res = 1 
        }else{
            p.pipeline.Alu[stage].Res = 0
        }
    }
    if p.DebugMode {
        fmt.Printf("\nEXECUTE\nOpCode: %v\nALU:\n %v\n", numToCommand(opCode), p.pipeline.Alu[stage].ToString())
    }
}

//WRITEBACK
func (p *Pennywise700) stageFour() {
    stage := 4
    //fetching current command on stage four 
    p.pipeline.Mtx.Lock()
    defer p.pipeline.Mtx.Unlock()
    cmd := p.pipeline.Pipe[4]
    opCode := (p.pipeline.Pipe[4]&0x00F00000)>>20


    //If we jump to other command, we should skip commands those are already in Pipeline
    //Thus we need to ignore those commands, we will use ignorence counter
    if p.ignoreWR > 0 {
        p.ignoreWR -= 1
        if p.DebugMode {
            fmt.Printf("WRITEBACK was ignored! Next %v commands will be ignored\n", p.ignoreWR)
        }
        p.pc += 1
        return
    }

    //execute command with alu
    switch opCode{
    case LTM:
        adr_m1 := cmd&0x000003FF
        p.mem[adr_m1] = p.pipeline.Alu[stage].Res
        p.pc += 1

    case MTR, RTR:
        adr_r1 := cmd&0x000F0000>>16
        p.RF[adr_r1] = p.pipeline.Alu[stage].Res
        p.pc += 1 

    case SUB, SUM:
        adr_r3 := cmd&0x00000F00>>8
        p.RF[adr_r3] = p.pipeline.Alu[stage].Res
        p.pc += 1

    case JUMP_LESS:
        if p.pipeline.Alu[stage].Res == 1 {
            p.pc = uint16(cmd&0x000003FF)
            p.ignoreWR = 4
        }else {
            p.pc += 1
        }

    case MTRK:
        adr_r1 := cmd&0x000F0000>>16
        p.RF[adr_r1] = p.mem[p.pipeline.Alu[stage].Res]
        p.pc += 1

    case RTMK:
        adr_r2 := cmd&0x0000F000>>12
        p.mem[p.pipeline.Alu[stage].Res] = p.RF[adr_r2]
        p.pc += 1

    case JMP:
        p.ignoreWR = 4
        p.pc = p.pipeline.Alu[stage].Res

    default:
        p.pc += 1
    }
    if p.DebugMode {
        fmt.Printf("\nWRITEBACK\nOpCode: %v\nALU:\n %v\n", numToCommand(opCode), p.pipeline.Alu[stage].ToString())
    }
}

func (p *Pennywise700) Load(path string) {
    prog, err := os.Open(path)
    if err != nil {
        log.Println(err)
        return
    }
    scanner := bufio.NewScanner(prog)
    i := 0
    for scanner.Scan() {
        line := scanner.Text()
        cmd, err := strconv.ParseUint(line, 2, 32)
        if err != nil {
            log.Println(err)
            return
        }
        p.cmd_mem[i] = uint32(cmd)
        i++
    }
}

//SOME DEBUG PURPOSE FUNCTIONS
func (p *Pennywise700) GetMem() [1024]uint16 {
    return p.mem
}
func (p *Pennywise700) GetPc() uint16 {
    return p.pc
}
func (p *Pennywise700) GetCurCommand() uint32 {
    return p.cmd_mem[p.pc]
}
func (p *Pennywise700) GetPipeline() []uint32 {
    return p.pipeline.Pipe
}
func (p *Pennywise700) GetCommands() [1024]uint32 {
    return p.cmd_mem
}
func numToCommand(num uint32) string {
    switch num {
    case 0:
        return "NOP"
    case 1:
        return "LTM"
    case 2:
        return "MTR"
    case 3:
        return "RTR"
    case 4:
        return "SUB"
    case 5:
        return "JUMP_LESS"
    case 6:
        return "MTRK"
    case 7:
        return "RTMK"
    case 8:
        return "JMP"
    case 9:
        return "SUM"
    default:
        return "ERROR"
    }
}

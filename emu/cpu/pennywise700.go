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
    pc_stop  bool
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
    if p.pc_stop {
        p.pc-=1
        p.pc_stop = false
    }
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

    //fetching current command on stage one
    p.pipeline.Mtx.Lock()
    defer p.pipeline.Mtx.Unlock()
    opCode := p.pipeline.FetchOpcode(stage)

    //Read potential operands to be read from current stage 
    r_adr_r, r_skip := p.pipeline.GetReadOpsD1(stage) 

    //Read potential operands to be writen from next stage (Decode 2)
    _, wD2_adr_r, d2_skip := p.pipeline.GetWriteOps(stage+1)
    //Read potential operands to be writen from execute stage
    _, wE_adr_r, e_skip := p.pipeline.GetWriteOps(stage+2)
    //Read potential operands to be writen from WB stage
    _, wB_adr_r, wB_skip := p.pipeline.GetWriteOps(stage+3)

    //Checking read-write conflicts
    if ((r_adr_r == wD2_adr_r && !d2_skip) || (r_adr_r == wE_adr_r && !e_skip)) && !r_skip { 
        p.pipeline.M3 = true
        p.pc_stop = true 
        return
    }

    //execute command with alu
    switch opCode {
    case LTM:
        p.pipeline.Alu[stage].Op1 = p.pipeline.DecodeLiteral(stage) 

    case RTR, MTRK:
        if r_adr_r == wB_adr_r && !r_skip && !wB_skip{
            p.pipeline.Alu[stage].Op1 = p.pipeline.Alu[stage+3].Res
            if p.pipeline.FetchOpcode(stage+3) == MTRK {
                p.pipeline.Alu[stage].Op1 = p.mem[p.pipeline.Alu[stage+3].Res]
            }
            if p.DebugMode {
                fmt.Println("Write back was executed")
            }
        } else {
            adr_r2 := p.pipeline.DecodeAdrR2(stage)
            p.pipeline.Alu[stage].Op1 = p.RF[adr_r2]
        }

    case SUB, JUMP_LESS, SUM, RTMK:
        if r_adr_r == wB_adr_r {
            p.pipeline.Alu[stage].Op1 = p.pipeline.Alu[stage+3].Res
            if p.DebugMode {
                fmt.Println("Write back was executed")
            }
        } else {
            adr_r1 := p.pipeline.DecodeAdrR1(stage)
            p.pipeline.Alu[stage].Op1 = p.RF[adr_r1]
        }

    case JMP:
        p.pipeline.Alu[stage].Op1 = p.pipeline.DecodeAdrToJump(stage) 
    }
    if p.DebugMode {
         fmt.Printf("\nDECODE OP1\nCMD: %v\nALU:\n %v\n", p.pipeline.CommandToString(stage), p.pipeline.Alu[stage].ToString())
    }
}

//DECODE OP 2
func (p *Pennywise700) stageTwo() {
    stage := 2
    p.pipeline.Mtx.Lock()
    defer p.pipeline.Mtx.Unlock()
    opCode := p.pipeline.FetchOpcode(stage)

    //Read potential operands to be read from current stage 
    r_adr_m, r_adr_r, r_skip := p.pipeline.GetReadOpsD2(stage)
    //Read potential operands to be writen from execute stage
    wE_adr_m, wE_adr_r, e_skip := p.pipeline.GetWriteOps(stage+1)
    //Read potential operands to be writen from WB stage
    wB_adr_m, wB_adr_r, wB_skip := p.pipeline.GetWriteOps(stage+2)

    if !r_skip && !e_skip && (r_adr_r == wE_adr_r)  {
        p.pipeline.M4 = true  
        p.pc_stop = true
        return
    }
    if opCode == MTR && (p.pipeline.FetchOpcode(stage+1) == RTMK) && (r_adr_m == p.RF[wE_adr_m]) && !r_skip && !e_skip{
        p.pipeline.M4 = true  
        p.pc_stop = true
        return
    }

    //execute command with alu
    switch opCode {
    case SUB, SUM, JUMP_LESS:
        if r_adr_r == wB_adr_r && !r_skip && !wB_skip {
            p.pipeline.Alu[stage].Op2 = p.pipeline.Alu[stage+2].Res
            if p.pipeline.FetchOpcode(stage+2) == MTRK {
                p.pipeline.Alu[stage].Op2 = p.mem[p.pipeline.Alu[stage+2].Res]
            }
            if p.DebugMode {
                fmt.Println("Write back was executed")
            }
        }else {
            adr_r2 := p.pipeline.DecodeAdrR2(stage) 
            p.pipeline.Alu[stage].Op2 = p.RF[adr_r2]
        }

    case MTR:
        if !r_skip && !wB_skip && 
            (r_adr_m == wB_adr_m || p.pipeline.FetchOpcode(stage+2) == RTMK && r_adr_m == p.RF[wB_adr_m]) {
            p.pipeline.Alu[stage].Op1 = p.pipeline.Alu[stage+2].Res
            if p.DebugMode {
                fmt.Println("Write back was executed")
            }
        }else {
            adr_m := p.pipeline.DecodeAdrM(stage)
            p.pipeline.Alu[stage].Op1 = p.mem[adr_m] //In this command can write to any operand
        }
    }

    if p.DebugMode {
        fmt.Printf("\nDECODE OP2\nOpCode: %v\nALU:\n %v\n", p.pipeline.CommandToString(stage), p.pipeline.Alu[stage].ToString())
    }
}

//EXECUTE
func (p *Pennywise700) stageThree() {
    stage := 3

    //fetching current command on stage three 
    p.pipeline.Mtx.Lock()
    defer p.pipeline.Mtx.Unlock()
    opCode := p.pipeline.FetchOpcode(stage)


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
        fmt.Printf("\nEXECUTE\nOpCode: %v\nALU:\n %v\n", p.pipeline.CommandToString(stage), p.pipeline.Alu[stage].ToString())
    }
}

//WRITEBACK
func (p *Pennywise700) stageFour() {
    stage := 4
    //fetching current command on stage four 
    p.pipeline.Mtx.Lock()
    defer p.pipeline.Mtx.Unlock()
    opCode := p.pipeline.FetchOpcode(stage) 

    //execute command with alu
    switch opCode{
    case LTM:
        adr_m1 := p.pipeline.DecodeAdrM(stage) 
        p.mem[adr_m1] = p.pipeline.Alu[stage].Res
        p.pc += 1

    case MTR, RTR:
        adr_r1 := p.pipeline.DecodeAdrR1(stage)
        p.RF[adr_r1] = p.pipeline.Alu[stage].Res
        p.pc += 1 

    case SUB, SUM:
        adr_r3 := p.pipeline.DecodeAdrR3(stage)
        p.RF[adr_r3] = p.pipeline.Alu[stage].Res
        p.pc += 1

    case JUMP_LESS:
        if p.pipeline.Alu[stage].Res == 1 {
            p.pc = p.pipeline.DecodeAdrToJump(stage) 
            //Ignore all commands in pipe if jump
            p.pipeline.DropPipe()
        }else {
            p.pc += 1
        }

    case MTRK:
        adr_r1 := p.pipeline.DecodeAdrR1(stage)
        p.RF[adr_r1] = p.mem[p.pipeline.Alu[stage].Res]
        p.pc += 1

    case RTMK:
        adr_r2 := p.pipeline.DecodeAdrR2(stage) 
        p.mem[p.pipeline.Alu[stage].Res] = p.RF[adr_r2]
        p.pc += 1

    case JMP:
        //Ignore all commands in pipe if jump
        p.pipeline.DropPipe()
        p.pc = p.pipeline.Alu[stage].Res

    default:
        p.pc += 1
    }
    if p.DebugMode {
        fmt.Printf("\nWRITEBACK\nOpCode: %v\nALU:\n %v\n", p.pipeline.CommandToString(stage), p.pipeline.Alu[stage].ToString())
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
func (p *Pennywise700) GetPipeline() [5]string {
    return p.pipeline.PipeToString()
}
func (p *Pennywise700) GetCommands() [1024]uint32 {
    return p.cmd_mem
}

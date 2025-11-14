package main 

import (
	"fmt"
	"os"
    "github.com/Tyulenb/Pennywise700/cpu"
)

func main() {
    args := os.Args
    path := "program.txt"
    debugMode := false
    if len(args) >= 2 {
        path = args[1]
        if len(args) > 2 && args[2] == "d" {
            debugMode = true
        }
    }else {
        fmt.Println("FORMAT cmd.go 'path to your program' 'd (optionaly for debug)'\n"+
        "go run cmd.go program.txt\ngo run cmd.go program.txt d (for debug)")
        return
    }
    p := cpu.NewPennywise700()
    p.DebugMode = debugMode
    p.Load(path)
    if debugMode {
        Debug(p)
    }else {
        Run(p)
    }
}

func Run(p *cpu.Pennywise700) {
    for range 1024 {
        p.EmulateCycle()
    }
    mem := p.GetMem()
    fmt.Println("MEM[0:10]",mem[0:10])
}

func Debug(p *cpu.Pennywise700) {
    fmt.Println("Enter - step for next command\nq - to exit")
    var arg string = "s"
    commands := map[uint32]string{
        0x0: "NOP",
        0x1: "LTM",
        0x2: "MTR",
        0x3: "RTR",
        0x4: "SUB",
        0x5: "JUMP_LESS",
        0x6: "MTRK",
        0x7: "RTMK",
        0x8: "JMP",       
        0x9: "SUM",     
    }
    for { 
        fmt.Scanln(&arg)
        switch(arg) {
        default:
            fmt.Println("PC:", p.GetPc())
            cmd_mem := p.GetCommands()
            fmt.Println("Commands [0:10]", cmd_mem[0:10])
            fmt.Println("Cur command", commands[p.GetCurCommand()>>20])
            p.EmulateCycle()
            fmt.Println("Cycle Results:")
            mem := p.GetMem()
            pipe := p.GetPipeline()
            fmt.Println("MEM[0:10]",mem[0:10])
            fmt.Println("REGS:", p.RF)
            fmt.Println("PIPE:", pipe)

        case "q":
            return
        }
    }
}

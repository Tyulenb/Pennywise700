package main

import (
	"bufio"
	"fmt"
	"os"
    "github.com/Tyulenb/Pennywise700/translator/internal"
)

func main() {
    args := os.Args
    if len(args) != 3 {
        fmt.Println("FORMAT main.go 'path to your assembly language' 'output file'")
        return
    }
    in := args[1]
    out := args[2]
    coms, err := internal.Assemble(in)
    if err != nil {
        fmt.Println(err)
        return
    }

    file, err := os.Create(out)
    if err != nil {
        fmt.Println(err)
        return
    }
    defer file.Close()

    writer := bufio.NewWriter(file)

    for i := range coms {
        bin := toBin(coms[i])
        _, err := writer.WriteString(bin+"\n")
        if err != nil {
            fmt.Println(err)
            return
        }
    }
    writer.Flush()
}

func toBin(num uint32) string {
    var result string
    for i := 31; i >= 8; i-- {
        if num&(1<<i) != 0 {
            result += "1"
        }else {
            result += "0"
        }
    }
    return result
}

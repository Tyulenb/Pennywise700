package main

import (
	"bufio"
	"fmt"
	"os"
)

func main() {
    coms, err := Assemble("insertion_sort9.txt")
    if err != nil {
        fmt.Println(err)
        return
    }

    file, err := os.Create("program.txt")
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

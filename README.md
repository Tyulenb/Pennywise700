# Pennywise700 - Pipeline Processor Emulator

Pennywise700 is a simple pipeline processor emulator designed for educational purposes, allowing users to understand basic pipeline processor design.
The main feature of pipeline processors is simultaneous processing of multiple commands. However there are some read-write conflicts need to be handled.

## Commands Description

| Command     | Description                                        | Format                                | Pseudocode                                        |
| ----------- | -------------------------------------------------- | ------------------------------------- | ------------------------------------------------- |
| `NOP`       | Do Nothing                                         | [OpCode]                              | void do(){}                                       |
| `LTM`       | Load literal to memory                             | [OpCode][literal][adr_m]              | mem[adr_m] = literal                              |
| `MTR`       | Load memory to register                            | [OpCode][adr_r1][adr_m]               | RF[adr_r1] = mem[adr_m]                           |
| `RTR`       | Load register to register                          | [OpCode][adr_r1][adr_r2]              | RF[adr_r1] = RF[adr_r2]                           |
| `SUB`       | Subtract                                           | [OpCode][adr_r1][adr_r2][adr_r3]      | RF[adr_r3] = RF[adr_r1] - RF[adr_r2]              |
| `JUMP_LESS` | Jump to another operation on condition             | [OpCode][adr_r1][adr_r2][adr_to_jump] | if(RF[adr_r1] >= RF[adr_r2]) { pc = adr_to_jump } |
| `MTRK`      | Load register from memory by address from register | [OpCode][adr_r1][adr_r2]              | RF[adr_r1] = mem[RF[adr_r2]]                      |
| `RTMK`      | Load memory from register by address from register | [OpCode][adr_r1][adr_r2]              | mem[RF[adr_r1]] = RF[adr_r2]                      |
| `JMP`       | Jump to another operation                          | [OpCode][adr_to_jump]                 | pc = adr_to_jump                                  |
| `SUM`       | Add                                                | [OpCode][adr_r1][adr_r2][adr_r3]      | RF[adr_r3] = RF[adr_r1]+RF[adr_r2]                |

## Command Stage Description

| Command     | Fetch       | Decode 1       | Decode 2       | Execute     | Write Back                                                      |
| ----------- | ----------- | -------------- | -------------- | ----------- | --------------------------------------------------------------- |
| `NOP`       | cmd_mem[pc] | -              | -              | -           | -                                                               |
| `LTM`       | cmd_mem[pc] | op1=literal    | -              | res=op1     | mem[adr_m]=res<br>pc+=1                                         |
| `MTR`       | cmd_mem[pc] | -              | op1=mem[adr_m] | res=op1     | RF[adr_r1]=res<br>pc+=1                                         |
| `RTR`       | cmd_mem[pc] | op1=RF[adr_r2] | -              | res=op1     | RF[adr_r1]=res<br>pc+=1                                         |
| `SUB`       | cmd_mem[pc] | op1=RF[adr_r1] | op2=RF[adr_r2] | res=op1-op2 | RF[adr_r3]=res<br>pc+=1                                         |
| `JUMP_LESS` | cmd_mem[pc] | op1=RF[adr_r1] | op2=RF[adr_r2] | res=op1<op2 | if(RF[adr_r1] >= RF[adr_r2]) { pc = adr_to_jump }<br>else pc+=1 |
| `MTRK`      | cmd_mem[pc] | op1=RF[adr_r2] | -              | res=op1     | RF[adr_r1]=mem[res]<br>pc+=1                                    |
| `RTMK`      | cmd_mem[pc] | op1=RF[adr_r1] | -              | res=op1     | mem[res]=RF[adr_r2]<br>pc+=1                                    |
| `JMP`       | cmd_mem[pc] | op1=adr_to_jmp | -              | res=op1     | pc=res                                                          |
| `SUM`       | cmd_mem[pc] | op1=RF[adr_r1] | op2=RF[adr_r2] | res=op1     | RF[adr_r3]=res<br>pc+=1                                         |


## Program Example
[Insertion_sort](https://github.com/Tyulenb/Pennywise700/blob/main/docs/insertion_sort.txt)
<br>
[Pipeline specifications of example](https://github.com/Tyulenb/Pennywise700/blob/main/docs/insertion_sort_pipe_specs.md)
## Installation
### Requirements
Go version 1.20 or higher
### Steps
```bash
git clone https://github.com/Tyulenb/Pennywise700.git
cd Pennywise700/emu
go run cmd/cmd.go "path to your program"
```
### Debug Mode
To enable debug mode, you can use the d flag:
```bash
go run cmd/cmd.go "path to your program" d
```

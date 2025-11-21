# Insertion Sort Pipeline specification
**BOLD** - Take data from write back

| Fetch     | Decode 1      | Decode 2    | Execute   | Write Back    |
| --------- | ------------- | ----------- | --------- | ------------- |
| LTM 0,5   |               |             |           |               |
| LTM 7,1   | LTM 0,5       |             |           |               |
| LTM 4,2   | LTM 7,1       | LTM 0,5     |           |               |
| LTM 3,3   | LTM 4,2       | LTM 7,1     | LTM 0,5   |               |
| RTR 2,1   | LTM 3,3       | LTM 4,2     | LTM 7,1   | LTM 0,5       |
| MTR 3,3   | RTR 2,1       | LTM 3,3     | LTM 4,2   | LTM 7,1       |
| NOP       | MTR 3,3       | RTR 2,1     | LTM 3,3   | LTM 4,2       |
| J_L 2,3   | NOP           | **MTR 3,3** | RTR 2,1   | **LTM 3,3**   |
| MTRK 4,2  | **J_L 2,3**   | NOP         | MTR 3,3   | **RTR 2,1**   |
| RTR 5,2   | MTRK 4,2      | **J_L 2,3** | NOP       | **MTR 3,3**   |
| NOP       | RTR 5,2       | MTRK 4,2    | J_L 2,3   | NOP           |
| J_L 0,5   | NOP           | RTR 5,2     | MTRK 4,2  | J_L 2,3       |
| SUB 5,1,6 | J_L 0,5       | NOP         | RTR 5,2   | MTRK 4,2      |
| NOP       | **SUB 5,1,6** | **J_L 0,5** | NOP       | **RTR 5,2**   |
| NOP       | NOP           | SUB 5,1,6   | J_L 0,5   | NOP           |
| MTRK 7,6  | NOP           | NOP         | SUB 5,1,6 | J_L 0,5       |
| NOP       | **MTRK 7,6**  | NOP         | NOP       | **SUB 5,1,6** |
| J_L 4,7   | NOP           | MTRK 7,6    | NOP       | NOP           |
| MTRK 8,5  | J_L 4,7       | NOP         | MTRK 7,6  | NOP           |
| RTMK 6,8  | MTRK 8,5      | J_L 4,7     | NOP       | MTRK 7,6      |
| RTMK 5,7  | RTMK 6,8      | MTRK 8,5    | J_L 4,7   | NOP           |
| RTR 5,6   | RTMK 5,7      | RTMK 6,8    | MTRK 8,5  | J_L 4,7       |
| JMP 9     | RTR 5,6       | RTMK 5,7    | RTMK 6,8  | MTRK 8,5      |
|           | JMP 9         | RTR 5,6     | RTMK 5,7  | RTMK 6,8      |
|           |               | JMP 9       | RTR 5,6   | RTMK 5,7      |
|           |               |             | JMP 9     | RTR 5,6       |
| ...       | ...           | ...         | ...       | JMP 9         |
| RTMK 5,4  |               |             |           |               |
| SUM 2,1,2 | RTMK 5,4      |             |           |               |
| JMP 6     | SUM 2,1,2     | RTMK 5,4    |           |               |
|           | JMP 6         | SUM 2,1,2   | RTMK 5,4  |               |
|           |               | JMP 6       | SUM 2,1,2 | RTMK 5,4      |
|           |               |             | JMP 6     | SUM 2,1,2     |
|           |               |             |           | JMP 6         |
| ...       | ...           | ...         | ...       | ...           |
| NOP       | NOP           | NOP         | NOP       | NOP           |


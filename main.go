package main

import (
	"fmt"
	"go-4004/cpu"
)

func main() {
	c := cpu.NewCPU4004()

	program := []byte{
		cpu.LDM(7),      // A = 7
		cpu.XCH(cpu.R1), // R1 = 7, A = 0

		cpu.LDM(3),      // A = 3
		cpu.XCH(cpu.R2), // R2 = 3, A = 0

		cpu.LD(cpu.R1),  // A = R1 = 7
		cpu.ADD(cpu.R2), // A = 7 + 3 = 10
	}

	fmt.Println("=== BEFORE ===")
	printCPU(c)

	for i, op := range program {
		fmt.Printf("\nSTEP %d\n", i)
		fmt.Printf("Executing opcode: 0x%02X\n", op)

		if err := c.Execute(op); err != nil {
			panic(err)
		}

		printCPU(c)
	}

	fmt.Println("\n=== FINAL STATE ===")
	printCPU(c)
}

func printCPU(c *cpu.CPU4004) {
	fmt.Printf("A=%d C=%v\n", c.A, c.C)

	for i := 0; i < 16; i++ {
		fmt.Printf("R%X=%d ", i, c.R[i])

		if (i+1)%4 == 0 {
			fmt.Println()
		}
	}
}

package main

import (
	"fmt"
	"go-4004/cpu"
)

func main() {
	c := cpu.NewCPU4004()

	// addizione BCD: 8 + 5 = 13 (decimale)
	program := []byte{
		cpu.LDM(8),      // A = 8
		cpu.XCH(cpu.R0), // R0 = 8
		cpu.LDM(5),      // A = 5
		cpu.ADD(cpu.R0), // A = 13 (0xD), C = false — risultato binario
		cpu.DAA(),       // A = 3, C = true  — corretto in BCD
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

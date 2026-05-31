package main

import (
	"fmt"
	"go-4004/cpu"
)

func main() {
	c := cpu.NewCPU4004()

	// 30 + 18 = 48
	// Le cifre sono memorizzate in registri separati (BCD: una cifra per nibble)
	// 30 → R0=3 (decine), R1=0 (unità)
	// 18 → R2=1 (decine), R3=8 (unità)

	// moltiplica 3 per 2 usando RAL (shift left = *2)
	program := []byte{
		cpu.LDM(3), // A = 0011 (3)
		cpu.CLC(),  // C = false
		cpu.RAL(),  // A = 0110 (6), C = false
		cpu.RAL(),  // A = 1100 (12), C = false
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

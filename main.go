package main

import (
	"fmt"
	"go-4004/cpu"
)

func main() {
	c := cpu.NewCPU4004()

	// KBP: decodifica tasti da one-hot a posizione
	// Simula la lettura di tre colonne di tastiera
	program := []byte{
		cpu.LDM(0b0001), // A = colonna 1 premuta (one-hot)
		cpu.KBP(),       // A = 1
		cpu.XCH(cpu.R0), // salva in R0

		cpu.LDM(0b0100), // A = colonna 3 premuta (one-hot)
		cpu.KBP(),       // A = 3
		cpu.XCH(cpu.R1), // salva in R1

		cpu.LDM(0b0110), // A = due colonne (input non valido)
		cpu.KBP(),       // A = 0xF (errore)
		cpu.XCH(cpu.R2), // salva in R2
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

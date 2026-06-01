package main

import (
	"fmt"
	"go-4004/cpu"
)

func main() {
	c := cpu.NewCPU4004()

	// 30 + 18 = 48 (BCD, cifra per cifra)
	rom := cpu.NewROM([]byte{
		cpu.LDM(3), cpu.XCH(cpu.R0), // R0 = 3 (decine di 30)
		cpu.LDM(0), cpu.XCH(cpu.R1), // R1 = 0 (unità di 30)
		cpu.LDM(1), cpu.XCH(cpu.R2), // R2 = 1 (decine di 18)
		cpu.LDM(8), cpu.XCH(cpu.R3), // R3 = 8 (unità di 18)

		cpu.LD(cpu.R1), cpu.ADD(cpu.R3), cpu.DAA(), cpu.XCH(cpu.R5), // unità: 0+8=8
		cpu.LD(cpu.R0), cpu.ADD(cpu.R2), cpu.DAA(), cpu.XCH(cpu.R4), // decine: 3+1=4
	})

	fmt.Println("=== BEFORE ===")
	printCPU(c)

	for i := range rom.Data {
		fmt.Printf("\nSTEP %d — PC=%03X OP=0x%02X\n", i, c.PC, rom.Data[c.PC])
		if err := c.Step(rom); err != nil {
			panic(err)
		}
		printCPU(c)
	}

	fmt.Println("\n=== FINAL STATE ===")
	printCPU(c)
	fmt.Printf("Risultato BCD: R4=%d (decine) R5=%d (unità) → %d%d\n",
		c.R[cpu.R4], c.R[cpu.R5], c.R[cpu.R4], c.R[cpu.R5])
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

package main

import (
	"fmt"
	"go-4004/cpu"
)

func main() {
	c := cpu.NewCPU4004()

	// Demo BBL: ritorno da subroutine con valore in A.
	//
	// Layout ROM:
	//   0x000  LDM 9    ← programma principale (non eseguito in questo demo)
	//   0x001  LDM 3    ← (qui andrebbe JMS 0x002, non ancora implementato)
	//   0x002  BBL 7    ← corpo subroutine: ritorna con A=7
	//   0x003  NOP      ← prima istruzione dopo il ritorno
	//
	// Simuliamo JMS manualmente:
	//   Push(0x003) → salva l'indirizzo di ritorno nello stack
	//   PC = 0x002  → puntiamo direttamente alla subroutine
	//   Step()      → esegue BBL 7: A=7, PC ripristinato a 0x003

	rom := cpu.NewROM([]byte{
		cpu.LDM(9), // 0x000
		cpu.LDM(3), // 0x001
		cpu.BBL(7), // 0x002 — subroutine: ritorna con A=7
		cpu.NOP(),  // 0x003 — prima istruzione dopo il ritorno
	})

	c.Push(0x003) // simula JMS: salva indirizzo di ritorno
	c.PC = 0x002  // salta alla subroutine

	fmt.Println("=== BEFORE BBL ===")
	fmt.Printf("PC=0x%03X  A=%d  SP=%d\n", c.PC, c.A, c.SP)

	if err := c.Step(rom); err != nil {
		panic(err)
	}

	fmt.Println("\n=== AFTER BBL ===")
	fmt.Printf("PC=0x%03X  A=%d  SP=%d\n", c.PC, c.A, c.SP)
	fmt.Println("→ PC ripristinato a 0x003, A=7 (valore di ritorno)")
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

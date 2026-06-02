package main

import (
	"fmt"
	"go-4004/cpu"
)

func main() {
	// Demo: loop con ISZ + subroutine via JMS/BBL.
	//
	// Programma:
	//   0x000  FIM R2, 0x03  — R2=0, R3=3 (R3 = contatore loop = 3)
	//   0x002  JMS 0x010     — chiama subroutine a 0x010
	//   0x004  NOP           — prima istruzione dopo il ritorno
	//   0x005  ISZ R3, 0x02  — incrementa R3; se != 0 torna a 0x002 (top loop)
	//   0x007  JUN 0x000     — fine: ricomincia da capo (halt simulato)
	//
	//   subroutine @ 0x010:
	//   0x010  LDM 7         — A = 7 (risultato subroutine)
	//   0x011  BBL 5         — ritorna con A = 5

	rom := cpu.NewROM(make([]byte, 4096))

	// programma principale
	rom.Data[0x000] = cpu.FIM(cpu.R2) // 0x24
	rom.Data[0x001] = 0x0D            // R2=0x0, R3=0xD → loop termina dopo 3 incrementi (0xD→0xE→0xF→0)
	rom.Data[0x002] = cpu.JMS(0x0)   // 0x50: JMS 0x010
	rom.Data[0x003] = 0x10
	rom.Data[0x004] = cpu.NOP()
	rom.Data[0x005] = cpu.ISZ(cpu.R3) // salta a 0x002 se R3 != 0 dopo l'incremento
	rom.Data[0x006] = 0x02
	rom.Data[0x007] = cpu.JUN(0x0) // JUN 0x000 (halt loop)
	rom.Data[0x008] = 0x00

	// subroutine
	rom.Data[0x010] = cpu.LDM(7)
	rom.Data[0x011] = cpu.BBL(5)

	c := cpu.NewCPU4004()
	fmt.Println("=== Demo JMS / BBL / ISZ / FIM ===")
	fmt.Printf("Start: PC=0x%03X  A=%d  R2=%d R3=%d\n", c.PC, c.A, c.R[cpu.R2], c.R[cpu.R3])

	for step := 0; step < 30; step++ {
		pc := c.PC
		if err := c.Step(rom); err != nil {
			fmt.Printf("Errore al passo %d (PC=0x%03X): %v\n", step, pc, err)
			break
		}
		fmt.Printf("step %2d  PC=0x%03X → 0x%03X  A=%d  R2=%d R3=%d  SP=%d\n",
			step+1, pc, c.PC, c.A, c.R[cpu.R2], c.R[cpu.R3], c.SP)

		// halt: JUN 0x000 — il programma è tornato all'inizio
		if pc == 0x007 {
			fmt.Println("--- halt ---")
			break
		}
	}

	fmt.Printf("\nFine: A=%d (valore di ritorno BBL)\n", c.A)
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

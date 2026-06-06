package main

import (
	"fmt"
	"go-4004/cpu"
)

func main() {
	// Demo: WR0/WR3 — scrive nei nibble di stato della RAM.
	//
	//   Usa lo status per salvare due flag: segno (WR0) e overflow (WR3).

	rom := cpu.NewROM(make([]byte, 4096))
	ram := cpu.NewRAM()

	rom.Data[0x000] = cpu.LDM(0)
	rom.Data[0x001] = cpu.DCL()
	rom.Data[0x002] = cpu.FIM(cpu.R0)
	rom.Data[0x003] = 0x00
	rom.Data[0x004] = cpu.SRC(cpu.R0)
	rom.Data[0x005] = cpu.LDM(1) // flag segno = 1 (negativo)
	rom.Data[0x006] = cpu.WR0()
	rom.Data[0x007] = cpu.LDM(0) // flag overflow = 0
	rom.Data[0x008] = cpu.WR3()

	c := cpu.NewCPU4004()
	fmt.Println("=== Demo WR0 + WR3 ===")

	for i := 0; i < 9; i++ {
		if err := c.Step(rom, ram); err != nil {
			fmt.Printf("Errore: %v\n", err)
			break
		}
	}

	fmt.Printf("Status[0][0][0] = %d (segno, atteso 1)\n", ram.Status[0][0][0])
	fmt.Printf("Status[0][0][3] = %d (overflow, atteso 0)\n", ram.Status[0][0][3])
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

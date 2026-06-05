package main

import (
	"fmt"
	"go-4004/cpu"
)

func main() {
	// Demo: ADM — somma RAM + A con carry.
	//
	//   LDM 0 / DCL       → banco 0
	//   FIM R0, 0x00      → R0=0, R1=0 (registro 0, carattere 0)
	//   SRC R0            → SRCAddr=0x00
	//   LDM 6 / WRM       → ram.Data[0][0][0] = 6
	//   LDM 7             → A=7
	//   CLC               → C=false
	//   ADM               → A = 7 + 6 = 13 → nibble=13, C=false
	//   DAA               → A = 13+6=19 → nibble=3, C=true (correzione BCD)

	rom := cpu.NewROM(make([]byte, 4096))
	ram := cpu.NewRAM()

	rom.Data[0x000] = cpu.LDM(0)
	rom.Data[0x001] = cpu.DCL()
	rom.Data[0x002] = cpu.FIM(cpu.R0)
	rom.Data[0x003] = 0x00
	rom.Data[0x004] = cpu.SRC(cpu.R0)
	rom.Data[0x005] = cpu.LDM(6)
	rom.Data[0x006] = cpu.WRM()
	rom.Data[0x007] = cpu.LDM(7)
	rom.Data[0x008] = cpu.CLC()
	rom.Data[0x009] = cpu.ADM()
	rom.Data[0x00A] = cpu.DAA()

	c := cpu.NewCPU4004()
	fmt.Println("=== Demo ADM + DAA (7 + 6 BCD) ===")

	for range 11 {
		if err := c.Step(rom, ram); err != nil {
			fmt.Printf("Errore: %v\n", err)
			break
		}
	}

	fmt.Printf("A = %d, C = %v  (atteso A=3 C=true: 7+6=13 → BCD 3 con carry)\n", c.A, c.C)
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

package main

import (
	"fmt"
	"go-4004/cpu"
)

func main() {
	// Demo: SBM + WMP
	//
	//   Scrive 9 in RAM, poi calcola 5 - 9 (con borrow)
	//   e invia il risultato sulla porta di output.
	//
	//   LDM 0 / DCL       → banco 0
	//   FIM R0, 0x00 / SRC R0
	//   LDM 9 / WRM       → ram.Data[0][0][0] = 9
	//   LDM 5             → A=5
	//   STC               → C=true (nessun borrow iniziale)
	//   SBM               → A = 5 - 9 = -4 → nibble(12), C=false (borrow)
	//   WMP               → ram.Port[0] = 12

	rom := cpu.NewROM(make([]byte, 4096))
	ram := cpu.NewRAM()

	rom.Data[0x000] = cpu.LDM(0)
	rom.Data[0x001] = cpu.DCL()
	rom.Data[0x002] = cpu.FIM(cpu.R0)
	rom.Data[0x003] = 0x00
	rom.Data[0x004] = cpu.SRC(cpu.R0)
	rom.Data[0x005] = cpu.LDM(9)
	rom.Data[0x006] = cpu.WRM()
	rom.Data[0x007] = cpu.LDM(5)
	rom.Data[0x008] = cpu.STC()
	rom.Data[0x009] = cpu.SBM()
	rom.Data[0x00A] = cpu.WMP()

	c := cpu.NewCPU4004()
	fmt.Println("=== Demo SBM + WMP (5 - 9) ===")

	for i := 0; i < 11; i++ {
		if err := c.Step(rom, ram); err != nil {
			fmt.Printf("Errore: %v\n", err)
			break
		}
	}

	fmt.Printf("A = %d, C = %v  (atteso A=12 C=false: borrow generato)\n", c.A, c.C)
	fmt.Printf("Port[0] = %d     (atteso 12: valore inviato su porta output)\n", ram.Port[0])
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

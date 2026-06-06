// Programma: moltiplicazione 3 × 4 = 12 tramite addizioni ripetute.
//
// Dimostra:
//   - loop con ISZ (pattern base del 4004)
//   - uso dei registri come accumulatore e contatore
//   - scrittura del risultato in RAM
//
// Algoritmo:
//
//	R1 = 3   (addendo)
//	R4 = 12  (contatore loop = 16-4, così dopo 4 ISZ raggiunge 0 e il loop termina)
//	A  = 0   (accumulatore: sommerà R1 per 4 volte → 3+3+3+3 = 12)
package main

import (
	"fmt"
	"go-4004/cpu"
)

func main() {
	rom := cpu.NewROM(make([]byte, 256))
	ram := cpu.NewRAM()

	// Setup: seleziona banco RAM 0
	rom.Data[0x000] = cpu.LDM(0) // A = 0
	rom.Data[0x001] = cpu.DCL()  // CL = 0 (banco RAM 0)

	// Carica dati nei registri
	rom.Data[0x002] = cpu.FIM(cpu.R0) // FIM R0, 0x03 → R0=0, R1=3 (addendo)
	rom.Data[0x003] = 0x03
	rom.Data[0x004] = cpu.FIM(cpu.R2) // FIM R2, 0x00 → R2=0, R3=0 (indirizzo RAM)
	rom.Data[0x005] = 0x00
	rom.Data[0x006] = cpu.SRC(cpu.R2) // SRCAddr = 0x00 (banco 0, registro 0, pos 0)

	// Prepara loop: R4=12 (contatore), A=0 (accumulatore)
	rom.Data[0x007] = cpu.LDM(12)     // A = 12
	rom.Data[0x008] = cpu.XCH(cpu.R4) // R4=12, A=0

	// LOOP (indirizzo 0x009)
	rom.Data[0x009] = cpu.ADD(cpu.R1) // A = A + R1 (= A + 3)
	rom.Data[0x00A] = cpu.ISZ(cpu.R4) // R4++; se R4 != 0 → salta a 0x009
	rom.Data[0x00B] = 0x09            //   target: stesso indirizzo 0x009

	// Salva risultato
	rom.Data[0x00C] = cpu.WRM() // ram.Data[0][0][0] = A

	c := cpu.NewCPU4004()
	c.Trace = true

	fmt.Println("=== Moltiplicazione: 3 × 4 tramite addizioni ripetute ===")
	fmt.Println()

	// 7 step di setup + 4 iterazioni × 2 step (ADD + ISZ) + 1 WRM = 16 step totali
	for i := range 16 {
		if err := c.Step(rom, ram); err != nil {
			fmt.Printf("Errore a step %d: %v\n", i+1, err)
			return
		}
	}

	fmt.Println()
	fmt.Printf("A             = %d  (atteso: 12)\n", c.A)
	fmt.Printf("RAM[0][0][0]  = %d  (atteso: 12)\n", ram.Data[0][0][0])

	if c.A == 12 && ram.Data[0][0][0] == 12 {
		fmt.Println("✓ Corretto!")
	} else {
		fmt.Println("✗ Errore!")
	}
}

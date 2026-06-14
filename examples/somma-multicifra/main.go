// Programma: addizione BCD multi-cifra — calcola 47 + 58 = 105 con propagazione del carry.
package main

import (
	"fmt"
	"go-4004/cpu"
)

func main() {
	rom := cpu.NewROM(make([]byte, 256))
	ram := cpu.NewRAM()

	// --- SETUP ---
	rom.Data[0x000] = cpu.LDM(0) // A = 0 (per DCL)
	rom.Data[0x001] = cpu.DCL()  // CL = 0 → banco RAM 0

	// --- A = 47 nel registro RAM 0: char0=7 (unità), char1=4 (decine) ---
	rom.Data[0x002] = cpu.FIM(cpu.R0) // R0=0 (reg 0), R1=0 (char 0)
	rom.Data[0x003] = 0x00
	rom.Data[0x004] = cpu.SRC(cpu.R0) // seleziona [0][0][0]
	rom.Data[0x005] = cpu.LDM(7)      // unità di 47
	rom.Data[0x006] = cpu.WRM()       // RAM[0][0][0] = 7
	rom.Data[0x007] = cpu.FIM(cpu.R0) // R0=0, R1=1 (char 1)
	rom.Data[0x008] = 0x01
	rom.Data[0x009] = cpu.SRC(cpu.R0) // seleziona [0][0][1]
	rom.Data[0x00A] = cpu.LDM(4)      // decine di 47
	rom.Data[0x00B] = cpu.WRM()       // RAM[0][0][1] = 4

	// --- B = 58 nel registro RAM 1: char0=8 (unità), char1=5 (decine) ---
	rom.Data[0x00C] = cpu.FIM(cpu.R2) // R2=1 (reg 1), R3=0 (char 0)
	rom.Data[0x00D] = 0x10
	rom.Data[0x00E] = cpu.SRC(cpu.R2) // seleziona [0][1][0]
	rom.Data[0x00F] = cpu.LDM(8)      // unità di 58
	rom.Data[0x010] = cpu.WRM()       // RAM[0][1][0] = 8
	rom.Data[0x011] = cpu.FIM(cpu.R2) // R2=1, R3=1 (char 1)
	rom.Data[0x012] = 0x11
	rom.Data[0x013] = cpu.SRC(cpu.R2) // seleziona [0][1][1]
	rom.Data[0x014] = cpu.LDM(5)      // decine di 58
	rom.Data[0x015] = cpu.WRM()       // RAM[0][1][1] = 5

	// --- INIT LOOP: prepara i tre puntatori (tutti a char 0) ---
	rom.Data[0x016] = cpu.FIM(cpu.R0) // R0=0 (reg A), R1=0 (char) — riporta A a char 0
	rom.Data[0x017] = 0x00
	rom.Data[0x018] = cpu.FIM(cpu.R2) // R2=1 (reg B), R3=0 (char) — riporta B a char 0
	rom.Data[0x019] = 0x10
	rom.Data[0x01A] = cpu.FIM(cpu.R4) // R4=2 (reg risultato), R5=0 (char)
	rom.Data[0x01B] = 0x20
	rom.Data[0x01C] = cpu.FIM(cpu.R6) // R6=14 (contatore: 16-2 cifre), R7=0
	rom.Data[0x01D] = 0xE0
	rom.Data[0x01E] = cpu.CLC() // C = 0 → nessun riporto entra nelle unità

	// --- LOOP (0x01F): per ogni cifra → A_cifra + B_cifra + riporto, corretto BCD ---
	rom.Data[0x01F] = cpu.SRC(cpu.R0) // seleziona la cifra di A (reg 0, char R1)
	rom.Data[0x020] = cpu.RDM()       // A = cifra di A
	rom.Data[0x021] = cpu.SRC(cpu.R2) // seleziona la cifra di B (reg 1, char R3)
	rom.Data[0x022] = cpu.ADM()       // A = cifra_A + cifra_B + C  (C = riporto precedente)
	rom.Data[0x023] = cpu.DAA()       // correzione BCD: cifra giusta + nuovo riporto in C
	rom.Data[0x024] = cpu.SRC(cpu.R4) // seleziona la cifra del risultato (reg 2, char R5)
	rom.Data[0x025] = cpu.WRM()       // RAM[risultato][char] = cifra
	rom.Data[0x026] = cpu.INC(cpu.R1) // avanza il puntatore di A
	rom.Data[0x027] = cpu.INC(cpu.R3) // avanza il puntatore di B
	rom.Data[0x028] = cpu.INC(cpu.R5) // avanza il puntatore del risultato
	rom.Data[0x029] = cpu.ISZ(cpu.R6) // R6++; se != 0 → torna al LOOP
	rom.Data[0x02A] = 0x1F            //   target: 0x01F

	// --- dopo il loop: l'ultimo riporto è la cifra delle centinaia ---
	rom.Data[0x02B] = cpu.TCC()       // A = C (1 se c'è riporto), poi C = 0
	rom.Data[0x02C] = cpu.SRC(cpu.R4) // R4=2, R5=2 (già avanzato) → [0][2][2] = centinaia
	rom.Data[0x02D] = cpu.WRM()       // RAM[0][2][2] = cifra centinaia

	// --- HALT ---
	rom.Data[0x02E] = cpu.JUN(0x0) // JUN 0x02E → salta a se stesso
	rom.Data[0x02F] = 0x2E

	const haltAddr = 0x02E

	c := cpu.NewCPU4004()
	c.Trace = true

	fmt.Println("=== Addizione BCD multi-cifra: 47 + 58 ===")
	fmt.Println()

	for {
		if err := c.Step(rom, ram); err != nil {
			fmt.Printf("Errore: %v\n", err)
			return
		}
		if c.PC == haltAddr {
			break
		}
	}

	u := ram.Data[0][2][0] // unità
	d := ram.Data[0][2][1] // decine
	h := ram.Data[0][2][2] // centinaia

	fmt.Println()
	fmt.Printf("Risultato in RAM (reg 2): centinaia=%d decine=%d unità=%d\n", h, d, u)

	got := int(h)*100 + int(d)*10 + int(u)
	fmt.Printf("47 + 58 = %d  (atteso: 105)\n", got)

	if got == 105 {
		fmt.Println("✓ Corretto!")
	} else {
		fmt.Println("✗ Errore!")
	}
}

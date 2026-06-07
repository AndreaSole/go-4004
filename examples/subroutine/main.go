// Programma: subroutine con JMS/BBL — calcola 3 + 5 chiamando una funzione.
//
// Dimostra:
//   - chiamata a subroutine (JMS: push PC sullo stack, salto)
//   - ritorno da subroutine (BBL: pop PC dallo stack, carica immediato in A)
//   - la trappola di BBL: sovrascrive sempre A con un valore immediato,
//     quindi il risultato calcolato va salvato in un registro PRIMA del ritorno
//   - la convenzione "halt" (JUN a se stesso) per fermare l'esecuzione
//
// Algoritmo:
//
//	MAIN:
//	  setup RAM (banco 0, indirizzo 0x00) e operandi (R0=3, R1=5)
//	  JMS SOMMA        → chiama la subroutine
//	  LD R5            → recupera il risultato salvato dalla subroutine
//	  WRM              → scrive il risultato in RAM
//	HALT:
//	  JUN HALT         → loop infinito su se stesso (fine programma)
//
//	SOMMA:
//	  A = R0 + R1      → 3 + 5 = 8
//	  XCH R5           → salva il risultato in R5 (sopravvive al ritorno)
//	  BBL 0            → torna al chiamante (A viene sovrascritto con 0)
package main

import (
	"fmt"
	"go-4004/cpu"
)

func main() {
	rom := cpu.NewROM(make([]byte, 256))
	ram := cpu.NewRAM()

	// --- MAIN ---

	// Setup: seleziona banco RAM 0, indirizzo 0x00
	rom.Data[0x000] = cpu.LDM(0)      // A = 0
	rom.Data[0x001] = cpu.DCL()       // CL = 0 (banco RAM 0)
	rom.Data[0x002] = cpu.FIM(cpu.R2) // FIM R2, 0x00 → R2=0, R3=0
	rom.Data[0x003] = 0x00
	rom.Data[0x004] = cpu.SRC(cpu.R2) // SRCAddr = 0x00

	// Operandi della somma: R0=3, R1=5
	rom.Data[0x005] = cpu.FIM(cpu.R0) // FIM R0, 0x35 → R0=3, R1=5
	rom.Data[0x006] = 0x35

	// Chiama la subroutine SOMMA (a 0x00D)
	rom.Data[0x007] = cpu.JMS(0x0) // JMS 0x00D → push PC=0x009, salta a 0x00D
	rom.Data[0x008] = 0x0D

	// Al ritorno: recupera il risultato salvato dalla subroutine e lo scrive in RAM
	rom.Data[0x009] = cpu.LD(cpu.R5) // A = R5 (risultato)
	rom.Data[0x00A] = cpu.WRM()      // ram.Data[0][0][0] = A

	// --- HALT: loop infinito su se stesso ---
	const haltAddr = 0x00B
	rom.Data[0x00B] = cpu.JUN(0x0) // JUN 0x00B → salta a se stesso
	rom.Data[0x00C] = 0x0B

	// --- SUBROUTINE: SOMMA (calcola R0 + R1) ---
	rom.Data[0x00D] = cpu.LD(cpu.R0)  // A = R0 (3)
	rom.Data[0x00E] = cpu.ADD(cpu.R1) // A = A + R1 = 3 + 5 = 8
	rom.Data[0x00F] = cpu.XCH(cpu.R5) // R5 = 8 — salva il risultato PRIMA di tornare!
	rom.Data[0x010] = cpu.BBL(0)      // pop stack, torna al chiamante (A sovrascritto a 0)

	c := cpu.NewCPU4004()
	c.Trace = true

	fmt.Println("=== Subroutine: 3 + 5 tramite JMS/BBL ===")
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

	fmt.Println()
	fmt.Printf("RAM[0][0][0] = %d  (atteso: 8)\n", ram.Data[0][0][0])
	fmt.Printf("R5           = %d  (atteso: 8 — il risultato sopravvive nel registro)\n", c.R[5])
	fmt.Printf("A            = %d  (atteso: 8 — BBL lo aveva azzerato, ma LD R5 lo sovrascrive subito dopo)\n", c.A)

	if ram.Data[0][0][0] == 8 && c.R[5] == 8 {
		fmt.Println("✓ Corretto!")
	} else {
		fmt.Println("✗ Errore!")
	}
}

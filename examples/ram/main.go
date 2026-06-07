// Programma: riempire un array in RAM — scrive 1, 2, 3, 4 in celle consecutive.
//
// Dimostra:
//   - aggiornamento dinamico di SRCAddr dentro un loop (pattern "scrivi un array")
//   - INC come strumento per far avanzare sia il valore che la posizione
//   - perché SRC va richiamato ogni volta che cambia l'indirizzo RAM
//   - la convenzione "halt" (JUN a se stesso)
//
// Algoritmo:
//
//	setup:
//	  CL = 0
//	  R2:R3 = 0x00     (indirizzo di partenza: registro 0, posizione 0)
//	  R0:R1 = 0x01     (R1 = primo valore da scrivere = 1)
//	  R4 = 12          (contatore loop = 16-4, tramite FIM R4, 0xC0)
//
//	LOOP:
//	  SRC R2           → invia l'indirizzo corrente (R2:R3) alla RAM
//	  LD R1            → A = valore da scrivere
//	  WRM              → scrivi A nella cella puntata da SRCAddr
//	  INC R1           → valore++   (prossimo numero da scrivere)
//	  INC R3           → posizione++ (prossima cella)
//	  ISZ R4, LOOP     → ripeti finché R4 non torna a 0
//
//	HALT:
//	  JUN HALT
package main

import (
	"fmt"
	"go-4004/cpu"
)

func main() {
	rom := cpu.NewROM(make([]byte, 256))
	ram := cpu.NewRAM()

	// --- Setup ---
	rom.Data[0x000] = cpu.LDM(0)      // A = 0
	rom.Data[0x001] = cpu.DCL()       // CL = 0 (banco RAM 0)
	rom.Data[0x002] = cpu.FIM(cpu.R2) // FIM R2, 0x00 → R2=0, R3=0 (indirizzo iniziale)
	rom.Data[0x003] = 0x00
	rom.Data[0x004] = cpu.FIM(cpu.R0) // FIM R0, 0x01 → R0=0, R1=1 (primo valore)
	rom.Data[0x005] = 0x01
	rom.Data[0x006] = cpu.FIM(cpu.R4) // FIM R4, 0xC0 → R4=12, R5=0 (contatore loop = 16-4)
	rom.Data[0x007] = 0xC0

	// --- LOOP (indirizzo 0x008) ---
	rom.Data[0x008] = cpu.SRC(cpu.R2) // invia (R2:R3) alla RAM come indirizzo corrente
	rom.Data[0x009] = cpu.LD(cpu.R1)  // A = valore da scrivere
	rom.Data[0x00A] = cpu.WRM()       // RAM[CL][R2][R3] = A
	rom.Data[0x00B] = cpu.INC(cpu.R1) // valore++
	rom.Data[0x00C] = cpu.INC(cpu.R3) // posizione++
	rom.Data[0x00D] = cpu.ISZ(cpu.R4) // R4++; se !=0 → torna all'inizio del loop
	rom.Data[0x00E] = 0x08            //   target: 0x008 (SRC R2, inizio LOOP)

	// --- HALT: loop infinito su se stesso ---
	const haltAddr = 0x00F
	rom.Data[0x00F] = cpu.JUN(0x0) // JUN 0x00F → salta a se stesso
	rom.Data[0x010] = 0x0F

	c := cpu.NewCPU4004()
	c.Trace = true

	fmt.Println("=== Riempire un array in RAM: scrive 1,2,3,4 in celle consecutive ===")
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
	for i := 0; i < 4; i++ {
		fmt.Printf("RAM[0][0][%d] = %d  (atteso: %d)\n", i, ram.Data[0][0][i], i+1)
	}

	ok := true
	for i := 0; i < 4; i++ {
		if ram.Data[0][0][i] != uint8(i+1) {
			ok = false
		}
	}
	if ok {
		fmt.Println("✓ Corretto!")
	} else {
		fmt.Println("✗ Errore!")
	}
}

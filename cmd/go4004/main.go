package main

import (
	"fmt"
	"go-4004/cpu"
)

func main() {
	// Demo: addizione BCD  8 + 9 = 17
	//
	// Il processore 4004 non ha istruzione "somma decimale" nativa.
	// Si usa ADD (binario) seguito da DAA (Decimal Adjust Accumulator)
	// che corregge il risultato in formato BCD.
	//
	// Programma (5 byte, 4 istruzioni):
	//
	//   FIM R0, 0x09  → R0=0, R1=9  (secondo operando nei registri)
	//   LDM 8         → A = 8        (primo operando nell'accumulatore)
	//   ADD R1        → A = 8+9 = 17 → nibble: A=1, C=true (overflow)
	//   DAA           → se A>9 o C=1: A+=6 → 1+6=7, C=true (riporto)
	//
	// Risultato: A=7, C=1
	// Interpretazione BCD: cifra bassa = 7, cifra alta = C = 1 → numero: 17

	rom := cpu.NewROM(make([]byte, 256))
	ram := cpu.NewRAM()

	rom.Data[0x000] = cpu.FIM(cpu.R0) // FIM R0, 0x09
	rom.Data[0x001] = 0x09            //   → R0=0, R1=9
	rom.Data[0x002] = cpu.LDM(8)      // A = 8
	rom.Data[0x003] = cpu.ADD(cpu.R1) // A = 8+9, C=true se overflow nibble
	rom.Data[0x004] = cpu.DAA()       // correzione BCD

	c := cpu.NewCPU4004()
	c.Trace = true // abilita il debugger: stampa ogni istruzione eseguita

	fmt.Println("=== Demo: addizione BCD  8 + 9 ===")
	fmt.Println()

	for i := 0; i < 4; i++ {
		if err := c.Step(rom, ram); err != nil {
			fmt.Printf("Errore a step %d: %v\n", i+1, err)
			return
		}
	}

	cifraAlta := uint8(0)
	if c.C {
		cifraAlta = 1
	}
	fmt.Println()
	fmt.Printf("Risultato BCD: %d%d  (atteso: 17)\n", cifraAlta, c.A)
}

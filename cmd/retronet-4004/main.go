// Comando retronet-4004: carica una ROM da file ed esegue il programma
// sull'emulatore Intel 4004, con trace e dump opzionali dello stato.
//
// Uso:
//
//	retronet-4004 [flags] <file.rom>
//
// Flag:
//
//	-trace       stampa ogni istruzione eseguita (trace mode)
//	-max N       limite di step di sicurezza (default 100000)
//	-dump-ram    a fine esecuzione stampa le celle RAM non-zero
//
// Il programma si ferma quando il PC raggiunge un JUN che salta su se stesso
// (la convenzione "halt" usata dagli esempi) oppure al raggiungimento di -max.
package main

import (
	"flag"
	"fmt"
	"os"

	"go-4004/cpu"
)

func main() {
	trace := flag.Bool("trace", false, "stampa ogni istruzione eseguita (trace mode)")
	maxSteps := flag.Int("max", 100000, "limite di step di sicurezza (anti loop infinito)")
	dumpRAM := flag.Bool("dump-ram", false, "a fine esecuzione stampa le celle RAM non-zero")
	flag.Usage = func() {
		fmt.Fprintln(os.Stderr, "uso: retronet-4004 [flags] <file.rom>")
		flag.PrintDefaults()
	}
	flag.Parse()

	if flag.NArg() != 1 {
		flag.Usage()
		os.Exit(2)
	}
	path := flag.Arg(0)

	code, err := os.ReadFile(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "errore lettura ROM %q: %v\n", path, err)
		os.Exit(1)
	}
	if len(code) == 0 {
		fmt.Fprintf(os.Stderr, "ROM vuota: %s\n", path)
		os.Exit(1)
	}
	if len(code) > 4096 {
		fmt.Fprintf(os.Stderr, "ROM troppo grande: %d byte (massimo 4096)\n", len(code))
		os.Exit(1)
	}

	// La ROM del 4004 copre 4096 byte (PC a 12 bit). Carichiamo il programma
	// all'inizio; il resto dello spazio resta 0x00 (NOP).
	buf := make([]byte, 4096)
	copy(buf, code)
	rom := cpu.NewROM(buf)
	ram := cpu.NewRAM()

	c := cpu.NewCPU4004()
	c.Trace = *trace

	fmt.Printf("=== retronet-4004 — ROM: %s (%d byte) ===\n", path, len(code))
	if *trace {
		fmt.Println()
	}

	steps := 0
	for {
		// Convenzione "halt": un JUN che punta a se stesso ferma il programma.
		if isHalt(rom, c.PC) {
			break
		}
		if steps >= *maxSteps {
			fmt.Fprintf(os.Stderr, "\ninterrotto: superati %d step senza HALT (loop infinito?)\n", *maxSteps)
			os.Exit(1)
		}
		if err := c.Step(rom, ram); err != nil {
			fmt.Fprintf(os.Stderr, "\nerrore durante l'esecuzione (PC=0x%03X): %v\n", c.PC, err)
			os.Exit(1)
		}
		steps++
	}

	printState(c, steps)
	if *dumpRAM {
		dumpRam(ram)
	}
}

// isHalt riconosce l'idioma di arresto: un JUN (0x4n) all'indirizzo pc il cui
// indirizzo di destinazione coincide con pc stesso (salto su se stesso).
func isHalt(rom *cpu.ROM, pc uint16) bool {
	if int(pc)+1 >= len(rom.Data) {
		return false
	}
	op := rom.Data[pc]
	if op&0xF0 != cpu.OP_JUN {
		return false
	}
	target := (uint16(op&0x0F) << 8) | uint16(rom.Data[pc+1])
	return target == pc
}

// printState stampa lo stato finale della CPU: PC di arresto, accumulatore,
// carry, registri e numero di step eseguiti.
func printState(c *cpu.CPU4004, steps int) {
	fmt.Printf("\nHALT a PC=0x%03X dopo %d step\n", c.PC, steps)
	fmt.Printf("A=%X  C=%v\n", c.A, c.C)
	fmt.Print("registri: ")
	for i, v := range c.R {
		fmt.Printf("R%X=%X ", i, v)
	}
	fmt.Println()
}

// dumpRam stampa le celle dati RAM diverse da zero, indicizzate da
// banco/registro/carattere.
func dumpRam(ram *cpu.RAM) {
	fmt.Println("RAM (celle dati non-zero):")
	found := false
	for b := range ram.Data {
		for r := range ram.Data[b] {
			for ch := range ram.Data[b][r] {
				if v := ram.Data[b][r][ch]; v != 0 {
					fmt.Printf("  [%d][%d][%2d] = %X\n", b, r, ch, v)
					found = true
				}
			}
		}
	}
	if !found {
		fmt.Println("  (nessuna)")
	}
}

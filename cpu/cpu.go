package cpu

import (
	"fmt"
	"io"
	"os"
)

// CPU4004 rappresenta lo stato interno del processore Intel 4004.
//
// Il 4004 è un processore a 4 bit: A, i registri e le operazioni ALU
// lavorano su nibble (0–15). Go non ha un tipo a 4 bit, quindi si usa
// uint8 mascherato con nibble().
type CPU4004 struct {
	A  uint8     // Accumulator — registro di lavoro principale, 4 bit
	C  bool      // Carry flag — usato da ADD, SUB, rotazioni e istruzioni BCD
	R  [16]uint8 // Registri R0–RF — 16 registri da 4 bit ciascuno
	PC uint16    // Program Counter — indirizzo prossima istruzione, 12 bit (0x000–0xFFF)
	CL uint8     // Command Line — seleziona il banco RAM attivo, impostato da DCL

	// Stack hardware a 3 livelli.
	// Sul 4004 reale lo stack non è RAM: è un registro a scorrimento interno
	// con esattamente 3 slot. Non è indirizzabile né ispezionabile dal firmware.
	// Il 4° push sovrascrive il primo slot (comportamento ciclico, nessun errore).
	Stack [3]uint16 // indirizzi di ritorno salvati da JMS, ripristinati da BBL
	SP    uint8     // stack pointer — conta i livelli occupati (0 = stack vuoto)

	// SRCAddr conserva l'indirizzo inviato dall'istruzione SRC sul bus esterno.
	//
	// Sul 4004 reale, SRC non scrive in nessun registro interno della CPU:
	// mette il valore sui pin del bus esterno, e il chip RAM Intel 4002 collegato
	// lo riceve e lo salva al suo interno. L'indirizzo non vive nella CPU, vive nel chip RAM.
	//
	// Nell'emulatore non esiste un bus fisico, quindi conserviamo qui il valore
	// come comodo placeholder. Quando verrà implementata la struct RAM,
	// SRCAddr verrà spostato dentro RAM — che è il posto architetturalmente corretto —
	// e questo campo verrà rimosso dalla CPU.
	//
	// Formato: nibble alto (bit 7-4) = chip/banco RAM, nibble basso (bit 3-0) = registro nel chip.
	SRCAddr uint8

	// Trace e TraceWriter controllano la modalità debugger.
	// Se Trace = true, Step() stampa una riga per ogni istruzione eseguita.
	// TraceWriter è la destinazione dell'output; se nil viene usato os.Stdout.
	Trace       bool
	TraceWriter io.Writer
}

// NewCPU4004 crea una nuova istanza del CPU4004 con valori iniziali
// Tutti i registri e l'accumulatore sono inizializzati a 0, e il carry è false
// Questo è importante per garantire che il comportamento del CPU sia prevedibile e corretto fin dall'inizio
func NewCPU4004() *CPU4004 {
	return &CPU4004{}
}

// nibble estrae i 4 bit meno significativi di un byte
// garantendo che i valori rimangano entro 0-15
// Questo è importante per simulare correttamente il comportamento del 4004
// e per evitare overflow nei calcoli che devono rimanere a 4 bit
// Ad esempio, se A = 0x0F (15) e aggiungiamo 1, il risultato dovrebbe essere 0x00 (0) con carry = true
// Senza questa funzione, potremmo ottenere un risultato errato come 0x10 (16) che non è valido per un registro a 4 bit
func nibble(v uint8) uint8 {
	return v & 0x0F
}

// Step esegue un singolo ciclo fetch-execute:
// legge l'opcode dalla ROM all'indirizzo PC, incrementa PC, esegue l'istruzione.
// PC è mascherato a 12 bit (range 0x000–0xFFF) come sul 4004 reale.
// Le istruzioni a 2 byte (JCN, FIM, JUN, JMS, ISZ) leggono un secondo byte prima di eseguire.
// Se Trace è true, stampa una riga di debug su TraceWriter (o os.Stdout) dopo ogni esecuzione riuscita.
func (c *CPU4004) Step(rom *ROM, ram *RAM) (err error) {
	pcBefore := c.PC
	var op byte

	// Il defer viene eseguito al termine della funzione, qualunque sia il percorso di ritorno.
	// Poiché err è un named return, la closure vede il valore finale assegnato dalla funzione.
	defer func() {
		if c.Trace && err == nil {
			c.traceLine(pcBefore, op)
		}
	}()

	op, err = readROM(rom, c.PC)
	if err != nil {
		return
	}
	c.PC = (c.PC + 1) & 0x0FFF

	switch op & 0xF0 {
	case OP_JCN, OP_JUN, OP_JMS, OP_ISZ:
		var arg byte
		arg, err = readROM(rom, c.PC)
		if err != nil {
			return
		}
		c.PC = (c.PC + 1) & 0x0FFF
		err = c.executeWithArg(op, arg)
		return
	case OP_FIM & 0xF0: // 0x20: FIM (bit 0 = 0, 2 byte) o SRC (bit 0 = 1, 1 byte)
		if op&0x01 == 0 {
			var arg byte
			arg, err = readROM(rom, c.PC)
			if err != nil {
				return
			}
			c.PC = (c.PC + 1) & 0x0FFF
			err = c.executeWithArg(op, arg)
			return
		}
		err = c.Execute(op) // SRC: 1 byte
		return
	case OP_FIN & 0xF0: // 0x30: FIN (bit 0 = 0) o JIN (bit 0 = 1)
		if op&0x01 == 0 {
			// FIN Rr: fetch indirect da ROM usando R0:R1 come indirizzo (nella pagina corrente)
			// La pagina arriva dal PC post-fetch: questo replica anche l'eccezione Intel
			// per FIN posto all'ultimo byte di una pagina ROM.
			rp := op & 0x0E // primo registro della coppia (sempre pari)
			addr := (c.PC & 0x0F00) | (uint16(c.R[0]) << 4) | uint16(c.R[1])
			var data byte
			data, err = readROM(rom, addr)
			if err != nil {
				return
			}
			c.R[rp] = data >> 4
			c.R[rp+1] = data & 0x0F
			return
		}
		err = c.Execute(op) // JIN: 1 byte, gestito in Execute
		return
	case 0xE0: // gruppo I/O e RAM (0xEX) — richiede accesso alla RAM e alla ROM virtuale
		err = c.executeIO(op, rom, ram)
		return
	default:
		err = c.Execute(op)
		return
	}
}

// Push è la versione esportata di push, usata da main e dai test di integrazione
// finché JMS non sarà implementato. Simula il salvataggio dell'indirizzo di ritorno
// che normalmente avviene automaticamente con l'istruzione JMS.
func (c *CPU4004) Push(addr uint16) { c.push(addr) }

// push salva un indirizzo sullo stack prima di un salto a subroutine (JMS).
// L'indice è calcolato modulo 3: se lo stack è pieno il valore più vecchio
// viene sovrascritto, replicando il comportamento hardware del 4004 reale.
func (c *CPU4004) push(addr uint16) {
	c.Stack[c.SP%3] = addr & 0x0FFF
	c.SP++
}

// pop recupera l'indirizzo di ritorno dallo stack (usato da BBL).
// Decrementa SP prima di leggere, speculare a push.
func (c *CPU4004) pop() uint16 {
	if c.SP == 0 {
		return 0
	}
	c.SP--
	return c.Stack[c.SP%3]
}

// traceLine stampa una riga di trace per l'istruzione appena eseguita.
// Formato: PC=003 OP=F2 IAC        A=8  C=true
// Per le istruzioni I/O (0xEX) aggiunge anche il banco RAM attivo e SRCAddr.
func (c *CPU4004) traceLine(pc uint16, op byte) {
	w := c.TraceWriter
	if w == nil {
		w = os.Stdout
	}
	mnem := Disassemble(op)
	fmt.Fprintf(w, "PC=%03X OP=%02X %-10s A=%X  C=%-5v", pc, op, mnem, c.A, c.C)
	if op&0xF0 == 0xE0 {
		fmt.Fprintf(w, "  CL=%X SRC=%02X", c.CL, c.SRCAddr)
	}
	fmt.Fprintln(w)
}

func readROM(rom *ROM, addr uint16) (byte, error) {
	if rom == nil {
		return 0, fmt.Errorf("ROM non inizializzata")
	}
	if int(addr) >= len(rom.Data) {
		return 0, fmt.Errorf("indirizzo ROM 0x%03X fuori dalla ROM (len=%d)", addr, len(rom.Data))
	}
	return rom.Data[addr], nil
}

package cpu

import "fmt"

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
func (c *CPU4004) Step(rom *ROM) error {
	if int(c.PC) >= len(rom.Data) {
		return fmt.Errorf("PC 0x%03X fuori dalla ROM (dimensione %d)", c.PC, len(rom.Data))
	}
	// page viene salvato PRIMA dell'incremento: JCN, ISZ e FIN devono usare
	// la pagina dell'istruzione corrente, non quella del byte successivo.
	page := c.PC & 0x0F00
	op := rom.Data[c.PC]
	c.PC = (c.PC + 1) & 0x0FFF

	switch op & 0xF0 {
	case OP_JCN, OP_JUN, OP_JMS, OP_ISZ:
		if int(c.PC) >= len(rom.Data) {
			return fmt.Errorf("PC 0x%03X fuori dalla ROM durante fetch secondo byte", c.PC)
		}
		arg := rom.Data[c.PC]
		c.PC = (c.PC + 1) & 0x0FFF
		return c.executeWithArg(op, arg, page)
	case OP_FIM & 0xF0: // 0x20: FIM (bit 0 = 0, 2 byte) o SRC (bit 0 = 1, 1 byte)
		if op&0x01 == 0 {
			if int(c.PC) >= len(rom.Data) {
				return fmt.Errorf("PC 0x%03X fuori dalla ROM durante fetch secondo byte FIM", c.PC)
			}
			arg := rom.Data[c.PC]
			c.PC = (c.PC + 1) & 0x0FFF
			return c.executeWithArg(op, arg, page)
		}
		return c.Execute(op) // SRC: 1 byte
	case OP_FIN & 0xF0: // 0x30: FIN (bit 0 = 0) o JIN (bit 0 = 1)
		if op&0x01 == 0 {
			// FIN Rr: fetch indirect da ROM usando R0:R1 come indirizzo nella pagina di FIN.
			// Usa `page` (calcolato prima dell'incremento) per restare nella pagina corretta
			// anche quando FIN si trova all'ultimo byte di una pagina.
			rp := op & 0x0E // primo registro della coppia (sempre pari)
			addr := page | (uint16(c.R[0]) << 4) | uint16(c.R[1])
			if int(addr) >= len(rom.Data) {
				return fmt.Errorf("indirizzo FIN 0x%03X fuori dalla ROM", addr)
			}
			data := rom.Data[addr]
			c.R[rp] = data >> 4
			c.R[rp+1] = data & 0x0F
			return nil
		}
		return c.Execute(op) // JIN: 1 byte, gestito in Execute
	default:
		return c.Execute(op)
	}
}

// Push è la versione esportata di push, usata dai test di integrazione.
func (c *CPU4004) Push(addr uint16) { c.push(addr) }

// push salva un indirizzo sullo stack prima di un salto a subroutine (JMS).
// L'indice è calcolato modulo 3: se lo stack è pieno il valore più vecchio
// viene sovrascritto, replicando il comportamento hardware del 4004 reale.
func (c *CPU4004) push(addr uint16) {
	c.Stack[c.SP%3] = addr
	c.SP++
}

// pop recupera l'indirizzo di ritorno dallo stack (usato da BBL).
// Decrementa SP prima di leggere, speculare a push.
// Se lo stack è vuoto (SP==0) restituisce 0: sul 4004 reale il comportamento
// è indefinito; nell'emulatore scegliamo un risultato deterministico.
func (c *CPU4004) pop() uint16 {
	if c.SP == 0 {
		return 0
	}
	c.SP--
	return c.Stack[c.SP%3]
}

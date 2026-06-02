package cpu

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
func (c *CPU4004) Step(rom *ROM) error {
	op := rom.Data[c.PC]
	c.PC = (c.PC + 1) & 0x0FFF
	return c.Execute(op)
}

// push salva un indirizzo sullo stack prima di un salto a subroutine (JMS).
// L'indice è calcolato modulo 3: se lo stack è pieno il valore più vecchio
// viene sovrascritto, replicando il comportamento hardware del 4004 reale.
func (c *CPU4004) push(addr uint16) {
	c.Stack[c.SP%3] = addr
	c.SP++
}

// pop recupera l'indirizzo di ritorno dallo stack (usato da BBL).
// Decrementa SP prima di leggere, speculare a push.
func (c *CPU4004) pop() uint16 {
	c.SP--
	return c.Stack[c.SP%3]
}

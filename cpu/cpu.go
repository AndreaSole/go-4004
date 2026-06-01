package cpu

// CPU4004 rappresenta lo stato del processore Intel 4004
// Include l'accumulatore (A), il flag di carry (C), i registri (R) e il program counter (PC)
type CPU4004 struct {
	A  uint8     // Accumulator, 4 bit
	C  bool      // Carry flag
	R  [16]uint8 // 16 registri da 4 bit
	PC uint16    // Program Counter, 12 bit
	CL uint8     // Command Line — banco RAM attivo (0-7), impostato da DCL
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
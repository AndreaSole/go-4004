package cpu

import "fmt"

// Execute esegue un'istruzione data dal byte opcode
// Il metodo interpreta l'opcode e aggiorna lo stato del CPU di conseguenza
// Supporta un sottoinsieme di istruzioni del 4004, come LDM, XCH, INC e ADD
// Se viene passato un opcode non implementato, restituisce un errore
func (c *CPU4004) Execute(op byte) error {
	low := op & 0x0F

	switch {
	case op == OP_NOP:
		// NOP
		// Non fa nulla, semplicemente incrementa il program counter

	case op&0xF0 == OP_LDM:
		// LDM data: carica valore immediato in A
		// LDM 0-15: carica il valore immediato (0-15) nell'accumulatore (A)
		c.A = nibble(low)

	// L'istruzione LD Rn: carica il valore del registro specificato (R0-R15) nell'accumulatore (A)
	// Ad esempio, se R0 = 0x05, dopo LD R0, A sarà 0x05
	case op&0xF0 == OP_LD:
		n := low
		c.A = nibble(c.R[n])

	case op&0xF0 == OP_XCH:
		// XCH R0-R15: scambia il valore dell'accumulatore (A) con quello del registro specificato (R0-R15)
		// Ad esempio, se A = 0x02 e R0 = 0x00, dopo XCH R0, A sarà 0x00 e R0 sarà 0x02
		n := low
		c.A, c.R[n] = c.R[n], c.A

	case op&0xF0 == OP_INC:
		// INC R0-R15: incrementa il valore del registro specificato (R0-R15) di 1
		// Ad esempio, se R0 = 0x0F (15), dopo INC R0, R0 sarà 0x00 (0) e non ci sarà carry, poiché i registri sono a 4 bit
		n := low

		// Incrementa il registro specificato e assicura che rimanga a 4 bit
		c.R[n] = nibble(c.R[n] + 1)

	case op&0xF0 == OP_ADD:
		// ADD R0-R15: aggiunge il valore del registro specificato (R0-R15) all'accumulatore (A) e al carry
		// Ad esempio, se A = 0x03, R0 = 0x02 e C = false, dopo ADD R0, A sarà 0x05 e C sarà false
		// Se A = 0x0F, R0 = 0x01 e C = true, dopo ADD R0, A sarà 0x01 (0 + 1 + 1) e C sarà true (carry)
		n := low

		// Calcola il risultato dell'addizione considerando il carry
		// Il carry viene trattato come 1 se è true, altrimenti 0
		// Questo è importante per simulare correttamente il comportamento del 4004, dove il carry influisce sull'addizione
		// Ad esempio, se A = 0x0F, R0 = 0x01 e C = true, il risultato sarà 0x11 (17), ma poiché A è a 4 bit, diventa 0x01 con carry = true
		carry := uint8(0)
		if c.C {
			carry = 1
		}

		result := c.A + c.R[n] + carry
		c.A = nibble(result)

		// Il carry è true se il risultato dell'addizione supera 0x0F (15), altrimenti è false
		c.C = result > 0x0F

	default:
		return fmt.Errorf("opcode non implementato: 0x%02X", op)
	}

	return nil
}

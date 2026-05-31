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

	// SUB R0-R15: sottrae il valore del registro specificato (R0-R15) dall'accumulatore (A) considerando il borrow (carry)
	// La formula 16 + A - Rr - borrow evita underflow su uint8. Se il risultato è < 16, significa che senza il "prestito del 16" sarebbe stato negativo → borrow avvenuto → C=1.
	case op&0xF0 == OP_SUB:
		n := low
		borrow := uint8(0)
		if c.C {
			borrow = 1
		}
		sum := uint8(16) + c.A - c.R[n] - borrow
		c.A = nibble(sum)
		c.C = sum < 16

	// IAC: Increment Accumulator, incrementa l'accumulatore (A) di 1 considerando il carry
	case op == OP_IAC:
		result := c.A + 1
		c.A = nibble(result)
		c.C = result > 0x0F

	// DAC: Decrement Accumulator, decrementa l'accumulatore (A) di 1 considerando il borrow (carry)
	// La formula 16 + A - 1 evita underflow su uint8. Se il risultato è < 16, significa che senza il "prestito del 16" sarebbe stato negativo → borrow avvenuto → C=1.
	case op == OP_DAC:
		result := uint8(16) + c.A - 1
		c.A = nibble(result)
		c.C = result < 16

	// CMA: Complement Accumulator, inverte tutti i bit dell'accumulatore (A)
	case op == OP_CMA:
		c.A = nibble(^c.A)

	// CLB: Clear Accumulator and Borrow, azzera l'accumulatore (A) e il carry (C)
	case op == OP_CLB:
		c.A = 0
		c.C = false

	// CLC: Clear Carry, azzera solo il carry (C) lasciando intatto l'accumulatore (A)
	case op == OP_CLC:
		c.C = false

	// STC: Set Carry, imposta il carry (C) a true senza modificare l'accumulatore (A)
	case op == OP_STC:
		c.C = true

	// CMC: Complement Carry, inverte lo stato del carry (C) senza modificare l'accumulatore (A)
	case op == OP_CMC:
		c.C = !c.C

	// RAL: Rotate Accumulator Left, ruota i bit dell'accumulatore (A) a sinistra e sposta il bit più significativo nel carry (C)
	// Ad esempio, se A = 0b1011 (11) e C = false, dopo RAL, A sarà 0b0110 (6) e C sarà true (il bit più significativo 1 è stato spostato nel carry)
	// Se A = 0b1011 (11) e C = true, dopo RAL, A sarà 0b0111 (7) e C sarà true (il bit più significativo 1 è stato spostato nel carry e il vecchio carry true è stato spostato in A)
	case op == OP_RAL:
		newCarry := c.A&0x08 != 0
		oldCarry := uint8(0)
		if c.C {
			oldCarry = 1
		}
		// La rotazione a sinistra sposta i bit di A a sinistra e inserisce il vecchio carry nel bit meno significativo
		c.A = nibble(c.A<<1) | oldCarry
		c.C = newCarry

	// RAR: Rotate Accumulator Right, ruota i bit dell'accumulatore (A) a destra e sposta il bit più significativo nel carry (C)
	// Ad esempio, se A = 0b1011 (11) e C = false, dopo RAR, A sarà 0b1101 (13) e C sarà false (il bit più significativo 1 è stato spostato nel carry)
	// Se A = 0b1011 (11) e C = true, dopo RAR, A sarà 0b1101 (13) e C sarà true (il bit più significativo 1 è stato spostato nel carry e il vecchio carry true è stato spostato in A)
	case op == OP_RAR:
		newCarry := c.A&0x01 != 0
		oldCarry := uint8(0)
		if c.C {
			oldCarry = 1
		}
		// La rotazione a destra sposta i bit di A a destra e inserisce il vecchio carry nel bit più significativo
		c.A = (c.A >> 1) | (oldCarry << 3)
		c.C = newCarry

	default:
		return fmt.Errorf("opcode non implementato: 0x%02X", op)
	}

	return nil
}

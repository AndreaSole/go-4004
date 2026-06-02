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
		c.A = nibble(c.R[low])

	case op&0xF0 == OP_XCH:
		// XCH R0-R15: scambia il valore dell'accumulatore (A) con quello del registro specificato (R0-R15)
		// Ad esempio, se A = 0x02 e R0 = 0x00, dopo XCH R0, A sarà 0x00 e R0 sarà 0x02
		c.A, c.R[low] = c.R[low], c.A

	case op&0xF0 == OP_INC:
		// INC R0-R15: incrementa il valore del registro specificato (R0-R15) di 1
		// Ad esempio, se R0 = 0x0F (15), dopo INC R0, R0 sarà 0x00 (0) e non ci sarà carry, poiché i registri sono a 4 bit
		// Incrementa il registro specificato e assicura che rimanga a 4 bit
		c.R[low] = nibble(c.R[low] + 1)

	case op&0xF0 == OP_ADD:
		// ADD R0-R15: aggiunge il valore del registro specificato (R0-R15) all'accumulatore (A) e al carry
		// Ad esempio, se A = 0x03, R0 = 0x02 e C = false, dopo ADD R0, A sarà 0x05 e C sarà false
		// Se A = 0x0F, R0 = 0x01 e C = true, dopo ADD R0, A sarà 0x01 (0 + 1 + 1) e C sarà true (carry)
		carry := uint8(0)
		if c.C {
			carry = 1
		}

		result := c.A + c.R[low] + carry
		c.A = nibble(result)

		// Il carry è true se il risultato dell'addizione supera 0x0F (15), altrimenti è false
		c.C = result > 0x0F

	// SUB R0-R15: sottrae il valore del registro specificato (R0-R15) dall'accumulatore (A) considerando il borrow (carry)
	// La formula 16 + A - Rr - borrow evita underflow su uint8. Se il risultato è < 16, significa che senza il "prestito del 16" sarebbe stato negativo → borrow avvenuto → C=1.
	case op&0xF0 == OP_SUB:
		borrow := uint8(0)
		if c.C {
			borrow = 1
		}
		sum := uint8(16) + c.A - c.R[low] - borrow
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

	// RAR: Rotate Accumulator Right, ruota i bit dell'accumulatore (A) a destra e sposta bit meno significativo (bit 0) nel carry (C)
	case op == OP_RAR:
		newCarry := c.A&0x01 != 0
		oldCarry := uint8(0)
		if c.C {
			oldCarry = 1
		}
		// La rotazione a destra sposta i bit di A a destra e inserisce il vecchio carry nel bit più significativo
		c.A = (c.A >> 1) | (oldCarry << 3)
		c.C = newCarry

	// TCC: Transfer Carry to Accumulator and Clear, carica 1 in A se il carry (C) è true, altrimenti carica 0 in A, e azzera il carry (C)
	case op == OP_TCC:
		if c.C {
			c.A = 1
		} else {
			c.A = 0
		}
		c.C = false

	// TCS: Transfer Carry to Accumulator and Set, carica 10 in A se il carry (C) è true, altrimenti carica 9 in A, e azzera il carry (C)
	case op == OP_TCS:
		if c.C {
			c.A = 10
		} else {
			c.A = 9
		}
		c.C = false

	// DAA: Decimal Adjust Accumulator, regola il valore dell'accumulatore (A) per ottenere un risultato corretto in formato BCD dopo un'addizione o sottrazione
	// Se il carry (C) è true o A è maggiore di 9, aggiunge 6 a A per correggere il risultato in BCD e imposta C a true se c'è stato un overflow oltre 0x0F
	// Ad esempio, se A = 0x09 e C = false, dopo DAA, A sarà 0x09 (9) e C sarà false (nessuna correzione necessaria)
	// Se A = 0x0A (10) e C = false, dopo DAA, A sarà 0x00 (0) e C sarà true (correzione necessaria perché 10 non è un singolo digit BCD)
	case op == OP_DAA:
		if c.C || c.A > 9 {
			c.A = nibble(c.A + 6)
			c.C = true
		}

	// DCL: Designate Command Line, copia il valore di A nel registro CL per selezionare il banco RAM attivo.
	// Usato prima delle istruzioni RAM (WRM, RDM, ecc.) per indicare quale banco di chip Intel 4002 risponde.
	// Non modifica A né il carry.
	case op == OP_DCL:
		c.CL = c.A & 0x07 // CL può essere solo 0-7, quindi prendiamo solo i 3 bit meno significativi di A

	// KBP: Keyboard Process, converte un valore one-hot dell'accumulatore nel numero di posizione del bit attivo.
	// Usato per decodificare la colonna attiva durante la scansione della tastiera a matrice.
	// Se A ha più di un bit a 1 (input non valido), imposta A = 0xF (errore).
	// Il carry non viene modificato.
	case op == OP_KBP:
		switch c.A {
		case 0b0000:
			c.A = 0
		case 0b0001:
			c.A = 1
		case 0b0010:
			c.A = 2
		case 0b0100:
			c.A = 3
		case 0b1000:
			c.A = 4
		default:
			c.A = 0xF
		}

	// BBL 0-15: branch back and load.
	// Ripristina PC dallo stack (ritorno da subroutine) e carica il nibble basso in A.
	// Non modifica il carry.
	case op&0xF0 == OP_BBL:
		c.A = nibble(low)
		c.PC = c.pop()

	default:
		return fmt.Errorf("opcode non implementato: 0x%02X", op)
	}

	return nil
}

package cpu

import "fmt"

// executeWithArg esegue le istruzioni a 2 byte del 4004.
// op è il primo byte (opcode + nibble), arg è il secondo byte già letto dalla ROM.
func (c *CPU4004) executeWithArg(op, arg byte) error {
	high := op & 0x0F // nibble alto dell'indirizzo o numero di registro

	switch op & 0xF0 {

	// JUN a: salta incondizionatamente all'indirizzo a 12 bit.
	// Primo byte: 0x4n (n = bit 11-8 dell'indirizzo)
	// Secondo byte: AB (bit 7-0 dell'indirizzo)
	// Indirizzo finale: n<<8 | AB
	case OP_JUN:
		c.PC = (uint16(high) << 8) | uint16(arg)

	// JMS a: salta a subroutine all'indirizzo a 12 bit.
	// Salva il PC corrente sullo stack prima del salto.
	// Il PC a questo punto punta già all'istruzione successiva (post-fetch dei 2 byte).
	case OP_JMS:
		c.push(c.PC)
		c.PC = (uint16(high) << 8) | uint16(arg)

	// JCN c,a: salta se la condizione c e' vera, a e' il byte basso dell'indirizzo.
	// La pagina e' quella del PC post-fetch: sul 4004 questo include l'eccezione
	// documentata per JCN posto sugli ultimi due byte di una pagina ROM.
	// Condizione codificata in 4 bit (nibble basso del primo byte):
	//   bit 3 (C1): inverte l'intera condizione (NOT)
	//   bit 2 (C2): salta se A == 0
	//   bit 1 (C3): salta se carry == 1
	//   bit 0 (C4): salta se TEST == 0 (TEST pin non emulato, trattato come 1)
	case OP_JCN:
		cond := high
		taken := false
		if cond&0x04 != 0 && c.A == 0 {
			taken = true
		}
		if cond&0x02 != 0 && c.C {
			taken = true
		}
		// C4 (bit 0, TEST pin) non emulato: considerato sempre HIGH, quindi mai vero.
		if cond&0x08 != 0 { // C1 (bit 3): NOT della condizione composta.
			taken = !taken
		}
		if taken {
			// L'indirizzo di salto condivide i 4 bit alti con il PC post-fetch.
			c.PC = (c.PC & 0x0F00) | uint16(arg)
		}

	// ISZ Rr,a: incrementa il registro Rr; se il risultato e' diverso da 0,
	// salta a 'a' usando la stessa pagina post-fetch di JCN.
	case OP_ISZ:
		c.R[high] = nibble(c.R[high] + 1)
		if c.R[high] != 0 {
			c.PC = (c.PC & 0x0F00) | uint16(arg)
		}

	// FIM Rr,d: carica il byte immediato 'd' nella coppia di registri Rr/Rr+1.
	// Rr riceve il nibble alto di d, Rr+1 riceve il nibble basso.
	// 'high' è sempre pari (bit 0 del nibble = 0, bit 0 distingue FIM da SRC).
	case OP_FIM:
		c.R[high] = arg >> 4
		c.R[high+1] = arg & 0x0F

	default:
		return fmt.Errorf("executeWithArg: opcode non implementato: 0x%02X arg=0x%02X", op, arg)
	}

	return nil
}

// Execute esegue un'istruzione a singolo byte dato l'opcode.
// Le istruzioni a 2 byte usano executeWithArg (chiamato da Step).
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
		c.A, c.R[low] = nibble(c.R[low]), nibble(c.A)

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

	// SUB R0-R15: A = A + ~Rr + CY. Sul 4004 CY=1 significa nessun borrow
	// precedente; dopo l'operazione CY=1 significa nessun borrow generato.
	case op&0xF0 == OP_SUB:
		carry := uint8(0)
		if c.C {
			carry = 1
		}
		sum := c.A + nibble(^c.R[low]) + carry
		c.A = nibble(sum)
		c.C = sum > 0x0F

	// IAC: Increment Accumulator, incrementa l'accumulatore (A) di 1 considerando il carry
	case op == OP_IAC:
		result := c.A + 1
		c.A = nibble(result)
		c.C = result > 0x0F

	// DAC: Decrement Accumulator, decrementa A di 1.
	// Un borrow imposta CY=0; nessun borrow imposta CY=1.
	case op == OP_DAC:
		result := c.A + 0x0F
		c.A = nibble(result)
		c.C = result > 0x0F

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

	// SRC Rr: send register control — imposta l'indirizzo del registro RAM per le operazioni I/O.
	// Il byte SRC è formato da Rr (nibble alto = chip/banco RAM) e Rr+1 (nibble basso = registro).
	// Memorizzato in SRCAddr; verrà letto dalle istruzioni del gruppo 0xEX (WRM, RDM...) in Step 7.
	// Bit 0 del nibble basso dell'opcode è sempre 1 (distingue SRC da FIM a 2 byte).
	case op&0xF1 == OP_SRC:
		rp := low &^ 1 // forza Rr pari
		c.SRCAddr = (c.R[rp] << 4) | c.R[rp+1]

	// FIN Rr: fetch indirect da ROM — richiede accesso alla ROM, non supportato da Execute.
	// Usare Step() che gestisce FIN direttamente con accesso alla ROM.
	case op&0xF1 == OP_FIN:
		return fmt.Errorf("FIN richiede accesso alla ROM: usare Step()")

	// JIN Rr: jump indirect — salta all'indirizzo formato da Rr (nibble alto) e Rr+1 (nibble basso),
	// nella stessa pagina del PC corrente (bit 11-8 invariati).
	case op&0xF1 == OP_JIN:
		rp := low &^ 1
		addr := (uint16(c.R[rp]) << 4) | uint16(c.R[rp+1])
		c.PC = (c.PC & 0x0F00) | addr

	default:
		return fmt.Errorf("opcode non implementato: 0x%02X", op)
	}

	return nil
}

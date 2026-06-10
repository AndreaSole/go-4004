package cpu

// Il 4004 ha 16 registri da 4 bit, identificati da R0 a RF.
// Vengono usati per memorizzare dati temporanei durante l'esecuzione delle istruzioni.
const (
	R0 = 0x0
	R1 = 0x1
	R2 = 0x2
	R3 = 0x3
	R4 = 0x4
	R5 = 0x5
	R6 = 0x6
	R7 = 0x7
	R8 = 0x8
	R9 = 0x9
	RA = 0xA
	RB = 0xB
	RC = 0xC
	RD = 0xD
	RE = 0xE
	RF = 0xF
)

// Gruppo registro (0x00, 0x6X–0xDX)
// Istruzioni che operano su un registro specificato nel nibble basso dell'opcode.
// Decodifica: op & 0xF0 == OP_XXX
const (
	OP_NOP = 0x00 // No Operation

	OP_INC = 0x60 // INC Rr — incrementa registro
	OP_ADD = 0x80 // ADD Rr — somma registro ad A con carry
	OP_SUB = 0x90 // SUB Rr — sottrae registro da A con borrow
	OP_LD  = 0xA0 // LD Rr  — carica registro in A
	OP_XCH = 0xB0 // XCH Rr — scambia A con registro
	OP_BBL = 0xC0 // BBL n  — branch back and load (ritorno da subroutine)
	OP_LDM = 0xD0 // LDM n  — carica immediato in A
)

// Gruppo accumulatore (0xFX)
// Istruzioni a byte singolo fisso che operano solo su A e/o C.
// Decodifica: op == OP_XXX
const (
	OP_CLB = 0xF0 // Clear Both         — A=0, C=false
	OP_CLC = 0xF1 // Clear Carry        — C=false
	OP_IAC = 0xF2 // Increment Acc      — A++
	OP_CMC = 0xF3 // Complement Carry   — C=!C
	OP_CMA = 0xF4 // Complement Acc     — A=~A
	OP_RAL = 0xF5 // Rotate Left        — ruota A sinistra attraverso C
	OP_RAR = 0xF6 // Rotate Right       — ruota A destra attraverso C
	OP_TCC = 0xF7 // Transfer Carry     — A=C, C=false
	OP_DAC = 0xF8 // Decrement Acc      — A--
	OP_TCS = 0xF9 // Transfer Carry Sub — A=9 o 10 in base a C, C=false
	OP_STC = 0xFA // Set Carry          — C=true
	OP_DAA = 0xFB // Decimal Adjust     — corregge A per BCD dopo ADD
	OP_KBP = 0xFC // Keyboard Process   — decodifica one-hot in posizione
	OP_DCL = 0xFD // Designate CL       — CL=A, seleziona banco RAM
)

// Gruppo salti e indirizzamento (0x1X–0x7X)
// Istruzioni a 2 byte: il primo byte contiene opcode + nibble, il secondo l'indirizzo/dato.
// Decodifica: op & 0xF0 == OP_XXX
const (
	OP_JCN = 0x10 // JCN c,a  — jump condizionale (c=condizione, a=indirizzo basso)
	OP_FIM = 0x20 // FIM Rr,d — fetch immediate in register pair (d=dato)
	OP_SRC = 0x21 // SRC Rr   — send register control (1 byte, ma same high byte as FIM)
	OP_FIN = 0x30 // FIN Rr   — fetch indirect da ROM via R0R1
	OP_JIN = 0x31 // JIN Rr   — jump indirect via registro pair
	OP_JUN = 0x40 // JUN a    — jump unconditional (a=indirizzo 12 bit)
	OP_JMS = 0x50 // JMS a    — jump to subroutine (push PC, poi salta)
	OP_ISZ = 0x70 // ISZ Rr,a — increment register, skip if zero
)

// Gruppo I/O e RAM (0xEX)
// Istruzioni a byte singolo che operano sulla RAM virtuale (chip 4002)
// all'indirizzo selezionato da DCL + SRC. Implementate in executeIO.
const (
	OP_WRM = 0xE0 // Write RAM data     — scrivi A in RAM
	OP_WMP = 0xE1 // Write RAM port     — scrivi A su porta output RAM
	OP_WRR = 0xE2 // Write ROM port     — scrivi A su porta output ROM
	OP_WPM = 0xE3 // Write prog memory  — scrivi A in program memory
	OP_WR0 = 0xE4 // Write RAM status 0
	OP_WR1 = 0xE5 // Write RAM status 1
	OP_WR2 = 0xE6 // Write RAM status 2
	OP_WR3 = 0xE7 // Write RAM status 3
	OP_SBM = 0xE8 // Sub RAM from Acc   — A=A-RAM con borrow
	OP_RDM = 0xE9 // Read RAM data      — A=RAM data
	OP_RDR = 0xEA // Read ROM port      — A=porta ROM
	OP_ADM = 0xEB // Add RAM to Acc     — A=A+RAM con carry
	OP_RD0 = 0xEC // Read RAM status 0  — A=status 0
	OP_RD1 = 0xED // Read RAM status 1  — A=status 1
	OP_RD2 = 0xEE // Read RAM status 2  — A=status 2
	OP_RD3 = 0xEF // Read RAM status 3  — A=status 3
)

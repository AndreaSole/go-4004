package cpu

// Il 4004 ha 16 registri da 4 bit, identificati da R0 a RF
// Questi registri vengono utilizzati per memorizzare dati temporanei durante l'esecuzione delle istruzioni
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

// OP_NOP: No Operation, non fa nulla
// OP_INC: Incrementa il registro specificato
// OP_ADD: Aggiunge il valore del registro specificato all'accumulatore (A) e al carry
// OP_XCH: Scambia il valore dell'accumulatore con quello del registro specificato
// OP_LDM: Carica un valore immediato nell'accumulatore (A)
// OP_LD: Carica il valore del registro specificato nell'accumulatore (A)
const (
	OP_NOP = 0x00
	OP_INC = 0x60
	OP_ADD = 0x80
	OP_LD  = 0xA0
	OP_XCH = 0xB0
	OP_LDM = 0xD0
)

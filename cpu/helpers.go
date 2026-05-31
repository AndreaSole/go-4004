package cpu

// Funzioni per creare istruzioni
// Queste funzioni semplificano la creazione di istruzioni per il CPU4004
func NOP() byte {
	return OP_NOP
}

func LDM(v byte) byte {
	return OP_LDM | nibble(v)
}

func XCH(r byte) byte {
	return OP_XCH | nibble(r)
}

func INC(r byte) byte {
	return OP_INC | nibble(r)
}

func ADD(r byte) byte {
	return OP_ADD | nibble(r)
}

func LD(r byte) byte {
	return OP_LD | nibble(r)
}

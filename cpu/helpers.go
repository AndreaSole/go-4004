package cpu

// Queste funzioni semplificano la creazione di istruzioni restituendo il byte opcode corretto per ciascuna istruzione
func NOP() byte { return OP_NOP }

func LDM(v byte) byte { return OP_LDM | nibble(v) }

func XCH(r byte) byte { return OP_XCH | nibble(r) }

func INC(r byte) byte { return OP_INC | nibble(r) }

func ADD(r byte) byte { return OP_ADD | nibble(r) }

func LD(r byte) byte { return OP_LD | nibble(r) }

func SUB(r byte) byte { return OP_SUB | nibble(r) }

func IAC() byte { return OP_IAC }

func DAC() byte { return OP_DAC }

func CMA() byte { return OP_CMA }

func CLB() byte { return OP_CLB }

func CLC() byte { return OP_CLC }

func STC() byte { return OP_STC }

func CMC() byte { return OP_CMC }

func RAL() byte { return OP_RAL }

func RAR() byte { return OP_RAR }

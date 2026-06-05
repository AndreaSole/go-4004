package cpu

// Queste funzioni semplificano la creazione di istruzioni restituendo il byte opcode corretto per ciascuna istruzione
func NOP() byte       { return OP_NOP }
func LDM(v byte) byte { return OP_LDM | nibble(v) }
func XCH(r byte) byte { return OP_XCH | nibble(r) }
func INC(r byte) byte { return OP_INC | nibble(r) }
func ADD(r byte) byte { return OP_ADD | nibble(r) }
func LD(r byte) byte  { return OP_LD | nibble(r) }
func SUB(r byte) byte { return OP_SUB | nibble(r) }
func IAC() byte       { return OP_IAC }
func DAC() byte       { return OP_DAC }
func CMA() byte       { return OP_CMA }
func CLB() byte       { return OP_CLB }
func CLC() byte       { return OP_CLC }
func STC() byte       { return OP_STC }
func CMC() byte       { return OP_CMC }
func RAL() byte       { return OP_RAL }
func RAR() byte       { return OP_RAR }
func TCC() byte       { return OP_TCC }
func TCS() byte       { return OP_TCS }
func DAA() byte       { return OP_DAA }
func KBP() byte       { return OP_KBP }
func DCL() byte       { return OP_DCL }
func BBL(v byte) byte { return OP_BBL | nibble(v) }

// Istruzioni di salto — restituiscono il primo byte; il secondo byte (indirizzo/dato) va aggiunto separatamente nella ROM.
func JUN(addrHigh byte) byte { return OP_JUN | (addrHigh & 0x0F) }
func JMS(addrHigh byte) byte { return OP_JMS | (addrHigh & 0x0F) }
func JCN(cond byte) byte     { return OP_JCN | (cond & 0x0F) }
func ISZ(r byte) byte        { return OP_ISZ | nibble(r) }
func FIM(rp byte) byte       { return OP_FIM | (nibble(rp) &^ 1) } // rp pari: coppia Rr/Rr+1
func SRC(rp byte) byte       { return OP_SRC | (nibble(rp) &^ 1) }
func FIN(rp byte) byte       { return OP_FIN | (nibble(rp) &^ 1) }
func JIN(rp byte) byte       { return OP_JIN | (nibble(rp) &^ 1) }

// istruzioni per la ram
func WRM() byte { return OP_WRM }
func RDM() byte { return OP_RDM }
func ADM() byte { return OP_ADM }

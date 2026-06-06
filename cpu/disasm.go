package cpu

import "fmt"

// Disassemble restituisce il mnemonic leggibile di un opcode Intel 4004.
//
// Per le istruzioni "famiglia" (registro nel nibble basso) mostra anche
// il numero del registro: "ADD R3", "XCH RF".
// Per le istruzioni a 2 byte mostra ".." al posto dell'argomento mancante:
// "FIM R0,..", "JUN 4..", "ISZ R2,..".
func Disassemble(op byte) string {
	switch {

	// ── Byte singolo ──────────────────────────────────────────────────────────
	case op == OP_NOP:
		return "NOP"

	// ── Gruppo registro (0x6X–0xDX) ──────────────────────────────────────────
	case op&0xF0 == OP_INC:
		return fmt.Sprintf("INC R%X", op&0x0F)
	case op&0xF0 == OP_ADD:
		return fmt.Sprintf("ADD R%X", op&0x0F)
	case op&0xF0 == OP_SUB:
		return fmt.Sprintf("SUB R%X", op&0x0F)
	case op&0xF0 == OP_LD:
		return fmt.Sprintf("LD  R%X", op&0x0F)
	case op&0xF0 == OP_XCH:
		return fmt.Sprintf("XCH R%X", op&0x0F)
	case op&0xF0 == OP_BBL:
		return fmt.Sprintf("BBL %X", op&0x0F)
	case op&0xF0 == OP_LDM:
		return fmt.Sprintf("LDM %X", op&0x0F)

	// ── Gruppo accumulatore (0xFX) ────────────────────────────────────────────
	case op == OP_CLB:
		return "CLB"
	case op == OP_CLC:
		return "CLC"
	case op == OP_IAC:
		return "IAC"
	case op == OP_CMC:
		return "CMC"
	case op == OP_CMA:
		return "CMA"
	case op == OP_RAL:
		return "RAL"
	case op == OP_RAR:
		return "RAR"
	case op == OP_TCC:
		return "TCC"
	case op == OP_DAC:
		return "DAC"
	case op == OP_TCS:
		return "TCS"
	case op == OP_STC:
		return "STC"
	case op == OP_DAA:
		return "DAA"
	case op == OP_KBP:
		return "KBP"
	case op == OP_DCL:
		return "DCL"

	// ── Salti e indirizzamento (0x1X–0x7X) ───────────────────────────────────
	case op&0xF0 == OP_JCN:
		return fmt.Sprintf("JCN %X,..", op&0x0F)
	case op&0xF0 == 0x20 && op&0x01 == 0: // FIM Rr,d — 2 byte
		return fmt.Sprintf("FIM R%X,..", op&0x0E)
	case op&0xF0 == 0x20 && op&0x01 == 1: // SRC Rr — 1 byte
		return fmt.Sprintf("SRC R%X", op&0x0E)
	case op&0xF0 == 0x30 && op&0x01 == 0: // FIN Rr — 1 byte
		return fmt.Sprintf("FIN R%X", op&0x0E)
	case op&0xF0 == 0x30 && op&0x01 == 1: // JIN Rr — 1 byte
		return fmt.Sprintf("JIN R%X", op&0x0E)
	case op&0xF0 == OP_JUN:
		return fmt.Sprintf("JUN %X..", op&0x0F)
	case op&0xF0 == OP_JMS:
		return fmt.Sprintf("JMS %X..", op&0x0F)
	case op&0xF0 == OP_ISZ:
		return fmt.Sprintf("ISZ R%X,..", op&0x0F)

	// ── I/O e RAM (0xEX) ─────────────────────────────────────────────────────
	case op == OP_WRM:
		return "WRM"
	case op == OP_WMP:
		return "WMP"
	case op == OP_WRR:
		return "WRR"
	case op == OP_WPM:
		return "WPM"
	case op == OP_WR0:
		return "WR0"
	case op == OP_WR1:
		return "WR1"
	case op == OP_WR2:
		return "WR2"
	case op == OP_WR3:
		return "WR3"
	case op == OP_SBM:
		return "SBM"
	case op == OP_RDM:
		return "RDM"
	case op == OP_RDR:
		return "RDR"
	case op == OP_ADM:
		return "ADM"
	case op == OP_RD0:
		return "RD0"
	case op == OP_RD1:
		return "RD1"
	case op == OP_RD2:
		return "RD2"
	case op == OP_RD3:
		return "RD3"

	default:
		return fmt.Sprintf("??? %02X", op)
	}
}

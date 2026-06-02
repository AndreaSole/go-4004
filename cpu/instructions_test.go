package cpu

import "testing"

// TestNOP verifica che NOP non modifichi alcuno stato della CPU
func TestNOP(t *testing.T) {
	c := NewCPU4004()
	c.A = 9
	c.C = true
	c.R[R0] = 4
	c.PC = 12

	if err := c.Execute(NOP()); err != nil {
		t.Fatal(err)
	}

	if c.A != 9 {
		t.Fatalf("expected A unchanged, got A=%d", c.A)
	}
	if !c.C {
		t.Fatal("expected carry unchanged")
	}
	if c.R[R0] != 4 {
		t.Fatalf("expected R0 unchanged, got R0=%d", c.R[R0])
	}
	if c.PC != 12 {
		t.Fatalf("expected PC unchanged, got PC=%d", c.PC)
	}
}

// --- LDM ---

func TestLDM(t *testing.T) {
	c := NewCPU4004()
	if err := c.Execute(LDM(7)); err != nil {
		t.Fatal(err)
	}
	if c.A != 7 {
		t.Fatalf("expected A=7, got A=%d", c.A)
	}
}

// --- LD ---

func TestLD(t *testing.T) {
	c := NewCPU4004()
	c.A = 1
	c.R[R2] = 9
	if err := c.Execute(LD(R2)); err != nil {
		t.Fatal(err)
	}
	if c.A != 9 {
		t.Fatalf("expected A=9, got A=%d", c.A)
	}
	if c.R[R2] != 9 {
		t.Fatalf("expected R2 unchanged, got R2=%d", c.R[R2])
	}
}

// --- XCH ---

func TestXCH(t *testing.T) {
	c := NewCPU4004()
	c.A = 5
	c.R[R0] = 2
	if err := c.Execute(XCH(R0)); err != nil {
		t.Fatal(err)
	}
	if c.A != 2 {
		t.Fatalf("expected A=2, got A=%d", c.A)
	}
	if c.R[R0] != 5 {
		t.Fatalf("expected R0=5, got R0=%d", c.R[R0])
	}
}

// --- INC ---

func TestINC(t *testing.T) {
	c := NewCPU4004()
	c.R[R1] = 3
	if err := c.Execute(INC(R1)); err != nil {
		t.Fatal(err)
	}
	if c.R[R1] != 4 {
		t.Fatalf("expected R1=4, got R1=%d", c.R[R1])
	}
}

func TestINCWrapsToNibble(t *testing.T) {
	c := NewCPU4004()
	c.R[R1] = 0x0F
	if err := c.Execute(INC(R1)); err != nil {
		t.Fatal(err)
	}
	if c.R[R1] != 0 {
		t.Fatalf("expected R1=0 after wrap, got R1=%d", c.R[R1])
	}
}

// --- ADD ---

func TestADD(t *testing.T) {
	c := NewCPU4004()
	c.A = 3
	c.R[R0] = 2
	if err := c.Execute(ADD(R0)); err != nil {
		t.Fatal(err)
	}
	if c.A != 5 {
		t.Fatalf("expected A=5, got A=%d", c.A)
	}
	if c.C {
		t.Fatal("expected carry=false")
	}
}

func TestADDWithCarryOut(t *testing.T) {
	c := NewCPU4004()
	c.A = 0x0F
	c.R[R0] = 1
	if err := c.Execute(ADD(R0)); err != nil {
		t.Fatal(err)
	}
	if c.A != 0 {
		t.Fatalf("expected A=0 after overflow, got A=%d", c.A)
	}
	if !c.C {
		t.Fatal("expected carry=true")
	}
}

func TestADDWithExistingCarry(t *testing.T) {
	c := NewCPU4004()
	c.A = 2
	c.R[R0] = 3
	c.C = true
	if err := c.Execute(ADD(R0)); err != nil {
		t.Fatal(err)
	}
	if c.A != 6 {
		t.Fatalf("expected A=6, got A=%d", c.A)
	}
	if c.C {
		t.Fatal("expected carry=false")
	}
}

// --- SUB ---

func TestSUB(t *testing.T) {
	c := NewCPU4004()
	c.A = 7
	c.R[R2] = 3
	if err := c.Execute(SUB(R2)); err != nil {
		t.Fatal(err)
	}
	if c.A != 4 {
		t.Errorf("A = %d, want 4", c.A)
	}
	if c.C {
		t.Error("C = true, want false")
	}
}

func TestSUBWithBorrow(t *testing.T) {
	c := NewCPU4004()
	c.A = 3
	c.R[R2] = 7
	if err := c.Execute(SUB(R2)); err != nil {
		t.Fatal(err)
	}
	if c.A != 12 { // 3 - 7 = -4 → nibble(12)
		t.Errorf("A = %d, want 12", c.A)
	}
	if !c.C {
		t.Error("C = false, want true")
	}
}

func TestSUBWithInitialBorrow(t *testing.T) {
	c := NewCPU4004()
	c.A = 5
	c.R[R2] = 3
	c.C = true
	if err := c.Execute(SUB(R2)); err != nil {
		t.Fatal(err)
	}
	if c.A != 1 { // 5 - 3 - 1 = 1
		t.Errorf("A = %d, want 1", c.A)
	}
	if c.C {
		t.Error("C = true, want false")
	}
}

// --- IAC ---

func TestIAC(t *testing.T) {
	c := NewCPU4004()
	c.A = 5
	if err := c.Execute(IAC()); err != nil {
		t.Fatal(err)
	}
	if c.A != 6 {
		t.Errorf("A = %d, want 6", c.A)
	}
	if c.C {
		t.Error("C = true, want false")
	}
}

func TestIACOverflow(t *testing.T) {
	c := NewCPU4004()
	c.A = 0x0F
	if err := c.Execute(IAC()); err != nil {
		t.Fatal(err)
	}
	if c.A != 0 {
		t.Errorf("A = %d, want 0", c.A)
	}
	if !c.C {
		t.Error("C = false, want true")
	}
}

// --- DAC ---

func TestDAC(t *testing.T) {
	c := NewCPU4004()
	c.A = 5
	if err := c.Execute(DAC()); err != nil {
		t.Fatal(err)
	}
	if c.A != 4 {
		t.Errorf("A = %d, want 4", c.A)
	}
	if c.C {
		t.Error("C = true, want false")
	}
}

func TestDACUnderflow(t *testing.T) {
	c := NewCPU4004()
	c.A = 0
	if err := c.Execute(DAC()); err != nil {
		t.Fatal(err)
	}
	if c.A != 0x0F {
		t.Errorf("A = %d, want 15", c.A)
	}
	if !c.C {
		t.Error("C = false, want true")
	}
}

// --- CMA ---

func TestCMA(t *testing.T) {
	c := NewCPU4004()
	c.A = 0b0101 // 5
	if err := c.Execute(CMA()); err != nil {
		t.Fatal(err)
	}
	if c.A != 0b1010 {
		t.Errorf("A = %d, want 10", c.A)
	}
	if c.C {
		t.Error("C = true, want false (CMA does not affect carry)")
	}
}

// --- CLB ---

func TestCLB(t *testing.T) {
	c := NewCPU4004()
	c.A = 9
	c.C = true
	if err := c.Execute(CLB()); err != nil {
		t.Fatal(err)
	}
	if c.A != 0 {
		t.Errorf("A = %d, want 0", c.A)
	}
	if c.C {
		t.Error("C = true, want false")
	}
}

// --- CLC ---

func TestCLC(t *testing.T) {
	c := NewCPU4004()
	c.A = 7
	c.C = true
	if err := c.Execute(CLC()); err != nil {
		t.Fatal(err)
	}
	if c.A != 7 {
		t.Errorf("A = %d, want 7 (unchanged)", c.A)
	}
	if c.C {
		t.Error("C = true, want false")
	}
}

// --- STC ---

func TestSTC(t *testing.T) {
	c := NewCPU4004()
	c.C = false
	if err := c.Execute(STC()); err != nil {
		t.Fatal(err)
	}
	if !c.C {
		t.Error("C = false, want true")
	}
}

// --- CMC ---

func TestCMCSetToFalse(t *testing.T) {
	c := NewCPU4004()
	c.C = true
	if err := c.Execute(CMC()); err != nil {
		t.Fatal(err)
	}
	if c.C {
		t.Error("C = true, want false")
	}
}

func TestCMCSetToTrue(t *testing.T) {
	c := NewCPU4004()
	c.C = false
	if err := c.Execute(CMC()); err != nil {
		t.Fatal(err)
	}
	if !c.C {
		t.Error("C = false, want true")
	}
}

// --- RAL ---

func TestRAL(t *testing.T) {
	c := NewCPU4004()
	c.A = 0b0110
	c.C = false
	if err := c.Execute(RAL()); err != nil {
		t.Fatal(err)
	}
	if c.A != 0b1100 {
		t.Errorf("A = %04b, want 1100", c.A)
	}
	if c.C {
		t.Error("C = true, want false")
	}
}

func TestRALCarryIn(t *testing.T) {
	c := NewCPU4004()
	c.A = 0b0110
	c.C = true
	if err := c.Execute(RAL()); err != nil {
		t.Fatal(err)
	}
	if c.A != 0b1101 {
		t.Errorf("A = %04b, want 1101", c.A)
	}
	if c.C {
		t.Error("C = true, want false")
	}
}

func TestRALCarryOut(t *testing.T) {
	c := NewCPU4004()
	c.A = 0b1010
	c.C = false
	if err := c.Execute(RAL()); err != nil {
		t.Fatal(err)
	}
	if c.A != 0b0100 {
		t.Errorf("A = %04b, want 0100", c.A)
	}
	if !c.C {
		t.Error("C = false, want true")
	}
}

// --- RAR ---

func TestRAR(t *testing.T) {
	c := NewCPU4004()
	c.A = 0b0110
	c.C = false
	if err := c.Execute(RAR()); err != nil {
		t.Fatal(err)
	}
	if c.A != 0b0011 {
		t.Errorf("A = %04b, want 0011", c.A)
	}
	if c.C {
		t.Error("C = true, want false")
	}
}

func TestRARCarryIn(t *testing.T) {
	c := NewCPU4004()
	c.A = 0b0100
	c.C = true
	if err := c.Execute(RAR()); err != nil {
		t.Fatal(err)
	}
	if c.A != 0b1010 {
		t.Errorf("A = %04b, want 1010", c.A)
	}
	if c.C {
		t.Error("C = true, want false")
	}
}

func TestRARCarryOut(t *testing.T) {
	c := NewCPU4004()
	c.A = 0b0101
	c.C = false
	if err := c.Execute(RAR()); err != nil {
		t.Fatal(err)
	}
	if c.A != 0b0010 {
		t.Errorf("A = %04b, want 0010", c.A)
	}
	if !c.C {
		t.Error("C = false, want true")
	}
}

// --- TCC ---

func TestTCCWithCarrySet(t *testing.T) {
	c := NewCPU4004()
	c.A = 9
	c.C = true
	if err := c.Execute(TCC()); err != nil {
		t.Fatal(err)
	}
	if c.A != 1 {
		t.Errorf("A = %d, want 1", c.A)
	}
	if c.C {
		t.Error("C = true, want false")
	}
}

func TestTCCWithCarryClear(t *testing.T) {
	c := NewCPU4004()
	c.A = 9
	c.C = false
	if err := c.Execute(TCC()); err != nil {
		t.Fatal(err)
	}
	if c.A != 0 {
		t.Errorf("A = %d, want 0", c.A)
	}
	if c.C {
		t.Error("C = true, want false")
	}
}

// --- TCS ---

func TestTCSWithCarrySet(t *testing.T) {
	c := NewCPU4004()
	c.C = true
	if err := c.Execute(TCS()); err != nil {
		t.Fatal(err)
	}
	if c.A != 10 {
		t.Errorf("A = %d, want 10", c.A)
	}
	if c.C {
		t.Error("C = true, want false")
	}
}

func TestTCSWithCarryClear(t *testing.T) {
	c := NewCPU4004()
	c.C = false
	if err := c.Execute(TCS()); err != nil {
		t.Fatal(err)
	}
	if c.A != 9 {
		t.Errorf("A = %d, want 9", c.A)
	}
	if c.C {
		t.Error("C = true, want false")
	}
}

// --- DAA ---

func TestDAANoAdjust(t *testing.T) {
	c := NewCPU4004()
	c.A = 7
	c.C = false
	if err := c.Execute(DAA()); err != nil {
		t.Fatal(err)
	}
	if c.A != 7 {
		t.Errorf("A = %d, want 7", c.A)
	}
	if c.C {
		t.Error("C = true, want false")
	}
}

func TestDAAInvalidBCD(t *testing.T) {
	c := NewCPU4004()
	c.A = 13 // 8+5, risultato invalido BCD
	c.C = false
	if err := c.Execute(DAA()); err != nil {
		t.Fatal(err)
	}
	if c.A != 3 {
		t.Errorf("A = %d, want 3", c.A)
	}
	if !c.C {
		t.Error("C = false, want true")
	}
}

func TestDAAWithCarry(t *testing.T) {
	c := NewCPU4004()
	c.A = 1 // 9+8=17 → A=1, C=true dopo ADD
	c.C = true
	if err := c.Execute(DAA()); err != nil {
		t.Fatal(err)
	}
	if c.A != 7 {
		t.Errorf("A = %d, want 7", c.A)
	}
	if !c.C {
		t.Error("C = false, want true")
	}
}

// --- KBP ---

func TestKBP(t *testing.T) {
	tests := []struct {
		input    uint8
		expected uint8
	}{
		{0b0000, 0},
		{0b0001, 1},
		{0b0010, 2},
		{0b0100, 3},
		{0b1000, 4},
		{0b0011, 0xF},
		{0b1111, 0xF},
	}
	for _, tt := range tests {
		c := NewCPU4004()
		c.A = tt.input
		if err := c.Execute(KBP()); err != nil {
			t.Fatalf("input=0b%04b: %v", tt.input, err)
		}
		if c.A != tt.expected {
			t.Errorf("input=0b%04b: A = %d, want %d", tt.input, c.A, tt.expected)
		}
	}
}

func TestKBPDoesNotAffectCarry(t *testing.T) {
	c := NewCPU4004()
	c.A = 0b0001
	c.C = true
	if err := c.Execute(KBP()); err != nil {
		t.Fatal(err)
	}
	if !c.C {
		t.Error("C = false, want true (KBP should not affect carry)")
	}
}

// --- DCL ---

func TestDCL(t *testing.T) {
	c := NewCPU4004()
	c.A = 3
	if err := c.Execute(DCL()); err != nil {
		t.Fatal(err)
	}
	if c.CL != 3 {
		t.Errorf("CL = %d, want 3", c.CL)
	}
	if c.A != 3 {
		t.Errorf("A = %d, want 3 (unchanged)", c.A)
	}
}

func TestDCLDoesNotAffectCarry(t *testing.T) {
	c := NewCPU4004()
	c.A = 2
	c.C = true
	if err := c.Execute(DCL()); err != nil {
		t.Fatal(err)
	}
	if !c.C {
		t.Error("C = false, want true (DCL should not affect carry)")
	}
}

func TestDCLMasksHighBits(t *testing.T) {
	c := NewCPU4004()
	c.A = 9 // 0b1001 → & 0x07 = 1
	if err := c.Execute(DCL()); err != nil {
		t.Fatal(err)
	}
	if c.CL != 1 {
		t.Errorf("CL = %d, want 1 (A=9, 9 & 0x07 = 1)", c.CL)
	}
}

// --- BBL ---

func TestBBLRestoresPC(t *testing.T) {
	c := NewCPU4004()
	c.push(0x123) // simula un JMS che ha salvato l'indirizzo di ritorno
	if err := c.Execute(BBL(5)); err != nil {
		t.Fatal(err)
	}
	if c.PC != 0x123 {
		t.Errorf("PC = 0x%03X, want 0x123", c.PC)
	}
	if c.A != 5 {
		t.Errorf("A = %d, want 5", c.A)
	}
}

func TestBBLZero(t *testing.T) {
	c := NewCPU4004()
	c.push(0x050)
	if err := c.Execute(BBL(0)); err != nil {
		t.Fatal(err)
	}
	if c.PC != 0x050 {
		t.Errorf("PC = 0x%03X, want 0x050", c.PC)
	}
	if c.A != 0 {
		t.Errorf("A = %d, want 0", c.A)
	}
}

func TestBBLDoesNotAffectCarry(t *testing.T) {
	c := NewCPU4004()
	c.C = true
	c.push(0x001)
	if err := c.Execute(BBL(0)); err != nil {
		t.Fatal(err)
	}
	if !c.C {
		t.Error("C = false, want true (BBL should not affect carry)")
	}
}

func TestBBLOnEmptyStackReturnsZero(t *testing.T) {
	c := NewCPU4004()
	// SP=0: BBL senza JMS precedente — comportamento indefinito sull'hardware,
	// l'emulatore restituisce PC=0 senza underflow di SP.
	if err := c.Execute(BBL(0)); err != nil {
		t.Fatal(err)
	}
	if c.PC != 0 {
		t.Errorf("PC = 0x%03X, want 0x000", c.PC)
	}
	if c.SP != 0 {
		t.Errorf("SP = %d, want 0 (non deve andare in underflow)", c.SP)
	}
}

// --- JUN ---

func TestJUN(t *testing.T) {
	c := NewCPU4004()
	rom := NewROM(make([]byte, 4096))
	// JUN 0x3AB: primo byte 0x43, secondo byte 0xAB
	rom.Data[0x000] = JUN(0x3) // opcode: 0x43
	rom.Data[0x001] = 0xAB
	if err := c.Step(rom); err != nil {
		t.Fatal(err)
	}
	if c.PC != 0x3AB {
		t.Errorf("PC = 0x%03X, want 0x3AB", c.PC)
	}
}

func TestJUNDoesNotModifyRegisters(t *testing.T) {
	c := NewCPU4004()
	c.A = 7
	c.C = true
	rom := NewROM(make([]byte, 4096))
	rom.Data[0x000] = JUN(0x0)
	rom.Data[0x001] = 0x10
	if err := c.Step(rom); err != nil {
		t.Fatal(err)
	}
	if c.A != 7 {
		t.Errorf("A = %d, want 7 (JUN should not touch A)", c.A)
	}
	if !c.C {
		t.Error("C = false, want true (JUN should not touch carry)")
	}
}

// --- JMS ---

func TestJMS(t *testing.T) {
	c := NewCPU4004()
	c.PC = 0x100
	rom := NewROM(make([]byte, 4096))
	// JMS 0x2CD: primo byte 0x52, secondo byte 0xCD. PC parte da 0x100.
	rom.Data[0x100] = JMS(0x2) // opcode: 0x52
	rom.Data[0x101] = 0xCD
	if err := c.Step(rom); err != nil {
		t.Fatal(err)
	}
	if c.PC != 0x2CD {
		t.Errorf("PC = 0x%03X, want 0x2CD", c.PC)
	}
	// L'indirizzo di ritorno deve essere 0x102 (dopo i 2 byte di JMS)
	c.SP-- // pop manuale per verificare
	ret := c.Stack[c.SP%3]
	if ret != 0x102 {
		t.Errorf("return address on stack = 0x%03X, want 0x102", ret)
	}
}

func TestJMSAndBBLRoundtrip(t *testing.T) {
	c := NewCPU4004()
	rom := NewROM(make([]byte, 4096))
	// JMS a 0x010, poi BBL 3 al ritorno
	rom.Data[0x000] = JMS(0x0) // 0x50
	rom.Data[0x001] = 0x10
	rom.Data[0x002] = NOP() // istruzione al ritorno
	rom.Data[0x010] = BBL(3)

	if err := c.Step(rom); err != nil { // JMS
		t.Fatal(err)
	}
	if c.PC != 0x010 {
		t.Fatalf("after JMS: PC = 0x%03X, want 0x010", c.PC)
	}
	if err := c.Step(rom); err != nil { // BBL 3
		t.Fatal(err)
	}
	if c.PC != 0x002 {
		t.Errorf("after BBL: PC = 0x%03X, want 0x002", c.PC)
	}
	if c.A != 3 {
		t.Errorf("after BBL: A = %d, want 3", c.A)
	}
}

// --- JCN ---

func TestJCNCarrySet(t *testing.T) {
	c := NewCPU4004()
	c.C = true
	rom := NewROM(make([]byte, 4096))
	// JCN con C2=1 (salta se carry=1): cond = 0b0010 = 2
	rom.Data[0x000] = JCN(0x2)
	rom.Data[0x001] = 0x50 // target: pagina 0, offset 0x50
	if err := c.Step(rom); err != nil {
		t.Fatal(err)
	}
	if c.PC != 0x050 {
		t.Errorf("PC = 0x%03X, want 0x050", c.PC)
	}
}

func TestJCNCarryClearNoJump(t *testing.T) {
	c := NewCPU4004()
	c.C = false
	rom := NewROM(make([]byte, 4096))
	// JCN con C2=1 (salta se carry=1): carry è falso → nessun salto
	rom.Data[0x000] = JCN(0x2)
	rom.Data[0x001] = 0x50
	if err := c.Step(rom); err != nil {
		t.Fatal(err)
	}
	if c.PC != 0x002 { // PC avanzato di 2 (i 2 byte di JCN)
		t.Errorf("PC = 0x%03X, want 0x002 (no jump)", c.PC)
	}
}

func TestJCNAccZero(t *testing.T) {
	c := NewCPU4004()
	c.A = 0
	rom := NewROM(make([]byte, 4096))
	// JCN con C3=1 (salta se A=0): cond = 0b0100 = 4
	rom.Data[0x000] = JCN(0x4)
	rom.Data[0x001] = 0x30
	if err := c.Step(rom); err != nil {
		t.Fatal(err)
	}
	if c.PC != 0x030 {
		t.Errorf("PC = 0x%03X, want 0x030", c.PC)
	}
}

func TestJCNInvertedCondition(t *testing.T) {
	c := NewCPU4004()
	c.C = false
	rom := NewROM(make([]byte, 4096))
	// JCN con C4=1 C2=1 (salta se NOT carry=1, cioè se carry=0): cond = 0b1010 = 0xA
	rom.Data[0x000] = JCN(0xA)
	rom.Data[0x001] = 0x40
	if err := c.Step(rom); err != nil {
		t.Fatal(err)
	}
	if c.PC != 0x040 {
		t.Errorf("PC = 0x%03X, want 0x040 (inverted: jump because carry=0)", c.PC)
	}
}

// --- ISZ ---

func TestISZNoJumpWhenZero(t *testing.T) {
	c := NewCPU4004()
	c.R[R2] = 0x0F // 0xF + 1 = 0 → non salta (zero = uscita dal loop)
	rom := NewROM(make([]byte, 4096))
	rom.Data[0x000] = ISZ(R2)
	rom.Data[0x001] = 0x50
	if err := c.Step(rom); err != nil {
		t.Fatal(err)
	}
	if c.R[R2] != 0 {
		t.Errorf("R2 = %d, want 0", c.R[R2])
	}
	if c.PC != 0x002 { // nessun salto
		t.Errorf("PC = 0x%03X, want 0x002 (no jump on zero)", c.PC)
	}
}

func TestISZJumpWhenNotZero(t *testing.T) {
	c := NewCPU4004()
	c.R[R2] = 3 // 3 + 1 = 4 ≠ 0 → salta (continua il loop)
	rom := NewROM(make([]byte, 4096))
	rom.Data[0x000] = ISZ(R2)
	rom.Data[0x001] = 0x50
	if err := c.Step(rom); err != nil {
		t.Fatal(err)
	}
	if c.R[R2] != 4 {
		t.Errorf("R2 = %d, want 4", c.R[R2])
	}
	if c.PC != 0x050 {
		t.Errorf("PC = 0x%03X, want 0x050 (jump because not zero)", c.PC)
	}
}

func TestISZDoesNotAffectCarry(t *testing.T) {
	c := NewCPU4004()
	c.C = true
	c.R[R2] = 3
	rom := NewROM(make([]byte, 4096))
	rom.Data[0x000] = ISZ(R2)
	rom.Data[0x001] = 0x50
	if err := c.Step(rom); err != nil {
		t.Fatal(err)
	}
	if !c.C {
		t.Error("C = false, want true (ISZ non deve modificare il carry)")
	}
}

// --- FIM ---

func TestFIM(t *testing.T) {
	c := NewCPU4004()
	rom := NewROM(make([]byte, 4096))
	// FIM R0, 0xAB: R0 = 0xA, R1 = 0xB
	rom.Data[0x000] = FIM(R0) // 0x20
	rom.Data[0x001] = 0xAB
	if err := c.Step(rom); err != nil {
		t.Fatal(err)
	}
	if c.R[R0] != 0xA {
		t.Errorf("R0 = %X, want A", c.R[R0])
	}
	if c.R[R1] != 0xB {
		t.Errorf("R1 = %X, want B", c.R[R1])
	}
	if c.PC != 0x002 {
		t.Errorf("PC = 0x%03X, want 0x002", c.PC)
	}
}

func TestFIMPair4(t *testing.T) {
	c := NewCPU4004()
	rom := NewROM(make([]byte, 4096))
	// FIM R4, 0x37: R4 = 0x3, R5 = 0x7
	rom.Data[0x000] = FIM(R4) // 0x24
	rom.Data[0x001] = 0x37
	if err := c.Step(rom); err != nil {
		t.Fatal(err)
	}
	if c.R[R4] != 0x3 {
		t.Errorf("R4 = %X, want 3", c.R[R4])
	}
	if c.R[R5] != 0x7 {
		t.Errorf("R5 = %X, want 7", c.R[R5])
	}
}

// --- SRC ---

func TestSRC(t *testing.T) {
	c := NewCPU4004()
	c.R[R2] = 0x5 // chip/banco
	c.R[R3] = 0x3 // registro RAM
	if err := c.Execute(SRC(R2)); err != nil {
		t.Fatal(err)
	}
	if c.SRCAddr != 0x53 {
		t.Errorf("SRCAddr = 0x%02X, want 0x53", c.SRCAddr)
	}
}

func TestSRCDoesNotModifyRegisters(t *testing.T) {
	c := NewCPU4004()
	c.A = 9
	c.C = true
	c.R[R0] = 0x2
	c.R[R1] = 0x7
	if err := c.Execute(SRC(R0)); err != nil {
		t.Fatal(err)
	}
	if c.A != 9 {
		t.Errorf("A = %d, want 9 (SRC should not touch A)", c.A)
	}
	if !c.C {
		t.Error("C = false, want true (SRC should not touch carry)")
	}
	if c.R[R0] != 0x2 || c.R[R1] != 0x7 {
		t.Error("SRC should not modify registers")
	}
}

// --- FIN ---

func TestFIN(t *testing.T) {
	c := NewCPU4004()
	rom := NewROM(make([]byte, 4096))
	// R0=0x5, R1=0x8 → indirizzo pagina 0: 0x058
	// rom.Data[0x058] = 0xAB → R2=0xA, R3=0xB
	c.R[R0] = 0x5
	c.R[R1] = 0x8
	rom.Data[0x000] = FIN(R2) // 0x32
	rom.Data[0x058] = 0xAB    // byte da leggere
	c.PC = 0x000
	if err := c.Step(rom); err != nil {
		t.Fatal(err)
	}
	if c.R[R2] != 0xA {
		t.Errorf("R2 = %X, want A", c.R[R2])
	}
	if c.R[R3] != 0xB {
		t.Errorf("R3 = %X, want B", c.R[R3])
	}
	if c.PC != 0x001 { // FIN è 1 byte: PC avanza di 1
		t.Errorf("PC = 0x%03X, want 0x001", c.PC)
	}
}

func TestFINDoesNotChangeR0R1(t *testing.T) {
	c := NewCPU4004()
	rom := NewROM(make([]byte, 4096))
	c.R[R0] = 0x2
	c.R[R1] = 0x0
	rom.Data[0x000] = FIN(R4) // 0x34: carica in R4/R5, non in R0/R1
	rom.Data[0x020] = 0xCD
	c.PC = 0x000
	if err := c.Step(rom); err != nil {
		t.Fatal(err)
	}
	if c.R[R0] != 0x2 || c.R[R1] != 0x0 {
		t.Error("FIN should not modify R0/R1 (used as address, not destination)")
	}
	if c.R[R4] != 0xC || c.R[R5] != 0xD {
		t.Errorf("R4=%X R5=%X, want C D", c.R[R4], c.R[R5])
	}
}

// --- JIN ---

func TestJIN(t *testing.T) {
	c := NewCPU4004()
	c.R[R0] = 0x7
	c.R[R1] = 0x3
	c.PC = 0x000
	if err := c.Execute(JIN(R0)); err != nil {
		t.Fatal(err)
	}
	// indirizzo = (0x7 << 4) | 0x3 = 0x73, pagina corrente (PC & 0x0F00 = 0x000)
	if c.PC != 0x073 {
		t.Errorf("PC = 0x%03X, want 0x073", c.PC)
	}
}

func TestJINPreservesPage(t *testing.T) {
	c := NewCPU4004()
	c.R[R2] = 0x4
	c.R[R3] = 0x8
	c.PC = 0x200 // pagina 2
	if err := c.Execute(JIN(R2)); err != nil {
		t.Fatal(err)
	}
	// indirizzo = pagina 2 + (0x4 << 4 | 0x8) = 0x248
	if c.PC != 0x248 {
		t.Errorf("PC = 0x%03X, want 0x248", c.PC)
	}
}

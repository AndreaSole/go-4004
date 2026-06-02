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

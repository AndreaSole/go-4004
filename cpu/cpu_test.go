package cpu

import "testing"

// TestLDM verifica che l'istruzione LDM carichi correttamente un valore immediato nell'accumulatore (A)
func TestLDM(t *testing.T) {
	c := NewCPU4004()

	err := c.Execute(LDM(7))
	if err != nil {
		t.Fatal(err)
	}

	if c.A != 7 {
		t.Fatalf("expected A=7, got A=%d", c.A)
	}
}

// TestXCH verifica che l'istruzione XCH scambi correttamente i valori tra l'accumulatore (A) e un registro (Rn)
func TestXCH(t *testing.T) {
	c := NewCPU4004()

	c.A = 5
	c.R[R0] = 2

	err := c.Execute(XCH(R0))
	if err != nil {
		t.Fatal(err)
	}

	if c.A != 2 {
		t.Fatalf("expected A=2, got A=%d", c.A)
	}

	if c.R[R0] != 5 {
		t.Fatalf("expected R0=5, got R0=%d", c.R[R0])
	}
}

// TestINC verifica che l'istruzione INC incrementi correttamente il valore di un registro (Rn)
func TestINC(t *testing.T) {
	c := NewCPU4004()

	c.R[R1] = 3

	err := c.Execute(INC(R1))
	if err != nil {
		t.Fatal(err)
	}

	if c.R[R1] != 4 {
		t.Fatalf("expected R1=4, got R1=%d", c.R[R1])
	}
}

// TestINCWrapsToNibble verifica che l'istruzione INC incrementi correttamente un registro (Rn) e si assicuri che il valore rimanga entro 4 bit (0-15)
// Ad esempio, se R1 = 0x0F (15) e viene incrementato, dovrebbe tornare a 0x00 (0) senza overflow, poiché i registri sono a 4 bit
func TestINCWrapsToNibble(t *testing.T) {
	c := NewCPU4004()

	c.R[R1] = 0x0F

	err := c.Execute(INC(R1))
	if err != nil {
		t.Fatal(err)
	}

	if c.R[R1] != 0 {
		t.Fatalf("expected R1=0 after wrap, got R1=%d", c.R[R1])
	}
}

// TestADD verifica che l'istruzione ADD sommi correttamente i valori tra l'accumulatore (A) e un registro (Rn)
func TestADD(t *testing.T) {
	c := NewCPU4004()

	c.A = 3
	c.R[R0] = 2

	err := c.Execute(ADD(R0))
	if err != nil {
		t.Fatal(err)
	}

	if c.A != 5 {
		t.Fatalf("expected A=5, got A=%d", c.A)
	}

	if c.C {
		t.Fatal("expected carry=false")
	}
}

// TestADDWithCarryOut verifica che l'istruzione ADD gestisca correttamente il carry quando la somma supera 0x0F (15)
// Ad esempio, se A = 0x0F (15) e R0 = 1, il risultato dovrebbe essere A = 0x00 (0) con carry = true, poiché A è a 4 bit
func TestADDWithCarryOut(t *testing.T) {
	c := NewCPU4004()

	c.A = 0x0F
	c.R[R0] = 1

	err := c.Execute(ADD(R0))
	if err != nil {
		t.Fatal(err)
	}

	if c.A != 0 {
		t.Fatalf("expected A=0 after overflow, got A=%d", c.A)
	}

	if !c.C {
		t.Fatal("expected carry=true")
	}
}

// TestADDWithExistingCarry verifica che l'istruzione ADD consideri correttamente il carry esistente durante l'addizione
// Ad esempio, se A = 2, R0 = 3 e C = true, il risultato dovrebbe essere A = 6 (2 + 3 + 1) con carry = false, poiché il risultato è inferiore a 0x10 (16)
func TestADDWithExistingCarry(t *testing.T) {
	c := NewCPU4004()

	c.A = 2
	c.R[R0] = 3
	c.C = true

	err := c.Execute(ADD(R0))
	if err != nil {
		t.Fatal(err)
	}

	if c.A != 6 {
		t.Fatalf("expected A=6, got A=%d", c.A)
	}

	if c.C {
		t.Fatal("expected carry=false")
	}
}

// TestNOP verifica che l'istruzione NOP non modifichi lo stato del CPU, inclusi A, C, i registri e il program counter (PC)
// Ad esempio, se A = 9, C = true, R0 = 4 e PC = 12, dopo l'esecuzione di NOP, tutti questi valori dovrebbero rimanere invariati
func TestNOP(t *testing.T) {
	c := NewCPU4004()

	c.A = 9
	c.C = true
	c.R[R0] = 4
	c.PC = 12

	err := c.Execute(NOP())
	if err != nil {
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

// TestLD verifica che l'istruzione LD carichi correttamente il valore di un registro (Rn) nell'accumulatore (A)
func TestLD(t *testing.T) {
	c := NewCPU4004()

	c.A = 1
	c.R[R2] = 9

	err := c.Execute(LD(R2))
	if err != nil {
		t.Fatal(err)
	}

	if c.A != 9 {
		t.Fatalf("expected A=9, got A=%d", c.A)
	}

	if c.R[R2] != 9 {
		t.Fatalf("expected R2 unchanged, got R2=%d", c.R[R2])
	}
}

// TestSUB verifica che l'istruzione SUB sottragga correttamente il valore di un registro (Rn) dall'accumulatore (A) e aggiorni il carry (C) di conseguenza
// Ad esempio, se A = 7 e R2 = 3, dopo SUB R2, A dovrebbe essere 4 e C dovrebbe essere false, poiché non c'è borrow
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
	if c.C != false {
		t.Errorf("C = true, want false")
	}
}

// TestSUBWithBorrow verifica che l'istruzione SUB gestisca correttamente il borrow quando il valore del registro (Rn) è maggiore dell'accumulatore (A)
// Ad esempio, se A = 3 e R2 = 7, dopo SUB R2, A dovrebbe essere 12 (nibble(-4)) e C dovrebbe essere true, poiché c'è un borrow
func TestSUBWithBorrow(t *testing.T) {
	c := NewCPU4004()
	c.A = 3
	c.R[R2] = 7
	if err := c.Execute(SUB(R2)); err != nil {
		t.Fatal(err)
	}
	if c.A != 12 {
		t.Errorf("A = %d, want 12", c.A)
	} // 3 - 7 = -4 → nibble(12)
	if c.C != true {
		t.Errorf("C = false, want true")
	}
}

// TestIAC verifica che l'istruzione IAC incrementi correttamente l'accumulatore (A) di 1 e aggiorni il carry (C) se necessario
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

// TestIACOverflow verifica che l'istruzione IAC gestisca correttamente l'overflow quando l'accumulatore (A) è a 0x0F (15) e viene incrementato
// Ad esempio, se A = 0x0F (15), dopo IAC, A dovrebbe tornare a 0x00 (0) e C dovrebbe essere true, poiché c'è un overflow
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

// TestDAC verifica che l'istruzione DAC decrementi correttamente l'accumulatore (A) di 1 e aggiorni il carry (C) se necessario
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

// TestDACUnderflow verifica che l'istruzione DAC gestisca correttamente l'underflow quando l'accumulatore (A) è a 0x00 (0) e viene decrementato
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

// TestCMA verifica che l'istruzione CMA completi correttamente l'accumulatore (A) invertendo tutti i suoi bit	e che non modifichi il carry (C)
func TestCMA(t *testing.T) {
	c := NewCPU4004()
	c.A = 0b0101 // 5
	if err := c.Execute(CMA()); err != nil {
		t.Fatal(err)
	}
	if c.A != 0b1010 { // 10
		t.Errorf("A = %d, want 10", c.A)
	}
	if c.C {
		t.Error("C = true, want false (CMA does not affect carry)")
	}
}

// TestCLB verifica che l'istruzione CLB azzeri correttamente sia l'accumulatore (A) che il carry (C)
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

// TestCLC verifica che l'istruzione CLC azzeri correttamente solo il carry (C) lasciando intatto l'accumulatore (A)
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

// TestSTC verifica che l'istruzione STC imposti correttamente il carry (C) a true senza modificare l'accumulatore (A)
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

// TestCMC verifica che l'istruzione CMC completi correttamente il carry (C) invertendo il suo stato e che non modifichi l'accumulatore (A)
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

// TestCMCSetToTrue verifica che l'istruzione CMC completi correttamente il carry (C) invertendo il suo stato e che non modifichi l'accumulatore (A)
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

// TestRAL verifica che l'istruzione RAL ruoti correttamente i bit dell'accumulatore (A) a sinistra e sposti il bit più significativo nel carry (C)
// Ad esempio, se A = 0b0110 (6) e C = false, dopo RAL, A dovrebbe essere 0b1100 (12) e C dovrebbe essere true (il bit più significativo 1 è stato spostato nel carry)
func TestRAL(t *testing.T) {
	c := NewCPU4004()
	c.A = 0b0110 // 6
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

// TestRALCarryIn verifica che l'istruzione RAL consideri correttamente il carry esistente durante la rotazione a sinistra
// Ad esempio, se A = 0b0110 (6) e C = true, dopo RAL, A dovrebbe essere 0b1101 (13) e C dovrebbe essere true (il bit più significativo 1 è stato spostato nel carry e il vecchio carry true è stato spostato in A)
func TestRALCarryIn(t *testing.T) {
	c := NewCPU4004()
	c.A = 0b0110 // 6
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

// TestRALCarryOut verifica che l'istruzione RAL gestisca correttamente il carry quando il bit più significativo di A è 1o quando il vecchio carry è true
// Ad esempio, se A = 0b1010 (10) e C = false, dopo RAL, A dovrebbe essere 0b0100 (4) e C dovrebbe essere true (il bit più significativo 1 è stato spostato nel carry)
// Se A = 0b1010 (10) e C = true, dopo RAL, A dovrebbe essere 0b0101 (5) e C dovrebbe essere true (il bit più significativo 1 è stato spostato nel carry e il vecchio carry true è stato spostato in A)
func TestRALCarryOut(t *testing.T) {
	c := NewCPU4004()
	c.A = 0b1010 // 10
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

// TestRAR verifica che l'istruzione RAR ruoti correttamente i bit dell'accumulatore (A) a destra e sposti il bit più significativo nel carry (C)
// Ad esempio, se A = 0b0110 (6) e C = false, dopo RAR, A dovrebbe essere 0b0011 (3) e C dovrebbe essere false (il bit più significativo 0 è stato spostato nel carry)
func TestRAR(t *testing.T) {
	c := NewCPU4004()
	c.A = 0b0110 // 6
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

// TestRARCarryOut verifica che l'istruzione RAR gestisca correttamente il carry quando il bit più significativo di A è 1o quando il vecchio carry è true
// Ad esempio, se A = 0b0101 (5) e C = false, dopo RAR, A dovrebbe essere 0b0010 (2) e C dovrebbe essere true (il bit più significativo 1 è stato spostato nel carry)
// Se A = 0b0101 (5) e C = true, dopo RAR, A dovrebbe essere 0b0011 (3) e C dovrebbe essere true (il bit più significativo 1 è stato spostato nel carry e il vecchio carry true è stato spostato in A)
func TestRARCarryOut(t *testing.T) {
	c := NewCPU4004()
	c.A = 0b0101 // 5
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

func TestTCSWithCarrySet(t *testing.T) {
	c := NewCPU4004()
	c.A = 0
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
	c.A = 0
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

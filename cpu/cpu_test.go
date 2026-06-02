package cpu

import "testing"

// TestNewCPU4004 verifica che una nuova CPU sia inizializzata con tutti i valori a zero
func TestNewCPU4004(t *testing.T) {
	c := NewCPU4004()

	if c.A != 0 {
		t.Errorf("A = %d, want 0", c.A)
	}
	if c.C != false {
		t.Error("C = true, want false")
	}
	if c.PC != 0 {
		t.Errorf("PC = %d, want 0", c.PC)
	}
	if c.CL != 0 {
		t.Errorf("CL = %d, want 0", c.CL)
	}
	for i, v := range c.R {
		if v != 0 {
			t.Errorf("R[%d] = %d, want 0", i, v)
		}
	}
	if c.SP != 0 {
		t.Errorf("SP = %d, want 0", c.SP)
	}
	for i, v := range c.Stack {
		if v != 0 {
			t.Errorf("Stack[%d] = 0x%03X, want 0", i, v)
		}
	}
	if c.SRCAddr != 0 {
		t.Errorf("SRCAddr = 0x%02X, want 0", c.SRCAddr)
	}
}

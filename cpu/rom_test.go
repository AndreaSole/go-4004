package cpu

import "testing"

// TestStepExecutesOpcode verifica che Step legga e esegua l'opcode dalla ROM
func TestStepExecutesOpcode(t *testing.T) {
	c := NewCPU4004()
	rom := NewROM([]byte{LDM(7)})

	if err := c.Step(rom); err != nil {
		t.Fatal(err)
	}
	if c.A != 7 {
		t.Errorf("A = %d, want 7", c.A)
	}
}

// TestStepIncreasesPC verifica che Step incrementi PC ad ogni esecuzione
func TestStepIncreasesPC(t *testing.T) {
	c := NewCPU4004()
	rom := NewROM([]byte{LDM(1), LDM(2), LDM(3)})

	c.Step(rom)
	c.Step(rom)
	c.Step(rom)

	if c.PC != 3 {
		t.Errorf("PC = %d, want 3", c.PC)
	}
	if c.A != 3 {
		t.Errorf("A = %d, want 3", c.A)
	}
}

// TestStepPCWrapsAt12Bit verifica che PC torni a 0 dopo 0xFFF (limite 12 bit)
func TestStepPCWrapsAt12Bit(t *testing.T) {
	c := NewCPU4004()
	c.PC = 0x0FFF
	rom := NewROM(make([]byte, 0x1000)) // 4096 NOP

	if err := c.Step(rom); err != nil {
		t.Fatal(err)
	}
	if c.PC != 0 {
		t.Errorf("PC = 0x%X, want 0x000 (wrap)", c.PC)
	}
}

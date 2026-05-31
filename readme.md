# Go 4004 Emulator

Un emulatore dell'Intel 4004 scritto in Go, sviluppato passo passo con approccio didattico.

L'obiettivo finale è creare:

* un emulatore completo del processore Intel 4004
* una ROM virtuale
* RAM virtuale
* I/O virtuale
* un firmware che faccia funzionare il sistema come una calcolatrice da tavolo

Il progetto è sviluppato con focus su:

* comprensione dell'architettura CPU
* funzionamento dei primi microprocessori
* logica low-level
* emulatori
* sistemi a 4 bit
* organizzazione hardware/software

---

# Stato attuale

Attualmente il progetto include:

* CPU minima Intel 4004
* registri
* accumulator
* carry
* decoder opcode base
* helper functions stile mini assembler

Istruzioni implementate:

* NOP
* LDM
* LD
* XCH
* INC
* ADD
* SUB
* IAC
* DAC
* CMA
* CLB
* CLC
* STC
* CMC
* RAL
* RAR

---

# Obiettivo architetturale

L'emulatore deve rappresentare l'hardware.

La logica della calcolatrice NON sarà scritta direttamente in Go.

La calcolatrice sarà un programma caricato in una ROM virtuale, proprio come nei sistemi reali basati su Intel 4004.

---

# Struttura progetto

```text
go4004/
├── go.mod
├── main.go
├── README.md
└── cpu/
    ├── cpu.go
    ├── opcodes.go
    ├── helpers.go
    ├── instructions.go
    └── cpu_test.go
```

---

# File principali

## cpu/cpu.go

Contiene:

* struttura CPU
* accumulator
* carry
* registri
* program counter
* utility nibble()

---

## cpu/opcodes.go

Contiene:

* costanti registri
* costanti opcode

---

## cpu/helpers.go

Contiene helper functions per costruire opcode leggibili.

Esempio:

```go
cpu.LDM(2)
cpu.XCH(cpu.R0)
cpu.ADD(cpu.R0)
```

---

## cpu/instructions.go

Contiene il decoder ed executor delle istruzioni:

```go
func (c *CPU4004) Execute(op byte) error
```

---

## cpu/cpu_test.go

Conterrà i test automatici.

---

# Stato CPU attuale

## Registri

Il 4004 contiene:

* accumulator A
* carry flag C
* 16 registri da 4 bit
* program counter

Rappresentazione attuale:

```go
type CPU4004 struct {
	A  uint8
	C  bool
	R  [16]uint8
	PC uint16
}
```

---

# Gestione nibble

Il 4004 è un processore a 4 bit.

Go non supporta tipi da 4 bit, quindi viene usato:

```go
uint8
```

limitando i valori ai primi 4 bit tramite:

```go
func nibble(v uint8) uint8 {
	return v & 0x0F
}
```

---

# Opcode

Le istruzioni vengono riconosciute tramite maschere bitwise.

Esempio:

```go
case op&0xF0 == OP_ADD:
```

Questo approccio è più vicino a come funzionano i decoder hardware reali.

---

# Helper functions

Per evitare opcode poco leggibili come:

```go
0xD2
0xB0
0x80
```

vengono usate helper functions:

```go
cpu.LDM(2)
cpu.XCH(cpu.R0)
cpu.ADD(cpu.R0)
```

Questo rende il codice:

* più leggibile
* più vicino all'assembly reale
* più facile da estendere
* preparato per un futuro assembler

---

# Esempio programma attuale

```go
program := []byte{
	cpu.LDM(2),
	cpu.XCH(cpu.R0),
	cpu.LDM(3),
	cpu.ADD(cpu.R0),
}
```

Questo programma esegue:

```text
2 + 3 = 5
```

---

# Esecuzione

Avvio progetto:

```bash
go run .
```

---

# Output atteso

```text
A = 5
Carry = false
R0 = 2
```

---

# Filosofia del progetto

Il progetto NON vuole essere solo un emulatore funzionante.

Vuole essere anche:

* uno strumento didattico
* un percorso di apprendimento
* una ricostruzione storica del funzionamento dei primi microprocessori

Lo sviluppo avviene in piccoli step:

1. implementazione
2. test
3. verifica
4. estensione

---

# Roadmap

## Step attuale

CPU minima:

* NOP
* LDM
* LD
* XCH
* INC
* ADD
* SUB
* IAC
* DAC
* CMA
* CLB
* CLC
* STC
* CMC
* RAL
* RAR


* test LDM
* test LD
* test XCH
* test INC
* test ADD
* test carry
* test NOP
* test SUB
* test SUB con borrow
* test IAC
* test IAC con overflow

## Step futuri

* tutte le 46 istruzioni Intel 4004
* ROM virtuale
* fetch/decode/execute reale
* stack
* RAM virtuale
* I/O virtuale
* debugger
* assembler minimale
* firmware calcolatrice
* display virtuale
* tastiera virtuale

---

# Informazioni sul set istruzioni

Il target finale è:

```text
46 istruzioni Intel 4004 complete
```

Molti opcode sono varianti della stessa istruzione.

Esempio:

```text
D0 = LDM 0
D1 = LDM 1
...
DF = LDM 15
```

Questa rappresenta:

```text
1 istruzione logica
16 encoding diversi
```

---

# Obiettivo finale firmware

La CPU dovrà eseguire un firmware da ROM virtuale che implementi una calcolatrice.

Loop firmware previsto:

```text
leggi tastiera
interpreta input
aggiorna stato
esegui operazione
aggiorna display
ripeti
```

---

# Note

Questo progetto è volutamente sviluppato senza scorciatoie ad alto livello.

L'obiettivo è comprendere realmente:

* fetch
* decode
* execute
* gestione memoria
* bus logici
* ALU
* stack
* architettura CPU
* funzionamento storico dei microprocessori

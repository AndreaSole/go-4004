# Go 4004 Emulator

Un emulatore dell'Intel 4004 scritto in Go, sviluppato passo passo con approccio didattico.

L'obiettivo finale è creare:

* un emulatore completo del processore Intel 4004
* una ROM virtuale
* RAM virtuale (chip Intel 4002)
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

**46 / 46 istruzioni implementate — set completo** 🎉

Gruppo registro — completo (8/8):

| Istruzione | Opcode | Descrizione |
|------------|--------|-------------|
| NOP        | 0x00   | No operation |
| INC Rr     | 0x6r   | Incrementa registro |
| ADD Rr     | 0x8r   | Somma registro ad A con carry |
| SUB Rr     | 0x9r   | Sottrae registro da A con borrow |
| LD Rr      | 0xAr   | Carica registro in A |
| XCH Rr     | 0xBr   | Scambia A con registro |
| BBL n      | 0xCr   | Ritorno da subroutine, carica n in A |
| LDM n      | 0xDr   | Carica immediato in A |

Gruppo accumulatore — completo (14/14):

| Istruzione | Opcode | Descrizione |
|------------|--------|-------------|
| CLB        | 0xF0   | A=0, C=false |
| CLC        | 0xF1   | C=false |
| IAC        | 0xF2   | A++ |
| CMC        | 0xF3   | C=!C |
| CMA        | 0xF4   | A=~A |
| RAL        | 0xF5   | Ruota A a sinistra attraverso C |
| RAR        | 0xF6   | Ruota A a destra attraverso C |
| TCC        | 0xF7   | A=C, C=false |
| DAC        | 0xF8   | A-- |
| TCS        | 0xF9   | A=9 o 10 in base a C, C=false |
| STC        | 0xFA   | C=true |
| DAA        | 0xFB   | Correzione BCD dopo addizione |
| KBP        | 0xFC   | Decodifica one-hot in posizione |
| DCL        | 0xFD   | CL=A, seleziona banco RAM |

Gruppo salti e indirizzamento — completo (8/8):

| Istruzione | Opcode      | Byte | Descrizione |
|------------|-------------|------|-------------|
| JCN c,a    | 0x1c + byte | 2    | Salto condizionale |
| FIM Rr,d   | 0x2r + byte | 2    | Fetch immediato in coppia registro |
| SRC Rr     | 0x2r+1      | 1    | Imposta indirizzo RAM (SRCAddr) |
| FIN Rr     | 0x3r        | 1    | Fetch indiretto da ROM via R0:R1 |
| JIN Rr     | 0x3r+1      | 1    | Salto indiretto via coppia registro |
| JUN a      | 0x4r + byte | 2    | Salto incondizionale a 12 bit |
| JMS a      | 0x5r + byte | 2    | Salto a subroutine (push PC) |
| ISZ Rr,a   | 0x7r + byte | 2    | Incrementa registro, salta se != 0 |

Gruppo I/O e RAM — completo (16/16):

| Istruzione | Opcode | Descrizione |
|------------|--------|-------------|
| WRM        | 0xE0   | Scrive A nella RAM data |
| WMP        | 0xE1   | Scrive A sulla porta di output RAM |
| WR0        | 0xE4   | Scrive A nel nibble di stato 0 |
| WR1        | 0xE5   | Scrive A nel nibble di stato 1 |
| WR2        | 0xE6   | Scrive A nel nibble di stato 2 |
| WR3        | 0xE7   | Scrive A nel nibble di stato 3 |
| SBM        | 0xE8   | A = A - RAM - borrow |
| RDM        | 0xE9   | Legge RAM data in A |
| ADM        | 0xEB   | A = A + RAM + carry |
| RD0        | 0xEC   | A = nibble di stato 0 |
| RD1        | 0xED   | A = nibble di stato 1 |
| RD2        | 0xEE   | A = nibble di stato 2 |
| RD3        | 0xEF   | A = nibble di stato 3 |
| WRR        | 0xE2   | Scrive A sulla porta output ROM |
| RDR        | 0xEA   | A = porta input ROM (tastiera) |
| WPM        | 0xE3   | Write program memory (no-op) |

---

# Infrastruttura

* CPU Intel 4004 con 33/46 istruzioni
* ROM virtuale — `cpu/rom.go`
* RAM virtuale — `cpu/ram.go` (modello chip Intel 4002)
* Ciclo fetch-execute — `Step(rom, ram)`
* Program Counter a 12 bit (0x000–0xFFF)
* Stack hardware a 3 livelli

---

# Struttura progetto

```text
go-4004/
├── go.mod
├── readme.md
├── cmd/
│   └── go4004/
│       └── main.go         ← entry point / demo
├── docs/
│   ├── bcd.md              ← spiegazione codifica BCD
│   ├── istruzioni-registro.md
│   ├── istruzioni-accumulatore.md
│   ├── istruzioni-salto.md
│   └── istruzioni-io-ram.md
└── cpu/
    ├── cpu.go              ← struct CPU, Step(), stack
    ├── opcodes.go          ← costanti opcode in 4 gruppi
    ├── helpers.go          ← helper functions (mini assembler)
    ├── instructions.go     ← Execute(), executeWithArg(), executeIO()
    ├── rom.go              ← struct ROM
    ├── ram.go              ← struct RAM (Intel 4002)
    ├── cpu_test.go         ← test inizializzazione CPU
    ├── instructions_test.go ← test tutte le istruzioni
    └── rom_test.go         ← test Step(), PC wrap
```

---

# Struct CPU

```go
type CPU4004 struct {
    A       uint8     // Accumulator, 4 bit
    C       bool      // Carry flag
    R       [16]uint8 // Registri R0–RF, 4 bit ciascuno
    PC      uint16    // Program Counter, 12 bit (0x000–0xFFF)
    CL      uint8     // Command Line — banco RAM attivo (da DCL)
    Stack   [3]uint16 // Stack hardware a 3 livelli
    SP      uint8     // Stack pointer
    SRCAddr uint8     // Indirizzo RAM corrente (da SRC)
}
```

---

# Struct RAM

```go
type RAM struct {
    Data   [4][4][20]uint8 // [banco][registro][carattere]
    Status [4][4][4]uint8  // [banco][registro][status nibble]
    Port   [4]uint8        // porta di output per banco
}
```

---

# Gestione nibble

Il 4004 è un processore a 4 bit. Go usa `uint8` mascherato:

```go
func nibble(v uint8) uint8 {
    return v & 0x0F
}
```

---

# Opcode

Le istruzioni vengono riconosciute tramite maschere bitwise:

```go
case op&0xF0 == OP_ADD:   // famiglia di istruzioni (registro variabile)
case op == OP_IAC:         // istruzione singola (byte fisso)
```

---

# Helper functions

Ogni istruzione ha un helper che costruisce l'opcode corretto:

```go
cpu.LDM(7)        // 0xD7
cpu.ADD(cpu.R2)   // 0x82
cpu.JMS(0x3)      // 0x53 (primo byte di JMS 0x3AB)
cpu.WRM()         // 0xE0
```

---

# Esempio programma

Scrittura e lettura di un valore in RAM:

```go
rom.Data[0x000] = cpu.LDM(0)
rom.Data[0x001] = cpu.DCL()           // banco 0
rom.Data[0x002] = cpu.FIM(cpu.R0)
rom.Data[0x003] = 0x05                // R0=0, R1=5
rom.Data[0x004] = cpu.SRC(cpu.R0)    // SRCAddr = 0x05
rom.Data[0x005] = cpu.LDM(7)
rom.Data[0x006] = cpu.WRM()           // ram.Data[0][0][5] = 7
rom.Data[0x007] = cpu.LDM(0)
rom.Data[0x008] = cpu.RDM()           // A = 7
```

---

# Esecuzione

```bash
go run ./cmd/go4004
go test ./...
```

---

# Filosofia del progetto

Il progetto NON vuole essere solo un emulatore funzionante.

Vuole essere anche:

* uno strumento didattico
* un percorso di apprendimento
* una ricostruzione storica del funzionamento dei primi microprocessori

Lo sviluppo avviene in piccoli step atomici:
un'istruzione alla volta, con test, demo in main e commit dedicato.

---

# Roadmap

| Step | Stato | Contenuto |
|------|-------|-----------|
| 1  | ✅ | CPU minima (NOP, LDM, ADD, SUB, ...) |
| 2  | ✅ | Test automatici |
| 3  | ✅ | ROM virtuale + fetch-execute |
| 4  | ✅ | PC a 12 bit |
| 5  | ✅ | Stack hardware (JMS/BBL) |
| 6  | ✅ | Istruzioni di salto (JUN, JMS, JCN, ISZ, FIM, SRC, FIN, JIN) |
| 7  | ✅ | RAM virtuale — tutte le 46 istruzioni implementate |
| 8  | 🔲 | Debugger — trace `PC / opcode / A / C` per ogni step |
| 9  | 🔲 | Programmi reali — validare la CPU con programmi non banali |
| 10 | 🔲 | Calcolatrice BCD cifra singola — primo firmware reale |
| 11 | 🔲 | Numeri multi-cifra — addizione con carry tra cifre in RAM |
| 12 | 🔲 | Assembler minimale — testo → ROM binaria |
| 13 | 🔲 | Operazioni complete — sottrazione, moltiplicazione, divisione BCD |
| 14 | 🔲 | I/O virtuale — callback tastiera (RDR) e display (WMP) |
| 15 | 🔲 | Firmware loop completo — calcolatrice interattiva |

---

# Obiettivo finale firmware

La CPU eseguirà un firmware da ROM che implementa una calcolatrice.
La logica della calcolatrice è scritta come programma nella ROM — non in Go.

```text
leggi tastiera  →  interpreta input  →  esegui operazione  →  aggiorna display  →  ripeti
```

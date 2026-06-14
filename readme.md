# retronet-4004 — Emulatore Intel 4004

Un emulatore dell'Intel 4004 scritto in Go, sviluppato passo passo con approccio
didattico. Implementa tutte le 46 istruzioni del processore, una ROM e una RAM
virtuali (chip Intel 4002), il ciclo fetch-execute, lo stack hardware a 3 livelli
e un trace/debugger integrato.

È il primo modulo dell'ecosistema **RetroNet**, una piattaforma open source
educativa che ricostruisce l'evoluzione dell'informatica
(4004 → 8008 → 8080 → CP/M-like → terminali → BBS → HTTP storico → primo Web).

Focus del progetto: architettura CPU, funzionamento dei primi microprocessori,
logica low-level, sistemi a 4 bit, organizzazione hardware/software.

---

# Quick Start

```bash
# esegui una ROM (con trace istruzione per istruzione)
go run ./cmd/retronet-4004 -trace testdata/bcd-add.rom

# lancia tutti i test
go test ./...

# compila il binario della CLI
go build -o retronet-4004 ./cmd/retronet-4004
```

---

# La CLI

La CLI carica una **ROM** da file ed esegue il programma sull'emulatore,
stampando lo stato finale (e, su richiesta, il trace e la RAM).

```
retronet-4004 [flags] <file.rom>
```

> ⚠️ I flag vanno **prima** del nome del file (è la convenzione del package
> `flag` di Go): `retronet-4004 -trace rom` ✅, non `retronet-4004 rom -trace`.

### Flag

| Flag         | Default | Descrizione |
|--------------|---------|-------------|
| `-trace`     | off     | stampa ogni istruzione eseguita (PC, opcode, mnemonico, A, C) |
| `-max N`     | 100000  | limite di step di sicurezza (interrompe i loop infiniti) |
| `-dump-ram`  | off     | a fine esecuzione stampa le celle RAM diverse da zero |

### Cos'è un file `.rom`

Un `.rom` è una sequenza grezza di **byte = opcode**, caricata a partire
dall'indirizzo `0x000`. Lo spazio non usato (fino a 4096 byte) resta `0x00`
(NOP). Esempio — `testdata/bcd-add.rom` calcola `8 + 9` in BCD:

```
20 09   FIM R0, 0x09   → R0=0, R1=9
D8      LDM 8          → A=8
81      ADD R1         → A=8+9=17 → A=1, C=1 (overflow nibble)
FB      DAA            → correzione BCD: A=7, C=1
40 05   JUN 0x005      → salto su se stesso (HALT)
```

Per ora le ROM si producono a mano (byte) o dai programmi Go in `examples/`.
La generazione da sorgente testuale (`LDM 8` ecc.) sarà compito di un modulo
dedicato, `retronet-asm`.

### Convenzione di arresto (HALT)

Il 4004 non ha un'istruzione HALT. Per convenzione un programma termina con un
`JUN` che salta **su se stesso**; la CLI rileva questo idioma e si ferma (senza
eseguire il salto). In assenza di HALT interviene il limite `-max`.

### Esempio di output

```
$ go run ./cmd/retronet-4004 -trace -dump-ram testdata/bcd-add.rom
=== retronet-4004 — ROM: testdata/bcd-add.rom (7 byte) ===

PC=000 OP=20 FIM R0,..  A=0  C=false
PC=002 OP=D8 LDM 8      A=8  C=false
PC=003 OP=81 ADD R1     A=1  C=true
PC=004 OP=FB DAA        A=7  C=true

HALT a PC=0x005 dopo 4 step
A=7  C=true
registri: R0=0 R1=9 R2=0 R3=0 R4=0 R5=0 R6=0 R7=0 R8=0 R9=0 RA=0 RB=0 RC=0 RD=0 RE=0 RF=0
RAM (celle dati non-zero):
  (nessuna)
```

Il risultato BCD si legge come cifra alta = carry (1) e cifra bassa = A (7) → **17**.

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

* CPU Intel 4004 con 46/46 istruzioni
* ROM virtuale — `cpu/rom.go`
* RAM virtuale — `cpu/ram.go` (modello chip Intel 4002)
* Ciclo fetch-execute — `Step(rom, ram)`
* Program Counter a 12 bit (0x000–0xFFF)
* Stack hardware a 3 livelli
* Trace/debugger integrato + `Disassemble(op)`
* CLI che carica ed esegue ROM da file — `cmd/retronet-4004`

---

# Struttura progetto

```text
go-4004/
├── go.mod
├── readme.md
├── cmd/
│   └── retronet-4004/
│       └── main.go         ← CLI: carica ed esegue una ROM
├── testdata/
│   └── bcd-add.rom         ← ROM di esempio (8 + 9 in BCD)
├── examples/               ← programmi dimostrativi (Go che popola la ROM)
│   ├── moltiplicazione/    ← 3×4 con loop ISZ
│   ├── subroutine/         ← 3+5 con JMS/BBL
│   ├── ram/                ← riempire un array in RAM
│   ├── somma-bcd/          ← calcolatrice BCD a cifra singola
│   └── somma-multicifra/   ← 47+58=105 con propagazione del carry
├── docs/
│   ├── bcd.md              ← spiegazione codifica BCD
│   ├── debugger.md         ← uso del trace mode
│   ├── istruzioni-registro.md
│   ├── istruzioni-accumulatore.md
│   ├── istruzioni-salto.md
│   └── istruzioni-io-ram.md
└── cpu/
    ├── cpu.go              ← struct CPU, Step(), stack
    ├── opcodes.go          ← costanti opcode in 4 gruppi
    ├── helpers.go          ← helper functions (mini assembler)
    ├── instructions.go     ← Execute(), executeWithArg(), executeIO()
    ├── disasm.go           ← Disassemble(op) → mnemonico
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
    Data   [4][4][16]uint8 // [banco][registro][carattere]
    Status [4][4][4]uint8  // [banco][registro][status nibble]
    Port   [4]uint8        // porta di output per banco
}
```

Indirizzamento: `banco = CL & 0x3` (da DCL), `registro = (SRCAddr>>4) & 0x3`,
`carattere = SRCAddr & 0x0F` (entrambi da SRC).

---

# Gestione nibble

Il 4004 è un processore a 4 bit. Go usa `uint8` mascherato:

```go
func nibble(v uint8) uint8 {
    return v & 0x0F
}
```

---

# Helper functions (mini assembler)

Ogni istruzione ha un helper che costruisce l'opcode corretto. Tutti gli helper
hanno un commento di documentazione (visibile in hover) con semantica ed effetto
sul carry.

```go
cpu.LDM(7)        // 0xD7
cpu.ADD(cpu.R2)   // 0x82
cpu.JMS(0x3)      // 0x53 (primo byte di JMS 0x3AB)
cpu.WRM()         // 0xE0
```

Esempio — scrittura e lettura di un valore in RAM:

```go
rom.Data[0x000] = cpu.LDM(0)
rom.Data[0x001] = cpu.DCL()           // banco 0
rom.Data[0x002] = cpu.FIM(cpu.R0)
rom.Data[0x003] = 0x05                // R0=0, R1=5
rom.Data[0x004] = cpu.SRC(cpu.R0)     // SRCAddr = 0x05
rom.Data[0x005] = cpu.LDM(7)
rom.Data[0x006] = cpu.WRM()           // ram.Data[0][0][5] = 7
rom.Data[0x007] = cpu.LDM(0)
rom.Data[0x008] = cpu.RDM()           // A = 7
```

---

# Esempi

I programmi in `examples/` (uno per cartella, con README) mostrano la CPU al
lavoro su casi reali:

| Esempio | Dimostra |
|---------|----------|
| `moltiplicazione/` | loop con contatore ISZ (3×4 per addizioni ripetute) |
| `subroutine/`      | chiamata/ritorno con JMS/BBL e convenzione halt |
| `ram/`             | aggiornamento dinamico dell'indirizzo RAM (SRC nel loop) |
| `somma-bcd/`       | prima calcolatrice BCD a cifra singola (ADM + DAA) |
| `somma-multicifra/`| addizione multi-cifra con propagazione del carry (47+58=105) |

```bash
go run ./examples/somma-multicifra
```

---

# Debugger / Trace

Abilita il trace per seguire l'esecuzione istruzione per istruzione:

```go
c := cpu.NewCPU4004()
c.Trace = true   // stampa su os.Stdout per default
```

Output:

```
PC=000 OP=20 FIM R0,..  A=0  C=false
PC=002 OP=D8 LDM 8      A=8  C=false
PC=003 OP=81 ADD R1     A=1  C=true
PC=004 OP=FB DAA        A=7  C=true
```

Per reindirizzare l'output (utile nei test):

```go
var buf strings.Builder
c.TraceWriter = &buf   // qualsiasi io.Writer
```

La funzione `cpu.Disassemble(op byte) string` decodifica un singolo opcode in
mnemonico leggibile. Documentazione completa in `docs/debugger.md`.

Dalla CLI il trace si attiva con il flag `-trace`.

---

# Filosofia del progetto

Il progetto NON vuole essere solo un emulatore funzionante. Vuole essere anche:

* uno strumento didattico
* un percorso di apprendimento
* una ricostruzione storica del funzionamento dei primi microprocessori

Lo sviluppo avviene in piccoli step atomici: un'istruzione (o una feature) alla
volta, con test, demo ed esempi, e commit dedicato. In linea con la regola di
RetroNet: *ogni modulo piccolo, testabile, documentato, eseguibile e integrabile.*

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
| 8  | ✅ | Debugger — trace `PC / opcode / A / C` per ogni step |
| 9  | ✅ | Programmi reali — esempi in `examples/` |
| 10 | ✅ | Calcolatrice BCD cifra singola |
| 11 | ✅ | Numeri multi-cifra — addizione con carry tra cifre in RAM |
| —  | 🔲 | **Verso v0.1.0**: CLI ✅, Dockerfile, README, `testdata/` ✅, tag |

L'**assembler** (testo → ROM) e gli **8008/8080** vivranno in moduli RetroNet
separati (`retronet-asm`, `retronet-8080`), dopo la release `v0.1.0` di questo
modulo. La calcolatrice BCD interattiva resta un demo didattico.

---

# Obiettivo finale firmware (demo didattico)

La CPU eseguirà un firmware da ROM che implementa una calcolatrice. La logica è
scritta come programma nella ROM — non in Go.

```text
leggi tastiera  →  interpreta input  →  esegui operazione  →  aggiorna display  →  ripeti
```

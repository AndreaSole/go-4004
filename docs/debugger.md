# Debugger — Modalità Trace

Il debugger permette di seguire l'esecuzione del firmware istruzione per istruzione,
stampando a ogni step il Program Counter, l'opcode, il mnemonic leggibile e lo
stato dell'accumulatore e del carry.

---

## Attivazione

```go
c := cpu.NewCPU4004()
c.Trace = true   // disabilitato per default
```

Per default l'output va su `os.Stdout`. Per reindirizzarlo, imposta `TraceWriter`:

```go
import "strings"

var buf strings.Builder
c.Trace = true
c.TraceWriter = &buf   // qualsiasi io.Writer
```

---

## Formato output

```
PC=000 OP=20 FIM R0,..    A=0  C=false
PC=002 OP=D8 LDM 8        A=8  C=false
PC=003 OP=81 ADD R1       A=1  C=true
PC=004 OP=FB DAA          A=7  C=true
```

| Campo    | Descrizione |
|----------|-------------|
| `PC=NNN` | Indirizzo hex (3 cifre) dell'istruzione **prima** della fetch |
| `OP=XX`  | Opcode byte in esadecimale (2 cifre) |
| mnemonic | Nome dell'istruzione con registro o argomento (se applicabile) |
| `A=N`    | Valore dell'accumulatore **dopo** l'esecuzione (0–F) |
| `C=bool` | Carry flag **dopo** l'esecuzione |
| `CL=N`   | Banco RAM attivo — mostrato solo per istruzioni del gruppo I/O (0xEX) |
| `SRC=XX` | SRCAddr corrente — mostrato solo per istruzioni del gruppo I/O (0xEX) |

Lo stato è **post-esecuzione**: ogni riga mostra cosa è successo,
non cosa sta per succedere.

---

## Istruzioni a 2 byte

FIM, JUN, JMS, JCN, ISZ occupano 2 byte in ROM ma producono **una sola riga** di
trace. Il PC mostrato è quello del primo byte; l'argomento non è visibile nel trace
ma viene consumato e applicato.

```
PC=000 OP=20 FIM R0,..    A=0  C=false   ← primo byte (0x20), il secondo (0x09) è incluso
PC=002 OP=D8 LDM 8        A=8  C=false   ← PC salta a 002, conferma che FIM ha usato 2 byte
```

---

## Uso nei test

Per catturare e verificare il trace in un test automatico:

```go
func TestMioFirmware(t *testing.T) {
    var buf strings.Builder
    c := cpu.NewCPU4004()
    c.Trace = true
    c.TraceWriter = &buf

    // esegui il firmware...

    t.Log(buf.String()) // trace visibile solo con go test -v, o se il test fallisce
}
```

---

## Disassembly standalone

`cpu.Disassemble(op byte) string` è disponibile indipendentemente dal trace:

```go
for i, op := range rom.Data[:8] {
    fmt.Printf("%03X: %02X  %s\n", i, op, cpu.Disassemble(op))
}
```

**Attenzione**: `Disassemble` non ha contesto della ROM. Se usata linearmente su
una ROM con istruzioni a 2 byte, il secondo byte viene interpretato come opcode
autonomo — il che produce un mnemonic fuorviante. Il trace di `Step()` non ha
questo problema perché conosce la struttura reale delle istruzioni.

---

## Impatto sulle prestazioni

Abilitare il trace rallenta l'emulazione perché ogni step formatta stringhe e
scrive su un `io.Writer`. Per benchmark o esecuzione di firmware lunghi:

```go
c.Trace = false  // disabilita il trace (default)
```

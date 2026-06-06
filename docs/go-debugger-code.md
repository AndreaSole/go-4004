# Il codice del Debugger — Spiegazione Go

Questo documento spiega i concetti Go usati per implementare il debugger
(`cpu/disasm.go` e la modifica di `cpu/cpu.go`).

---

## 1. Named return values

In Go una funzione può assegnare un nome ai suoi valori di ritorno:

```go
// senza named return — stile normale
func divide(a, b int) (int, error) {
    if b == 0 {
        return 0, errors.New("divisione per zero")
    }
    return a / b, nil
}

// con named return — il valore di ritorno ha un nome
func divideNamed(a, b int) (risultato int, err error) {
    if b == 0 {
        err = errors.New("divisione per zero")
        return   // "bare return": restituisce i valori correnti di risultato ed err
    }
    risultato = a / b
    return
}
```

Nel nostro `Step()`:

```go
func (c *CPU4004) Step(rom *ROM, ram *RAM) (err error) {
```

`err` è il nome del valore di ritorno. Questo permette due cose:
- Usare `return` senza argomenti (restituisce il valore corrente di `err`)
- Riferirsi a `err` dall'interno di funzioni anonime annidate (closure)

### Perché lo usiamo qui

Avevamo bisogno che una funzione anonima nel `defer` potesse leggere il
valore finale di `err` **dopo** che la funzione principale ha finito.
I named return lo rendono possibile: `err` è una variabile condivisa
tra `Step()` e tutto ciò che è definito dentro di essa.

---

## 2. defer

`defer` schedula l'esecuzione di una funzione **al termine della funzione
corrente**, qualunque sia il percorso di ritorno (normale o con errore).

```go
func esempio() {
    defer fmt.Println("terzo")   // registrato per dopo
    fmt.Println("primo")
    fmt.Println("secondo")
}
// output:
// primo
// secondo
// terzo
```

`defer` è usato tipicamente per:
- chiudere file o connessioni: `defer f.Close()`
- sbloccare mutex: `defer mu.Unlock()`
- cleanup garantito indipendentemente dagli errori
- **logging post-esecuzione** — il nostro caso

Nel nostro `Step()`:

```go
defer func() {
    if c.Trace && err == nil {
        c.traceLine(pcBefore, op)
    }
}()
```

Tradotto: "quando `Step()` finisce — per qualsiasi ragione, con o senza errore —
esegui questo blocco."

### Perché defer risolve il problema

`Step()` ha molti punti di uscita (`return` sparsi nei vari `case` dello switch).
Senza `defer`, avremmo dovuto chiamare `c.traceLine(...)` prima di ognuno di essi:

```go
// SENZA defer — brutto e pericoloso (facile dimenticare un return)
case OP_JCN, ...:
    // ...
    if c.Trace { c.traceLine(pcBefore, op) }  // ← da ripetere ovunque
    return
case OP_FIM & 0xF0:
    // ...
    if c.Trace { c.traceLine(pcBefore, op) }  // ← di nuovo
    return
// ...
```

Con `defer`, una sola dichiarazione copre tutti i percorsi. Il codice è più
sicuro e più leggibile.

---

## 3. Closure — la funzione anonima nel defer

```go
pcBefore := c.PC   // catturiamo il PC prima che venga modificato
var op byte         // dichiarato qui perché la closure ne ha bisogno

defer func() {                  // ← funzione anonima
    if c.Trace && err == nil {
        c.traceLine(pcBefore, op)
    }
}()                             // ← () finale: "passala a defer (da eseguire dopo)"
```

La funzione anonima usa variabili definite fuori di lei:
- `c` — il ricevitore di `Step()`
- `err` — il named return
- `pcBefore` — locale di `Step()`
- `op` — locale di `Step()`

Questo si chiama **closure**: la funzione "si chiude" attorno alle variabili
del contesto in cui è definita. Non copia i valori al momento della dichiarazione
— **li legge al momento dell'esecuzione** (quando defer scatta).

### La differenza tra catturare un valore e catturare una variabile

```go
// Cattura il VALORE — snapshot al momento del defer
x := 0
defer fmt.Println(x)   // stamperà 0, anche se x viene cambiato dopo
x = 42                 // troppo tardi

// Cattura la VARIABILE — legge il valore finale
x := 0
defer func() { fmt.Println(x) }()   // stamperà 42
x = 42
```

Nel nostro caso, `var op byte` è dichiarato **prima** del defer. La closure
cattura la variabile `op` (non il suo valore iniziale zero). Quando defer
esegue, `op` ha già ricevuto il suo valore da `op, err = readROM(...)`.

```go
var op byte               // op = 0 qui

defer func() {
    // ...
    c.traceLine(pcBefore, op)   // legge op QUI, al momento dell'esecuzione del defer
}()

op, err = readROM(rom, c.PC)   // op riceve il valore reale (es. 0xFB per DAA)
// ...
// defer scatta → op è 0xFB → corretto ✓
```

---

## 4. io.Writer — interfaccia per output generico

Invece di scrivere direttamente su `*os.File` o `*bytes.Buffer`, `TraceWriter`
è dichiarato come `io.Writer`:

```go
TraceWriter io.Writer
```

`io.Writer` è un'**interfaccia** della libreria standard Go:

```go
// definizione (da package io)
type Writer interface {
    Write(p []byte) (n int, err error)
}
```

Un'interfaccia in Go è un contratto: qualsiasi tipo che ha un metodo
`Write([]byte) (int, error)` lo soddisfa automaticamente — senza dichiararlo
esplicitamente ("duck typing" statico).

Questi tipi implementano tutti `io.Writer`:
- `*os.File` (file su disco, incluso `os.Stdout`)
- `*bytes.Buffer` (buffer in memoria)
- `*strings.Builder` (costruisce una stringa incrementalmente)
- `net.Conn` (connessione TCP/UDP)

```go
// Tutti questi funzionano senza cambiare il codice di traceLine:
c.TraceWriter = os.Stdout          // terminale
c.TraceWriter = os.Stderr          // stderr
c.TraceWriter = &myBuf             // *bytes.Buffer
c.TraceWriter = &myBuilder         // *strings.Builder
c.TraceWriter = conn               // net.Conn
```

Nel codice:

```go
func (c *CPU4004) traceLine(pc uint16, op byte) {
    w := c.TraceWriter
    if w == nil {
        w = os.Stdout   // fallback se l'utente non ha impostato TraceWriter
    }
    fmt.Fprintf(w, "PC=%03X ...", pc, ...)  // Fprintf accetta qualsiasi io.Writer
}
```

`fmt.Fprintf` accetta `io.Writer` come primo argomento: funziona con qualsiasi
destinazione senza cambiare una sola riga.

### Perché TraceWriter è nil per default

Quando `NewCPU4004()` crea la struct, Go inizializza ogni campo al suo
**zero value**: `bool → false`, puntatori e interfacce → `nil`.

Quindi di default: `Trace = false`, `TraceWriter = nil`.

Se imposti `c.Trace = true` senza specificare `TraceWriter`, la funzione
usa `os.Stdout` come fallback — il comportamento più comodo per il caso comune.

---

## 5. switch con espressioni booleane (disasm.go)

Go permette un `switch` senza valore a sinistra:

```go
switch {
case op == OP_NOP:
    return "NOP"
case op&0xF0 == OP_ADD:
    return fmt.Sprintf("ADD R%X", op&0x0F)
case op&0xF0 == 0x20 && op&0x01 == 0:
    return fmt.Sprintf("FIM R%X,..", op&0x0E)
}
```

È equivalente a `switch true { case condizione: ... }`. Ogni `case` è
un'espressione booleana; il primo `true` viene eseguito.

Questa forma è utile quando le condizioni non hanno struttura uniforme:
alcune controllano l'intero opcode (`op == X`), altre la famiglia
(`op & 0xF0 == Y`), altre combinano più bit (`op&0xF0 == 0x20 && op&0x01 == 0`).

Con un `switch op { case X: ... }` non sarebbe possibile esprimere tutte
queste condizioni in modo pulito.

### L'ordine dei case conta

Go verifica i `case` in ordine dall'alto in basso e si ferma al primo `true`
(a differenza del C, non c'è fallthrough automatico).

Per FIM e SRC (stesso nibble alto, bit 0 diverso) i case sono mutuamente
esclusivi, quindi l'ordine non cambia il risultato:

```go
case op&0xF0 == 0x20 && op&0x01 == 0:   // FIM: bit 0 = 0
    return fmt.Sprintf("FIM R%X,..", op&0x0E)
case op&0xF0 == 0x20 && op&0x01 == 1:   // SRC: bit 0 = 1
    return fmt.Sprintf("SRC R%X", op&0x0E)
```

In generale però: metti sempre le condizioni più specifiche prima di
quelle più generali per evitare che una condizione "larga" catturi un
opcode destinato a una condizione "stretta" più in basso.

---

## Riepilogo

| Concetto | Dove | Perché |
|----------|------|--------|
| Named return `(err error)` | firma di `Step()` | Condividere `err` con la closure nel `defer` |
| `defer` | all'inizio del corpo di `Step()` | Eseguire il trace a **ogni** uscita dalla funzione, senza ripetizioni |
| Closure | la funzione anonima nel `defer` | Catturare `pcBefore`, `op`, `err` dal contesto di `Step()` |
| `io.Writer` | campo `TraceWriter` nella struct | Scrivere su terminale, file o stringa con lo stesso codice |
| `switch` booleano | `Disassemble()` | Condizioni eterogenee su un singolo valore |

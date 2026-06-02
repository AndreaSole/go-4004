# Gruppo Salti e Indirizzamento — Istruzioni Intel 4004

Queste istruzioni controllano il **flusso del programma** (dove va il PC)
e gestiscono l'**indirizzamento** di ROM e RAM.

Prima di leggere questo documento, è utile capire come è organizzata la memoria.

---

## Organizzazione della ROM — Pagine

La ROM del 4004 è 4096 byte (12 bit di indirizzo: 0x000–0xFFF).
È divisa in **16 pagine** da 256 byte ciascuna:

```
Pagina 0:  0x000 – 0x0FF   (256 byte)
Pagina 1:  0x100 – 0x1FF
Pagina 2:  0x200 – 0x2FF
...
Pagina F:  0xF00 – 0xFFF
```

I 12 bit dell'indirizzo si leggono così:

```
indirizzo:  pppp oooo oooo
            ^^^^             ← bit 11-8: numero pagina (0-15)
                 ^^^^ ^^^^   ← bit 7-0:  offset dentro la pagina (0-255)
```

Alcune istruzioni (JCN, ISZ, JIN) possono saltare **solo dentro la pagina corrente**
perché usano solo 8 bit per l'indirizzo. I 4 bit della pagina vengono dal PC.

---

## Mappa opcode del gruppo

```
0x10–0x1F  → JCN c,a    (2 byte) — jump condizionale
0x20–0x2E  → FIM Rr,d   (2 byte) — fetch immediate (opcode pari)
0x21–0x2F  → SRC Rr     (1 byte) — send register control (opcode dispari)
0x30–0x3E  → FIN Rr     (1 byte) — fetch indirect ROM (opcode pari)
0x31–0x3F  → JIN Rr     (1 byte) — jump indirect (opcode dispari)
0x40–0x4F  → JUN a      (2 byte) — jump unconditional
0x50–0x5F  → JMS a      (2 byte) — jump to subroutine
0x70–0x7F  → ISZ Rr,a   (2 byte) — increment and skip if zero
```

FIM e SRC condividono il range 0x20–0x2F: il bit 0 del nibble basso distingue i due.
FIN e JIN condividono il range 0x30–0x3F: stesso meccanismo.

---

## JUN — Jump Unconditional

**Cosa fa:** salta a qualsiasi indirizzo a 12 bit nella ROM. Nessuna condizione.

**Formato:** 2 byte
```
Byte 1:  0100 nnnn   ← codice JUN + bit 11-8 dell'indirizzo
Byte 2:  oooo oooo   ← bit 7-0 dell'indirizzo
```

L'indirizzo finale è: `n << 8 | byte2`

**Come decodificare:**

Voglio saltare a `0x2C5`:
```
0x2C5 in binario a 12 bit:
  0010 1100 0101
  ^^^^             ← pagina = 0010 = 2   → n = 2
       ^^^^ ^^^^   ← offset = 1100 0101  = 0xC5

Byte 1: 0100 0010  = 0x42   (JUN + n=2)
Byte 2: 1100 0101  = 0xC5
```

**Esempi:**

| Destinazione | Byte 1 | Byte 2 |
|--------------|--------|--------|
| 0x000        | 0x40   | 0x00   |
| 0x100        | 0x41   | 0x00   |
| 0x3AB        | 0x43   | 0xAB   |
| 0xFFF        | 0x4F   | 0xFF   |

**Effetti:**

| Cosa | Prima | Dopo |
|------|-------|------|
| PC   | X     | n<<8 \| byte2 |
| A    | X     | invariato |
| C    | X     | invariato |

---

## JMS — Jump to Subroutine

**Cosa fa:** chiama una subroutine. Salva il PC corrente nello **stack hardware**,
poi salta all'indirizzo specificato. BBL farà il contrario al ritorno.

**Formato:** 2 byte (identico a JUN ma opcode `0x5`)
```
Byte 1:  0101 nnnn   ← codice JMS + bit 11-8 dell'indirizzo
Byte 2:  oooo oooo   ← bit 7-0 dell'indirizzo
```

**Come funziona passo per passo:**

```
Stato prima di JMS:
  PC = 0x010   (qui c'è JMS)
  SP = 0       (stack vuoto)
  Stack = [?, ?, ?]

Istruzione JMS 0x050:
  Byte 1 a 0x010: 0x50
  Byte 2 a 0x011: 0x50

Esecuzione:
  1. fetch byte 1 → PC diventa 0x011
  2. fetch byte 2 → PC diventa 0x012
  3. push(0x012) sullo stack → Stack[0] = 0x012, SP = 1
  4. PC = 0x050

Stato dopo:
  PC = 0x050   (stiamo eseguendo la subroutine)
  SP = 1
  Stack = [0x012, ?, ?]
```

L'indirizzo salvato è `0x012` (quello DOPO i 2 byte di JMS), cioè la prima istruzione
che verrà eseguita al ritorno.

**Lo stack è a 3 livelli.** Puoi annidare al massimo 3 subroutine. Il 4° JMS
sovrascrive il più vecchio senza errori (comportamento hardware reale).

**Esempio con BBL:**
```
ROM:
  0x000: JMS 0x020   → byte 1: 0x50, byte 2: 0x20
  0x001:              → (secondo byte di JMS)
  0x002: NOP          ← qui torniamo dopo BBL

  0x020: LDM 7       ← corpo subroutine
  0x021: BBL 3       ← ritorna, A = 3

Esecuzione:
  PC=0x000 → JMS: push(0x002), PC=0x020
  PC=0x020 → LDM 7: A=7
  PC=0x021 → BBL 3: A=3, PC=pop()=0x002
  PC=0x002 → NOP
```

**Effetti:**

| Cosa  | Prima | Dopo |
|-------|-------|------|
| PC    | X     | indirizzo subroutine |
| SP    | N     | N + 1 |
| Stack | ...   | ...+ PC di ritorno |
| A, C  | X     | invariati |

---

## JCN — Jump Conditional

**Cosa fa:** salta a un indirizzo nella stessa pagina se una condizione è vera.

**Formato:** 2 byte
```
Byte 1:  0001 cccc   ← codice JCN + nibble condizione
Byte 2:  oooo oooo   ← offset di destinazione (8 bit, stessa pagina)
```

Il nibble condizione `c` ha 4 bit, uno per ogni condizione:

```
c = C4 C3 C2 C1
     │   │   │  └── bit 0: C1 — salta se TEST pin = 0  (non emulato)
     │   │   └───── bit 1: C2 — salta se carry = 1
     │   └───────── bit 2: C3 — salta se A = 0
     └───────────── bit 3: C4 — inverte il risultato finale
```

**Logica:** C1, C2, C3 si combinano con OR (basta che una sia vera).
C4 inverte tutto il risultato.

**Indirizzo di salto:**
```
PC = (PC corrente & 0x0F00) | byte2

Esempio: PC = 0x150, byte2 = 0x80
  0x150 & 0x0F00 = 0x100   (pagina 1)
  0x100 | 0x80  = 0x180    ← destinazione
```

**Tabella delle condizioni più usate:**

| c (hex) | c (bin) | Condizione | Salta se... |
|---------|---------|-----------|-------------|
| 0x2     | 0010    | C2        | carry = 1 |
| 0x4     | 0100    | C3        | A = 0 |
| 0x6     | 0110    | C2 OR C3  | carry=1 oppure A=0 |
| 0xA     | 1010    | NOT C2    | carry = 0 |
| 0xC     | 1100    | NOT C3    | A ≠ 0 |
| 0x8     | 1000    | NOT (niente) = sempre | salta sempre (stessa pagina) |

**Esempio — salta se carry è 1:**
```
ROM[0x000] = 0x12   ← JCN, c=2 (C2: carry=1)
ROM[0x001] = 0x50   ← destinazione: offset 0x50

A = 5, C = true

Esecuzione:
  C2 attivo, carry = true → condizione vera → salta
  PC = (0x002 & 0x0F00) | 0x50 = 0x050
```

**Esempio — salta se A è zero:**
```
ROM[0x100] = 0x14   ← JCN, c=4 (C3: A=0), siamo a pagina 1
ROM[0x101] = 0x30   ← destinazione: offset 0x30

A = 0

Esecuzione:
  C3 attivo, A = 0 → condizione vera → salta
  PC = (0x102 & 0x0F00) | 0x30 = 0x130
```

**Esempio — condizione falsa, nessun salto:**
```
ROM[0x000] = 0x12   ← JCN, c=2 (carry=1)
ROM[0x001] = 0x50

C = false

Esecuzione:
  carry = false → condizione falsa → NON salta
  PC = 0x002   (prosegue normalmente)
```

**Effetti:**

| Cosa | Se salta | Se non salta |
|------|----------|--------------|
| PC   | stessa pagina + byte2 | PC + 2 |
| A, C | invariati | invariati |

---

## ISZ — Increment and Skip if Zero

**Cosa fa:** incrementa un registro di 1. Se il risultato è **diverso da zero**, salta.
Se il risultato è **zero**, non salta (continua alla prossima istruzione).

**Formato:** 2 byte
```
Byte 1:  0111 rrrr   ← codice ISZ + numero registro
Byte 2:  oooo oooo   ← offset di destinazione (stessa pagina)
```

**Logica apparentemente strana (ma ha senso):**
- salta → continua il loop
- non salta → esce dal loop

ISZ è pensato per **contatori di loop**. Il registro parte da un valore negativo
(lontano da 0) e viene incrementato ogni volta. Quando raggiunge 0, il loop finisce.

**Formula per N iterazioni:** inizializza il registro a `16 - N`

| N iterazioni | Valore iniziale | Sequenza |
|-------------|-----------------|----------|
| 1  | 15 (0xF) | 0xF → 0 (stop) |
| 2  | 14 (0xE) | 0xE → 0xF → 0 (stop) |
| 3  | 13 (0xD) | 0xD → 0xE → 0xF → 0 (stop) |
| 8  | 8  (0x8) | 0x8 → ... → 0xF → 0 (stop) |
| 16 | 0  (0x0) | 0x0 → 0x1 → ... → 0xF → 0 (stop) |

**Esempio — loop 3 iterazioni:**
```
ROM[0x000] = FIM R2, 0x0D  → R3 = 0xD (13)
ROM[0x002] = ... corpo loop ...
ROM[0x005] = ISZ R3         ← byte 1: 0x73
ROM[0x006] = 0x02           ← byte 2: torna a 0x002

Esecuzione:
  Iter 1: R3 = 0xD → ISZ → R3 = 0xE ≠ 0 → salta a 0x002
  Iter 2: R3 = 0xE → ISZ → R3 = 0xF ≠ 0 → salta a 0x002
  Iter 3: R3 = 0xF → ISZ → R3 = 0x0 = 0 → NON salta → esce dal loop
```

**In binario (il momento critico — R3 = 0xF):**
```
R3 = 1111  (15)
+    0001  (1)
─────────
1 0000  →  teniamo 4 bit → R3 = 0000 = 0  →  NON salta
```

**Effetti:**

| Cosa | Valore |
|------|--------|
| Rr   | Rr + 1 (mod 16) |
| PC   | stessa pagina + byte2  (se Rr ≠ 0) oppure  PC + 2 (se Rr = 0) |
| A, C | invariati |

---

## FIM — Fetch Immediate

**Cosa fa:** carica un byte immediato (nel codice) in una **coppia di registri**.
Il nibble alto del byte va nel registro pari, il nibble basso in quello dispari successivo.

**Formato:** 2 byte
```
Byte 1:  0010 rr00   ← codice FIM + numero coppia (r = 0,2,4,6,8,A,C,E)
Byte 2:  HHHHLLLL   ← H = nibble alto → Rr,  L = nibble basso → Rr+1
```

Il bit 0 del nibble basso di byte 1 è sempre 0 (distingue FIM da SRC).

**Coppie di registri:**

| Coppia | Rr | Rr+1 | Byte 1 (FIM) |
|--------|----|------|--------------|
| 0      | R0 | R1   | 0x20 |
| 1      | R2 | R3   | 0x22 |
| 2      | R4 | R5   | 0x24 |
| 3      | R6 | R7   | 0x26 |
| 4      | R8 | R9   | 0x28 |
| 5      | RA | RB   | 0x2A |
| 6      | RC | RD   | 0x2C |
| 7      | RE | RF   | 0x2E |

**Esempio — caricare 0xAB in R0/R1:**
```
ROM[0x000] = 0x20   ← FIM coppia 0 (R0/R1)
ROM[0x001] = 0xAB   ← dato

In binario:
  Byte 2: 1010 1011
          ^^^^       ← nibble alto = 0xA = 10  → R0 = 10
               ^^^^  ← nibble basso = 0xB = 11 → R1 = 11

Risultato: R0 = 0xA, R1 = 0xB
```

**Esempio — caricare l'indirizzo 0x37 in R4/R5:**
```
ROM[0x000] = 0x24   ← FIM coppia 2 (R4/R5)
ROM[0x001] = 0x37

In binario:
  0x37 = 0011 0111
         ^^^^       → R4 = 3
              ^^^^  → R5 = 7

Risultato: R4 = 3, R5 = 7
```

**Quando si usa:** caricare indirizzi RAM (per SRC), valori di lookup, inizializzare contatori a 8 bit.

**Effetti:**

| Cosa  | Valore |
|-------|--------|
| Rr    | nibble alto del byte2 |
| Rr+1  | nibble basso del byte2 |
| A, C  | invariati |
| altri registri | invariati |

---

## SRC — Send Register Control

**Cosa fa:** imposta l'**indirizzo del registro RAM** che le successive istruzioni I/O useranno.
Non scrive né legge dati — prepara solo il "puntatore" per WRM, RDM, ecc.

**Formato:** 1 byte
```
Byte 1:  0010 rr01   ← codice SRC + numero coppia (bit 0 = 1 distingue da FIM)
```

Il byte SRC viene costruito così: `SRCAddr = (Rr << 4) | Rr+1`

```
SRCAddr: CCCC RRRR
         ^^^^       ← nibble alto = Rr  = numero chip/banco RAM
              ^^^^  ← nibble basso = Rr+1 = numero registro nel chip
```

**Esempio:**
```
FIM R0, 0x23   → R0 = 2, R1 = 3
SRC R0         → SRCAddr = (2 << 4) | 3 = 0x23
                 → chip RAM 2, registro 3

Poi:
WRM            → scrive A nel registro 3 del chip RAM 2
RDM            → legge il nibble del registro 3 del chip RAM 2 in A
```

**In binario:**
```
R0 = 0010  (2)
R1 = 0011  (3)

SRC R0:
  SRCAddr = [0010][0011] = 0010 0011 = 0x23
```

**Nota:** SRC e DCL lavorano insieme per selezionare la RAM:
- `DCL` seleziona il gruppo di chip (banco 0-7)
- `SRC` seleziona il registro specifico dentro il chip

**Effetti:**

| Cosa    | Valore |
|---------|--------|
| SRCAddr | (Rr << 4) \| Rr+1 |
| A, C    | invariati |
| R0–RF   | invariati |

---

## FIN — Fetch Indirect from ROM

**Cosa fa:** usa R0 e R1 come indirizzo per leggere un byte dalla ROM,
poi carica quel byte nella coppia di registri specificata.

È il meccanismo delle **lookup table** — tabelle di valori costanti nella ROM
che il programma può consultare a runtime.

**Formato:** 1 byte
```
Byte 1:  0011 rr00   ← codice FIN + numero coppia destinazione (bit 0 = 0)
```

**Come viene calcolato l'indirizzo da leggere:**
```
addr = (pagina corrente) | (R0 << 4) | R1
     = (PC & 0x0F00) | (R0 << 4) | R1
```

- i 4 bit alti vengono dalla pagina corrente del PC
- R0 fornisce i 4 bit centrali
- R1 fornisce i 4 bit bassi

L'indirizzo è quindi sempre nella **stessa pagina** in cui si trova FIN.

**Esempio — lookup table dei quadrati:**
```
Supponiamo di voler calcolare il quadrato di un numero 0-4.
Mettiamo la tabella a partire da ROM[0x010]:

ROM[0x010] = 0x00   ← 0² = 0
ROM[0x011] = 0x01   ← 1² = 1
ROM[0x012] = 0x04   ← 2² = 4
ROM[0x013] = 0x09   ← 3² = 9
ROM[0x014] = 0x10   ← non è 16, è 0x10: nibble alto=1, nibble basso=0

Per leggere 3²:
  R0 = 0x1   (nibble alto dell'indirizzo 0x13)
  R1 = 0x3   (nibble basso dell'indirizzo 0x13)

ROM[0x000] = FIN R2   ← 0x32

Esecuzione (PC = 0x000, pagina = 0):
  addr = (0x000 & 0x0F00) | (0x1 << 4) | 0x3
       = 0x000 | 0x10 | 0x03
       = 0x013

  legge ROM[0x013] = 0x09 = 0000 1001
  nibble alto = 0 → R2 = 0
  nibble basso = 9 → R3 = 9

Risultato: R2 = 0, R3 = 9   (cioè 3² = 9, un nibble ciascuno)
```

**Importante:** FIN **non modifica R0 e R1** — sono solo l'indirizzo sorgente.

**Effetti:**

| Cosa  | Valore |
|-------|--------|
| Rr    | nibble alto del byte letto da ROM |
| Rr+1  | nibble basso del byte letto da ROM |
| R0, R1 | invariati (usati come indirizzo, non modificati) |
| A, C  | invariati |
| PC    | PC + 1 (FIN è 1 byte) |

---

## JIN — Jump Indirect

**Cosa fa:** salta all'indirizzo contenuto in una coppia di registri, nella stessa pagina.
È come JUN ma l'indirizzo viene letto dai registri invece di essere hardcoded.

**Formato:** 1 byte
```
Byte 1:  0011 rr01   ← codice JIN + numero coppia sorgente (bit 0 = 1)
```

**Calcolo dell'indirizzo di salto:**
```
PC = (PC corrente & 0x0F00) | (Rr << 4) | Rr+1
```

La pagina viene preservata (stesso meccanismo di JCN e ISZ).

**Esempio:**
```
R0 = 0x7
R1 = 0x3
PC = 0x200   (siamo in pagina 2)

JIN R0  (opcode 0x31)

addr = (0x200 & 0x0F00) | (0x7 << 4) | 0x3
     = 0x200 | 0x70 | 0x03
     = 0x273

PC = 0x273
```

**In binario:**
```
R0 = 0111   (7)
R1 = 0011   (3)

(R0 << 4) | R1 = 0111 0011 = 0x73

PC alta parte: pagina 2 = 0x200 = 0010 0000 0000

Risultato: 0010 0111 0011 = 0x273
```

**Quando si usa — jump table (tabella di salti):**
Immagina un programma che deve fare cose diverse in base a un valore 0–3
(come uno switch/case). Si costruisce una tabella di indirizzi nella ROM,
si carica l'indirizzo giusto con FIM, e si salta con JIN:

```
FIM R0, 0x50   ← se caso 0, vai a 0x050
JIN R0         ← salta a 0x050

FIM R0, 0x60   ← se caso 1, vai a 0x060
JIN R0         ← salta a 0x060
... ecc.
```

**Differenza con JUN:**
- `JUN` ha l'indirizzo fisso nel codice — deciso quando scrivi il programma
- `JIN` ha l'indirizzo nei registri — può cambiare a runtime

**Effetti:**

| Cosa   | Valore |
|--------|--------|
| PC     | (pagina corrente) \| (Rr << 4) \| Rr+1 |
| Rr, Rr+1 | invariati |
| A, C   | invariati |

---

## Riepilogo del gruppo

| Istruzione | Byte | Opcode  | Indirizzo | Condizione | Uso principale |
|------------|------|---------|-----------|------------|----------------|
| JUN a      | 2    | `0x4n`  | 12 bit (ovunque) | nessuna | goto |
| JMS a      | 2    | `0x5n`  | 12 bit (ovunque) | nessuna | chiama funzione |
| JCN c,a    | 2    | `0x1c`  | 8 bit (stessa pagina) | carry, A=0, NOT | if/while |
| ISZ Rr,a   | 2    | `0x7r`  | 8 bit (stessa pagina) | Rr ≠ 0 dopo++ | loop con contatore |
| FIM Rr,d   | 2    | `0x2r`  | — (dato immediato) | nessuna | carica 8 bit in registro pair |
| SRC Rr     | 1    | `0x2r+1`| — (imposta SRCAddr) | nessuna | seleziona registro RAM |
| FIN Rr     | 1    | `0x3r`  | 8 bit (stessa pagina, da R0:R1) | nessuna | lookup table ROM |
| JIN Rr     | 1    | `0x3r+1`| 8 bit (stessa pagina, da Rr:Rr+1) | nessuna | jump table |

### Istruzioni a 2 byte — come viene letto il secondo byte

Sul 4004 reale, `Step()` legge l'opcode, poi legge il secondo byte prima di eseguire.
Nel nostro emulatore è implementato esattamente così in `cpu.go`:

```go
switch op & 0xF0 {
case OP_JCN, OP_JUN, OP_JMS, OP_ISZ:
    arg := rom.Data[c.PC]       // leggi secondo byte
    c.PC = (c.PC + 1) & 0x0FFF // avanza PC
    return c.executeWithArg(op, arg)
...
}
```

### Indirizzi "stessa pagina" — schema

```
Indirizzo a 12 bit:  PPPP OOOO OOOO
                     ^^^^             ← pagina (4 bit, dal PC corrente)
                          ^^^^ ^^^^   ← offset (8 bit, dal secondo byte)

JCN/ISZ/JIN usano solo i bit offset (8 bit).
I bit pagina vengono copiati dal PC → salto sempre nella stessa pagina.

JUN/JMS usano tutti e 12 i bit → possono saltare ovunque nella ROM.
```

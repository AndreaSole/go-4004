# Esempio: Moltiplicazione 3 × 4 = 12

Dimostra il pattern base del loop sul 4004: usare ISZ come contatore di iterazioni,
accumulare un valore in A, e salvare il risultato in RAM.

---

## Algoritmo

La moltiplicazione `3 × 4` viene calcolata come 4 addizioni successive di 3:

```
0 + 3 = 3
3 + 3 = 6
6 + 3 = 9
9 + 3 = 12
```

Sul 4004 questo si traduce in:

```
R1 = 3    ← addendo (valore da sommare ad ogni iterazione)
R4 = 12   ← contatore loop (inizializzato a 16-4, vedi sotto)
A  = 0    ← accumulatore

LOOP:
  A = A + R1   → somma l'addendo
  R4++         → incrementa il contatore (ISZ)
  se R4 != 0   → torna al LOOP

WRM  → salva A in RAM
```

---

## Layout ROM

```
0x000  LDM 0          A = 0 (serve per DCL)
0x001  DCL            CL = 0 (seleziona banco RAM 0)
0x002  FIM R0, 0x03   R0=0, R1=3  (addendo nei registri)
0x003   └── 0x03
0x004  FIM R2, 0x00   R2=0, R3=0  (indirizzo RAM per SRC)
0x005   └── 0x00
0x006  SRC R2         SRCAddr = 0x00 → scriveremo in RAM[0][0][0]
0x007  LDM 12         A = 12 (contatore loop iniziale)
0x008  XCH R4         R4=12, A=0
       ── LOOP ──
0x009  ADD R1         A = A + R1
0x00A  ISZ R4, ...    R4++; se R4 != 0 → salta a 0x009
0x00B   └── 0x09
       ── FINE LOOP ──
0x00C  WRM            ram.Data[0][0][0] = A
```

---

## Il trucco del contatore ISZ: inizializzare a 16 − N

ISZ **incrementa** il registro e salta se il risultato è diverso da zero.
Per fare esattamente N iterazioni, il registro si inizializza a `16 − N`,
in modo che dopo N incrementi raggiunga 0 (overflow nibble) e il loop termini.

Per 4 iterazioni: `16 − 4 = 12`

```
R4 = 12 → ISZ → R4=13  ≠ 0 → continua (iter 1)
R4 = 13 → ISZ → R4=14  ≠ 0 → continua (iter 2)
R4 = 14 → ISZ → R4=15  ≠ 0 → continua (iter 3)
R4 = 15 → ISZ → R4= 0  = 0 → esce     (iter 4)
```

---

## Due approcci per caricare un valore in un registro

### Approccio 1 — FIM (usato in questo programma)

```
FIM R0, 0x03   → R0=0, R1=3
```

FIM carica **due registri contemporaneamente** in un solo step CPU (2 byte, 1 istruzione).
È l'istruzione progettata per caricare dati immediati nelle coppie di registri.

### Approccio 2 — LDM + XCH

```
LDM 3    → A = 3
XCH R1   → R1 = 3, A = 0
```

Carica prima A con LDM, poi scambia A con il registro target via XCH.
Richiede 2 step CPU invece di 1, ma è più esplicito e carica un singolo registro.

| | Step CPU | Byte ROM | Registri toccati |
|---|---|---|---|
| FIM R0, 0x03 | 1 | 2 | R0 e R1 |
| LDM 3 + XCH R1 | 2 | 2 | solo R1 (A viene azzerato) |

**Quando usare FIM:** quando vuoi caricare una coppia di registri correlati
(per esempio un indirizzo RAM a 8 bit, o due nibble di dati correlati).

**Quando usare LDM + XCH:** quando vuoi caricare un singolo registro e non hai
bisogno di toccare il suo compagno di coppia. Lo stesso programma usa questo
pattern per il contatore: `LDM 12` → `XCH R4`.

---

## Risultato atteso

```
A            = 12
RAM[0][0][0] = 12
```

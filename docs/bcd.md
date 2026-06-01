# BCD — Binary Coded Decimal

## Cos'è BCD

BCD (Binary Coded Decimal) è un sistema di codifica in cui ogni **cifra decimale** (0–9) viene rappresentata separatamente in binario, usando 4 bit (un nibble).

| Cifra decimale | Codifica BCD (4 bit) |
|:--------------:|:--------------------:|
| 0              | 0000                 |
| 1              | 0001                 |
| 2              | 0010                 |
| 3              | 0011                 |
| 4              | 0100                 |
| 5              | 0101                 |
| 6              | 0110                 |
| 7              | 0111                 |
| 8              | 1000                 |
| 9              | 1001                 |

I valori da 10 a 15 (1010–1111) **non esistono in BCD** — sono "buchi" nella codifica.

---

## Perché BCD e non binario puro?

Con 4 bit in binario puro potresti rappresentare valori da 0 a 15.

Il problema è che **visualizzare un numero binario puro su un display decimale richiede una conversione** — e quella conversione è costosa su hardware semplice come il 4004.

Con BCD invece ogni cifra è già "pronta" per essere mostrata su un display a 7 segmenti: la cifra `7` si codifica come `0111` e il display la legge direttamente.

Il 4004 fu progettato specificamente per le calcolatrici da tavolo — BCD non è una scelta di convenienza, è un requisito architetturale.

---

## Come si rappresenta un numero multi-cifra

Ogni cifra decimale occupa un registro separato (o una cella RAM).

Il numero **47** viene memorizzato così:

```
R0 = 4   (0100)  ← cifra delle decine
R1 = 7   (0111)  ← cifra delle unità
```

Il numero **382** richiederebbe tre nibble:

```
R0 = 3   ← centinaia
R1 = 8   ← decine
R2 = 2   ← unità
```

---

## Il problema dell'addizione BCD

Quando sommi due cifre BCD con il 4004, la CPU esegue una **somma binaria** su 4 bit. Il risultato può cadere fuori dal range BCD valido (0–9).

### Caso 1 — Risultato valido (0–9)

```
  3 + 4 = 7
  0011 + 0100 = 0111  →  A=7, C=false
```

Il risultato `7` è una cifra BCD valida. Nessuna correzione necessaria.

### Caso 2 — Risultato invalido (10–15)

```
  8 + 5 = 13
  1000 + 0101 = 1101  →  A=13 (0xD), C=false
```

Il risultato `13` (0xD) **non è una cifra BCD valida**. In decimale dovrebbe essere "cifra 3, porta 1 alle decine".

### Caso 3 — Overflow binario (carry)

```
  9 + 8 = 17
  1001 + 1000 = 10001  →  A=1, C=true
```

Il carry indica che la somma ha superato 15. Il risultato binario `1` con carry vale effettivamente `17` in decimale — ma il nibble da solo non lo dice.

---

## La correzione: aggiungere 6

La correzione BCD consiste nell'aggiungere **6** al risultato quando:
- Il risultato è > 9 (caso 2), oppure
- C'è stato carry (caso 3)

Perché 6? Perché da 0 a 15 ci sono 16 valori, ma BCD usa solo 10 (0–9). I 6 valori "saltati" (10, 11, 12, 13, 14, 15) sono esattamente il gap da colmare.

```
Caso 2:  A=13, C=false
         13 + 6 = 19  →  nibble=3, carry=true
         Risultato BCD: cifra 3, porta 1  →  "13" ✓

Caso 3:  A=1, C=true
         1 + 6 = 7, carry=true (ereditato)
         Risultato BCD: cifra 7, porta 1  →  "17" ✓
```

Sul 4004 questa correzione si esegue con l'istruzione **DAA** (Decimal Adjust Accumulator), che la applica automaticamente leggendo lo stato di A e C.

---

## Addizione multi-cifra: 38 + 47 = 85

L'addizione si esegue cifra per cifra, dalle unità verso le decine, propagando il carry.

```
Cifre delle unità:
  8 + 7 = 15  →  A=15 (0xF), C=false   ← invalido BCD
  DAA          →  A=5, C=true            ← corretto: cifra 5, porta 1

Cifre delle decine (con carry=1):
  3 + 4 + 1 = 8  →  A=8, C=false
  DAA              →  A=8, C=false        ← già valido BCD

Risultato: decine=8, unità=5  →  85 ✓
```

In codice con questo emulatore:

```go
// 38 + 47 = 85

// Carica i valori
// 38: R0=3 (dec), R1=8 (uni)
// 47: R2=4 (dec), R3=7 (uni)

// Step 1: somma unità
cpu.LD(cpu.R1),   // A = 8
cpu.ADD(cpu.R3),  // A = 8+7 = 15, C = false
cpu.DAA(),        // A = 5, C = true
cpu.XCH(cpu.R5),  // salva cifra unità in R5

// Step 2: somma decine (carry già impostato da DAA)
cpu.LD(cpu.R0),   // A = 3
cpu.ADD(cpu.R2),  // A = 3+4+1(carry) = 8, C = false
cpu.DAA(),        // A = 8, nessuna correzione
cpu.XCH(cpu.R4),  // salva cifra decine in R4

// Risultato: R4=8, R5=5  →  85
```

---

## La sottrazione BCD: TCS

Per la sottrazione BCD esiste un meccanismo analogo che usa l'istruzione **TCS** (Transfer Carry Subtract).

Dopo una SUB, TCS carica in A il valore di correzione corretto:
- `A = 10` se c'è stato borrow (C=true) — la cifra ha "preso in prestito" 10 dalla cifra successiva
- `A = 9`  se non c'è stato borrow (C=false)

Questo valore viene poi usato nel passo di correzione della cifra successiva.

---

## Riepilogo istruzioni BCD del 4004

| Istruzione | Uso BCD |
|------------|---------|
| ADD        | Somma binaria di due cifre |
| DAA        | Corregge A dopo ADD per BCD |
| SUB        | Sottrae binario con borrow |
| TCS        | Carica correzione (9 o 10) dopo SUB |
| TCC        | Trasferisce carry in A (utile per propagare il riporto) |

---

## Perché il 4004 usa BCD invece di binario puro

Il 4004 fu progettato nel 1971 da Federico Faggin per la calcolatrice Busicom 141-PF.

Le calcolatrici dell'epoca dovevano:
1. leggere tasti numerici (già in formato decimale)
2. calcolare in decimale
3. mostrare risultati su display decimali

Con BCD il firmware non deve mai convertire tra decimale e binario. Ogni cifra che entra dalla tastiera è già BCD. Ogni cifra che va al display è già BCD. La CPU non fa altro che sommare/sottrarre nibble con correzione DAA/TCS.

Su hardware con poche centinaia di transistor e 4 bit di larghezza del bus, questa era la scelta giusta.

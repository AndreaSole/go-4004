# Gruppo Registro — Istruzioni Intel 4004

Queste istruzioni lavorano sui **16 registri R0–RF** del 4004.
Ogni registro contiene 4 bit (un nibble), quindi può valere da 0 a 15.

Il nibble basso dell'opcode indica **quale registro** usare.
Ad esempio `ADD R3` ha opcode `0x83` — gli ultimi 4 bit (`0011` = 3) sono il numero del registro.

---

## Mappa opcode del gruppo

```
0x00        → NOP
0x60–0x6F   → INC Rr
0x80–0x8F   → ADD Rr
0x90–0x9F   → SUB Rr
0xA0–0xAF   → LD  Rr
0xB0–0xBF   → XCH Rr
0xC0–0xCF   → BBL n
0xD0–0xDF   → LDM n
```

---

## NOP — No Operation

**Cosa fa:** non fa niente. Occupa 1 ciclo di clock senza modificare nessuno stato.

**Opcode:** `0x00` = `0000 0000`

**Quando si usa:**
- riempire spazi nella ROM (padding)
- introdurre ritardi temporali (delay loop)
- segnaposto durante lo sviluppo

**Esempio:**
```
ROM[0x000] = 0x00   ← NOP
ROM[0x001] = 0xD5   ← LDM 5
```
Esecuzione: non succede nulla, poi A = 5.

**Effetti:**

| Registro | Prima | Dopo |
|----------|-------|------|
| A        | qualsiasi | invariato |
| C        | qualsiasi | invariato |
| R0–RF    | qualsiasi | invariati |
| PC       | X    | X + 1 |

---

## LDM — Load Immediate to Accumulator

**Cosa fa:** carica un valore fisso (0–15) direttamente nell'accumulatore A.
Il valore è **hardcoded nell'opcode** — non viene letto dai registri.

**Opcode:** `0xDn` dove `n` è il valore da caricare (0–F)

**Formato in binario:**
```
1101 nnnn
^^^^ ^^^^
│    └── valore immediato (0-15)
└─────── codice LDM
```

**Esempi:**

```
LDM 0  →  0xD0  =  1101 0000   →  A = 0
LDM 5  →  0xD5  =  1101 0101   →  A = 5
LDM 9  →  0xD9  =  1101 1001   →  A = 9
LDM 15 →  0xDF  =  1101 1111   →  A = 15
```

**Esempio pratico:**
```
A = 7, C = true

Eseguo: LDM 3  (opcode 0xD3)

Dopo:   A = 3, C = true (invariato)
```

**Effetti:**

| Registro | Cambia? | Valore |
|----------|---------|--------|
| A        | ✅ sì   | valore n |
| C        | ❌ no   | invariato |
| R0–RF    | ❌ no   | invariati |

---

## LD — Load Register to Accumulator

**Cosa fa:** copia il valore di un registro nell'accumulatore A.
Il registro sorgente NON viene modificato.

**Opcode:** `0xAr` dove `r` è il numero del registro (0–F)

**Formato in binario:**
```
1010 rrrr
^^^^ ^^^^
│    └── numero registro (0-15)
└─────── codice LD
```

**Esempi:**

```
LD R0  →  0xA0  =  1010 0000   →  A = R0
LD R5  →  0xA5  =  1010 0101   →  A = R5
LD RF  →  0xAF  =  1010 1111   →  A = RF
```

**Esempio pratico:**
```
A = 0, R3 = 9

Eseguo: LD R3  (opcode 0xA3)

Dopo:   A = 9, R3 = 9 (invariato)
```

**Differenza con LDM:**
- `LDM 5` carica il numero 5 — valore fisso nel codice
- `LD R5` carica il contenuto di R5 — valore variabile a runtime

**Effetti:**

| Registro | Cambia? | Valore |
|----------|---------|--------|
| A        | ✅ sì   | valore di Rr |
| C        | ❌ no   | invariato |
| Rr       | ❌ no   | invariato (sorgente, non modificato) |

---

## XCH — Exchange Accumulator and Register

**Cosa fa:** scambia il valore dell'accumulatore A con quello del registro specificato.
Nessun valore viene perso — è uno scambio completo.

**Opcode:** `0xBr` dove `r` è il numero del registro (0–F)

**Formato in binario:**
```
1011 rrrr
^^^^ ^^^^
│    └── numero registro
└─────── codice XCH
```

**Esempio pratico:**
```
A = 5, R2 = 9

Eseguo: XCH R2  (opcode 0xB2)

Dopo:   A = 9, R2 = 5
```

**In binario:**
```
Prima:
  A  = 0101  (5)
  R2 = 1001  (9)

Dopo XCH R2:
  A  = 1001  (9)
  R2 = 0101  (5)
```

**Quando si usa:** XCH è utile per salvare temporaneamente A senza perdere dati.
Ad esempio, se vuoi fare un calcolo e poi confrontare il risultato col valore precedente:
```
XCH R0   ← salva A in R0, metti il vecchio R0 in A
... calcoli su A ...
XCH R0   ← rimetti A in R0, recupera il vecchio valore
```

**Effetti:**

| Registro | Cambia? | Valore |
|----------|---------|--------|
| A        | ✅ sì   | vecchio valore di Rr |
| Rr       | ✅ sì   | vecchio valore di A |
| C        | ❌ no   | invariato |

---

## INC — Increment Register

**Cosa fa:** aggiunge 1 al valore del registro specificato.
Il registro è a 4 bit: se vale 15 (`1111`) e lo incrementi, torna a 0 (`0000`). Non c'è carry.

**Opcode:** `0x6r` dove `r` è il numero del registro (0–F)

**Formato in binario:**
```
0110 rrrr
^^^^ ^^^^
│    └── numero registro
└─────── codice INC
```

**Esempi:**
```
R1 = 3   →  INC R1  →  R1 = 4
R1 = 9   →  INC R1  →  R1 = 10
R1 = 14  →  INC R1  →  R1 = 15
R1 = 15  →  INC R1  →  R1 = 0   ← wrap, nessun carry
```

**In binario — il caso wrap:**
```
Prima:  R1 = 1111  (15)
+            0001  (1)
        ─────────
        1 0000  →  teniamo solo i 4 bit bassi → R1 = 0000  (0)
```

Il bit di overflow (il 5°) viene scartato. **Il carry C non viene toccato.**

**Confronto con IAC:**
- `INC R3` incrementa il registro R3, non tocca A
- `IAC` incrementa A, non tocca i registri

**Effetti:**

| Registro | Cambia? | Valore |
|----------|---------|--------|
| Rr       | ✅ sì   | valore precedente + 1 (mod 16) |
| A        | ❌ no   | invariato |
| C        | ❌ no   | invariato (INC non genera carry) |

---

## ADD — Add Register to Accumulator

**Cosa fa:** somma il valore del registro Rr e il carry C all'accumulatore A.
Il risultato è a 4 bit: se supera 15, il carry viene impostato e A contiene il resto.

**Formula:** `A = A + Rr + C`  (poi tronca a 4 bit, e aggiorna C)

**Opcode:** `0x8r` dove `r` è il numero del registro (0–F)

**Formato in binario:**
```
1000 rrrr
^^^^ ^^^^
│    └── numero registro
└─────── codice ADD
```

**Esempio senza carry:**
```
A = 3  (0011)
R0 = 2 (0010)
C = false

ADD R0:
  0011
+ 0010
──────
  0101 = 5   →   A = 5, C = false
```

**Esempio con carry out:**
```
A = 9  (1001)
R0 = 8 (1000)
C = false

ADD R0:
  1001
+ 1000
──────
1 0001  →  il 5° bit è 1 → C = true, A = 0001 = 1
```

**Esempio con carry in:**
```
A = 3  (0011)
R0 = 4 (0100)
C = true  ← già c'era un carry precedente

ADD R0:
  0011
+ 0100
+ 0001  ← il carry vale 1
──────
  1000 = 8   →   A = 8, C = false
```

**A cosa serve il carry in ADD?** Permette di sommare numeri a più di 4 bit.
Per sommare due numeri a 8 bit (due nibble ciascuno), si sommano prima i nibble bassi
con ADD, poi i nibble alti con ADD (che include automaticamente il carry del passo precedente).

**Effetti:**

| Registro | Cambia? | Valore |
|----------|---------|--------|
| A        | ✅ sì   | (A + Rr + C) mod 16 |
| C        | ✅ sì   | 1 se c'è overflow, 0 altrimenti |
| Rr       | ❌ no   | invariato |

---

## SUB — Subtract Register from Accumulator

**Cosa fa:** sottrae il valore del registro Rr da A usando il carry/link come borrow in ingresso.

**Formula Intel:** `A = A + ~Rr + ~C`  (poi tronca a 4 bit, e aggiorna C)

Il carry in ingresso viene **complementato** prima della somma:
- se `C = false`: `A = A - Rr` (sottrazione esatta)
- se `C = true`: `A = A - Rr - 1` (borrow dalla cifra precedente)

**Regola del carry/borrow in SUB:**
- Dopo SUB, se c'è stato borrow, C = **false**
- Se non c'è stato borrow, C = **true**

Attenzione: le convenzioni in ingresso e in uscita sono **opposte** — in
ingresso il borrow è `C = true`, in uscita il borrow è `C = false`. Per questo
nelle sottrazioni multi-cifra il manuale MCS-4 richiede di complementare il
carry con `CMC` tra una SUB e la successiva.

**Opcode:** `0x9r` dove `r` è il numero del registro (0–F)

**Formato in binario:**
```
1001 rrrr
^^^^ ^^^^
│    └── numero registro
└─────── codice SUB
```

**Esempio senza borrow:**
```
A = 7  (0111)
R2 = 3 (0011)
C = false

SUB R2:
  0111  (7)
- 0011  (3)
──────
  0100  (4)   →   A = 4, C = true  (nessun prestito)
```

**Esempio con borrow out (risultato negativo):**
```
A = 3  (0011)
R2 = 7 (0111)
C = false

SUB R2:
  3 - 7 = -4

Il 4004 tratta i 4 bit come un ciclo, quindi -4 diventa 12 (16 - 4 = 12):
  0011  (3)
- 0111  (7)
──────
  1100  (12)   →   A = 12, C = false  (c'è stato un prestito)
```

**Esempio con borrow in:**
```
A = 5
R2 = 3
C = true  ← borrow dal calcolo precedente (es. dopo CMC)

SUB R2:
  5 - 3 - 1 = 1   →   A = 1, C = true
```

**Effetti:**

| Registro | Cambia? | Valore |
|----------|---------|--------|
| A        | ✅ sì   | A + ~Rr + ~C (mod 16) |
| C        | ✅ sì   | true se non c'è borrow, false se c'è borrow |
| Rr       | ❌ no   | invariato |

---

## BBL — Branch Back and Load

**Cosa fa:** ritorna da una subroutine. Estrae l'indirizzo di ritorno dallo stack (salvato da JMS),
ripristina PC, e carica un valore di ritorno nell'accumulatore A.

**Opcode:** `0xCn` dove `n` è il valore da caricare in A al ritorno (0–F)

**Formato in binario:**
```
1100 nnnn
^^^^ ^^^^
│    └── valore da mettere in A
└─────── codice BBL
```

**Come funziona in coppia con JMS:**

```
ROM[0x002] = JMS 0x010   ← chiama subroutine a 0x010, salva 0x004 sullo stack
ROM[0x003] = ...
ROM[0x004] = ...          ← PC torna qui dopo BBL

ROM[0x010] = LDM 7       ← corpo subroutine: fa qualcosa
ROM[0x011] = BBL 5       ← ritorna con A = 5, PC = 0x004
```

**BBL carica il valore n in A** — è il meccanismo per "restituire un valore" da una subroutine,
simile a `return 5` in Go. Il chiamante può leggere A subito dopo il ritorno.

**Esempio pratico:**
```
Prima di BBL 5:
  A  = qualsiasi
  SP = 1 (c'è un indirizzo nello stack: 0x004)

Eseguo: BBL 5  (opcode 0xC5)

Dopo:
  A  = 5           ← valore di ritorno
  PC = 0x004       ← indirizzo estratto dallo stack
  SP = 0           ← stack svuotato
  C  = invariato
```

**In binario:**
```
Opcode BBL 5:
  1100 0101
  ^^^^ ^^^^
  │    └── n = 0101 = 5 → A = 5
  └─────── codice BBL
```

**Effetti:**

| Registro | Cambia? | Valore |
|----------|---------|--------|
| A        | ✅ sì   | valore n (0–15) |
| PC       | ✅ sì   | indirizzo di ritorno dallo stack |
| SP       | ✅ sì   | SP - 1 |
| C        | ❌ no   | invariato |

---

## Riepilogo del gruppo

| Istruzione | Opcode  | Cosa cambia         | Cosa NON cambia |
|------------|---------|---------------------|-----------------|
| NOP        | `0x00`  | niente              | tutto           |
| LDM n      | `0xDn`  | A = n               | C, registri     |
| LD Rr      | `0xAr`  | A = Rr              | C, Rr           |
| XCH Rr     | `0xBr`  | A ↔ Rr              | C               |
| INC Rr     | `0x6r`  | Rr = Rr + 1 (mod 16)| A, C            |
| ADD Rr     | `0x8r`  | A = A+Rr+C, aggiorna C | Rr            |
| SUB Rr     | `0x9r`  | A = A+~Rr+~C, aggiorna C | Rr            |
| BBL n      | `0xCn`  | A = n, PC←stack     | C               |

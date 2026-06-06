# Gruppo I/O e RAM — Istruzioni Intel 4004

Queste istruzioni permettono al 4004 di **leggere e scrivere dati** nella memoria RAM
e di comunicare con dispositivi esterni tramite porte I/O.

Il 4004 **non ha RAM interna**: accede a chip di memoria esterni chiamati **Intel 4002**.
Ogni istruzione di questo gruppo opera sul chip e sul registro selezionato
dalla coppia di istruzioni `DCL` + `SRC`.

Tutti gli opcode di questo gruppo sono nella forma `0xEX` (byte singolo fisso).

---

## Il chip Intel 4002 — come funziona la RAM

Ogni chip 4002 contiene:

```
4 registri × 20 nibble di dati  = 80 nibble (area dati)
4 registri × 4 nibble di stato  = 16 nibble (area status)
1 porta di output da 4 bit
```

Un sistema 4004 può collegare fino a **4 chip** (banco 0–3).

**Area dati:** usata per conservare i valori principali (cifre BCD, risultati, ecc.)  
**Area status:** usata per flag applicativi (segno, overflow, metadati)  
**Porta output:** usata per inviare dati a dispositivi esterni (display, LED, ecc.)

---

## Come si seleziona l'indirizzo

Prima di ogni accesso alla RAM il firmware deve indicare dove leggere/scrivere.
Si usano due istruzioni:

**1. DCL — seleziona il banco**
```
LDM 0    ← A = 0
DCL      ← CL = 0 (banco 0 attivo)
```

**2. SRC Rp — seleziona registro e carattere**
```
FIM R0, 0x05   ← R0 = 0 (registro 0), R1 = 5 (carattere 5)
SRC R0         ← SRCAddr = (R0 << 4) | R1 = 0x05
```

**Formato di SRCAddr:**
```
7 6 5 4  │  3 2 1 0
─────────┼─────────
nibble   │  nibble
alto     │  basso
registro │  carattere
(0–3)    │  (0–15)
```

**Indirizzo finale usato da WRM/RDM/ADM/SBM:**
```
banco     = CL & 0x3
registro  = (SRCAddr >> 4) & 0x3
carattere = SRCAddr & 0x0F
```

---

## Mappa opcode del gruppo

```
0xE0  → WRM  — scrive A nella RAM data
0xE1  → WMP  — scrive A sulla porta output RAM
0xE2  → WRR  — scrive A sulla porta output ROM
0xE3  → WPM  — scrive A in program memory
0xE4  → WR0  — scrive A nello status nibble 0
0xE5  → WR1  — scrive A nello status nibble 1
0xE6  → WR2  — scrive A nello status nibble 2
0xE7  → WR3  — scrive A nello status nibble 3
0xE8  → SBM  — A = A - RAM - borrow
0xE9  → RDM  — A = RAM data
0xEA  → RDR  — A = porta input ROM
0xEB  → ADM  — A = A + RAM + carry
0xEC  → RD0  — A = status nibble 0
0xED  → RD1  — A = status nibble 1
0xEE  → RD2  — A = status nibble 2
0xEF  → RD3  — A = status nibble 3
```

---

## WRM — Write RAM Main Memory

**Cosa fa:** scrive il nibble dell'accumulatore nella cella di dati RAM selezionata.
È l'istruzione fondamentale per salvare un valore in memoria.

**Opcode:** `0xE0`

**Formula:** `RAM.Data[banco][registro][carattere] = nibble(A)`

**Esempio:**
```
CL = 0, SRCAddr = 0x05  (registro 0, carattere 5)
A = 7

WRM

Risultato: RAM.Data[0][0][5] = 7
           A = 7 (invariato)
           C = invariato
```

**Uso tipico — salvare una cifra BCD:**
```
LDM 0 / DCL           ← banco 0
FIM R0, 0x00          ← R0=0 (registro 0), R1=0 (carattere 0)
SRC R0                ← SRCAddr = 0x00

LDM 3                 ← A = 3 (cifra da salvare)
WRM                   ← RAM.Data[0][0][0] = 3

FIM R0, 0x01          ← prossimo carattere
SRC R0
LDM 7
WRM                   ← RAM.Data[0][0][1] = 7
```

**Effetti:**

| Cosa         | Cambia? |
|--------------|---------|
| RAM.Data[b][r][c] | ✅ sì — riceve nibble(A) |
| A            | ❌ no   |
| C            | ❌ no   |

---

## RDM — Read RAM Main Memory

**Cosa fa:** legge il nibble dalla cella di dati RAM selezionata e lo carica in A.
È l'inverso di WRM.

**Opcode:** `0xE9`

**Formula:** `A = nibble(RAM.Data[banco][registro][carattere])`

**Esempio:**
```
RAM.Data[0][0][5] = 9
CL = 0, SRCAddr = 0x05
A = 0

RDM

Risultato: A = 9
           C = invariato
```

**Uso tipico — leggere una cifra BCD:**
```
FIM R0, 0x03 / SRC R0   ← seleziona carattere 3
RDM                      ← A = cifra al carattere 3
ADD R2                   ← somma al registro R2
DAA                      ← correzione BCD
```

**Effetti:**

| Cosa | Cambia? |
|------|---------|
| A    | ✅ sì — valore letto dalla RAM |
| C    | ❌ no   |

---

## ADM — Add RAM to Accumulator

**Cosa fa:** somma il valore della cella RAM corrente e il carry all'accumulatore.
È identica a `ADD Rr` ma legge l'operando dalla RAM invece che da un registro.

**Opcode:** `0xEB`

**Formula:** `result = A + RAM.Data[b][r][c] + carry;  A = nibble(result);  C = result > 0x0F`

**Esempio:**
```
A = 7, C = false
RAM.Data[0][0][0] = 6

ADM

result = 7 + 6 + 0 = 13  →  A = 13, C = false

DAA  →  13 + 6 = 19 → nibble = 3, C = true
```

**Quando si usa:** addizione BCD cifra per cifra — somma la cifra in RAM a quella in A.

**Effetti:**

| Cosa | Cambia? |
|------|---------|
| A    | ✅ sì — (A + RAM + C) mod 16 |
| C    | ✅ sì — 1 se risultato > 15 |

---

## SBM — Subtract RAM from Accumulator

**Cosa fa:** sottrae il valore della cella RAM corrente da A usando il carry come link.
Identica a `SUB Rr` ma legge l'operando dalla RAM.

**Opcode:** `0xE8`

**Formula:** `result = A + ~RAM + carry;  A = nibble(result);  C = result > 0x0F`

Come in `SUB`: `C = true` significa nessun borrow, `C = false` significa borrow generato.

**Esempio senza borrow:**
```
A = 7, C = true (nessun borrow precedente)
RAM.Data[0][0][0] = 3

SBM

result = 7 + ~3 + 1 = 7 + 12 + 1 = 20 > 15  →  A = nibble(20) = 4, C = true
```

**Esempio con borrow (risultato negativo):**
```
A = 3, C = true
RAM.Data[0][0][0] = 7

SBM

result = 3 + ~7 + 1 = 3 + 8 + 1 = 12  →  A = 12, C = false (borrow!)
```

**Effetti:**

| Cosa | Cambia? |
|------|---------|
| A    | ✅ sì — (A + ~RAM + C) mod 16 |
| C    | ✅ sì — true = no borrow, false = borrow |

---

## WMP — Write RAM Port

**Cosa fa:** scrive A sulla **porta di output** del banco RAM attivo.
La porta è un registro a 4 bit che pilota direttamente pin hardware
(display a segmenti, LED, relè, ecc.).

**Opcode:** `0xE1`

**Formula:** `RAM.Port[banco] = nibble(A)`

**Esempio:**
```
CL = 0  (banco 0)
A = 0b0110  (segmenti b e c accesi su un display 7-segmenti)

WMP

Risultato: RAM.Port[0] = 6
           A = 6 (invariato)
```

**Uso tipico — display a 7 segmenti:**
```
LDM 0b0111    ← accendi segmenti a, b, c (cifra "7")
WMP           ← invia al display collegato alla porta
```

**Effetti:**

| Cosa         | Cambia? |
|--------------|---------|
| RAM.Port[b]  | ✅ sì — nibble(A) |
| A            | ❌ no   |
| C            | ❌ no   |

---

## WR0 / WR1 / WR2 / WR3 — Write RAM Status

**Cosa fanno:** scrivono A nell'area di **stato** del registro RAM selezionato.
L'area status è separata dall'area dati ed è usata per flag applicativi:
segno del numero, flag di overflow, stato della macchina, ecc.

**Opcode:**
```
WR0 → 0xE4
WR1 → 0xE5
WR2 → 0xE6
WR3 → 0xE7
```

**Formula:** `RAM.Status[banco][registro][n] = nibble(A)`  (n = 0, 1, 2 o 3)

**Esempio:**
```
CL = 0, SRCAddr = 0x00  (registro 0)
A = 1  (flag "numero negativo")

WR0

Risultato: RAM.Status[0][0][0] = 1
```

**Uso tipico — salvare il segno:**
```
; dopo una sottrazione che ha generato borrow:
LDM 1         ← 1 = negativo
WR0           ← salva il flag segno in status nibble 0

; dopo una sottrazione senza borrow:
LDM 0         ← 0 = positivo
WR0
```

**Effetti:**

| Cosa              | Cambia? |
|-------------------|---------|
| RAM.Status[b][r][n] | ✅ sì — nibble(A) |
| A                 | ❌ no   |
| C                 | ❌ no   |

---

## RD0 / RD1 / RD2 / RD3 — Read RAM Status

**Cosa fanno:** leggono i nibble di **stato** del registro RAM selezionato in A.
Simmetrici a WR0–WR3.

**Opcode:**
```
RD0 → 0xEC
RD1 → 0xED
RD2 → 0xEE
RD3 → 0xEF
```

**Formula:** `A = nibble(RAM.Status[banco][registro][n])`  (n = 0, 1, 2 o 3)

**Esempio:**
```
RAM.Status[0][0][0] = 1  (flag negativo salvato in precedenza)

RD0

Risultato: A = 1
           C = invariato
```

**Uso tipico — verificare il segno prima di visualizzare:**
```
RD0           ← legge il flag segno in A
JCN 0x4, neg  ← se A == 0, salta a "positivo"
; A = 1 → numero negativo, mostra '-'
```

**Effetti:**

| Cosa | Cambia? |
|------|---------|
| A    | ✅ sì — valore dello status nibble n |
| C    | ❌ no   |

---

## WRR — Write ROM Port

> ⚠️ Non ancora implementato nell'emulatore.

**Cosa fa:** scrive A sulla **porta di output** del chip ROM attivo (Intel 4001).
Ogni chip 4001 ha una porta I/O a 4 bit configurabile come input o output.

**Opcode:** `0xE2`

**Formula:** `ROM.Port[chipROM] = nibble(A)`

**Uso tipico:** in un sistema reale WRR controlla le linee di scansione di una
tastiera a matrice o pilota output generici collegati al chip ROM.

**Effetti:**

| Cosa         | Cambia? |
|--------------|---------|
| ROM.Port[n]  | ✅ sì — nibble(A) |
| A            | ❌ no   |
| C            | ❌ no   |

---

## RDR — Read ROM Port

> ⚠️ Non ancora implementato nell'emulatore.

**Cosa fa:** legge i 4 bit della **porta di input** del chip ROM attivo in A.
Usato per leggere lo stato dei tasti premuti durante la scansione della tastiera.

**Opcode:** `0xEA`

**Formula:** `A = nibble(ROM.Port[chipROM])`

**Uso tipico — lettura tastiera:**
```
; scansione riga per riga:
LDM 0b0001 / WRR    ← attiva riga 1
RDR                  ← leggi quali colonne sono attive
KBP                  ← decodifica one-hot → numero tasto
JCN ...              ← gestisci il tasto premuto
```

**Effetti:**

| Cosa | Cambia? |
|------|---------|
| A    | ✅ sì — stato della porta ROM |
| C    | ❌ no   |

---

## WPM — Write Program Memory

> ⚠️ Non ancora implementato nell'emulatore.

**Cosa fa:** scrive A in program memory (la ROM).
Sul hardware reale era usato durante la programmazione dei chip PROM 4001.
Nei sistemi con ROM mask-programmed (non modificabile) questa istruzione
non ha effetto pratico e viene trattata come no-op.

**Opcode:** `0xE3`

**Nota:** WPM richiede circuiti hardware speciali per la programmazione della ROM.
In un emulatore può essere implementata come stub (non fa nulla) senza impatto
sul firmware della calcolatrice.

**Effetti:** nessuno in un sistema con ROM fissa.

---

## Riepilogo del gruppo

| Istruzione | Opcode | Implementata | Cosa fa in breve |
|------------|--------|:---:|------------------|
| WRM        | `0xE0` | ✅ | RAM.Data[b][r][c] = A |
| WMP        | `0xE1` | ✅ | RAM.Port[b] = A |
| WRR        | `0xE2` | 🔲 | ROM.Port = A (output hardware) |
| WPM        | `0xE3` | 🔲 | scrive in program memory (raro) |
| WR0        | `0xE4` | ✅ | RAM.Status[b][r][0] = A |
| WR1        | `0xE5` | ✅ | RAM.Status[b][r][1] = A |
| WR2        | `0xE6` | ✅ | RAM.Status[b][r][2] = A |
| WR3        | `0xE7` | ✅ | RAM.Status[b][r][3] = A |
| SBM        | `0xE8` | ✅ | A = A - RAM - borrow |
| RDM        | `0xE9` | ✅ | A = RAM.Data[b][r][c] |
| RDR        | `0xEA` | 🔲 | A = ROM.Port (input tastiera) |
| ADM        | `0xEB` | ✅ | A = A + RAM + carry |
| RD0        | `0xEC` | ✅ | A = RAM.Status[b][r][0] |
| RD1        | `0xED` | ✅ | A = RAM.Status[b][r][1] |
| RD2        | `0xEE` | ✅ | A = RAM.Status[b][r][2] |
| RD3        | `0xEF` | ✅ | A = RAM.Status[b][r][3] |

**Nessuna istruzione di questo gruppo tocca i registri R0–RF.**  
**Le istruzioni di lettura (RDM, ADM, SBM, RD0–RD3) non modificano il carry**
**tranne ADM e SBM che lo aggiornano come ADD e SUB.**

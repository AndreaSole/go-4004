# Gruppo Accumulatore — Istruzioni Intel 4004

Queste istruzioni lavorano **solo sull'accumulatore A e/o sul carry C**.
Non toccano i registri R0–RF, né la ROM, né lo stack.

Tutte le istruzioni di questo gruppo sono a **1 byte fisso** — l'opcode non ha parametri variabili.
Il loro opcode è sempre nella forma `0xFX`.

---

## Mappa opcode del gruppo

```
0xF0  → CLB  — azzera A e C
0xF1  → CLC  — azzera solo C
0xF2  → IAC  — incrementa A
0xF3  → CMC  — complementa C
0xF4  → CMA  — complementa A
0xF5  → RAL  — ruota A a sinistra attraverso C
0xF6  → RAR  — ruota A a destra attraverso C
0xF7  → TCC  — trasferisce C in A, azzera C
0xF8  → DAC  — decrementa A
0xF9  → TCS  — trasferisce C in A come 9 o 10, azzera C
0xFA  → STC  — imposta C = 1
0xFB  → DAA  — corregge A per aritmetica BCD
0xFC  → KBP  — decodifica tastiera one-hot
0xFD  → DCL  — imposta il banco RAM attivo
```

---

## CLB — Clear Both

**Cosa fa:** azzera sia l'accumulatore A che il carry C in una sola istruzione.

**Opcode:** `0xF0` = `1111 0000`

**Esempio:**
```
Prima:  A = 9, C = true

Eseguo: CLB  (0xF0)

Dopo:   A = 0, C = false
```

**Quando si usa:** all'inizio di un calcolo per partire da uno stato pulito.

---

## CLC — Clear Carry

**Cosa fa:** azzera solo il carry C. Non tocca A.

**Opcode:** `0xF1` = `1111 0001`

**Esempio:**
```
Prima:  A = 7, C = true

Eseguo: CLC  (0xF1)

Dopo:   A = 7 (invariato), C = false
```

**Quando si usa:** prima di una sequenza di addizioni per assicurarsi che non ci sia
un carry "sporco" dal calcolo precedente.

---

## STC — Set Carry

**Cosa fa:** imposta il carry C a 1 (true). Non tocca A.

**Opcode:** `0xFA` = `1111 1010`

**Esempio:**
```
Prima:  A = 3, C = false

Eseguo: STC  (0xFA)

Dopo:   A = 3 (invariato), C = true
```

**Quando si usa:** forzare un carry prima di un'operazione che lo usa come input (ADD, SUB, RAL, RAR).

---

## CMC — Complement Carry

**Cosa fa:** inverte il carry C. Se era true diventa false, e viceversa. Non tocca A.

**Opcode:** `0xF3` = `1111 0011`

**Esempio:**
```
Prima:  C = false
Dopo:   C = true

Prima:  C = true
Dopo:   C = false
```

**Quando si usa:** capovolgere lo stato del carry senza passare da CLC/STC.

---

## IAC — Increment Accumulator

**Cosa fa:** aggiunge 1 ad A. Se A vale 15 (`1111`) torna a 0 (`0000`) e C diventa true.

**Opcode:** `0xF2` = `1111 0010`

**Formula:** `A = A + 1`  (4 bit), aggiorna C

**Esempi:**
```
A = 5, C = false   →  IAC  →  A = 6, C = false
A = 9, C = true    →  IAC  →  A = 10, C = false
A = 15, C = false  →  IAC  →  A = 0,  C = true   ← overflow
```

**In binario (caso overflow):**
```
  1111  (15)
+ 0001  (1)
──────
1 0000  →  teniamo 4 bit → A = 0000 = 0,  C = true (il 5° bit)
```

**Differenza con INC:**
- `IAC` incrementa A, non tocca i registri
- `INC R3` incrementa R3, non tocca A e non aggiorna C

---

## DAC — Decrement Accumulator

**Cosa fa:** sottrae 1 da A. Se A vale 0 (`0000`) va a 15 (`1111`) e C diventa false
(segnala un "borrow" — ha dovuto "prendere a prestito"). Se non c'è borrow, C diventa true.

**Opcode:** `0xF8` = `1111 1000`

**Formula:** `A = A - 1` (4 bit), aggiorna C

**Esempi:**
```
A = 5  →  DAC  →  A = 4,  C = true
A = 1  →  DAC  →  A = 0,  C = true
A = 0  →  DAC  →  A = 15, C = false   ← underflow/borrow
```

**In binario (caso underflow):**
```
  0000  (0)
- 0001  (1)
──────
Il 4004 risolve così: 16 + 0 - 1 = 15
  1111  (15)   →   A = 15, C = false
```

**Quando si usa:** contatori che devono decrementare, calcoli BCD.

---

## CMA — Complement Accumulator

**Cosa fa:** inverte tutti i 4 bit di A (complemento a 1). Non tocca C.

**Opcode:** `0xF4` = `1111 0100`

**Esempi:**
```
A = 0000  (0)   →  CMA  →  A = 1111  (15)
A = 1010  (10)  →  CMA  →  A = 0101  (5)
A = 0101  (5)   →  CMA  →  A = 1010  (10)
A = 1111  (15)  →  CMA  →  A = 0000  (0)
```

**In binario — passo per passo:**
```
A = 0110  (6)

CMA inverte ogni bit:
  0 → 1
  1 → 0
  1 → 0
  0 → 1

Risultato: 1001  (9)   →   A = 9
```

**Nota:** CMA + IAC insieme danno il **complemento a 2** (negazione in aritmetica binaria):
```
A = 5 = 0101
CMA  →  A = 1010  (complemento a 1)
IAC  →  A = 1011  (complemento a 2 = -5 in binario)
```

---

## RAL — Rotate Accumulator Left

**Cosa fa:** ruota i 4 bit di A verso sinistra di una posizione, passando attraverso il carry C.
Il bit che "esce" da sinistra (bit 3) entra nel carry. Il vecchio carry entra da destra (bit 0).

**Opcode:** `0xF5` = `1111 0101`

**Schema della rotazione:**
```
C  ←  [bit3][bit2][bit1][bit0]  ←  C
        ↑ esce                      ↑ entra
```

**In binario — esempio 1:**
```
A = 0110  (6)
C = false (0)

RAL:
  C_nuovo  = bit3 di A = 0
  A_nuovo  = [bit2][bit1][bit0][C_vecchio]
           = [1][1][0][0]
           = 1100  (12)

Risultato: A = 12, C = false
```

**In binario — esempio 2 (con carry che entra):**
```
A = 0110  (6)
C = true  (1)

RAL:
  C_nuovo  = bit3 di A = 0
  A_nuovo  = [1][1][0][1]   ← il vecchio C=1 entra a destra
           = 1101  (13)

Risultato: A = 13, C = false
```

**In binario — esempio 3 (carry che esce):**
```
A = 1010  (10)
C = false (0)

RAL:
  C_nuovo  = bit3 di A = 1  ← il bit più alto esce
  A_nuovo  = [0][1][0][0]
           = 0100  (4)

Risultato: A = 4, C = true
```

**Quando si usa:** moltiplicare per 2 (shift left), operazioni su bit singoli, serializzare dati.

---

## RAR — Rotate Accumulator Right

**Cosa fa:** come RAL ma nella direzione opposta.
Il bit che "esce" da destra (bit 0) va nel carry. Il vecchio carry entra da sinistra (bit 3).

**Opcode:** `0xF6` = `1111 0110`

**Schema della rotazione:**
```
C  →  [bit3][bit2][bit1][bit0]  →  C
  ↑ entra                             ↑ esce
```

**In binario — esempio:**
```
A = 0110  (6)
C = false (0)

RAR:
  C_nuovo  = bit0 di A = 0
  A_nuovo  = [C_vecchio][bit3][bit2][bit1]
           = [0][0][1][1]
           = 0011  (3)

Risultato: A = 3, C = false
```

**Esempio con carry:**
```
A = 0101  (5)
C = true  (1)

RAR:
  C_nuovo  = bit0 di A = 1  ← il bit più basso esce
  A_nuovo  = [1][0][1][0]   ← il vecchio C=1 entra a sinistra
           = 1010  (10)

Risultato: A = 10, C = true
```

**Quando si usa:** dividere per 2 (shift right), deserializzare dati, test di bit bassi.

---

## TCC — Transfer Carry to Accumulator and Clear

**Cosa fa:** copia il carry in A come 0 o 1, poi azzera il carry.

**Opcode:** `0xF7` = `1111 0111`

**Regola:**
```
Se C = true  →  A = 1, poi C = false
Se C = false →  A = 0, poi C = false
```

**Esempio:**
```
A = 9, C = true   →  TCC  →  A = 1, C = false
A = 3, C = false  →  TCC  →  A = 0, C = false
```

**Quando si usa:** leggere il carry come dato numerico (0 o 1) per usarlo in calcoli.
Ad esempio, in un'addizione multi-nibble, TCC permette di "recuperare" il carry come nibble da sommare.

---

## TCS — Transfer Carry Subtract and Clear

**Cosa fa:** come TCC, ma carica 10 invece di 1 e 9 invece di 0. Usato nella correzione BCD.

**Opcode:** `0xF9` = `1111 1001`

**Regola:**
```
Se C = true  →  A = 10, poi C = false
Se C = false →  A = 9,  poi C = false
```

**Esempio:**
```
C = true   →  TCS  →  A = 10, C = false
C = false  →  TCS  →  A =  9, C = false
```

**Perché 9 e 10?** TCS viene usato nella correzione BCD delle sottrazioni.
In BCD si lavora in base 10. Dopo una sottrazione, il 4004 usa C alto per indicare
che non c'è stato borrow: TCS carica 10 se C era true, 9 se C era false.
Vedere `docs/bcd.md` per dettagli.

---

## DAA — Decimal Adjust Accumulator

**Cosa fa:** corregge il valore in A per renderlo valido in **BCD** dopo un'addizione.

In BCD ogni cifra decimale (0–9) è rappresentata con 4 bit.
Ma ADD produce risultati 0–15, e i valori 10–15 non sono cifre BCD valide.
DAA aggiunge 6 ad A se A > 9 oppure se C è true, per "saltare" i valori non validi.

**Opcode:** `0xFB` = `1111 1011`

**Regola:**
```
Se (C = true) OPPURE (A > 9):
    A = A + 6
    C = true
Altrimenti:
    niente
```

**Perché +6?** In esadecimale i valori 0xA–0xF (10–15) non esistono in BCD.
Aggiungendo 6 si "salta" quei 6 valori non validi e si torna nel range BCD 0–9
(con eventuale carry che indica il riporto alla cifra successiva).

**Esempi:**
```
A = 7, C = false  →  DAA  →  niente (7 è BCD valido)  →  A = 7, C = false
A = 12, C = false →  DAA  →  A = 12+6 = 18 → nibble = 2, C = true  →  A = 2, C = true
A = 1, C = true   →  DAA  →  A = 1+6 = 7                            →  A = 7, C = true
```

**Esempio reale — somma BCD di 8 + 5:**
```
LDM 8    →  A = 8
ADD R0   →  A = 8+5 = 13  (0xD, non valido BCD), C = false
DAA      →  A = 13+6 = 19 → nibble basso = 3, C = true  →  A = 3, C = true

Risultato: cifra delle unità = 3, carry (cifra delle decine) = 1 → risultato = 13 ✓
```

Vedere `docs/bcd.md` per la spiegazione completa della codifica BCD.

---

## KBP — Keyboard Process

**Cosa fa:** converte un valore **one-hot** (un solo bit attivo) in un numero di posizione.
Usato per decodificare quale tasto di una tastiera a matrice è stato premuto.

**Opcode:** `0xFC` = `1111 1100`

**One-hot** significa che solo un bit alla volta è 1:
```
0001 = tasto 1 premuto
0010 = tasto 2 premuto
0100 = tasto 3 premuto
1000 = tasto 4 premuto
```

**Tabella di conversione:**
```
A = 0000  →  KBP  →  A = 0   (nessun tasto)
A = 0001  →  KBP  →  A = 1
A = 0010  →  KBP  →  A = 2
A = 0100  →  KBP  →  A = 3
A = 1000  →  KBP  →  A = 4
A = 0011  →  KBP  →  A = 15  (errore: due bit attivi)
A = 1111  →  KBP  →  A = 15  (errore)
```

**In binario:**
```
A = 0100  (tasto 3 premuto)
KBP
Risultato: A = 0011  (3)
```

Se A ha più di un bit a 1 (input non valido), KBP restituisce 0xF (15) come codice di errore.

**Non modifica C.**

---

## DCL — Designate Command Line

**Cosa fa:** imposta il **banco RAM attivo** copiando i bit bassi di A nel registro CL interno.
Le successive istruzioni RAM (WRM, RDM, ecc.) opereranno sul banco selezionato.

**Opcode:** `0xFD` = `1111 1101`

**Formula:** `CL = A & 0b0011`  (2 bit: banchi 0–3)

**Esempio:**
```
A = 3  (0011)
DCL
Risultato: CL = 3  →  banco RAM 3 selezionato
           A = 3  (invariato)
           C = invariato
```

**Nota:** il 4004 reale supporta fino a 8 banchi RAM tramite le linee CM-RAM; l'emulatore modella 4 banchi (un chip 4002 per banco), quindi CL usa 2 bit. I bit alti di A vengono ignorati.

**Schema di utilizzo:**
```
LDM 2    ← voglio usare il banco RAM 2
DCL      ← CL = 2, ora il banco 2 è attivo
...
WRM      ← scrive A nel banco RAM 2 (selezionato da DCL + SRC)
```

DCL seleziona il **gruppo** di chip RAM. SRC (vedi istruzioni di salto) seleziona
il registro specifico all'interno del gruppo. I due lavorano insieme.

---

## Riepilogo del gruppo

| Istruzione | Opcode | Cosa fa in breve |
|------------|--------|------------------|
| CLB        | `0xF0` | A=0, C=false |
| CLC        | `0xF1` | C=false |
| IAC        | `0xF2` | A++ (con carry) |
| CMC        | `0xF3` | C = !C |
| CMA        | `0xF4` | A = ~A (inverte bit) |
| RAL        | `0xF5` | ruota A sinistra attraverso C |
| RAR        | `0xF6` | ruota A destra attraverso C |
| TCC        | `0xF7` | A=C (0 o 1), C=false |
| DAC        | `0xF8` | A-- (con borrow) |
| TCS        | `0xF9` | A=9 o 10 in base a C, C=false |
| STC        | `0xFA` | C=true |
| DAA        | `0xFB` | corregge A per BCD (+6 se necessario) |
| KBP        | `0xFC` | decodifica one-hot → posizione bit |
| DCL        | `0xFD` | CL = A (seleziona banco RAM) |

Tutte le istruzioni di questo gruppo **non toccano i registri R0–RF**.

# Guida alle 46 Istruzioni Intel 4004 — Spiegazione Completa

Questa guida spiega tutte le 46 istruzioni del processore Intel 4004 partendo da zero,
con esempi in binario e analogie concrete. Non è un manuale tecnico — è scritta per
capire davvero cosa fa ogni istruzione e perché esiste.

---

## Parte 1 — Il modello mentale del 4004

Prima di parlare di istruzioni, devi avere chiaro il "teatro" in cui operano.

### Bit, nibble, e perché il 4004 è "a 4 bit"

Un **bit** è la cosa più piccola che esiste nell'informatica: o 0 o 1.
Un **nibble** sono 4 bit messi in fila. Con 4 bit puoi contare da 0 a 15:

```
0000 = 0       0100 = 4       1000 = 8       1100 = 12
0001 = 1       0101 = 5       1001 = 9       1101 = 13
0010 = 2       0110 = 6       1010 = 10      1110 = 14
0011 = 3       0111 = 7       1011 = 11      1111 = 15
```

Il 4004 è "a 4 bit" perché **tutte le sue unità di lavoro sono nibble**.
L'accumulatore è un nibble. Ogni registro è un nibble. Ogni cifra in RAM è un nibble.
Anche gli opcode sono gruppi di nibble (1 o 2 byte, cioè 2 o 4 nibble).

Quando vedi un numero in esadecimale come `0xB3`, significa:
```
0xB3  =  1011 0011
          ^^^^       ← nibble alto = B = 11
               ^^^^  ← nibble basso = 3 = 3
```

### L'accumulatore A — il foglio di carta dei calcoli

**A** è un registro a 4 bit. È l'unico registro direttamente connesso all'ALU
(la parte della CPU che fa i calcoli). Quasi ogni calcolo passa per A:

- vuoi sommare? il risultato finisce in A
- vuoi leggere dalla RAM? il valore arriva in A
- vuoi inviare dati all'esterno? prendi da A

Pensa ad A come al **display di una calcolatrice tascabile**: vedi sempre cosa c'è,
ma può contenere un solo valore alla volta (0–15).

### Il carry flag C — il riporto che non dimentichi

**C** è un singolo bit (vero/falso). Rappresenta il "riporto" delle operazioni aritmetiche.

Ricordi la somma in colonna a scuola?
```
  8
+ 9
───
 17  ← il "1" che scrivi a sinistra è il riporto
```

In un nibble a 4 bit non c'è posto per quel "1" extra, quindi il 4004 lo salva in C:
```
  1001  (9)
+ 1000  (8)
──────
1 0001  ← il 5° bit non entra nel nibble → va in C = true
         A = 0001 = 1
```

C può anche entrare **nei calcoli successivi**: le istruzioni ADD e SUB includono C
nella formula. Questo permette di sommare numeri più grandi di 15 (cifra per cifra).

Nella **sottrazione** C funziona al contrario: C=true significa "nessun prestito",
C=false significa "ho dovuto prendere a prestito" (borrow). Sembra strano ma ha senso
a livello hardware — ricordatelo per SUB, DAC, SBM.

### I registri R0–RF — i post-it della CPU

Il 4004 ha **16 registri** chiamati R0, R1, R2, ... R9, RA, RB, RC, RD, RE, RF.
Ognuno contiene un nibble (0–15). Non sono collegati all'ALU direttamente — per fare
calcoli su un registro devi prima spostarlo in A.

Pensali come **post-it**: ci scrivi valori temporanei, contatori, indirizzi, operandi.
Non puoi sommare R2 + R5 direttamente: devi copiare R2 in A, poi sommare R5 ad A.

I registri si usano spesso a coppie (R0+R1, R2+R3, ecc.) per rappresentare indirizzi
a 8 bit: il primo del paio contiene il nibble alto, il secondo il nibble basso.

### Il Program Counter PC — il segnalibro nel libro delle istruzioni

**PC** è un numero a 12 bit (0x000–0xFFF) che indica **quale istruzione eseguire dopo**.
Dopo ogni fetch, PC avanza automaticamente.

Pensa alla ROM come a un libro di istruzioni con 4096 pagine (0x000–0xFFF).
PC è il segnalibro: segna sempre la pagina corrente. Quando esegui JUN, JMS o JCN
stai spostando il segnalibro a una pagina diversa.

La ROM è divisa in **16 pagine da 256 byte** ciascuna:
```
Pagina 0:  0x000–0x0FF
Pagina 1:  0x100–0x1FF
...
Pagina F:  0xF00–0xFFF
```

Alcune istruzioni di salto (JCN, ISZ, JIN) possono saltare **solo dentro la pagina
corrente** — usano solo 8 bit per l'indirizzo. Per saltare ovunque nella ROM ci vuole
JUN o JMS (12 bit).

### Lo stack hardware — la pila di segnalibri

Quando chiami una subroutine con JMS, il 4004 deve ricordare dove tornare.
Lo **stack** è una piccola pila di 3 indirizzi di ritorno. JMS li mette nella pila,
BBL li toglie e torna all'indirizzo salvato.

```
Prima di JMS:   Stack = [_, _, _]   SP = 0
Dopo JMS:       Stack = [0x042, _, _]   SP = 1
Dopo BBL:       Stack = [_, _, _]   SP = 0   e PC = 0x042
```

Lo stack ha **solo 3 livelli**. Se fai 4 JMS senza BBL nel mezzo, il quarto JMS
sovrascrive il primo indirizzo senza errori (comportamento hardware reale).
Quindi: non annidare più di 3 subroutine.

### RAM e ROM — memoria di lavoro vs istruzioni

- **ROM** (Read-Only Memory): contiene il firmware. Non si modifica a runtime.
  4096 byte (0x000–0xFFF). Il PC la percorre sequenzialmente.
  
- **RAM** (chip Intel 4002): contiene i dati del programma — cifre BCD, risultati,
  flag. Si legge e si scrive con le istruzioni del gruppo 0xEX.
  La RAM è organizzata in chip: 4 banchi × 4 registri × 20 nibble dati + area status.

### Il CL e SRCAddr — come puntare alla RAM

Prima di leggere/scrivere in RAM devi dire al 4004 **dove** operare. Si usano due
registri speciali:

- **CL** (Command Line): selezionato da `DCL`, indica il banco RAM (0–3)
- **SRCAddr**: selezionato da `SRC`, indica registro e carattere dentro il banco

È come dare un indirizzo: CL è il "palazzo" (banco), SRCAddr è "piano e appartamento"
(registro e carattere).

---

## Parte 2 — Come si legge un opcode

Ogni istruzione ha un **opcode** (Operation Code): uno o due byte che il processore
legge per capire cosa fare. In questo progetto gli opcode sono in esadecimale.

### Istruzioni a byte fisso

Alcune istruzioni hanno un opcode preciso e non variabile:
```
NOP  = 0x00 = 0000 0000   (sempre uguale)
IAC  = 0xF2 = 1111 0010   (sempre uguale)
DAA  = 0xFB = 1111 1011   (sempre uguale)
WRM  = 0xE0 = 1110 0000   (sempre uguale)
```

### Istruzioni "famiglia" — il nibble basso cambia

Molte istruzioni contengono il numero del registro (o un argomento) nel nibble basso:
```
ADD R3 = 0x83 = 1000 0011
                ^^^^       ← codice ADD = 8
                     ^^^^  ← registro R3 = 3

ADD R7 = 0x87 = 1000 0111
                ^^^^       ← codice ADD = 8
                     ^^^^  ← registro R7 = 7
```

L'opcode base è `0x8_` — il `_` cambia da 0 a F a seconda del registro.

### Istruzioni a 2 byte

Alcune istruzioni hanno bisogno di un argomento più grande (un indirizzo, un dato).
Occupano 2 byte consecutivi in ROM:
```
JUN 0x3AB:
  Byte 1:  0x43  =  0100 0011   ← codice JUN + nibble alto dell'indirizzo (3)
  Byte 2:  0xAB  =  1010 1011   ← i due nibble bassi dell'indirizzo (AB)
  
  Indirizzo: 0x3AB → salta lì
```

---

## Parte 3 — Gruppo Registro

Queste istruzioni lavorano sui registri R0–RF e sull'accumulatore A.
L'opcode contiene sempre il numero del registro nel nibble basso.

---

### NOP — Non fare niente (No Operation)

**Opcode:** `0x00` = `0000 0000`

Il 4004 legge questa istruzione, la esegue... e non fa nulla. Solo il PC avanza di 1.

**Quando serve:** riempire spazio in ROM, creare piccoli ritardi (se ripetuta in loop),
segnaposto mentre sviluppi.

```
Esempio:
ROM[0x000] = 0x00   ← NOP (non succede niente)
ROM[0x001] = 0xD5   ← LDM 5 (A = 5)

Esecuzione:
  PC=0x000 → NOP → PC=0x001
  PC=0x001 → LDM 5 → A=5, PC=0x002
```

**Cambia:** solo PC (avanza di 1). **Non cambia:** A, C, R0–RF.

---

### LDM — Carica un numero fisso in A (Load iMmediate)

**Opcode:** `0xDn` dove `n` è il valore (0–F)

Metti direttamente un numero in A. Il numero è scritto nell'opcode stesso — non viene
letto dai registri o dalla RAM. È "immediato" nel senso che è lì, sul posto.

**Formato in binario:**
```
1101 nnnn
^^^^ ^^^^
│    └── il valore da mettere in A (0–15)
└─────── codice LDM
```

**Esempi:**
```
LDM 0  →  0xD0  =  1101 0000  →  A = 0
LDM 5  →  0xD5  =  1101 0101  →  A = 5
LDM 9  →  0xD9  =  1101 1001  →  A = 9
LDM 15 →  0xDF  =  1101 1111  →  A = 15
```

**Esempio con stato prima e dopo:**
```
Stato prima:  A = 12, C = true, R3 = 7

LDM 3  (opcode 0xD3)

Stato dopo:   A = 3,  C = true (invariato), R3 = 7 (invariato)
```

**Cambia:** A = n. **Non cambia:** C, R0–RF.

---

### LD — Copia un registro in A (LoaD)

**Opcode:** `0xAr` dove `r` è il numero del registro (0–F)

Copia il contenuto del registro Rr nell'accumulatore. Il registro di origine
**non viene modificato** — stai solo leggendo, non spostando.

**Formato in binario:**
```
1010 rrrr
^^^^ ^^^^
│    └── numero registro (0–15)
└─────── codice LD
```

**Esempio:**
```
Stato prima:  A = 0, R5 = 9

LD R5  (opcode 0xA5)

Stato dopo:   A = 9, R5 = 9 (invariato)
```

**Differenza con LDM:**
- `LDM 9` mette sempre il numero 9 in A — il 9 è scritto nel programma
- `LD R5` mette in A qualunque cosa contenga R5 — il valore è noto solo a runtime

**Cambia:** A = valore di Rr. **Non cambia:** C, Rr (sorgente), altri registri.

---

### XCH — Scambia A e un registro (eXCHange)

**Opcode:** `0xBr` dove `r` è il numero del registro

Scambia i contenuti di A e Rr. Nessun valore viene perso. È come tenere due carte
in mano e scambiarsele — entrambe rimangono, ma cambiano posizione.

**Formato in binario:**
```
1011 rrrr
```

**Esempio:**
```
Stato prima:  A = 5 (0101), R2 = 9 (1001)

XCH R2  (opcode 0xB2)

Stato dopo:   A = 9 (1001), R2 = 5 (0101)
```

**In binario — lo scambio:**
```
Prima:   A = 0101   R2 = 1001
Dopo:    A = 1001   R2 = 0101
```

**Quando si usa:** salvare temporaneamente A senza perderne il valore.
```
; vuoi fare un calcolo su A ma poi recuperare il valore originale
XCH R0       ← A va in R0, vecchio R0 viene in A
... calcoli ...
XCH R0       ← recupera il valore originale
```

**Cambia:** A ↔ Rr (scambio). **Non cambia:** C.

---

### INC — Aggiungi 1 a un registro (INCrement)

**Opcode:** `0x6r` dove `r` è il numero del registro

Aggiunge 1 al registro specificato. Se il registro vale 15 e incrementa, torna a 0
(wrap). **INC non tocca il carry C** — il carry viene aggiornato solo da ADD, SUB, IAC
e simili.

**Formato in binario:**
```
0110 rrrr
```

**Esempi:**
```
R1 = 3    →  INC R1  →  R1 = 4
R1 = 14   →  INC R1  →  R1 = 15
R1 = 15   →  INC R1  →  R1 = 0    ← wrap a 0, C invariato
```

**Il wrap in binario:**
```
R1 = 1111  (15)
  +  0001  (1)
  ──────────
  1 0000  →  teniamo solo i 4 bit bassi → R1 = 0000 = 0
```

Il quinto bit sparisce. C non si aggiorna. **Questo è il punto chiave** che distingue INC
da IAC: INC opera su un registro e non tocca il carry; IAC opera su A e aggiorna il carry.

**Cambia:** Rr = Rr + 1 (mod 16). **Non cambia:** A, C.

---

### ADD — Somma registro + carry in A (ADD with carry)

**Opcode:** `0x8r` dove `r` è il numero del registro

Somma il valore del registro Rr **e** il carry corrente all'accumulatore A.
Il risultato torna in A; se supera 15, C diventa true e A contiene il resto.

**Formula:** `A = A + Rr + C` (poi tronca a 4 bit, aggiorna C)

**Formato in binario:**
```
1000 rrrr
```

**Esempio 1 — somma semplice (C=false, nessun riporto):**
```
A = 3 (0011), R0 = 4 (0100), C = false (0)

ADD R0:
   0011  (3)
 + 0100  (4)
 + 0000  (carry = false = 0)
 ──────
   0111  (7)   →   A = 7, C = false
```

**Esempio 2 — overflow del nibble:**
```
A = 9 (1001), R0 = 8 (1000), C = false

ADD R0:
   1001  (9)
 + 1000  (8)
 ──────
 1 0001  →  il 5° bit = 1 diventa C = true
             A = 0001 = 1

Risultato: A = 1, C = true
```

**Esempio 3 — il carry entra nel calcolo:**
```
A = 3 (0011), R0 = 4 (0100), C = true (1)

ADD R0:
   0011  (3)
 + 0100  (4)
 + 0001  (carry = true = 1)
 ──────
   1000  (8)   →   A = 8, C = false
```

**Perché il carry entra?** Serve per sommare numeri a più di 4 bit.
Per sommare 47 + 58 in BCD, sommi prima le unità (7+8=15, overflow),
poi le decine (4+5 **più il riporto**). ADD include automaticamente quel riporto.

**Cambia:** A = (A + Rr + C) mod 16, C aggiornato. **Non cambia:** Rr.

---

### SUB — Sottrai registro da A con borrow (SUBtract)

**Opcode:** `0x9r` dove `r` è il numero del registro

Sottrae Rr da A. Il carry/link funziona in modo invertito rispetto ad ADD:
**C=true prima di SUB significa "nessun borrow precedente"** (la situazione normale).
**C=false prima di SUB significa "c'era già un borrow"** (cioè dobbiamo sottrarre 1 in più).

**Formula interna:** `A = A + (~Rr) + C` — il 4004 fa la sottrazione con il complemento a 1.

**Risultato del carry dopo SUB:**
- C = true → nessun borrow (risultato ≥ 0)
- C = false → c'è stato borrow (risultato era negativo, avvolto a 16)

**Formato in binario:**
```
1001 rrrr
```

**Esempio 1 — sottrazione normale:**
```
A = 7, R2 = 3, C = true (nessun borrow)

SUB R2:   7 - 3 = 4

Risultato: A = 4, C = true (nessun borrow)
```

**Esempio 2 — sottrazione con "prestito" (risultato negativo in nibble):**
```
A = 3, R2 = 7, C = true (nessun borrow precedente)

SUB R2:   3 - 7 = -4

In nibble a 4 bit non esistono i negativi: -4 diventa 16 - 4 = 12.
Il 4004 "prende in prestito" dalla cifra superiore.

Risultato: A = 12, C = false (c'è stato borrow)
```

**In binario (esempio 2):**
```
  0011  (3)
- 0111  (7)

Il 4004 usa il complemento: A + (~R2) + C
  0011   (3)
+ 1000   (~7 = 1000, perché ~0111 = 1000)
+ 0001   (C = true = 1)
──────
  1100   (12)   →   A = 12, C = false (borrow)
```

**Esempio 3 — con borrow precedente:**
```
A = 5, R2 = 3, C = false (c'era già un borrow)

SUB R2:   5 - 3 - 1 = 1

Risultato: A = 1, C = true (nessun nuovo borrow)
```

**Cambia:** A, C. **Non cambia:** Rr.

---

### BBL — Ritorna dalla subroutine con un valore (Branch Back and Load)

**Opcode:** `0xCn` dove `n` è il valore da mettere in A al ritorno

È la "chiusura" di ogni subroutine chiamata con JMS. Fa due cose insieme:
1. Estrae l'indirizzo di ritorno dallo stack e ci salta (torna al chiamante)
2. Carica il valore `n` in A (il meccanismo per "restituire un valore" dalla subroutine)

**Formato in binario:**
```
1100 nnnn
^^^^ ^^^^
│    └── valore da mettere in A (0–15) — il "valore di ritorno"
└─────── codice BBL
```

**Esempio completo con JMS:**
```
ROM[0x000] = 0x50   ← JMS (byte 1: codice JMS + pagina 0)
ROM[0x001] = 0x20   ← JMS (byte 2: offset 0x20)
ROM[0x002] = ...    ← qui continuiamo dopo BBL

ROM[0x020] = 0xD7   ← LDM 7  (corpo della subroutine)
ROM[0x021] = 0xC3   ← BBL 3  (ritorna con A = 3)

Esecuzione:
  PC=0x000 → JMS: salva 0x002 nello stack, salta a 0x020
  PC=0x020 → LDM 7: A = 7
  PC=0x021 → BBL 3: A = 3, PC = 0x002 (torna al chiamante)
  PC=0x002 → continua...  (A vale 3 — il valore restituito dalla subroutine)
```

**Pensa a BBL n come a `return n` in Go/C.**

**Cambia:** A = n, PC = indirizzo di ritorno dallo stack, SP--.
**Non cambia:** C.

---

## Parte 4 — Gruppo Accumulatore (0xFX)

Queste istruzioni operano solo su A e/o C. Sono tutte a 1 byte fisso.
Nessuna di loro tocca i registri R0–RF.

---

### CLB — Azzera A e C (CLear Both)

**Opcode:** `0xF0` = `1111 0000`

Azzera sia A che C in una botta sola. Utile per iniziare un calcolo da zero.

```
Prima:  A = 12, C = true
CLB
Dopo:   A = 0,  C = false
```

---

### CLC — Azzera solo il carry (CLear Carry)

**Opcode:** `0xF1` = `1111 0001`

Mette C a false senza toccare A.

```
Prima:  A = 7, C = true
CLC
Dopo:   A = 7, C = false
```

**Quando si usa:** prima di una serie di addizioni, per assicurarsi che non ci sia
un carry "sporco" da calcoli precedenti. Le addizioni BCD iniziano sempre con CLC.

---

### STC — Imposta il carry a 1 (SeT Carry)

**Opcode:** `0xFA` = `1111 1010`

Mette C a true senza toccare A.

```
Prima:  A = 5, C = false
STC
Dopo:   A = 5, C = true
```

**Quando si usa:** prima di SUB o SBM per iniziare una sottrazione senza borrow
precedente. SUB/SBM usano C come "nessun borrow" (C=true) — STC è il reset normale.

---

### CMC — Complementa il carry (CoMplement Carry)

**Opcode:** `0xF3` = `1111 0011`

Inverte il carry: true → false, false → true.

```
Prima:  C = true   →  CMC  →  C = false
Prima:  C = false  →  CMC  →  C = true
```

---

### IAC — Incrementa A (Increment ACcumulator)

**Opcode:** `0xF2` = `1111 0010`

Aggiunge 1 ad A. Se A = 15 → A = 0 e C = true (overflow).
A differenza di INC, **IAC aggiorna il carry**.

**Formula:** `A = A + 1` (tronca a 4 bit, aggiorna C)

```
A = 5,  C = false  →  IAC  →  A = 6,  C = false
A = 9,  C = true   →  IAC  →  A = 10, C = false
A = 15, C = false  →  IAC  →  A = 0,  C = true  ← overflow
```

**In binario (caso overflow):**
```
  1111  (15)
+ 0001  (1)
──────
1 0000  →  A = 0000 = 0,  C = true (il 5° bit)
```

**Differenza fondamentale rispetto a INC:**
| | Operando | Aggiorna C? |
|-|----------|------------|
| IAC | A | Sì |
| INC Rr | registro Rr | No |

---

### DAC — Decrementa A (Decrement ACcumulator)

**Opcode:** `0xF8` = `1111 1000`

Sottrae 1 da A. Se A = 0 → A = 15 e C = false (borrow/underflow).
Come SUB, il carry è invertito: C=true = nessun borrow, C=false = borrow.

```
A = 5  →  DAC  →  A = 4,  C = true   (nessun borrow)
A = 1  →  DAC  →  A = 0,  C = true   (nessun borrow)
A = 0  →  DAC  →  A = 15, C = false  (underflow/borrow)
```

**In binario (caso underflow):**
```
  0000  (0)
- 0001  (1)
Il 4004 calcola: 16 + 0 - 1 = 15
  1111  (15)   →   A = 15, C = false
```

---

### CMA — Complementa A bit per bit (CoMplement Accumulator)

**Opcode:** `0xF4` = `1111 0100`

Inverte tutti e 4 i bit di A (complemento a 1). Non tocca C.

```
A = 0000 (0)   →  CMA  →  A = 1111 (15)
A = 1010 (10)  →  CMA  →  A = 0101 (5)
A = 0101 (5)   →  CMA  →  A = 1010 (10)
A = 1111 (15)  →  CMA  →  A = 0000 (0)
```

**Bit per bit — esempio con A = 0110 (6):**
```
Bit 3: 0 → 1
Bit 2: 1 → 0
Bit 1: 1 → 0
Bit 0: 0 → 1

Risultato: 1001  (9)
```

**Curiosità: CMA + IAC = negazione (complemento a 2)**
```
A = 5 = 0101
CMA        → A = 1010  (complemento a 1)
IAC        → A = 1011  (complemento a 2 = "−5" in aritmetica binaria)
```

---

### RAL — Ruota A a sinistra attraverso C (Rotate A Left)

**Opcode:** `0xF5` = `1111 0101`

Ogni bit scorre di una posizione verso sinistra. Il bit che "esce" a sinistra (il bit 3)
va nel carry. Il vecchio carry entra da destra (posizione del bit 0).

**Schema:**
```
  C  ←  [bit3][bit2][bit1][bit0]  ←  C
  ↑ esce                             ↑ entra
```

**Esempio 1 (rotazione semplice, C=false):**
```
A = 0110  (6),  C = false (0)

RAL:
  bit3 (0) esce e va in C_nuovo
  A si sposta: [bit2][bit1][bit0][C_vecchio] = [1][1][0][0] = 1100 (12)

Risultato: A = 12,  C = false
```

**Esempio 2 (il carry entra):**
```
A = 0110  (6),  C = true (1)

RAL:
  bit3 (0) → C_nuovo = false
  A = [1][1][0][1] = 1101 (13)  ← il vecchio C=1 entra a destra

Risultato: A = 13,  C = false
```

**Esempio 3 (il carry esce):**
```
A = 1010  (10),  C = false (0)

RAL:
  bit3 (1) → C_nuovo = true  ← il bit alto "esce"
  A = [0][1][0][0] = 0100 (4)

Risultato: A = 4,  C = true
```

**Effetto pratico:** RAL moltiplica A per 2 (shift left), con il carry come bit di overflow.
Oppure, ripetuto 4 volte, ruota completamente il nibble passando per C.

---

### RAR — Ruota A a destra attraverso C (Rotate A Right)

**Opcode:** `0xF6` = `1111 0110`

Speculare a RAL: ogni bit scorre verso destra. Il bit 0 va nel carry, il vecchio carry
entra da sinistra (bit 3).

**Schema:**
```
  C  →  [bit3][bit2][bit1][bit0]  →  C
  ↑ entra                             ↑ esce
```

**Esempio:**
```
A = 0110  (6),  C = false (0)

RAR:
  bit0 (0) → C_nuovo = false
  A = [C_vecchio][bit3][bit2][bit1] = [0][0][1][1] = 0011 (3)

Risultato: A = 3,  C = false
```

**Esempio con carry che entra e esce:**
```
A = 0101  (5),  C = true (1)

RAR:
  bit0 (1) → C_nuovo = true  ← il bit basso "esce"
  A = [1][0][1][0] = 1010 (10)  ← il vecchio C=1 entra a sinistra

Risultato: A = 10,  C = true
```

**Effetto pratico:** RAR divide A per 2 (shift right). Utile per estrarre bit specifici.

---

### TCC — Trasferisci carry in A, poi azzeralo (Transfer Carry & Clear)

**Opcode:** `0xF7` = `1111 0111`

Copia il carry in A come numero (0 o 1), poi azzerare il carry.

```
C = true   →  TCC  →  A = 1, C = false
C = false  →  TCC  →  A = 0, C = false
```

**Quando si usa:** in un'addizione multi-nibble, dopo aver sommato le unità con ADD,
il carry viene salvato in A con TCC per poi aggiungerlo alle decine.

```
Esempio: sommare cifra-per-cifra due numeri a più nibble:

; somma le unità
RDM           ← leggi cifra unità dalla RAM
ADD R1        ← aggiungi cifra unità dell'altro numero
DAA           ← correzione BCD
WRM           ← salva il risultato

; il carry ora contiene il riporto alle decine
TCC           ← A = 1 (o 0) — il riporto diventa un numero normale
; ora puoi aggiungerlo alla somma delle decine
```

---

### TCS — Trasferisci carry per la sottrazione BCD (Transfer Carry Subtract)

**Opcode:** `0xF9` = `1111 1001`

Come TCC ma carica 10 (se C=true) o 9 (se C=false) invece di 1/0.
Poi azzera il carry.

```
C = true   →  TCS  →  A = 10, C = false
C = false  →  TCS  →  A = 9,  C = false
```

**Perché 9 e 10?** TCS è usato nella correzione BCD per la sottrazione.
Nella sottrazione BCD: se non c'è borrow (C=true) usi 10 come base di correzione;
se c'è borrow (C=false) usi 9. È il meccanismo opposto di DAA.
Vedi `docs/bcd.md` per i dettagli.

---

### DAA — Correggi A per aritmetica decimale (Decimal Adjust Accumulator)

**Opcode:** `0xFB` = `1111 1011`

Questa è l'istruzione magica della calcolatrice BCD.

**Il problema:** ADD fa la somma in binario. Ma noi vogliamo lavorare in decimale.
In decimale le cifre vanno da 0 a 9 — i valori 10–15 non esistono come cifre singole.

```
8 + 5 = 13 in binario  →  nibble = 1101
Ma in BCD "1101" non è una cifra valida! Ci aspettiamo "3 con riporto 1".
```

**La correzione:** DAA aggiunge 6 ad A se:
- A > 9 (il risultato è fuori dal range BCD)
- **oppure** C = true (c'era overflow)

Perché +6? Perché tra 9 e 16 (il prossimo carry) ci sono esattamente 6 valori non validi
(10, 11, 12, 13, 14, 15). Aggiungendo 6 "salti" quei valori e atterri nel posto giusto.

```
Regola DAA:
  se (A > 9) OPPURE (C = true):
      A = A + 6
      C = true    ← c'è un riporto alla cifra successiva
  altrimenti:
      niente
```

**Esempi:**
```
A = 7,  C = false  →  DAA  →  A = 7   (7 è BCD valido, niente da fare)
A = 12, C = false  →  DAA  →  A = 12+6 = 18 → nibble = 2, C = true  →  A = 2, C = true
A = 1,  C = true   →  DAA  →  A = 1+6 = 7                           →  A = 7, C = true
```

**Esempio reale — somma BCD di 8 + 9:**
```
LDM 8      → A = 8
            (R1 = 9 già caricato da FIM)
ADD R1     → A = 8 + 9 = 17 → nibble: A = 1, C = true (overflow nibble)
DAA        → C = true → A = 1 + 6 = 7, C = true

Risultato: A = 7, C = 1  →  interpretazione BCD: "17"
  Cifra bassa = A = 7
  Cifra alta  = C = 1 (riporto)
```

**In binario passo per passo:**
```
Dopo ADD: A = 0001, C = true

DAA: C = true → aggiungi 6:
   0001  (1)
 + 0110  (6)
 ──────
   0111  (7)   →   A = 7, C = true (invariato perché non c'è overflow in +6)
```

---

### KBP — Decodifica tasto premuto one-hot (KeyBoard Process)

**Opcode:** `0xFC` = `1111 1100`

Converte un valore **one-hot** (un solo bit attivo alla volta) in un numero di posizione.

**One-hot** è una codifica dove ogni tasto ha un solo bit dedicato:
```
0001 = tasto 1 (bit 0 attivo)
0010 = tasto 2 (bit 1 attivo)
0100 = tasto 3 (bit 2 attivo)
1000 = tasto 4 (bit 3 attivo)
```

**KBP converte la posizione del bit in un numero:**
```
A = 0000  →  KBP  →  A = 0   (nessun tasto)
A = 0001  →  KBP  →  A = 1   (bit 0)
A = 0010  →  KBP  →  A = 2   (bit 1)
A = 0100  →  KBP  →  A = 3   (bit 2)
A = 1000  →  KBP  →  A = 4   (bit 3)
A = qualsiasi con >1 bit attivo  →  KBP  →  A = 15  (errore)
```

**Perché esiste?** La tastiera del 4004 era a matrice: RDR legge le colonne come
un valore one-hot (es. `0100` = colonna 3 attiva). KBP converte quel valore nel
numero del tasto premuto, usabile in calcoli o confronti.

```
; workflow lettura tastiera:
LDM 0b0001 / WRR  ← attiva riga 1
RDR               ← A = 0100 (colonna 3 attiva)
KBP               ← A = 3   (numero tasto)
```

Non modifica C.

---

### DCL — Seleziona il banco RAM (Designate Command Line)

**Opcode:** `0xFD` = `1111 1101`

Copia i 3 bit bassi di A nel registro interno CL. Tutte le successive istruzioni
RAM (WRM, RDM, WMP, ecc.) useranno quel banco.

**Formula:** `CL = A & 0b0111` (3 bit → banchi 0–7)

```
A = 0  → DCL  →  CL = 0  (banco 0 attivo)
A = 1  → DCL  →  CL = 1  (banco 1 attivo)
A = 3  → DCL  →  CL = 3  (banco 3 attivo)
```

**A non viene modificato. C non viene modificato.**

**Esempio di uso:**
```
LDM 0    ← voglio usare il banco RAM 0
DCL      ← CL = 0, da ora le istruzioni RAM usano banco 0
FIM R0, 0x05
SRC R0   ← SRCAddr = 0x05 (registro 0, carattere 5 nel banco 0)
LDM 7
WRM      ← scrivi 7 in RAM[banco 0][registro 0][carattere 5]
```

---

## Parte 5 — Gruppo Salti e Indirizzamento

Queste istruzioni controllano il flusso del programma (dove va il PC) e gestiscono
l'indirizzamento di ROM e RAM.

---

### JUN — Salta incondizionatamente (Jump UNconditional)

**Opcode:** `0x4n` + 1 byte (2 byte totali) — salto a indirizzo a 12 bit

Sposta il PC a qualsiasi indirizzo nella ROM. Nessuna condizione, sempre salta.
Come un `goto` diretto.

**Formato:**
```
Byte 1:  0100 pppp   ← codice JUN + 4 bit alti dell'indirizzo (pagina)
Byte 2:  oooo oooo   ← 8 bit bassi dell'indirizzo (offset)

Indirizzo finale = (pppp << 8) | (oooo oooo)
```

**Esempio — salta a 0x3AB:**
```
0x3AB in binario a 12 bit:
  0011 1010 1011
  ^^^^             ← pagina = 0011 = 3
       ^^^^ ^^^^   ← offset = 1010 1011 = 0xAB

Byte 1 = 0x43  (0100 0011: JUN + pagina 3)
Byte 2 = 0xAB
```

| Destinazione | Byte 1 | Byte 2 | Note |
|-------------|--------|--------|------|
| 0x000       | 0x40   | 0x00   | inizio ROM |
| 0x100       | 0x41   | 0x00   | pagina 1 |
| 0x3AB       | 0x43   | 0xAB   | |
| 0xFFF       | 0x4F   | 0xFF   | fine ROM |

**Non cambia:** A, C, registri.

---

### JMS — Salta a subroutine (Jump to Main Subroutine)

**Opcode:** `0x5n` + 1 byte (2 byte totali) — stesso formato di JUN ma con push

Come JUN ma prima di saltare **salva l'indirizzo di ritorno nello stack**.
Il firmware può poi usare BBL per tornare.

**Cosa succede passo per passo:**
```
PC = 0x010  (JMS è all'indirizzo 0x010)
ROM[0x010] = 0x50  ← byte 1 di JMS 0x050
ROM[0x011] = 0x50  ← byte 2

Esecuzione:
1. fetch byte 1 → PC = 0x011
2. fetch byte 2 → PC = 0x012
3. push(0x012) nello stack  ← indirizzo dopo i 2 byte di JMS
4. PC = 0x050               ← salta alla subroutine

Stack = [0x012, ?, ?]   SP = 1
```

**Esempio completo JMS → BBL:**
```
ROM[0x000] = 0x50  ← JMS byte 1 (codice JMS, pagina 0)
ROM[0x001] = 0x20  ← JMS byte 2 (offset 0x20)
ROM[0x002] = NOP   ← prima istruzione al ritorno

ROM[0x020] = LDM 7 ← subroutine: carica 7
ROM[0x021] = BBL 3 ← ritorna con A = 3

Esecuzione:
PC=0x000 → JMS: push(0x002), PC=0x020
PC=0x020 → LDM 7: A = 7
PC=0x021 → BBL 3: A = 3, pop() = 0x002, PC = 0x002
PC=0x002 → NOP  (A = 3, il "valore di ritorno")
```

---

### JCN — Salta se condizione vera (Jump CoNditional)

**Opcode:** `0x1c` + 1 byte (2 byte totali) — salta solo nella pagina corrente

Salta a un offset nella stessa pagina se la condizione `c` è verificata.
Il nibble `c` ha 4 bit, ognuno attiva una condizione diversa.

**Il nibble condizione:**
```
c = C4 C3 C2 C1
     │   │   │  └── bit 0: C1 — salta se TEST pin = 0 (non emulato, sempre falso)
     │   │   └───── bit 1: C2 — salta se carry = 1
     │   └───────── bit 2: C3 — salta se A = 0
     └───────────── bit 3: C4 — inverte tutta la condizione (NOT)
```

C1, C2, C3 si combinano con OR: basta che uno sia vero.
C4 inverte il risultato finale.

**Indirizzo di salto (stessa pagina!):**
```
PC = (PC & 0x0F00) | byte2

Se PC = 0x150 e byte2 = 0x80:
  0x150 & 0x0F00 = 0x100  (pagina 1)
  0x100 | 0x80  = 0x180   (destinazione)
```

**Condizioni più usate:**

| c hex | c binario | Significato | Salta se... |
|-------|-----------|-------------|-------------|
| 0x2   | 0010      | C2          | carry = 1   |
| 0x4   | 0100      | C3          | A = 0       |
| 0x6   | 0110      | C2 OR C3    | carry=1 oppure A=0 |
| 0xA   | 1010      | NOT C2      | carry = 0   |
| 0xC   | 1100      | NOT C3      | A ≠ 0       |

**Esempio 1 — salta se carry è 1:**
```
ROM[0x000] = 0x12   ← JCN con c=2 (carry=1)
ROM[0x001] = 0x50   ← destinazione: offset 0x50 (stessa pagina)

A = 5, C = true

Esecuzione: carry = true → condizione vera → PC = (0x002 & 0x0F00) | 0x50 = 0x050
```

**Esempio 2 — salta se A è zero:**
```
ROM[0x100] = 0x14   ← JCN con c=4 (A=0), siamo in pagina 1
ROM[0x101] = 0x30   ← destinazione: offset 0x30

A = 0

Esecuzione: A = 0 → condizione vera → PC = (0x102 & 0x0F00) | 0x30 = 0x130
```

**Esempio 3 — condizione falsa, nessun salto:**
```
ROM[0x000] = 0x12   ← JCN con c=2 (carry=1)
ROM[0x001] = 0x50

C = false

Esecuzione: carry = false → condizione falsa → PC = 0x002 (continua normalmente)
```

**Costruzione if/else con JCN:**
```
; if A == 0: fai X, else: fai Y

JCN 0x4, ramo_zero   ← salta a ramo_zero se A = 0
; qui: A ≠ 0 → ramo else
... codice Y ...
JUN fine
ramo_zero:
... codice X ...
fine:
```

---

### ISZ — Incrementa registro e salta se non zero (Increment and Skip if Zero)

**Opcode:** `0x7r` + 1 byte (2 byte totali) — salta solo nella pagina corrente

Incrementa il registro Rr di 1. Se il risultato è **diverso da zero**, salta.
Se il risultato è **zero**, non salta (continua all'istruzione successiva).

**La logica sembra strana ma è pensata per i loop:**
- Salta → continua il loop
- Non salta → loop finito, esci

**Formula per N iterazioni:** inizializza il registro a `16 - N`

| Iterazioni N | Valore iniziale | Sequenza dei valori |
|-------------|-----------------|---------------------|
| 1  | 15 = 0xF | 0xF → 0 (stop) |
| 2  | 14 = 0xE | 0xE → 0xF → 0 (stop) |
| 3  | 13 = 0xD | 0xD → 0xE → 0xF → 0 (stop) |
| 8  | 8  = 0x8 | 0x8 → 0x9 → ... → 0xF → 0 (stop) |
| 16 | 0  = 0x0 | 0x0 → 0x1 → ... → 0xF → 0 (stop) |

**Esempio — loop di 3 iterazioni:**
```
ROM[0x000] = FIM R2, 0x0D  → R3 = 0xD = 13
ROM[0x002] = ... corpo del loop ...
ROM[0x005] = ISZ R3, 0x02  ← byte 1: 0x73, byte 2: torna a 0x002

Esecuzione:
  Iter 1: R3 = 0xD → ISZ → R3 = 0xE ≠ 0 → salta a 0x002  (continua)
  Iter 2: R3 = 0xE → ISZ → R3 = 0xF ≠ 0 → salta a 0x002  (continua)
  Iter 3: R3 = 0xF → ISZ → R3 = 0x0 = 0 → NON salta       (esce)
```

**In binario — il momento critico:**
```
R3 = 1111  (15)
  +  0001  (1)
  ─────────
  1 0000  →  4 bit → R3 = 0000 = 0  →  NON salta (loop finito)
```

**ISZ non tocca A né C.**

---

### FIM — Carica 8 bit in una coppia di registri (Fetch IMmediate)

**Opcode:** `0x2r` (r = numero coppia × 2) + 1 byte dati (2 byte totali)

Carica un byte completo (8 bit = 2 nibble) in una coppia di registri adiacenti.
Il nibble alto del byte va nel registro pari, il nibble basso in quello dispari.

**Formato:**
```
Byte 1:  0010 rr00   ← FIM + numero coppia (rr = 0,1,2,...,7)
Byte 2:  HHHH LLLL   ← H = nibble alto → Rr,  L = nibble basso → Rr+1
```

**Coppie disponibili:**

| Opcode byte 1 | Registro pari | Registro dispari |
|---------------|---------------|-----------------|
| 0x20          | R0            | R1              |
| 0x22          | R2            | R3              |
| 0x24          | R4            | R5              |
| 0x26          | R6            | R7              |
| ... | ... | ... |

**Esempio — caricare 0xAB in R0/R1:**
```
ROM[0x000] = 0x20   ← FIM coppia R0/R1
ROM[0x001] = 0xAB   ← dato

In binario: 0xAB = 1010 1011
                   ^^^^       → nibble alto = A = 10  → R0 = 10
                        ^^^^  → nibble basso = B = 11 → R1 = 11

Risultato: R0 = 0xA,  R1 = 0xB
```

**FIM è fondamentale per SRC**: si usa FIM per caricare l'indirizzo RAM (registro+carattere)
nella coppia R0/R1, poi SRC li legge per impostare SRCAddr.

**Non cambia:** A, C.

---

### SRC — Imposta il puntatore RAM (Send Register Control)

**Opcode:** `0x2r+1` — 1 byte, nibble basso dispari

Legge la coppia di registri Rr/Rr+1 e la imposta come indirizzo per le successive
operazioni RAM. Non legge né scrive nessun dato — prepara solo il "puntatore".

**Formula:** `SRCAddr = (Rr << 4) | Rr+1`

**Formato di SRCAddr:**
```
RRRR CCCC
^^^^       ← nibble alto = Rr  = numero registro nel chip RAM (0–3)
     ^^^^  ← nibble basso = Rr+1 = numero carattere nel registro (0–15)
```

**Esempio:**
```
FIM R0, 0x25   → R0 = 2, R1 = 5
SRC R0         → SRCAddr = (2 << 4) | 5 = 0x25
               → registro 2, carattere 5

WRM            → scrive A in RAM[banco][reg 2][char 5]
RDM            → legge da  RAM[banco][reg 2][char 5] in A
```

**In binario:**
```
R0 = 0010  (2)
R1 = 0101  (5)

SRC R0:
  SRCAddr = [0010][0101] = 0010 0101 = 0x25
```

SRC e DCL lavorano sempre insieme:
- `DCL` seleziona il banco (gruppo di chip RAM)
- `SRC` seleziona registro e carattere dentro quel banco

**Non cambia:** A, C, R0–RF.

---

### FIN — Carica dati da ROM usando R0:R1 come indirizzo (Fetch INdirect)

**Opcode:** `0x3r` (r = numero coppia destinazione) — 1 byte

Usa i valori di R0 e R1 come indirizzo (nella pagina corrente) per leggere un byte
dalla ROM, poi carica quel byte nella coppia di registri specificata.

È il meccanismo delle **lookup table**: tabelle di valori costanti nella ROM che il
programma consulta a runtime.

**Calcolo indirizzo:**
```
addr = (pagina corrente del PC) | (R0 << 4) | R1
     = (PC & 0x0F00) | (R0 << 4) | R1
```

**Esempio — tabella dei quadrati:**
```
La tabella è nella ROM a partire da 0x010:
  ROM[0x010] = 0x00  ← 0² = 0,0 (nibble alto=0, basso=0)
  ROM[0x011] = 0x01  ← 1² = 0,1 (nibble alto=0, basso=1)
  ROM[0x012] = 0x04  ← 2² = 0,4
  ROM[0x013] = 0x09  ← 3² = 0,9

Per leggere 3²:
  R0 = 0x1   R1 = 0x3   (indirizzo 0x13 nella pagina corrente)

ROM[0x000] = FIN R2  (opcode 0x32)

Esecuzione (PC = 0x001 dopo fetch, pagina = 0):
  addr = (0x000 & 0x0F00) | (0x1 << 4) | 0x3 = 0x013
  legge ROM[0x013] = 0x09 = 0000 1001
  nibble alto = 0 → R2 = 0
  nibble basso = 9 → R3 = 9

Risultato: R2 = 0, R3 = 9  (cioè 3² = 9)
```

**Nota:** FIN legge dalla **stessa pagina** del PC corrente. R0 e R1 non vengono modificati.

**Non cambia:** A, C, R0, R1.

---

### JIN — Salta all'indirizzo in una coppia di registri (Jump INdirect)

**Opcode:** `0x3r+1` — 1 byte, nibble basso dispari

Come JUN, ma l'indirizzo di destinazione viene dai registri invece di essere
hardcoded. Salta sempre nella pagina corrente (8 bit di offset da Rr/Rr+1).

**Formula:**
```
PC = (pagina corrente) | (Rr << 4) | Rr+1
```

**Esempio:**
```
R0 = 0x7  (0111)
R1 = 0x3  (0011)
PC = 0x200  (pagina 2)

JIN R0

addr = (0x200 & 0x0F00) | (0x7 << 4) | 0x3
     = 0x200 | 0x70 | 0x03
     = 0x273

PC = 0x273
```

**Quando si usa — jump table:**
```
; switch su un valore 0–3:
; in R0/R1 c'è già l'indirizzo del caso giusto (caricato con FIM prima)
JIN R0   ← salta al handler del caso
```

**Differenza con JUN:**
- `JUN` ha l'indirizzo fisso nel sorgente — noto a compile-time
- `JIN` ha l'indirizzo nei registri — può variare a runtime

---

## Parte 6 — Gruppo I/O e RAM (0xEX)

Queste istruzioni leggono e scrivono nella RAM e comunicano con dispositivi esterni.
Sono tutte a 1 byte nella forma `0xEX`.

**Prerequisito:** prima di usare qualsiasi istruzione RAM, il firmware deve avere:
1. Eseguito `DCL` per selezionare il banco (o usare il default banco 0)
2. Eseguito `FIM` + `SRC` per selezionare registro e carattere

---

### Il chip Intel 4002 — la RAM del 4004

Il 4004 non ha RAM interna. Usa chip esterni Intel 4002, ciascuno organizzato così:

```
Chip 4002:
  4 registri, ognuno con:
    ├── 20 nibble di dati     → area WRM/RDM/ADM/SBM
    └──  4 nibble di stato    → area WR0–WR3 / RD0–RD3
  1 porta di output a 4 bit   → WMP
```

Il sistema può avere fino a 4 chip (banchi 0–3).

**Area dati vs area stato:**
- **Dati:** i valori principali del calcolo (cifre BCD del numero)
- **Stato:** flag e metadati (es. il segno del numero, flag di overflow)

---

### WRM — Scrivi A nella cella dati RAM (Write RAM Memory)

**Opcode:** `0xE0`

Scrive il nibble di A nella cella dati indicata da CL+SRCAddr.

```
CL = 0, SRCAddr = 0x05  (registro 0, carattere 5)
A = 7

WRM

Risultato: RAM[banco 0][reg 0][char 5] = 7
           A invariato, C invariato
```

**Non cambia:** A, C.

---

### RDM — Leggi dalla RAM in A (Read RAM Memory)

**Opcode:** `0xE9`

Legge il nibble dalla cella dati indicata da CL+SRCAddr e lo carica in A.

```
RAM[banco 0][reg 0][char 5] = 9
CL = 0, SRCAddr = 0x05
A = 0

RDM

Risultato: A = 9
           C invariato
```

**Non cambia:** C.

---

### ADM — Somma RAM ad A (Add RAM to accumulator with carry)

**Opcode:** `0xEB`

Somma il valore dalla cella RAM e il carry ad A. Esattamente come ADD, ma
l'operando viene dalla RAM invece che da un registro.

**Formula:** `result = A + RAM[b][r][c] + C;  A = nibble(result);  C = result > 15`

```
A = 7, C = false
RAM[0][0][0] = 6

ADM

result = 7 + 6 + 0 = 13  →  A = 13, C = false

DAA  →  13+6 = 19 → nibble = 3, C = true
```

**Quando si usa:** in un'addizione BCD cifra-per-cifra, ADM somma la cifra in RAM
alla cifra in A, poi DAA corregge il risultato.

**Cambia:** A, C.

---

### SBM — Sottrai RAM da A (Subtract RAM from accumulator with borrow)

**Opcode:** `0xE8`

Sottrae il valore della cella RAM da A. Come SUB: C=true = nessun borrow,
C=false = borrow.

**Formula:** `result = A + (~RAM[b][r][c]) + C;  A = nibble(result);  C = result > 15`

```
A = 7, C = true (no borrow)
RAM[0][0][0] = 3

SBM

result = 7 + (~3) + 1 = 7 + 12 + 1 = 20  →  A = nibble(20) = 4, C = true
```

**Cambia:** A, C.

---

### WMP — Scrivi A sulla porta di output del banco RAM (Write Memory Port)

**Opcode:** `0xE1`

Scrive A sulla **porta di output** del banco RAM attivo (CL). La porta è collegata
direttamente a pin hardware del chip: display, LED, relè, ecc.

```
CL = 0  (banco 0)
A = 0b0110  (6)

WMP

Risultato: porta di output del banco 0 = 6
           Pin hardware aggiornati (es. segmenti di un display)
```

**Non cambia:** A, C.

---

### WR0 / WR1 / WR2 / WR3 — Scrivi A nei nibble di stato RAM

**Opcode:** WR0=`0xE4`, WR1=`0xE5`, WR2=`0xE6`, WR3=`0xE7`

Scrivono A nei 4 nibble di stato del registro RAM selezionato.
I nibble di stato sono separati dall'area dati e si usano per flag applicativi.

```
CL = 0, SRCAddr = 0x00  (registro 0)
A = 1  (flag "numero negativo")

WR0

Risultato: RAM.Status[banco 0][reg 0][nibble 0] = 1
```

**Uso tipico — salvare il segno:**
```
; la sottrazione ha generato borrow (C = false → numero negativo)
LDM 1      ← 1 = negativo
WR0        ← salva in status nibble 0

; ... più avanti ...
RD0        ← leggi il flag segno
JCN 0xC, mostra_meno   ← se A ≠ 0, mostra il segno −
```

**Non cambia:** A, C.

---

### RD0 / RD1 / RD2 / RD3 — Leggi nibble di stato in A

**Opcode:** RD0=`0xEC`, RD1=`0xED`, RD2=`0xEE`, RD3=`0xEF`

Leggono i nibble di stato del registro RAM selezionato in A. Opposto di WR0–WR3.

```
RAM.Status[0][0][0] = 1

RD0

Risultato: A = 1
```

**Non cambia:** C.

---

### WRR — Scrivi A sulla porta del chip ROM (Write ROM Register/port)

**Opcode:** `0xE2`

Scrive A sulla porta di output del chip ROM attivo (chip Intel 4001).
Usato per attivare le righe della tastiera durante la scansione, o come output generico.

```
A = 0b0001  (riga 1)

WRR

Risultato: ROM.Port = 1  (riga 1 della tastiera attivata)
```

**Non cambia:** A, C.

---

### RDR — Leggi dalla porta del chip ROM (Read ROM Register/port)

**Opcode:** `0xEA`

Legge la porta di input del chip ROM attivo in A. Usato per leggere le colonne
della tastiera dopo aver attivato una riga con WRR.

```
ROM.Port = 0b0100  (colonna 3 attiva — tasto premuto)

RDR

Risultato: A = 0b0100
```

**Ciclo completo di lettura tastiera:**
```
LDM 0b0001  ← scegli riga 1
WRR         ← attiva riga 1
RDR         ← leggi quali colonne sono attive → A = 0100 (col 3)
KBP         ← decodifica → A = 3 (tasto 3 della riga 1)
```

**Non cambia:** C.

---

### WPM — Scrivi in program memory (Write Program Memory)

**Opcode:** `0xE3`

Scriverebbe A nella ROM del sistema. Sul 4004 hardware originale era usata
per programmare chip PROM 4001. Su ROM fissa (il caso normale) è un **no-op**:
non fa niente.

```
WPM  →  non succede niente (nel nostro emulatore)
```

---

## Parte 7 — Mettere tutto insieme

Ecco come le istruzioni collaborano in scenari reali.

### Scenario 1 — Somma BCD di due cifre singole

```
; Calcola 8 + 9 e salva il risultato in RAM

; Prepara il banco RAM e l'indirizzo
LDM 0       ← A = 0
DCL         ← CL = 0 (banco 0)
FIM R0, 0x00
SRC R0      ← SRCAddr = 0x00 (reg 0, char 0)

; Esegui la somma
LDM 8       ← A = 8 (primo operando)
FIM R2, 0x09
            ← R2=0, R3=9 (secondo operando in R3)
ADD R3      ← A = 8+9 = 17 → nibble: A=1, C=true
DAA         ← A=1+6=7, C=true (correzione BCD)

; Salva il risultato
WRM         ← RAM[0][0][0] = 7 (cifra delle unità)
TCC         ← A = 1 (il carry = cifra delle decine), C = false
FIM R0, 0x01
SRC R0      ← SRCAddr = 0x01 (char 1)
WRM         ← RAM[0][0][1] = 1 (cifra delle decine)

; RAM[0][0][0] = 7,  RAM[0][0][1] = 1  → numero: 17 ✓
```

### Scenario 2 — Loop con contatore

```
; Esegui qualcosa 5 volte

LDM 0 / DCL          ← banco 0
FIM R2, 0x0B         ← R3 = 0xB = 11 = 16-5 (5 iterazioni)

inizio_loop:
  ; ... corpo del loop ...
  FIM R0, 0x05       ← (ricorda: FIM sovrascrive R0/R1, non R2/R3)
  SRC R0
  ... WRM, RDM, ...

  ISZ R3, inizio_loop   ← R3++; se R3≠0 torna all'inizio
; qui: R3 = 0, loop finito
```

### Scenario 3 — Subroutine con valore di ritorno

```
; Subroutine: calcola A + A (doppio)
; Input: A = valore
; Output: A = valore*2, C = carry

subroutine_doppio:
  XCH R0       ← salva A in R0, A ora = R0 (non importa)
  LD R0        ← A = valore originale
  ADD R0       ← A = A + R0 = valore + valore = valore*2
  BBL 0        ← ritorna con A = risultato (BBL 0 carica 0 in A... ma l'ADD ha già A)
```

Aspetta — BBL 0 sovrascrive A con 0! Per restituire il valore calcolato, il firmware
di solito usa i registri R0–RF per passare il risultato:
```
subroutine_doppio:
  XCH R0       ← R0 = input, A = vecchio R0
  LD R0        ← A = input
  ADD R0       ← A = input*2
  XCH R0       ← R0 = risultato (A*2), A = vecchio R0 (buttato via)
  BBL 0        ← ritorna. il chiamante legge R0 per il risultato
```

---

## Parte 8 — Tabella di riferimento rapida

### Registro (0x00, 0x6X–0xDX)

| Istruzione | Opcode | Cambia | Non cambia |
|------------|--------|--------|------------|
| NOP        | 0x00   | —      | tutto |
| LDM n      | 0xDn   | A=n    | C |
| LD Rr      | 0xAr   | A=Rr   | C, Rr |
| XCH Rr     | 0xBr   | A↔Rr   | C |
| INC Rr     | 0x6r   | Rr+1   | A, C |
| ADD Rr     | 0x8r   | A=(A+Rr+C)%16, C | Rr |
| SUB Rr     | 0x9r   | A=(A+~Rr+C)%16, C | Rr |
| BBL n      | 0xCn   | A=n, PC←stack, SP-- | C |

### Accumulatore (0xFX)

| Istruzione | Opcode | Cambia | Non cambia |
|------------|--------|--------|------------|
| CLB        | 0xF0   | A=0, C=false | — |
| CLC        | 0xF1   | C=false | A |
| IAC        | 0xF2   | A++, C | — |
| CMC        | 0xF3   | C=!C | A |
| CMA        | 0xF4   | A=~A | C |
| RAL        | 0xF5   | A ruota sinistra, C | — |
| RAR        | 0xF6   | A ruota destra, C | — |
| TCC        | 0xF7   | A=C (0/1), C=false | — |
| DAC        | 0xF8   | A--, C | — |
| TCS        | 0xF9   | A=9 o 10, C=false | — |
| STC        | 0xFA   | C=true | A |
| DAA        | 0xFB   | A+6 se >9 o C, C | — |
| KBP        | 0xFC   | A=posizione bit | C |
| DCL        | 0xFD   | CL=A&7 | A, C |

### Salti e indirizzamento

| Istruzione | Byte | Opcode | Salta... |
|------------|------|--------|---------|
| JUN a      | 2    | 0x4n   | sempre, ovunque (12 bit) |
| JMS a      | 2    | 0x5n   | sempre, ovunque (push PC) |
| JCN c,a    | 2    | 0x1c   | se condizione c, stessa pagina |
| ISZ Rr,a   | 2    | 0x7r   | se Rr+1≠0, stessa pagina |
| FIM Rr,d   | 2    | 0x2r   | non salta (carica byte in Rr/Rr+1) |
| SRC Rr     | 1    | 0x2r+1 | non salta (imposta SRCAddr) |
| FIN Rr     | 1    | 0x3r   | non salta (legge ROM[R0:R1] in Rr) |
| JIN Rr     | 1    | 0x3r+1 | sempre, stessa pagina (da Rr/Rr+1) |

### I/O e RAM (0xEX)

| Istruzione | Opcode | Operazione |
|------------|--------|-----------|
| WRM        | 0xE0   | RAM.Data[b][r][c] = A |
| WMP        | 0xE1   | RAM.Port[b] = A |
| WRR        | 0xE2   | ROM.Port = A |
| WPM        | 0xE3   | no-op (ROM fissa) |
| WR0–WR3    | 0xE4–E7| RAM.Status[b][r][0..3] = A |
| SBM        | 0xE8   | A = A−RAM−borrow, aggiorna C |
| RDM        | 0xE9   | A = RAM.Data[b][r][c] |
| RDR        | 0xEA   | A = ROM.Port |
| ADM        | 0xEB   | A = A+RAM+C, aggiorna C |
| RD0–RD3    | 0xEC–EF| A = RAM.Status[b][r][0..3] |

---

## Appendice — Regole d'oro da ricordare

1. **A è l'unico registro dell'ALU.** Tutto ciò che vuoi calcolare deve passare per A.

2. **Il carry ha due personalità.** In ADD/IAC/RAL: C=true = overflow (riporto normale).
   In SUB/DAC/SBM/RAR: C=true = nessun borrow (convenzione invertita).

3. **INC non tocca C, IAC sì.** INC è per contatori in registri. IAC è per incrementare A.

4. **DAA va sempre dopo ADD/ADM** nelle somme BCD. Non ha senso usarla da sola.

5. **Prima di ogni accesso RAM: DCL + SRC.** Senza questi, scrivi/leggi nel posto sbagliato.

6. **JCN, ISZ, JIN saltano solo nella stessa pagina.** Per saltare a pagine diverse usa JUN/JMS.

7. **Lo stack è solo 3 livelli.** Non annidare più di 3 subroutine senza ritornare.

8. **BBL sovrascrive A.** Se hai un risultato in A e chiami BBL n con n≠0, lo perdi.
   Salva il risultato in un registro prima di BBL.

9. **FIM usa sempre R(pari)/R(dispari+1).** Non puoi caricare in R1/R2 — solo in coppie allineate.

10. **SRC non scrive in RAM.** Imposta solo il puntatore per le istruzioni successive.

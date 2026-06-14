# Esempio: Addizione BCD multi-cifra (47 + 58 = 105)

Il salto dalla singola cifra al **numero**: la somma diventa un loop che lavora
cifra per cifra, dalle unità verso le decine, con il riporto che si propaga
automaticamente da una cifra alla successiva. È il pattern alla base di ogni
calcolatrice.

---

## Come stanno i numeri in RAM

Ogni numero è un array di cifre BCD, una per cella, **little-endian per cifra**
(unità nel char 0). Così il loop avanza con un semplice `INC` da char 0 in su.

| Registro RAM | char 0 (unità) | char 1 (decine) | char 2 (centinaia) |
|--------------|:--------------:|:---------------:|:------------------:|
| reg 0 → A = 47 | 7 | 4 | — |
| reg 1 → B = 58 | 8 | 5 | — |
| reg 2 → risultato | 5 | 0 | 1 |

I due operandi vivono in **registri RAM diversi** (0 e 1) per non sovrapporsi;
il risultato nel registro 2.

---

## I tre puntatori

`SRC` indirizza con una coppia di registri: nibble alto = registro RAM,
nibble basso = char. Ne servono tre, tutte sullo stesso char a ogni giro:

- `R0:R1` → A : `R0=0`, `R1` = char corrente
- `R2:R3` → B : `R2=1`, `R3` = char corrente
- `R4:R5` → risultato : `R4=2`, `R5` = char corrente
- `R6` → contatore loop (2 cifre → `16-2 = 14 = 0xE`, con `ISZ`)

`R1`, `R3`, `R5` partono a 0 e vengono incrementati insieme nel loop, così i
tre puntatori avanzano in colonna.

---

## Algoritmo

```
setup:
  RAM reg0 = [7, 4]   (A = 47)
  RAM reg1 = [8, 5]   (B = 58)
  puntatori R1=R3=R5=0, contatore R6=14
  CLC                 (nessun riporto entra nelle unità)

LOOP (2 volte):
  A = RAM[reg0][char]            (RDM)
  A = A + RAM[reg1][char] + C    (ADM — somma anche il riporto)
  DAA                            (correzione BCD, rialza C se trabocca)
  RAM[reg2][char] = A            (WRM)
  char++  (INC R1, R3, R5)
  ISZ R6 → LOOP

fine:
  A = C        (TCC: l'ultimo riporto diventa cifra)
  RAM[reg2][char2] = A   (centinaia)
```

---

## L'intuizione centrale: il riporto si propaga da solo

Nel loop **solo `ADM` e `DAA` toccano il carry**. `RDM`, `WRM`, `SRC`, `INC`,
`ISZ` lo lasciano intatto. Quindi il `C=1` prodotto dal `DAA` di una cifra
sopravvive a tutto ciò che viene dopo e arriva intatto all'`ADM` della cifra
successiva, che lo somma. È la propagazione del riporto in colonna — senza
scriverla a mano, viaggia nel flag.

Il trace lo mostra cifra per cifra:

```
unità:  ADM A=F (7+8=15) → DAA A=5, C=true     ← riporto generato
decine: ADM A=A (4+5+1=10, somma il C) → DAA A=0, C=true
fine:   TCC A=1 → centinaia                     ← l'ultimo riporto diventa cifra
```

`CLC` è l'unico azzeramento esplicito del carry di tutto il programma: serve
solo perché nessun riporto entri nelle unità. Da lì in poi il flag fa tutto.

---

## TCC: il riporto che esce diventa una cifra

Dopo l'ultima cifra il carry può essere ancora 1 (qui lo è: 4+5+1=10). Quel
riporto è la cifra delle centinaia, ma è "fuori" dall'accumulatore. `TCC`
(Transfer Carry and Clear) lo travasa in `A` (`A=1` se `C` era vero, `A=0`
altrimenti) e azzera `C`, così possiamo scriverlo in RAM come una cifra vera.

Nota: a fine loop `R5` vale già 2 (i due `INC R5`), quindi `SRC R4` punta
automaticamente al char 2 del risultato — la posizione delle centinaia.

---

## Layout ROM

```
       ── SETUP ──
0x000  LDM 0 / DCL
0x002  A=47 → reg0: char0=7, char1=4   (FIM/SRC/LDM/WRM ×2)
0x00C  B=58 → reg1: char0=8, char1=5   (FIM/SRC/LDM/WRM ×2)
       ── INIT LOOP ──
0x016  FIM R0,0x00   puntatore A → char 0
0x018  FIM R2,0x10   puntatore B → char 0
0x01A  FIM R4,0x20   puntatore risultato → char 0
0x01C  FIM R6,0xE0   contatore = 14 (16-2)
0x01E  CLC
       ── LOOP (0x01F) ──
0x01F  SRC R0 / RDM            A = cifra di A
0x021  SRC R2 / ADM            A += cifra di B + riporto
0x023  DAA                     correzione BCD
0x024  SRC R4 / WRM            scrivi cifra del risultato
0x026  INC R1 / INC R3 / INC R5   avanza i tre puntatori
0x029  ISZ R6, 0x01F           ripeti per 2 cifre
       ── RIPORTO FINALE ──
0x02B  TCC                     A = ultimo riporto
0x02C  SRC R4 / WRM            scrivi le centinaia in char 2
       ── HALT ──
0x02E  JUN 0x02E
```

---

## Risultato atteso

```
Risultato in RAM (reg 2): centinaia=1 decine=0 unità=5
47 + 58 = 105
✓ Corretto!
```

---

## Come eseguire

```
go run ./examples/somma-multicifra
```

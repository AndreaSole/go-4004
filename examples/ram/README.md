# Esempio: Riempire un array in RAM

Scrive i valori `1, 2, 3, 4` in quattro celle consecutive della RAM,
aggiornando dinamicamente l'indirizzo dentro un loop. È il pattern base
per qualsiasi manipolazione di sequenze/array sul 4004.

---

## Algoritmo

```
setup:
  CL = 0
  R2:R3 = 0x00     (indirizzo di partenza: registro 0, posizione 0)
  R0:R1 = 0x01     (R1 = primo valore da scrivere = 1)
  R4 = 12          (contatore loop = 16-4, caricato con FIM R4, 0xC0)

LOOP:
  SRC R2           → invia l'indirizzo corrente (R2:R3) alla RAM
  LD R1            → A = valore da scrivere
  WRM              → scrivi A nella cella puntata da SRCAddr
  INC R1           → valore++   (prossimo numero da scrivere)
  INC R3           → posizione++ (prossima cella)
  ISZ R4, LOOP     → ripeti finché R4 non torna a 0

HALT:
  JUN HALT
```

---

## Layout ROM

```
0x000  LDM 0            A = 0
0x001  DCL              CL = 0
0x002  FIM R2, 0x00     R2=0, R3=0   (indirizzo iniziale)
0x004  FIM R0, 0x01     R0=0, R1=1   (primo valore)
0x006  FIM R4, 0xC0     R4=12, R5=0  (contatore — un FIM solo!)
       ── LOOP (0x008) ──
0x008  SRC R2           SRCAddr = (R2<<4)|R3
0x009  LD R1            A = R1
0x00A  WRM              RAM[CL][R2][R3] = A
0x00B  INC R1           R1++  (prossimo valore)
0x00C  INC R3           R3++  (prossima posizione)
0x00D  ISZ R4, 0x008    R4++; se !=0 → torna all'inizio del loop
0x00E   └── 0x08
       ── HALT (0x00F) ──
0x00F  JUN 0x00F
0x010   └── 0x0F
```

---

## Il trucco: FIM può caricare un contatore in un colpo solo

Negli esempi precedenti il contatore del loop veniva preparato con due
istruzioni:

```
LDM 12   → A = 12
XCH R4   → R4 = 12, A = 0
```

Qui invece basta una `FIM`, che carica entrambi i registri della coppia
con un solo byte immediato:

```
FIM R4, 0xC0   → R4 = nibble alto di 0xC0 = 0xC = 12
                 R5 = nibble basso di 0xC0 = 0x0 = 0
```

Un'istruzione, un solo step, stesso risultato (più il bonus che azzera anche
R5). Vale la pena usarlo ogni volta che il valore desiderato si "incastra"
bene nei due nibble di un byte immediato.

---

## ⚠️ Punto chiave: SRC va richiamato ogni volta che cambia l'indirizzo

`SRC` non è un puntatore "vivo" che si aggiorna da solo — è un **segnale**
che la CPU manda al chip RAM una tantum: "d'ora in poi lavora su questa
cella, finché non ti dico diversamente". Cambiare `R3` (la posizione) senza
richiamare `SRC` non sposta la cella attiva nella RAM — bisogna rimandare
l'indirizzo aggiornato esplicitamente.

Per questo nel loop l'ordine è cruciale: **prima `SRC R2`** (aggiorna
l'indirizzo nel chip RAM), **poi `WRM`** (scrive nella cella appena
selezionata). Il trace lo conferma — l'indirizzo `SRC=` cambia ad ogni
iterazione subito prima della scrittura:

```
PC=008 OP=23 SRC R2     SRC=00   → poi WRM scrive in [0][0][0]
PC=008 OP=23 SRC R2     SRC=01   → poi WRM scrive in [0][0][1]
PC=008 OP=23 SRC R2     SRC=02   → poi WRM scrive in [0][0][2]
PC=008 OP=23 SRC R2     SRC=03   → poi WRM scrive in [0][0][3]
```

---

## Risultato atteso

```
RAM[0][0][0] = 1
RAM[0][0][1] = 2
RAM[0][0][2] = 3
RAM[0][0][3] = 4
```

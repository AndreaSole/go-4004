# Esempio: Subroutine con JMS/BBL

Calcola `3 + 5` chiamando una subroutine, e introduce due meccanismi
fondamentali per scrivere programmi reali sul 4004: le chiamate a funzione
e la convenzione di "fine programma".

---

## Algoritmo

```
MAIN:
  setup RAM (banco 0, indirizzo 0x00) e operandi (R0=3, R1=5)
  JMS SOMMA        → chiama la subroutine
  LD R5            → recupera il risultato salvato dalla subroutine
  WRM              → scrive il risultato in RAM
HALT:
  JUN HALT         → loop infinito su se stesso (fine programma)

SOMMA:
  A = R0 + R1      → 3 + 5 = 8
  XCH R5           → salva il risultato in R5 (sopravvive al ritorno)
  BBL 0            → torna al chiamante (A viene sovrascritto con 0)
```

---

## Layout ROM

```
0x000  LDM 0            A = 0
0x001  DCL              CL = 0 (banco RAM 0)
0x002  FIM R2, 0x00     R2=0, R3=0
0x003   └── 0x00
0x004  SRC R2           SRCAddr = 0x00
0x005  FIM R0, 0x35     R0=3, R1=5  (operandi)
0x006   └── 0x35
0x007  JMS 0x00D        push PC=0x009, salta a SOMMA
0x008   └── 0x0D
0x009  LD R5            A = R5  (recupera il risultato)
0x00A  WRM              ram.Data[0][0][0] = A
       ── HALT ──
0x00B  JUN 0x00B        salto infinito su se stesso
0x00C   └── 0x0B
       ── SUBROUTINE SOMMA (0x00D) ──
0x00D  LD R0            A = R0 (3)
0x00E  ADD R1           A = 3 + 5 = 8
0x00F  XCH R5           R5 = 8  (salva il risultato PRIMA di tornare!)
0x010  BBL 0            pop stack → torna a 0x009, A = 0
```

---

## ⚠️ La trappola di BBL: non restituisce A, lo sovrascrive

`BBL n` fa due cose contemporaneamente: estrae l'indirizzo di ritorno
dallo stack **e** carica il nibble immediato `n` in A.

Questo significa che qualunque cosa la subroutine abbia calcolato in A
viene **cancellata** al momento del ritorno. Se la subroutine facesse solo:

```
LD R0
ADD R1     → A = 8
BBL 0      → A diventa 0! Il risultato 8 è perso.
```

il chiamante non vedrebbe mai il risultato. Per questo la subroutine fa
`XCH R5` prima di `BBL`: sposta il risultato in un registro che sopravvive
al ritorno, e il chiamante lo recupera con `LD R5`.

Questo riflette l'uso reale di BBL sul 4004: serve per restituire **codici
di stato** (es. "operazione riuscita = 0", "errore = 1"), non valori
calcolati. I valori calcolati vanno sempre salvati in un registro o in RAM
prima del ritorno.

Il trace lo mostra chiaramente:

```
PC=00F OP=B5 XCH R5     A=0   ← R5=8 salvato, A azzerato (swap)
PC=010 OP=C0 BBL 0      A=0   ← torna, A resta a 0 (immediato di BBL)
PC=009 OP=A5 LD  R5     A=8   ← il chiamante recupera il risultato da R5
```

---

## La convenzione "halt": JUN a se stesso

Il 4004 non ha un'istruzione HALT. Senza un punto di arresto esplicito,
`Step()` continuerebbe a eseguire i byte successivi nella ROM (che sono
zeri = NOP) all'infinito.

La convenzione standard è terminare il programma con un salto su se stesso:

```
HALT:
  JUN HALT
```

Una volta che il PC raggiunge questo indirizzo, ci resta per sempre — è un
loop infinito intenzionale. Il programma host (il nostro `main.go`) rileva
questa condizione confrontando `c.PC` con l'indirizzo di halt e interrompe
il ciclo:

```go
const haltAddr = 0x00B

for {
    if err := c.Step(rom, ram); err != nil {
        fmt.Printf("Errore: %v\n", err)
        return
    }
    if c.PC == haltAddr {
        break
    }
}
```

A differenza del primo esempio (dove contavamo gli step a mano), qui il
programma host non ha bisogno di sapere in anticipo quanti step servono:
gira finché il programma stesso segnala di aver finito.

---

## Risultato atteso

```
RAM[0][0][0] = 8
R5           = 8
A            = 8   (LD R5 sovrascrive il valore azzerato da BBL)
```

package cpu

// RAM emula i chip di memoria esterni Intel 4002 collegati al 4004.
//
// Il 4004 non ha RAM interna: accede ai dati tramite chip 4002 separati.
// Ogni chip 4002 contiene:
//   - 4 registri × 16 nibble di dati (carattere = SRCAddr & 0x0F, range 0-15)
//   - 4 nibble di stato per registro, modellati separatamente (usati per flag applicativi)
//   - 1 porta di output da 4 bit (per display, buzzer, ecc.)
//
// Il sistema supporta fino a 4 chip, selezionati da DCL (banco) + SRC (registro/carattere).
//
// Indirizzamento usato dall'emulatore:
//   - banco     = CL & 0x3           (impostato da DCL)
//   - registro  = (SRCAddr >> 4) & 0x3  (nibble alto di SRCAddr, impostato da SRC)
//   - carattere = SRCAddr & 0x0F      (nibble basso di SRCAddr, impostato da SRC)
type RAM struct {
	Data   [4][4][16]uint8 // [banco][registro][carattere] — nibble dati
	Status [4][4][4]uint8  // [banco][registro][status]    — nibble di stato
	Port   [4]uint8        // porta di output per banco    — scritta da WMP
}

// NewRAM crea una RAM virtuale inizializzata a zero.
func NewRAM() *RAM {
	return &RAM{}
}

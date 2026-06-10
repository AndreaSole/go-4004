package cpu

// RAM emula i chip di memoria esterni Intel 4002 collegati al 4004.
//
// Il 4004 non ha RAM interna: accede ai dati tramite chip 4002 separati.
// Ogni chip 4002 contiene:
//   - 4 registri × 16 nibble di dati
//   - 4 nibble di stato per registro (usati per flag applicativi)
//   - 1 porta di output da 4 bit (per display, buzzer, ecc.)
//
// Semplificazione dell'emulatore: un chip 4002 per banco, 4 banchi.
// (Sul 4004 reale ogni banco può avere fino a 4 chip, selezionati dai
// bit 7-6 di SRC, e DCL può indirizzare fino a 8 banchi — qui i bit
// chip-select sono ignorati e i banchi sono 4.)
//
// Indirizzamento usato dall'emulatore:
//   - banco     = CL & 0x3              (impostato da DCL)
//   - registro  = (SRCAddr >> 4) & 0x3  (bit 5-4 di SRCAddr, impostato da SRC)
//   - carattere = SRCAddr & 0x0F        (nibble basso di SRCAddr, impostato da SRC)
type RAM struct {
	Data   [4][4][16]uint8 // [banco][registro][carattere] — nibble dati
	Status [4][4][4]uint8  // [banco][registro][status]    — nibble di stato
	Port   [4]uint8        // porta di output per banco    — scritta da WMP
}

// NewRAM crea una RAM virtuale inizializzata a zero.
func NewRAM() *RAM {
	return &RAM{}
}

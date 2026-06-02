package cpu

// ROM rappresenta la memoria di programma del sistema Intel 4004.
// Contiene le istruzioni che la CPU legge ed esegue sequenzialmente.
type ROM struct {
	Data []byte
}

// NewROM crea una ROM a partire da uno slice di byte (il programma).
func NewROM(data []byte) *ROM {
	return &ROM{Data: data}
}

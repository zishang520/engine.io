package engineio

type Packet struct {
	Type    string
	Data    string
	Options interface{}
}

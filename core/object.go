package core

var OBJ_TYPE_STRING uint8 = 0 << 4

var OBJ_ENCODING_RAW uint8 = 0
var OBJ_ENCODING_INT uint8 = 1
var OBJ_ENCODING_EMBSTR uint8 = 8

type Obj struct {
	TypeEncoding uint8
	Value        interface{}
	ValidTill    int
}

func NewObj(value interface{}, validTill int, oType uint8, oEncoding uint8) *Obj {

	return &Obj{
		Value:        value,
		ValidTill:    validTill,
		TypeEncoding: oType | oEncoding,
	}
}

func assertEncoding(oTypeEncoding uint8, expected uint8) bool {

	oEnc := (oTypeEncoding << 4) >> 4

	return oEnc == expected
}

func assertType(oTypeEncoding uint8, expected uint8) bool {
	oType := oTypeEncoding >> 4
	return oType == expected
}

package core

import "time"

var OBJ_TYPE_STRING uint8 = 0 << 4

var OBJ_ENCODING_RAW uint8 = 0
var OBJ_ENCODING_INT uint8 = 1
var OBJ_ENCODING_EMBSTR uint8 = 8

type Obj struct {
	TypeEncoding   uint8
	Value          interface{}
	ValidTill      int
	LastAccessedAt uint32
}

func NewObj(value interface{}, validTill int, oType uint8, oEncoding uint8) *Obj {

	return &Obj{
		Value:          value,
		ValidTill:      validTill,
		TypeEncoding:   oType | oEncoding,
		LastAccessedAt: uint32(time.Now().Unix()) & 0x00FFFFFF,
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

package proto

import "encoding/binary"

func AppendUint32(data []byte, v uint32) []byte {
	var buf [4]byte
	binary.BigEndian.PutUint32(buf[:], v)
	return append(data, buf[:]...)
}

func AppendUint16(data []byte, v uint16) []byte {
	var buf [2]byte
	binary.BigEndian.PutUint16(buf[:], v)
	return append(data, buf[:]...)
}

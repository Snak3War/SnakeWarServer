package proto

import (
	"encoding/binary"
	"errors"
)

var (
	ErrDataNotEnough = errors.New("data not enough")
)

const Version = 1

type Packet struct {
	Version byte
	Type    byte
	Count   byte
}

func (p Packet) Size() int {
	return 2
}

func (p Packet) MarshalBinary() (data []byte, err error) {
	data = append(data, (p.Version&0b111111)<<2|(p.Type&0b11), p.Count)
	return
}

func (p *Packet) UnmarshalBinary(data []byte) error {
	if len(data) < p.Size() {
		return ErrDataNotEnough
	}
	p.Version = data[0] >> 2
	p.Type = data[0] & 0b11
	p.Count = data[1]
	return nil
}

const (
	PacketTypeHandshake = iota
	PacketTypeEvent
)

type Handshake struct {
	Seed  uint16
	ID    byte
	Count byte
}

func (p Handshake) Size() int {
	return 4
}

func (p Handshake) MarshalBinary() (data []byte, err error) {
	data = AppendUint16(data, p.Seed)
	data = append(data, p.ID, p.Count)
	return
}

func (p *Handshake) UnmarshalBinary(data []byte) error {
	if len(data) < p.Size() {
		return ErrDataNotEnough
	}
	p.Seed = binary.ByteOrder.Uint16(binary.BigEndian, data)
	p.ID = data[2]
	p.Count = data[3]
	return nil
}

type Event struct {
	Tick     uint32
	Type     byte
	DataSize byte
}

func (p Event) Size() int {
	return 6
}

func (p Event) MarshalBinary() (data []byte, err error) {
	data = AppendUint32(data, p.Tick)
	data = append(data, p.Type, p.DataSize)
	return
}

func (p *Event) UnmarshalBinary(data []byte) error {
	if len(data) < p.Size() {
		return ErrDataNotEnough
	}
	p.Tick = binary.BigEndian.Uint32(data)
	p.Type = data[4]
	p.DataSize = data[5]
	return nil
}

const (
	EventTypeHeartbeat = iota
	EventTypeStart
	EventTypePause
	EventTypeMove
	EventTypeOver
)

type EventMove struct {
	ID        byte
	Direction byte
}

func (p EventMove) Size() int {
	return 2
}

func (p EventMove) MarshalBinary() (data []byte, err error) {
	data = append(data, p.ID, p.Direction)
	return
}

func (p *EventMove) UnmarshalBinary(data []byte) error {
	if len(data) < p.Size() {
		return ErrDataNotEnough
	}
	p.ID = data[0]
	p.Direction = data[1]
	return nil
}

type EventOver struct {
	ID byte
}

func (p EventOver) Size() int {
	return 1
}

func (p EventOver) MarshalBinary() (data []byte, err error) {
	data = append(data, p.ID)
	return
}

func (p *EventOver) UnmarshalBinary(data []byte) error {
	if len(data) < p.Size() {
		return ErrDataNotEnough
	}
	p.ID = data[0]
	return nil
}

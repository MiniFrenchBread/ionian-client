package kv

import (
	"crypto/sha256"
	"encoding/binary"

	"github.com/ethereum/go-ethereum/common"
)

// df2ff3bb0af36c6384e6206552a4ed807f6f6a26e7d0aa6bff772ddc9d4307aa
var StreamDomain = common.Hash(sha256.Sum256([]byte("STREAM")))

func CreateTags(streamIds ...common.Hash) []byte {
	result := make([]byte, 0, common.HashLength*(1+len(streamIds)))

	result = append(result, StreamDomain.Bytes()...)

	for _, v := range streamIds {
		result = append(result, v.Bytes()...)
	}

	return result
}

type AccessControlType uint8

const (
	// Admin role
	AclTypeGrantAdminRole    AccessControlType = 0x00
	AclTypeRenounceAdminRole AccessControlType = 0x01

	// set/unset special key
	AclTypeSetKeyToSpecial AccessControlType = 0x10
	AclTypeSetKeyToNormal  AccessControlType = 0x11

	// Write role for all keys
	AclTypeGrantWriteRole    AccessControlType = 0x20
	AclTypeRevokeWriteRole   AccessControlType = 0x21
	AclTypeRenounceWriteRole AccessControlType = 0x22

	// Write role for special key
	AclTypeGrantSpecialWriteRole    AccessControlType = 0x30
	AclTypeRevokeSpecialWriteRole   AccessControlType = 0x31
	AclTypeRenounceSpecialWriteRole AccessControlType = 0x32
)

type StreamRead struct {
	StreamId common.Hash
	Key      common.Hash
}

type StreamWrite struct {
	StreamId common.Hash
	Key      common.Hash
	Data     []byte
}

type AccessControl struct {
	Type     AccessControlType
	StreamId common.Hash
	Account  *common.Address
	Key      *common.Hash
}

type StreamData struct {
	Version  uint64
	Reads    []StreamRead
	Writes   []StreamWrite
	Controls []AccessControl
}

// Size returns the serialized data size in bytes.
func (sd *StreamData) Size() int {
	var size int

	size += 8                                     // version
	size += 4 + 2*common.HashLength*len(sd.Reads) // reads

	// writes
	size += 4 // size
	for _, v := range sd.Writes {
		size += 2*common.HashLength + 8 + len(v.Data)
	}

	// acls
	size += 4 // size
	for _, v := range sd.Controls {
		size += 1 + common.HashLength // type + streamId

		if v.Account != nil {
			size += common.AddressLength
		}

		if v.Key != nil {
			size += common.HashLength
		}
	}

	return size
}

func (sd *StreamData) encodeSize32(size int) []byte {
	var buf [4]byte
	binary.BigEndian.PutUint32(buf[:], uint32(size))
	return buf[:]
}

func (sd *StreamData) encodeSize64(size int) []byte {
	var buf [8]byte
	binary.BigEndian.PutUint64(buf[:], uint64(size))
	return buf[:]
}

func (sd *StreamData) Encode() []byte {
	// pre-allocate memory and init length for version
	encoded := make([]byte, 8, sd.Size())

	// version
	binary.BigEndian.PutUint64(encoded[:8], sd.Version)

	// reads
	encoded = append(encoded, sd.encodeSize32(len(sd.Reads))...)
	for _, v := range sd.Reads {
		encoded = append(encoded, v.StreamId.Bytes()...)
		encoded = append(encoded, v.Key.Bytes()...)
	}

	// writes
	encoded = append(encoded, sd.encodeSize32(len(sd.Writes))...)
	for _, v := range sd.Writes {
		encoded = append(encoded, v.StreamId.Bytes()...)
		encoded = append(encoded, v.Key.Bytes()...)
		encoded = append(encoded, sd.encodeSize64(len(v.Data))...)
	}

	for _, v := range sd.Writes {
		encoded = append(encoded, v.Data...)
	}

	// acls
	encoded = append(encoded, sd.encodeSize32(len(sd.Controls))...)
	for _, v := range sd.Controls {
		encoded = append(encoded, byte(v.Type))
		encoded = append(encoded, v.StreamId.Bytes()...)

		if v.Key != nil {
			encoded = append(encoded, v.Key.Bytes()...)
		}

		if v.Account != nil {
			encoded = append(encoded, v.Account.Bytes()...)
		}
	}

	return encoded
}
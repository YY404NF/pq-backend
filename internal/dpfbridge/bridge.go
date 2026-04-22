package dpfbridge

/*
#cgo CXXFLAGS: -std=c++20 -I${SRCDIR}/../../../pq-dpf-core/include
#cgo CPPFLAGS: -I${SRCDIR}/../../../pq-dpf-core/include
#cgo darwin LDFLAGS: -lc++
#include <stdlib.h>
#include "pq_dpf_core/c_api.h"
*/
import "C"

import (
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"runtime"
	"unsafe"
)

type Block128 struct {
	X uint32
	Y uint32
	Z uint32
	W uint32
}

type CorrectionWord struct {
	S  Block128
	Tr bool
}

type KeyShare struct {
	Seed            Block128
	CorrectionWords []CorrectionWord
	InBits          uint32
	DomainSize      uint64
}

type GeneratedKeys struct {
	Left  KeyShare
	Right KeyShare
}

func GenerateQueryKey(index uint64, domainSize uint64, randomBytes []byte) (GeneratedKeys, error) {
	if len(randomBytes) == 0 {
		return GeneratedKeys{}, errors.New("random seed bytes are required")
	}

	var cKeys C.PQGeneratedKeys
	result := C.generate_query_key(
		C.uint64_t(index),
		C.uint64_t(domainSize),
		(*C.uint8_t)(unsafe.Pointer(&randomBytes[0])),
		C.size_t(len(randomBytes)),
		&cKeys,
	)
	if result != 0 {
		return GeneratedKeys{}, errors.New("generate_query_key failed")
	}
	defer C.free_generated_query_key(&cKeys)

	return GeneratedKeys{
		Left:  fromCKeyShare(cKeys.left),
		Right: fromCKeyShare(cKeys.right),
	}, nil
}

func DecodeBlock128Hex(value string) (Block128, error) {
	bytes, err := hex.DecodeString(value)
	if err != nil {
		return Block128{}, err
	}
	if len(bytes) != 16 {
		return Block128{}, errors.New("block hex must encode 16 bytes")
	}
	return Block128{
		X: binary.LittleEndian.Uint32(bytes[0:4]),
		Y: binary.LittleEndian.Uint32(bytes[4:8]),
		Z: binary.LittleEndian.Uint32(bytes[8:12]),
		W: binary.LittleEndian.Uint32(bytes[12:16]),
	}, nil
}

func EncodeBlock128Hex(value Block128) string {
	buf := make([]byte, 16)
	binary.LittleEndian.PutUint32(buf[0:4], value.X)
	binary.LittleEndian.PutUint32(buf[4:8], value.Y)
	binary.LittleEndian.PutUint32(buf[8:12], value.Z)
	binary.LittleEndian.PutUint32(buf[12:16], value.W)
	return hex.EncodeToString(buf)
}

func EncodeU64Hex(value uint64) string {
	buf := make([]byte, 8)
	binary.LittleEndian.PutUint64(buf, value)
	return hex.EncodeToString(buf)
}

func AggregateQueryShare(
	party int,
	key KeyShare,
	payload []uint64,
	recordCount int,
	blockCount int,
	workerCount int,
) ([]uint64, error) {
	if len(payload) != recordCount*blockCount {
		return nil, fmt.Errorf("payload size mismatch: got %d want %d", len(payload), recordCount*blockCount)
	}

	cKey := C.PQKeyShare{
		seed:                  toCBlock(key.Seed),
		correction_word_count: C.size_t(len(key.CorrectionWords)),
		in_bits:               C.uint32_t(key.InBits),
		domain_size:           C.uint64_t(key.DomainSize),
	}
	if len(key.CorrectionWords) > 0 {
		cwsPtr := C.malloc(C.size_t(len(key.CorrectionWords)) * C.size_t(unsafe.Sizeof(C.PQCorrectionWord{})))
		if cwsPtr == nil {
			return nil, errors.New("failed to allocate correction word buffer")
		}
		defer C.free(cwsPtr)
		cws := unsafe.Slice((*C.PQCorrectionWord)(cwsPtr), len(key.CorrectionWords))
		for i, cw := range key.CorrectionWords {
			cws[i] = C.PQCorrectionWord{
				s:  toCBlock(cw.S),
				tr: boolToU8(cw.Tr),
			}
		}
		cKey.correction_words = (*C.PQCorrectionWord)(cwsPtr)
	}

	out := make([]uint64, blockCount)
	result := C.aggregate_query_share(
		C.int(party),
		&cKey,
		(*C.uint64_t)(unsafe.Pointer(&payload[0])),
		C.size_t(recordCount),
		C.size_t(blockCount),
		C.uint32_t(workerCount),
		(*C.uint64_t)(unsafe.Pointer(&out[0])),
	)
	runtime.KeepAlive(key)
	runtime.KeepAlive(payload)
	runtime.KeepAlive(out)
	if result != 0 {
		return nil, errors.New("aggregate_query_share failed")
	}
	return out, nil
}

func toCBlock(value Block128) C.PQBlock128 {
	return C.PQBlock128{
		x: C.uint32_t(value.X),
		y: C.uint32_t(value.Y),
		z: C.uint32_t(value.Z),
		w: C.uint32_t(value.W),
	}
}

func boolToU8(value bool) C.uint8_t {
	if value {
		return 1
	}
	return 0
}

func fromCKeyShare(value C.PQKeyShare) KeyShare {
	key := KeyShare{
		Seed:       fromCBlock(value.seed),
		InBits:     uint32(value.in_bits),
		DomainSize: uint64(value.domain_size),
	}
	count := int(value.correction_word_count)
	key.CorrectionWords = make([]CorrectionWord, 0, count)
	if count == 0 || value.correction_words == nil {
		return key
	}

	cws := unsafe.Slice(value.correction_words, count)
	for _, cw := range cws {
		key.CorrectionWords = append(key.CorrectionWords, CorrectionWord{
			S:  fromCBlock(cw.s),
			Tr: cw.tr != 0,
		})
	}
	return key
}

func fromCBlock(value C.PQBlock128) Block128 {
	return Block128{
		X: uint32(value.x),
		Y: uint32(value.y),
		Z: uint32(value.z),
		W: uint32(value.w),
	}
}

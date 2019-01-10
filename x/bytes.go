package x

import "encoding/binary"

//
// Author: 陈哈哈 chenyongjia@parkingwang.com, yoojiachen@gmail.com
//

// ByteWrapper 提供一个按字节顺序读写字节数据的封装类
type ByteWrapper struct {
	data   []byte
	offset int
	op     binary.ByteOrder
}

func (br *ByteWrapper) GetByte() byte {
	b := br.data[br.offset]
	br.offsetNext(1)
	return b
}

func (br *ByteWrapper) PutByte(v byte) {
	br.data[br.offset] = v
	br.offsetNext(1)
}

func (br *ByteWrapper) GetBytes(out []byte) {
	size := len(out)
	copy(out, br.data[br.offset:br.offset+size-1])
	br.offsetNext(size)
}

func (br *ByteWrapper) GetBytesSize(size int) []byte {
	out := make([]byte, size)
	copy(out, br.data[br.offset:br.offset+size-1])
	br.offsetNext(size)
	return out
}

func (br *ByteWrapper) PutBytes(in []byte) {
	size := len(in)
	copy(br.data[br.offset:], in)
	br.offsetNext(size)
}

func (br *ByteWrapper) GetUint16() uint16 {
	v := br.op.Uint16(br.data[br.offset:])
	br.offsetNext(2)
	return v
}

func (br *ByteWrapper) PutUint16(v uint16) {
	br.op.PutUint16(br.data[br.offset:], v)
	br.offsetNext(2)
}

func (br *ByteWrapper) GetUint32() uint32 {
	v := br.op.Uint32(br.data[br.offset:])
	br.offsetNext(4)
	return v
}

func (br *ByteWrapper) PutUint32(v uint32) {
	br.op.PutUint32(br.data[br.offset:], v)
	br.offsetNext(4)
}

func (br *ByteWrapper) GetUint64() uint64 {
	v := br.op.Uint64(br.data[br.offset:])
	br.offsetNext(8)
	return v
}

func (br *ByteWrapper) PutUint64(v uint64) {
	br.op.PutUint64(br.data[br.offset:], v)
	br.offsetNext(8)
}

func (br *ByteWrapper) offsetNext(step int) {
	br.offset += step
}

func (br *ByteWrapper) Reset() {
	br.offset = 0
}

////

func WrapByteReaderBigEndian(data []byte) *ByteWrapper {
	return &ByteWrapper{
		data:   data,
		offset: 0,
		op:     binary.BigEndian,
	}
}

func WrapByteReaderLittleEndian(data []byte) *ByteWrapper {
	return &ByteWrapper{
		data:   data,
		offset: 0,
		op:     binary.LittleEndian,
	}
}

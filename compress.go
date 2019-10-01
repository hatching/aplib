// Copyright (C) 2019 Hatching B.V.
// All rights reserved.

package aplib

import (
	"bytes"
)

type aPcompress struct {
	src      []byte
	index    int
	length   int
	pair     bool
	tag      int
	tagindex int
	bitcount uint32
	offset   int
	dst      bytes.Buffer
}

func (a *aPcompress) findLongestMatch() (int, int) {
	offset, length := 0, 0
	for idx := 0; idx < a.length; idx++ {
		word := a.src[a.index : a.index+idx+1]
		if index := bytes.LastIndex(a.src[:a.index+idx], word); index < 0 {
			break
		} else {
			offset = a.index - index
			length = idx + 1
		}
	}
	return offset, length
}

func (a *aPcompress) writeBit(bit int) {
	if a.bitcount != 0 {
		a.bitcount--
	} else {
		// TODO This is a bit ugly, but one approach is to patch the tag
		// byte in-place, and that's what we do here.
		if a.tagindex != 0 {
			a.dst.Bytes()[a.tagindex] = byte(a.tag)
		}
		a.tagindex = a.dst.Len()
		a.dst.Write([]byte{0})
		a.bitcount = 7
		a.tag = 0
	}

	if bit != 0 {
		a.tag |= 1 << a.bitcount
	}
}

func (a *aPcompress) binlen(value int) int {
	if value == 0 {
		return 1
	}
	ret := 0
	for value != 0 {
		value >>= 1
		ret++
	}
	return ret
}

func (a *aPcompress) lengthDelta(offset int) int {
	if offset < 0x80 || offset >= 0x7d00 {
		return 2
	}
	if offset >= 0x500 {
		return 1
	}
	return 0
}

func (a *aPcompress) writeNumber(value int) {
	length := a.binlen(value) - 2
	a.writeBit(value & (1 << uint(length)))
	for idx := length - 1; idx != -1; idx-- {
		a.writeBit(1)
		a.writeBit(value & (1 << uint(idx)))
	}
	a.writeBit(0)
}

func (a *aPcompress) block(offset, length int) {
	a.writeBit(1)
	a.writeBit(0)
	if a.pair && a.offset == offset {
		a.writeNumber(2)
		a.writeNumber(length)
	} else {
		high := (offset >> 8) + 2
		if a.pair {
			high++
		}
		a.writeNumber(high)
		a.dst.Write([]byte{uint8(offset)})
		a.writeNumber(length - a.lengthDelta(offset))
	}
	a.index += length
	a.length -= length
	a.offset = offset
	a.pair = false
}

func (a *aPcompress) shortBlock(offset, length int) {
	a.writeBit(1)
	a.writeBit(1)
	a.writeBit(0)
	a.dst.Write([]byte{uint8(offset + offset + length - 2)})
	a.index += length
	a.length -= length
	a.offset = offset
	a.pair = false
}

func (a *aPcompress) singleByte(offset int) {
	a.writeBit(1)
	a.writeBit(1)
	a.writeBit(1)
	a.writeBit((offset >> 3) & 1)
	a.writeBit((offset >> 2) & 1)
	a.writeBit((offset >> 1) & 1)
	a.writeBit((offset >> 0) & 1)
	a.index++
	a.length--
	a.pair = true
}

func Compress(buf []byte) []byte {
	a := aPcompress{
		src:    buf,
		length: len(buf),
		pair:   true,
	}

	if a.length < 1 {
		return nil
	}

	a.dst.Write([]byte{a.src[a.index]})
	a.index++
	a.length--

	for a.length != 0 {
		offset, length := a.findLongestMatch()
		switch {
		case length == 0:
			ch := a.src[a.index]
			if ch == 0 {
				a.singleByte(0)
			} else {
				a.writeBit(0)
				a.dst.Write([]byte{ch})
				a.index++
				a.length--
				a.pair = true
			}
		case length == 1 && offset < 16:
			a.singleByte(offset)
		case (length == 2 || length == 3) && offset < 128:
			a.shortBlock(offset, length)
		case length >= 3 && offset >= 2:
			a.block(offset, length)
		default:
			a.writeBit(0)
			a.dst.Write([]byte{a.src[a.index]})
			a.index++
			a.length--
			a.pair = true
		}
	}
	a.writeBit(1)
	a.writeBit(1)
	a.writeBit(0)
	a.dst.Write([]byte{0})
	a.dst.Bytes()[a.tagindex] = byte(a.tag)
	return a.dst.Bytes()
}

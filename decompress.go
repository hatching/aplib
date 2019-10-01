// Copyright (C) 2019 Hatching B.V.
// All rights reserved.

package aplib

import (
	"bytes"
)

type aPdecompress struct {
	src      []byte
	index    uint32
	length   uint32
	dst      bytes.Buffer
	tag      uint8
	bitcount uint32
}

func (a *aPdecompress) getBit() *int {
	if a.bitcount == 0 {
		if a.length == 0 {
			return nil
		}
		a.tag = a.src[a.index]
		a.index++
		a.length--
		a.bitcount = 7
	} else {
		a.bitcount--
	}

	bit := int((a.tag >> 7) & 1)
	a.tag <<= 1
	return &bit
}

func (a *aPdecompress) getGamma() *int {
	result := 1
	cont := true
	for cont {
		if result&0x80000000 != 0 {
			return nil
		}

		if bit := a.getBit(); bit == nil {
			return nil
		} else {
			result += result + *bit
		}

		if bit := a.getBit(); bit == nil {
			return nil
		} else {
			cont = *bit != 0
		}
	}
	return &result
}

func Decompress(buf []byte) []byte {
	a := aPdecompress{
		src:    buf,
		length: uint32(len(buf)),
	}

	r0, lwm, done := -1, false, false

	if a.length < 1 {
		return nil
	}

	a.dst.Write([]byte{a.src[a.index]})
	a.index++
	a.length--

	for !done {
		if bit := a.getBit(); bit == nil {
			return nil
		} else if *bit != 0 {
			if bit := a.getBit(); bit == nil {
				return nil
			} else if *bit != 0 {
				if bit := a.getBit(); bit == nil {
					return nil
				} else if *bit != 0 {
					offs := 0
					for i := 4; i != 0; i-- {
						if bit := a.getBit(); bit == nil {
							return nil
						} else {
							offs += offs + *bit
						}
					}
					if offs != 0 {
						if offs >= a.dst.Len() {
							return nil
						}
						ch := a.dst.Bytes()[a.dst.Len()-offs]
						a.dst.Write([]byte{ch})
					} else {
						a.dst.Write([]byte{0})
					}
					lwm = false
				} else {
					if a.length == 0 {
						return nil
					}

					offs := int(a.src[a.index])
					a.index++
					a.length--

					length := 2 + (offs & 1)
					offs >>= 1

					if offs != 0 {
						if offs >= a.dst.Len() {
							return nil
						}

						// off := a.dst.Len()-offs
						// a.dst.Write(a.dst.Bytes()[off:off+length])
						for ; length != 0; length-- {
							ch := a.dst.Bytes()[a.dst.Len()-offs]
							a.dst.Write([]byte{ch})
						}
					} else {
						done = true
					}
					r0 = offs
					lwm = true
				}
			} else {
				if offs := a.getGamma(); offs == nil {
					return nil
				} else if lwm == false && *offs == 2 {
					*offs = r0

					if length := a.getGamma(); length == nil {
						return nil
					} else if *offs >= a.dst.Len() {
						return nil
					} else {
						// off := a.dst.Len()-*offs
						// a.dst.Write(a.dst.Bytes()[off:off+*length])
						for ; *length != 0; *length -= 1 {
							ch := a.dst.Bytes()[a.dst.Len()-*offs]
							a.dst.Write([]byte{ch})
						}
					}
				} else {
					if lwm == false {
						*offs -= 3
					} else {
						*offs -= 2
					}

					if *offs > 0x00fffffe {
						return nil
					}

					*offs <<= 8

					if a.length == 0 {
						return nil
					}
					*offs += int(a.src[a.index])
					a.index++
					a.length--

					if length := a.getGamma(); length == nil {
						return nil
					} else {
						if *offs >= 32000 {
							*length++
						}
						if *offs >= 1280 {
							*length++
						}
						if *offs < 128 {
							*length += 2
						}
						if *offs >= a.dst.Len() {
							return nil
						}
						// off := a.dst.Len()-*offs
						// a.dst.Write(a.dst.Bytes()[off:off+*length])
						for ; *length != 0; *length -= 1 {
							ch := a.dst.Bytes()[a.dst.Len()-*offs]
							a.dst.Write([]byte{ch})
						}
						r0 = *offs
					}
					lwm = true
				}
			}
		} else {
			if a.length == 0 {
				return nil
			}
			a.dst.Write([]byte{a.src[a.index]})
			a.index++
			a.length--
			lwm = false
		}
	}
	return a.dst.Bytes()
}

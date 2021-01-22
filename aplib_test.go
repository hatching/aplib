// Copyright (C) 2019-2021 Hatching B.V.
// All rights reserved.

package aplib

import (
	"reflect"
	"testing"
)

func TestCompress(t *testing.T) {
	v := Compress([]byte("hello world"))
	if !reflect.DeepEqual(v, []byte("h8el\x8eo wnr\xccd\x00")) {
		t.Errorf("invalid compression of hello world")
	}
}

func TestDecompress(t *testing.T) {
	v := Decompress([]byte("h8el\x8eo wnr\xccd\x00"))
	if !reflect.DeepEqual(v, []byte("hello world")) {
		t.Errorf("invalid decompression of hello world")
	}

	Decompress([]byte("\xc2+\xed\xff\x02\xff\xff\xff\xff\xff\xff\xff\xe63}456620891834-41"))
	Decompress([]byte("0E0"))
	Decompress([]byte("n\x15ii\x15\x03\xef\xef\xbf\xef\xef\xbf\xff\x00\x04\xff"))
	Decompress([]byte("0U0\x00"))
}

func TestTwoWay(t *testing.T) {
	var v []byte
	for idx := 0; idx < 1024; idx++ {
		v = append(v, 'A')
	}
	if !reflect.DeepEqual(v, Decompress(Compress(v))) {
		t.Errorf("invalid decompression of hello world")
	}
}

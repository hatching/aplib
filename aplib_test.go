// Copyright (C) 2019 Hatching B.V.
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
}

// Copyright (c) 2016 Leonard Hecker
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package argon2

import (
	"bytes"
	"reflect"
	"strconv"
	"testing"
)

var (
	config = Config{
		HashLength:  32,
		SaltLength:  16,
		TimeCost:    3,
		MemoryCost:  1 << 12,
		Parallelism: 1,
		Mode:        ModeArgon2i,
		Version:     Version13,
	}

	password = []byte("password")
	salt     = []byte("saltsalt")

	expectedHash    = []byte{0x96, 0x5b, 0xd4, 0x76, 0xaa, 0x7a, 0xf7, 0x2d, 0x91, 0x07, 0xad, 0xbd, 0x74, 0x2b, 0x86, 0xe3, 0x69, 0x11, 0xe7, 0x2f, 0x8e, 0x71, 0xcf, 0xf3, 0x88, 0xa5, 0x79, 0x92, 0x7d, 0xeb, 0x48, 0xe3}
	expectedEncoded = []byte("$argon2i$v=19$m=4096,t=3,p=1$c2FsdHNhbHQ$llvUdqp69y2RB629dCuG42kR5y+Occ/ziKV5kn3rSOM")
)

func isFalsey(obj interface{}) bool {
	if obj == nil {
		return true
	}

	value := reflect.ValueOf(obj)
	kind := value.Kind()

	return kind >= reflect.Chan && kind <= reflect.Slice && value.IsNil()
}

func mustBeFalsey(t *testing.T, name string, obj interface{}) {
	if !isFalsey(obj) {
		t.Errorf("'%s' should be nil, but is: %v", name, obj)
	}
}

func mustBeTruthy(t *testing.T, name string, obj interface{}) {
	if isFalsey(obj) {
		t.Errorf("'%s' should be non nil, but is: %v", name, obj)
	}
}

func TestHashRaw(t *testing.T) {
	r, err := config.HashRaw(password)
	mustBeTruthy(t, "r.Config", r.Config)
	mustBeTruthy(t, "r.Salt", r.Salt)
	mustBeTruthy(t, "r.Hash", r.Hash)
	mustBeFalsey(t, "err", err)
}

func TestHashEncoded(t *testing.T) {
	enc, err := config.HashEncoded(password)
	mustBeTruthy(t, "encoded", enc)
	mustBeFalsey(t, "err", err)

	if len(enc) == 0 {
		t.Error("len(encoded) must be > 0")
	}

	for _, b := range enc {
		if b == 0 {
			t.Error("encoded must not contain 0x00")
		}
	}
}

func TestHashWithSalt(t *testing.T) {
	r, err := config.Hash(password, salt)
	mustBeTruthy(t, "r.Config", r.Config)
	mustBeTruthy(t, "r.Salt", r.Salt)
	mustBeTruthy(t, "r.Hash", r.Hash)
	mustBeFalsey(t, "err", err)

	if !bytes.Equal(r.Hash, expectedHash) {
		t.Logf("ref: %v", expectedHash)
		t.Logf("act: %v", r.Hash)
		t.Error("hashes do not match")
	}

	enc := r.Encode()
	mustBeTruthy(t, "encoded", enc)

	if !bytes.Equal(enc, expectedEncoded) {
		t.Logf("ref: %s", string(expectedEncoded))
		t.Logf("act: %s", string(enc))
		t.Error("encoded strings do not match")
	}
}

func TestVerifyRaw(t *testing.T) {
	r, err := config.HashRaw(password)
	mustBeTruthy(t, "r.Config", r.Config)
	mustBeTruthy(t, "r.Salt", r.Salt)
	mustBeTruthy(t, "r.Hash", r.Hash)
	mustBeFalsey(t, "err1", err)

	ok, err := r.Verify(password)
	mustBeTruthy(t, "ok", ok)
	mustBeFalsey(t, "err2", err)
}

func TestVerifyEncoded(t *testing.T) {
	encoded, err := config.HashEncoded(password)
	mustBeTruthy(t, "encoded", encoded)
	mustBeFalsey(t, "err1", err)

	ok, err := VerifyEncoded(password, encoded)
	mustBeTruthy(t, "ok", ok)
	mustBeFalsey(t, "err2", err)
}

func TestSecureZeroMemory(t *testing.T) {
	pwd := append([]byte(nil), password...)

	// SecureZeroMemory should erase up to cap(pwd) --> let's test that too
	SecureZeroMemory(pwd[0:0])

	for _, b := range pwd {
		if b != 0 {
			t.Error("pwd must only contain 0x00")
		}
	}
}

func BenchmarkHash(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, _ = config.Hash(password, salt)
	}
}

func BenchmarkVerify(b *testing.B) {
	r, err := config.Hash(password, salt)
	if err != nil {
		b.Error(err)
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, _ = r.Verify(password)
	}
}

func BenchmarkEncode(b *testing.B) {
	r, err := config.Hash(password, salt)
	if err != nil {
		b.Error(err)
	}

	b.SetBytes(int64(len(expectedEncoded)))
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = r.Encode()
	}
}

func BenchmarkDecode(b *testing.B) {
	b.SetBytes(int64(len(expectedEncoded)))
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, _ = Decode(expectedEncoded)
	}
}

func BenchmarkSecureZeroMemory(b *testing.B) {
	for _, n := range []int{16, 256, 4096, 65536} {
		b.Run(strconv.Itoa(n), func(b *testing.B) {
			buf := make([]byte, n)

			b.SetBytes(int64(n))
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				SecureZeroMemory(buf)
			}
		})
	}
}

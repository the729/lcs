package lcs

import (
	"errors"
	"io"
)

// readVarUint reads an unsigned integer of size n defined in https://webassembly.github.io/spec/core/binary/values.html#binary-int
// readVarUint panics if n>64.
func readVarUint(r io.Reader, n uint) (uint64, error) {
	if n > 64 {
		panic(errors.New("leb128: n must <= 64"))
	}
	p := make([]byte, 1)
	var res uint64
	var shift uint
	for {
		_, err := io.ReadFull(r, p)
		if err != nil {
			return 0, err
		}
		b := uint64(p[0])
		switch {
		// note: can not use b < 1<<n, when n == 64, 1<<n will overflow to 0
		case b < 1<<7 && b <= 1<<n-1:
			res += (1 << shift) * b
			return res, nil
		case b >= 1<<7 && n > 7:
			res += (1 << shift) * (b - 1<<7)
			shift += 7
			n -= 7
		default:
			return 0, errors.New("leb128: invalid uint")
		}
	}
}

// writeVarUint writes a LEB128 encoded unsigned 64-bit integer to w.
// It returns the integer value, the size of the encoded value (in bytes), and
// the error (if any).
func writeVarUint(w io.Writer, v uint64) (int, error) {
	var buf []byte
	for {
		c := uint8(v & 0x7f)
		v >>= 7
		if v != 0 {
			c |= 0x80
		}
		buf = append(buf, c)
		if c&0x80 == 0 {
			break
		}
	}
	return w.Write(buf)
}

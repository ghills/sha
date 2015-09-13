package mysha1

import "fmt"
import "io"

const outSz = 20
const debug = false

// RFC reads such that bitstream is interpreted as words
// that are big endian
func word(in []byte) uint32 {
	return uint32(in[0])<<24 |
		uint32(in[1])<<16 |
		uint32(in[2])<<8 |
		uint32(in[3])
}

func word_at(msg []byte, i int) uint32 {
	return word(msg[i : i+4])
}

// F function as defined in RFC
func f(i int, b uint32, c uint32, d uint32) uint32 {
	switch {
	case i <= 19:
		return (b & c) | (^b & d)
	case i <= 39:
		return b ^ c ^ d
	case i <= 59:
		return (b & c) | (b & d) | (c & d)
	case i <= 79:
		return b ^ c ^ d
	}
	return 0
}

// K constants as defined in RFC
func k(i int) uint32 {
	switch {
	case i <= 19:
		return 0x5A827999
	case i <= 39:
		return 0x6ED9EBA1
	case i <= 59:
		return 0x8F1BBCDC
	case i <= 79:
		return 0xCA62C1D6
	}
	return 0
}

// initialize H as defined in RFC
func init_h_array(h []uint32) {
	h[0] = 0x67452301
	h[1] = 0xEFCDAB89
	h[2] = 0x98BADCFE
	h[3] = 0x10325476
	h[4] = 0xC3D2E1F0

	if debug {
		fmt.Println("initial h values:")

		for i, value := range h {
			fmt.Printf("h[%d] = 0x%X\n", i, value)
		}
	}
}

// circular shift as defined in RFC
func s(x uint32, n uint) uint32 {
	return (x << n) | (x >> (32 - n))
}

// unpacks H array to output byte buffer
func unpack_data(h []uint32, buf []byte) {
	for i := 0; i < len(h); i++ {
		buf[4*i+0] = byte((h[i] >> 24) & 0xFF)
		buf[4*i+1] = byte((h[i] >> 16) & 0xFF)
		buf[4*i+2] = byte((h[i] >> 8) & 0xFF)
		buf[4*i+3] = byte(h[i] & 0xFF)
	}
}

func print_w(w []uint32, lower int, upper int) {
	for i := lower; i <= upper; i++ {
		fmt.Printf("w[%d] = 0x%08X\n", i, w[i])
	}
}

func Digest(reader io.Reader) []byte {
	// wrap passed in reader in block reader
	// it takes care of enforcing block and
	// padding rules
	reader = NewBlockReader(reader)

	var a, b, c, d, e, temp uint32
	h := make([]uint32, 5)
	w := make([]uint32, 80)

	init_h_array(h)

	// for each block
	buf := make([]byte, BlockSize)
	for n, err := reader.Read(buf); n > 0; n, err = reader.Read(buf) {
		if debug {
			fmt.Println("processing new block")
		}

		// step a - load up w
		for j := 0; j < (BlockSize / 4); j++ {
			w[j] = word_at(buf, 4*j)
		}
		if debug {
			fmt.Println("Block words")
			print_w(w, 0, 15)
		}

		// step b
		for t := 16; t <= 79; t++ {
			w[t] = s(w[t-3]^w[t-8]^w[t-14]^w[t-16], 1)
		}

		// step c
		a = h[0]
		b = h[1]
		c = h[2]
		d = h[3]
		e = h[4]

		// step d
		if debug {
			fmt.Println("\t\tA\t\tB\t\tC\t\tD\t\tE")
		}
		for t := 0; t <= 79; t++ {
			temp = s(a, 5) + f(t, b, c, d) + e + w[t] + k(t)
			e = d
			d = c
			c = s(b, 30)
			b = a
			a = temp

			if debug {
				fmt.Printf("t = %2d: %08X\t%08X\t%08X\t%08X\t%08X\n", t, a, b, c, d, e)
			}
		}

		//step e
		h[0] = h[0] + a
		h[1] = h[1] + b
		h[2] = h[2] + c
		h[3] = h[3] + d
		h[4] = h[4] + e

		if debug {
			fmt.Println("")
		}

		// according to docs, n > 0 should be processed before considering err
		if err == io.EOF {
			break
		}
	}

	out := make([]byte, outSz)
	unpack_data(h, out)

	return out
}

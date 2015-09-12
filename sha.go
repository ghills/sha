package main

import "fmt"
import "math/big"
import "encoding/hex"
import "os"
import "io/ioutil"

const BlockSize = 64
const BlockSizeInWords = 16
const Size = 20

func word(in []byte) uint32 {
	return uint32(in[0])<<24 | uint32(in[1])<<16 | uint32(in[2])<<8 | uint32(in[3])
}

func f(i int, b uint32, c uint32, d uint32) uint32 {
	if i <= 19 {
		return (b & c) | (^b & d)
	} else if i <= 39 {
		return b ^ c ^ d
	} else if i <= 59 {
		return (b & c) | (b & d) | (c & d)
	} else if i <= 79 {
		return b ^ c ^ d
	}
	return 0
}

func k(i int) uint32 {
	if i <= 19 {
		return 0x5A827999
	} else if i <= 39 {
		return 0x6ED9EBA1
	} else if i <= 59 {
		return 0x8F1BBCDC
	} else if i <= 79 {
		return 0xCA62C1D6
	}
	return 0
}

func pad(msg []byte) []byte {
	l := len(msg)
	to_add := BlockSize - (l % BlockSize)
	if to_add < 9 {
		// not enough room for pad and len, add another block
		to_add += BlockSize
	}
	padding := make([]byte, to_add)

	// set first pad bit to 0
	padding[0] = 0x80

	// set l at end of padding - change to bit length
	bitlength := uint64(l) * 8
	fmt.Println("bitlength", bitlength)
	for i := 0; i < 8; i++ {
		mask := uint64(0xFF) << uint64(8*i)
		padding[len(padding)-(i+1)] = uint8((bitlength & mask) >> uint64(8*i))
	}

	return append(msg, padding...)
}

func init_h_array(h []uint32) {
	h[0] = 0x67452301
	h[1] = 0xEFCDAB89
	h[2] = 0x98BADCFE
	h[3] = 0x10325476
	h[4] = 0xC3D2E1F0

	fmt.Println("initial h values:")
	for i, value := range h {
		fmt.Printf("h[%d] = 0x%X\n", i, value)
	}
}

func word_at(msg []byte, i int) uint32 {
	return word(msg[i : i+4])
}

func s(x uint32, n uint) uint32 {
	return (x << n) | (x >> (32 - n))
}

func unpack_data(h []uint32, buf []byte) {
	for i := 0; i < len(h); i++ {
		buf[4*i+0] = byte((h[i] >> 24) & 0xFF)
		buf[4*i+1] = byte((h[i] >> 16) & 0xFF)
		buf[4*i+2] = byte((h[i] >> 8) & 0xFF)
		buf[4*i+3] = byte(h[i] & 0xFF)
	}
}

func mod_add(x, y uint32) uint32 {
	var product, z big.Int
	bigx := big.NewInt(int64(x))
	bigy := big.NewInt(int64(y))
	product.Add(bigx, bigy)
	z.Mod(&product, big.NewInt(1<<32))

	return uint32(z.Int64())
}

func print_w(w []uint32, lower int, upper int) {
	for i := lower; i <= upper; i++ {
		fmt.Printf("w[%d] = 0x%08X\n", i, w[i])
	}
}

func digest(msg []byte) []byte {
	msg = pad(msg)

	// make into uint32s to process words
	data := make([]uint32, len(msg)/4)
	for i := 0; i < len(data); i++ {
		data[i] = word_at(msg, 4*i)
	}

	var a, b, c, d, e, temp uint32
	h := make([]uint32, 5)
	w := make([]uint32, 80)

	init_h_array(h)

	// for each block
	for i := 0; i < len(data)/BlockSizeInWords; i++ {
		fmt.Println("processing block", i)

		// step a - load up w
		for j := 0; j < BlockSizeInWords; j++ {
			w[j] = data[BlockSizeInWords*i+j]
		}
		fmt.Println("Block 1 words")
		print_w(w, 0, 15)

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
		fmt.Println("\t\tA\t\tB\t\tC\t\tD\t\tE")
		for t := 0; t <= 79; t++ {
			temp = mod_add(
				mod_add(
					mod_add(
						mod_add(s(a, 5), f(t, b, c, d)), e), w[t]), k(t))
			e = d
			d = c
			c = s(b, 30)
			b = a
			a = temp

			fmt.Printf("t = %2d: %08X\t%08X\t%08X\t%08X\t%08X\n", t, a, b, c, d, e)
		}

		//step e
		h[0] = mod_add(h[0], a)
		h[1] = mod_add(h[1], b)
		h[2] = mod_add(h[2], c)
		h[3] = mod_add(h[3], d)
		h[4] = mod_add(h[4], e)

		fmt.Println("")
	}

	//var out [Size]byte
	out := make([]byte, Size)
	unpack_data(h, out)

	return out
}

func main() {
	bytes, _ := ioutil.ReadAll(os.Stdin)
	hash := digest(bytes)
	fmt.Println(hex.EncodeToString(hash))
}

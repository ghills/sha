package mysha1

import "io"

const BlockSize = 64

type BlockReader struct {
	r       io.Reader
	count   uint64
	pending []byte
	eof     bool
}

func NewBlockReader(r io.Reader) *BlockReader {
	return &BlockReader{r, 0, make([]byte, BlockSize), false}
}

func (br *BlockReader) Read(p []byte) (int, error) {
	if len(p) < BlockSize {
		return 0, io.EOF
	}

	// check for eof, use pending
	if br.eof {
		copy(p, br.pending)
		return BlockSize, io.EOF
	}

	//do stuff
	n, err := br.r.Read(p)
	br.count += uint64(n)

	// set extra bytes to zero (padded)
	for i := n; i < len(p); i++ {
		p[i] = 0
	}

	if n < BlockSize || err == io.EOF {
		br.eof = true
		// check if there is room to fill in first 1 pad in current buffer
		pending_idx := 0
		if n < len(p) {
			p[n] = 0x80 // MSB set to 1
			n += 1
		} else {
			br.pending[0] = 0x80 // MSB set to 1
			pending_idx++
		}

		// check if room for size here
		var pad_buf []byte
		if len(p)-n >= 8 {
			pad_buf = p

			// everything fit in this word. EOF
			err = io.EOF
		} else {
			pad_buf = br.pending

			// overwrite error so caller will check again
			err = nil
		}

		// set size at the end of the buffer
		bitlength := uint64(br.count) * 8
		for i := 0; i < 8; i++ {
			mask := uint64(0xFF) << uint64(8*i)
			pad_buf[len(pad_buf)-(i+1)] = uint8((bitlength & mask) >> uint64(8*i))
		}
	}

	return BlockSize, err
}

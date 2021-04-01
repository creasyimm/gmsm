// +build amd64

package sm4

import (
	"crypto/cipher"

	smcipher "github.com/emmansun/gmsm/cipher"
	"golang.org/x/sys/cpu"
)

var supportsAES = cpu.X86.HasAES
var supportsGFMUL = cpu.X86.HasPCLMULQDQ

//go:noescape
func encryptBlocksAsm(xk *uint32, dst, src *byte)

//go:noescape
func encryptBlockAsm(xk *uint32, dst, src *byte)

//go:noescape
func expandKeyAsm(key *byte, ck, enc, dec *uint32)

type sm4CipherAsm struct {
	sm4Cipher
}

func newCipher(key []byte) (cipher.Block, error) {
	if !supportsAES {
		return newCipherGeneric(key)
	}
	c := sm4CipherAsm{sm4Cipher{make([]uint32, rounds), make([]uint32, rounds)}}
	expandKeyAsm(&key[0], &ck[0], &c.enc[0], &c.dec[0])
	if supportsAES && supportsGFMUL {
		return &sm4CipherGCM{c}, nil
	}
	return &c, nil
}

const FourBlocksSize = 64

func (c *sm4CipherAsm) BlockSize() int { return BlockSize }

func (c *sm4CipherAsm) Encrypt(dst, src []byte) {
	if len(src) < BlockSize {
		panic("sm4: input not full block")
	}
	if len(dst) < BlockSize {
		panic("sm4: output not full block")
	}
	if smcipher.InexactOverlap(dst[:BlockSize], src[:BlockSize]) {
		panic("sm4: invalid buffer overlap")
	}
	encryptBlockAsm(&c.enc[0], &dst[0], &src[0])
}

func (c *sm4CipherAsm) Decrypt(dst, src []byte) {
	if len(src) < BlockSize {
		panic("sm4: input not full block")
	}
	if len(dst) < BlockSize {
		panic("sm4: output not full block")
	}
	if smcipher.InexactOverlap(dst[:BlockSize], src[:BlockSize]) {
		panic("sm4: invalid buffer overlap")
	}
	encryptBlockAsm(&c.dec[0], &dst[0], &src[0])
}

// expandKey is used by BenchmarkExpand to ensure that the asm implementation
// of key expansion is used for the benchmark when it is available.
func expandKey(key []byte, enc, dec []uint32) {
	if supportsAES {
		expandKeyAsm(&key[0], &ck[0], &enc[0], &dec[0])
	} else {
		expandKeyGo(key, enc, dec)
	}
}
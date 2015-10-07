package totp

import (
	"crypto/hmac"
	"crypto/sha1"
	"fmt"
	"hash"
	"time"
)

// Options contains the different configurable values for a given TOTP
// invocation.
type Options struct {
	Time     func() time.Time
	TimeStep time.Duration
	Digits   uint8
	Hash     func() hash.Hash
}

// NewOptions constructs a pre-configured Options. The returned Options' uses
// time.Now to get the current time, has a window size of 30 seconds, and
// tries the currently active window, and the previous one. It expects 6 digits,
// and uses sha1 for its hash algorithm. These settings were chosen to be
// compatible with Google Authenticator.
func NewOptions() *Options {
	return &Options{
		Time:     time.Now,
		TimeStep: 30 * time.Second,
		Digits:   6,
		Hash:     sha1.New,
	}
}

var DefaultOptions = NewOptions()

var digit_power = []int64{
	1,          // 0
	10,         // 1
	100,        // 2
	1000,       // 3
	10000,      // 4
	100000,     // 5
	1000000,    // 6
	10000000,   // 7
	100000000,  // 8
	1000000000, // 9
}

// Authenticate verifies the TOTP userCode taking the key from secretKey and
// other options from o. If o is nil, then DefaultOptions is used instead.
func Authenticate(secretKey []byte, userCode string, o *Options) bool {
	return Key(secretKey, o) == userCode
}
func Key(secretKey []byte, o *Options) string {
	if o == nil {
		o = DefaultOptions
	}

	t := o.Time().Unix() / int64(o.TimeStep/time.Second)
	var tbuf [8]byte

	hm := hmac.New(o.Hash, secretKey)
	var hashbuf []byte

	tbuf[0] = byte(t >> 56)
	tbuf[1] = byte(t >> 48)
	tbuf[2] = byte(t >> 40)
	tbuf[3] = byte(t >> 32)
	tbuf[4] = byte(t >> 24)
	tbuf[5] = byte(t >> 16)
	tbuf[6] = byte(t >> 8)
	tbuf[7] = byte(t)

	hm.Reset()
	hm.Write(tbuf[:])
	hashbuf = hm.Sum(hashbuf[:0])

	offset := hashbuf[len(hashbuf)-1] & 0xf
	truncatedHash := hashbuf[offset:]

	code := int64(truncatedHash[0])<<24 |
		int64(truncatedHash[1])<<16 |
		int64(truncatedHash[2])<<8 |
		int64(truncatedHash[3])

	code &= 0x7FFFFFFF
	code %= digit_power[o.Digits]
	return fmt.Sprint(code)
}

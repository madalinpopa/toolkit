package toolkit

import "crypto/rand"

// randomSourceString is a collection of characters used as the basis for generating random strings.
var randomSourceString = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789_+"

// Tools is a utility type that provides various helper methods.
type Tools struct{}

// RandomString generates a random string of the specified length using characters from randomSourceString.
func (t *Tools) RandomString(length int) string {

	s, r := make([]rune, length), []rune(randomSourceString)
	for i := range s {
		p, _ := rand.Prime(rand.Reader, len(r))
		x, y := p.Uint64(), uint64(len(r))
		s[i] = r[x%y]
	}
	return string(s)
}

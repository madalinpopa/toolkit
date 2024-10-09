package toolkit

import "testing"

// TestTools_RandomString is a unit test function that validates the RandomString method of the Tools type.
func TestTools_RandomString(t *testing.T) {
	var testTools Tools

	s := testTools.RandomString(10)
	if len(s) != 10 {
		t.Error("wrong length random string return")
	}
}

package connect

import (
	"strings"
	"testing"
)

func TestParseCredientials(t *testing.T) {
	var tests = []struct {
		input       string
		expectCreds Credentials
		expectErr   error
	}{
		{"username=user1\npassword=pass1", Credentials{"user1", "pass1"}, nil},
		{" \n username = user1 \n password = pass1 \n", Credentials{"user1", "pass1"}, nil},
		{"username = user1 \n junk \n password = pass1 \n", Credentials{"user1", "pass1"}, nil},
		{"USERNAME = user1 \n passed = pass1", Credentials{}, ErrMalformedSccCredFile},
		{"username= \n password = \n", Credentials{}, ErrMalformedSccCredFile},
	}

	for _, test := range tests {
		got, err := parseCredientials(strings.NewReader(test.input))
		if err != test.expectErr || got != test.expectCreds {
			t.Errorf("ParseCredientials() == %+v, %s, expected %+v, %s", got, err, test.expectCreds, test.expectErr)
		}
	}
}

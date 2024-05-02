package services

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCheckLunh(t *testing.T) {
	cases := []struct {
		number string
		ok     bool
		err    bool
	}{
		{
			number: "4561261212345464",
			ok:     false,
			err:    false,
		},
		{
			number: "4561261212345467",
			ok:     true,
			err:    false,
		},
		{
			number: "2634",
			ok:     true,
			err:    false,
		},
		{
			number: "8334",
			ok:     true,
			err:    false,
		},
		{
			number: "йо-хо-хо",
			ok:     false,
			err:    true,
		},
	}
	for i := range cases {
		c := &cases[i]
		t.Run(c.number, func(t *testing.T) {
			ok, err := checkLuhn(c.number)
			require.Equal(t, c.ok, ok)
			if !c.err {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}
		})
	}
}

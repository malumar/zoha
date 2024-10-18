package mtp

import (
	"fmt"
	"github.com/malumar/zoha/api"
	"github.com/stretchr/testify/assert"

	"testing"
)

func FuzzValidEmail(f *testing.F) {
	for _, seed := range [][]string{
		{api.DontKnow.String(), "", "empty"},
		{api.No.String(), "@", "only at"},
		{api.No.String(), "user@", "don't have domain part"},
		{api.No.String(), "@example", "don't have user part, and domain is not fqdn (only hostname)"},
		{api.No.String(), "@example.com", "don't have user part, only domain part"},
		{api.DontKnow.String(), "user@example", "domain is not fqdn (only hostname)"},
		{api.Yes.String(), "user@example.tld", "username with fqdn"},
	} {
		f.Add(seed[0], seed[1], seed[2])
	}
	f.Fuzz(func(t *testing.T, valid, email, reason string) {
		switch valid {
		case api.No.String():
			assert.Equal(t, api.No, ValidEmail(email), email+" "+reason)
			break
		case api.Yes.String():
			assert.Equal(t, api.Yes, ValidEmail(email), email+" "+reason)
			break
		case api.DontKnow.String():
			assert.Equal(t, api.DontKnow, ValidEmail(email), email+" "+reason)
			break
		default:
			assert.Error(t, fmt.Errorf("unexpected valid value `%v`", valid))
		}

	})
}

package emailutil

import "testing"
import "github.com/stretchr/testify/assert"

func TestExtractEmail(t *testing.T) {

	matchExtractMail(t, "info", "xyz.tld", "info@xyz.tld")
	matchExtractMail(t, "info", "xyz.tld", "<info@xyz.tld>")
	matchExtractMail(t, "info", "xyzA.tld", "<info@xyzA.tld>")
	matchExtractMail(t, "info", "xn--2daa5pudb.pll", "<info@xn--2daa5pudb.pll>")

}

func matchExtractMail(t *testing.T, name string, host string, matchTO string) {
	n, h, e := ExtractEmail(matchTO)
	assert.NoError(t, e)
	assert.Equal(t, name, n)
	assert.Equal(t, host, h)
}

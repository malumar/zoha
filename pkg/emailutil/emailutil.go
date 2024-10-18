package emailutil

import (
	"errors"
	"regexp"
	"strings"
)

var errEmptyEmail = errors.New("empty email address")
var errInvalidAdresEmail = errors.New("invalid address email")

func IsEmptyErrEmail(err error) bool {
	if errors.Is(err, errEmptyEmail) {
		return true
	}
	return false
}

// ExtractEmail extract components (name and host) from the email address and possibly removing < ... >
func ExtractEmail(str string) (name string, host string, err error) {
	openTag := strings.Index(str, "<")
	closeTag := strings.Index(str, ">")

	if openTag > -1 && closeTag > -1 {

		if openTag >= closeTag {
			return "", "", errInvalidAdresEmail
		}

		if openTag+1 == closeTag {
			// this is for mtp who wants to know whether the address is empty or incorrectly formatted
			return "", "", errEmptyEmail
		} else {
			str = str[openTag+1 : closeTag]
		}
	} else {
		if openTag > -1 || closeTag > -1 {
			return "", "", errInvalidAdresEmail
		}
	}

	if len(str) == 0 {
		// t cannot also contain spaces, because that is a completely different thing,
		// it is for mtp which wants to know whether the address is empty or poorly formatted
		return "", "", errEmptyEmail
	} else {
		if len(strings.TrimSpace(str)) == 0 {
			return "", "", errInvalidAdresEmail
		}
	}

	name, host = extractNameAndHost(str)
	host = validHost(host)

	if host == "" || name == "" {
		err = errors.New("Invalid address, [" + name + "@" + host + "] address:" + str)
	}
	return name, host, err
}

func extractNameAndHost(str string) (name string, host string) {
	fstr := strings.TrimSpace(fixEmailString(str))
	fstrLen := len(fstr)
	altPos := strings.Index(fstr, "@")
	if altPos <= 0 {
		return "", fstr
	}
	if altPos >= fstrLen {
		return fstr, ""
	}
	return fstr[0:altPos], fstr[altPos+1:]
}

// fixEmailString remove < and >
func fixEmailString(str string) string {
	if len(str) > 3 {
		if str[0:1] == "<" {
			str = str[1:]
		}

		if str[len(str)-1:] == ">" {

			str = str[:len(str)-1]
		}
	}

	return str
}

func validHost(host string) string {
	if reValidHost.MatchString(host) &&
		strings.Index(host, "-.") == -1 &&
		strings.Index(host, ".-") == -1 &&
		strings.Index(host, "..") == -1 {
		return host
	}
	return ""
}

var reExtractMail *regexp.Regexp
var reValidHost *regexp.Regexp

func init() {

	reExtractMail, _ = regexp.Compile(`<(.+?)@(.+?)>`) // go home regex, you're drunk!
	reValidHost, _ = regexp.Compile(`^[a-zA-Z0-9][a-zA-Z0-9\-\.]+[a-zA-Z]+$`)
}

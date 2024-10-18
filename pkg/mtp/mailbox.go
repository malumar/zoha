package mtp

import (
	"github.com/malumar/strutils"
	"github.com/malumar/zoha/api"
)

// IsWhitelisted Whether the @asciiEmailLowerCase is marked as a white sender
func IsWhitelisted(mb *api.Mailbox, asciiEmailLowerCase string) bool {
	for _, v := range mb.WhiteList {
		if strutils.Match(v, asciiEmailLowerCase) {
			return true
		}
	}
	return false
}

// IsBlacklisted Whether the @asciiEmailLowerCase is marked as a white sender
func IsBlacklisted(mb *api.Mailbox, asciiEmailLowerCase string) bool {
	for _, v := range mb.BlackList {
		if strutils.Match(v, asciiEmailLowerCase) {
			return true
		}
	}
	return false
}

// HaveCapacity whether we will fit a message about the size of the @needCapacity in the user's account
func HaveCapacity(mb *api.Mailbox, needCapacity int64) bool {
	var ncc uint64
	if needCapacity > 0 {
		ncc = uint64(needCapacity)
	}

	if mb.Quota == 0 || mb.UsedQuota+ncc > mb.Quota {
		return true
	}

	return false
}

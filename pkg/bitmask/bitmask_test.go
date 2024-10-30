package bitmask

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

const (
	TESTFLAG_ONE Flag = 1 << iota
	TESTFLAG_TWO
	TESTFLAG_THREE
)
const (
	TESTFLAG64_ONE Flag64 = 1 << iota
	TESTFLAG64_TWO
	TESTFLAG64_THREE
)

func TestAddFlag(t *testing.T) {

	var mainFlag = TESTFLAG_TWO

	mainFlag.AddFlag(TESTFLAG_THREE)

	assert.Zero(t, mainFlag&(1<<TESTFLAG_THREE))

}

func TestClearFlag(t *testing.T) {

	var mainFlag = TESTFLAG_ONE | TESTFLAG_THREE

	mainFlag.ClearFlag(TESTFLAG_THREE)

	assert.NotZero(t, 1<<TESTFLAG_ONE)

}

func TestHasFlag(t *testing.T) {

	var mainFlag = TESTFLAG_ONE | TESTFLAG_THREE

	assert.True(t, mainFlag.HasFlag(TESTFLAG_THREE))

}

func TestToggleFlag(t *testing.T) {
	flag := TESTFLAG_ONE | TESTFLAG_THREE
	flag.ToggleFlag(TESTFLAG_ONE)
	assert.False(t, flag.HasFlag(TESTFLAG_ONE))
	flag.ToggleFlag(TESTFLAG_ONE)
	assert.True(t, flag.HasFlag(TESTFLAG_ONE))

}

func TestFlag64_AllCombinations(t *testing.T) {

	AllCombinations64([]FlagInfo64{
		{"TESTFLAG64_ONE", TESTFLAG64_ONE},
		{"TESTFLAG64_TWO", TESTFLAG64_TWO},
		{"TESTFLAG64_THREE", TESTFLAG64_THREE},
	})
}

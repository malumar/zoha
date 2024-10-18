package api

// Extended bool
type Maybe int

const (
	No Maybe = iota
	Yes
	// W chwili obecnej nie jestem tego w stanie sprawdzić
	DontKnow
)

func Answer(val bool) Maybe {
	if val {
		return Yes
	}
	return No
}

// HaveAnswer Do you know the answer?
func (m Maybe) HaveAnswer() bool {
	return m != DontKnow
}

// String for test purposes
func (b Maybe) String() string {
	switch b {
	case Yes:
		return "Yes"
	case No:
		return "No"
	case DontKnow:
		return "DontKnow"
	}
	return "unrecognized"
}
func (b Maybe) ToBool() bool {
	return b.True()
}

func (b Maybe) True() bool {
	return b == Yes
}

func (b Maybe) False() bool {
	return b == No
}

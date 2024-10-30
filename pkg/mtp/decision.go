package mtp

type Decision string

func (d Decision) String() string {
	return string(d)
}

const DecisionKey = "response"

const (
	// Dunno I don't know what to do, or I don't want to decide,
	// let the next person in the chain make that decision
	// this should be the default answer, which does not confirm anything,
	// it only informs that we are moving to the next stage
	Dunno Decision = "dunno"
	// Accept we finish the entire operation and confirm; other mechanisms after this are not important
	// (they will not be checked) if you want to confirm the correctness of the operation,
	// you should return dunno accept only finally
	// todo: Accept = "accept"
	// Reject rejection, we do not proceed
	Reject = "reject"
)

package mtp

//goland:noinspection GoCommentStart,GoCommentStart,GoCommentStart
const (
	FromField       = "from"
	ToField         = "to"
	MessageUidField = "mid"
	// to the public?
	// safe to pass on to an outsider (user police, etc.)
	ForPublicField = "pub"
	// action name
	ActionNameField = "action"
	// if the value was parried, give me the original so you know what was the reason
	OriginalValue = "original"

	ListenAddress = "listenAddress"
	// state of message rejected|queued
	DeliveryStatusField = "deliveryStatus"

	DeliveryRejected = "rejected"
	DeliveryQueued   = "queued"

	// when we have specified what parameter and limit determines its limit,
	// and has its value that the user has provided
	LimitField = "limit"
	HasField   = "has"
)

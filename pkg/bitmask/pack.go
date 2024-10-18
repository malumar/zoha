package bitmask

func PackTwoUint32(lowBits, highBits uint32) (packed uint64) {
	packed = (uint64(lowBits) << 32) | uint64(highBits)
	return
}

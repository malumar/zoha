package bitmask

type Flag uint32

func (f Flag) HasFlag(flag Flag) bool { return f&flag != 0 }
func (f *Flag) AddFlag(flag Flag)     { *f |= flag }
func (f *Flag) ClearFlag(flag Flag)   { *f &= ^flag }
func (f *Flag) ToggleFlag(flag Flag)  { *f ^= flag }

type Flag64 uint64
type FlagInfo64 struct {
	Name  string
	Value Flag64
}

func (f Flag64) HasFlag(flag Flag64) bool  { return f&flag != 0 }
func (f Flag64) IsNotSet(flag Flag64) bool { return f&flag == 0 }
func (f *Flag64) AddFlag(flag Flag64)      { *f |= flag }
func (f *Flag64) ClearFlag(flag Flag64)    { *f &= ^flag }
func (f *Flag64) ToggleFlag(flag Flag64)   { *f ^= flag }

// AllCombinations64 All returns all combinations for a given string array.
// This is essentially a powerset of the given set except that the empty set is disregarded.
func AllCombinations64(set []FlagInfo64) (subsets [][]FlagInfo64) {
	length := uint(len(set))
	for subsetBits := 1; subsetBits < (1 << length); subsetBits++ {
		//fmt.Println(subsetBits)
		var subset []FlagInfo64

		for object := uint(0); object < length; object++ {
			// checks if object is contained in subset
			// by checking if bit 'object' is set in subsetBits
			if (subsetBits>>object)&1 == 1 {
				// add object to subset
				subset = append(subset, set[object])
			}
		}
		subsets = append(subsets, subset)

	}

	return
}

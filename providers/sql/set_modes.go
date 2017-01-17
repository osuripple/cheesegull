package sql

const (
	modeSTD = 1 << iota
	modeTaiko
	modeCTB
	modeMania
)

var modeMappings = [...]int{
	modeSTD,
	modeTaiko,
	modeCTB,
	modeMania,
}

// build a single int based on the modes passed.
// for example, if 0 is passed, then the first value (osu! standard) of
// modeMappings will be OR'd into the int, which means 1 will be given. The
// second (taiko) will be 2, the third (ctb) 4, the fourth (mania) 8.
// this can also combine, e.g. ctb with mania = 12
func modesToEnum(i []int) int {
	var r int
	for _, v := range i {
		// check v can possibly be an array index of modeMappings
		if v < 0 || v >= len(modeMappings) {
			continue
		}
		// OR into r
		r |= modeMappings[v]
	}
	return r
}

func enumToModes(i int) []int {
	modes := make([]int, 0, len(modeMappings))
	// using uints because bit shifting in go is only allowed with uints on the
	// right
	for m := uint(0); m < uint(len(modeMappings)); m++ {
		if i&(1<<m) > 0 {
			modes = append(modes, int(m))
		}
	}
	return modes
}

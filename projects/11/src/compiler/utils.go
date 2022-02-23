package compiler

func stringInclude(strs []string, str string) bool {
	for _, s := range strs {
		if s == str {
			return true
		}
	}
	return false
}

func runeInclude(runes []rune, r rune) bool {
	for _, s := range runes {
		if s == r {
			return true
		}
	}
	return false
}

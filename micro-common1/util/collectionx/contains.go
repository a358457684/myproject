package collectionx

func Contains(a []string, s string) bool {
	for _, item := range a {
		if item == s {
			return true
		}
	}
	return false
}

func NotContains(a []string, s string) bool {
	return !Contains(a, s)
}

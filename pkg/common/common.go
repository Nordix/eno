package common

// SearchInSlice - Checks if String in Slice
func SearchInSlice(str string, list []string) bool {
	for _, item := range list {
		if str == item {
			return true
		}
	}
	return false
}

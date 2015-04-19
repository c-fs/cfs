package disk

func min(a, b int64) int64 {
	if a > b {
		return b
	}
	return a
}

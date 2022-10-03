package cross

// ValidateConnCredsData validadte the content from a data
func ValidateConnCredsData(v []byte, b bool) string {
	if b {
		return string(v)
	}
	return ""
}

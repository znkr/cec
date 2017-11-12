package cec

func isValidOsdName(s string) bool {
	// Must be between 1 and 14 bytes long
	if len(s) < 1 || len(s) > 14 {
		return false
	}
	// Each byte must be in [0x20, 0x7e] (ASCII).
	for _, b := range s {
		if b < 0x20 || b > 0x7e {
			return false
		}
	}
	return true
}

func isValidVendorId(id uint32) bool {
	return id <= 0xffffff
}

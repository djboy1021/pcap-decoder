package dictionary

// GetProductID returns the equivalent of a byte Product ID into a string
func GetProductID(id byte) string {
	switch id {
	case 0x21:
		return "HDL-32E"
	case 0x22:
		return "VLP-16/Puck LITE"
	case 0x24:
		return "Puck Hi-Res"
	case 0x28:
		return "VLP-32C"
	case 0x31:
		return "Velarray"
	case 0xA1:
		return "VLS-128"
	}
	return "Unknown"
}

// VLP32ElevationAngles is an array of the VLP32 elevation angles
var VLP32ElevationAngles = []int16{-25000, -1000, -1667, -15639, -11310, 0, -667, -8843, -7254, 333, -333,
	-6148, -5333, 1333, 667, -4000, -4667, 1667, 1000, -3667, -3333, 3333, 2333, -2667, -3000, 7000, 4667, -2333, -2000, 15000, 10333, -1333}

// VLP32AzimuthOffset is an array of the VLP32 azimuth offset angles
var VLP32AzimuthOffset = []int16{1400, -4200, 1400, -1400, 1400, -1400, 4200, -1400, 1400, -4200, 1400, -1400, 4200, -1400,
	4200, -1400, 1400, 4200, 1400, -4200, 4200, -1400, 1400, -1400, 1400, -1400, 1400, -4200, 4200, -1400, 1400, -1400}

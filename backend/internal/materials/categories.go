package materials

var Categories = []string{"CRT", "LCD Panel", "PCB", "Cable", "Battery", "Motor", "Magnet-bearing Assembly", "Mixed Plastics from EEE", "Other E-Waste"}

func Supported(category string) bool {
	for _, c := range Categories {
		if c == category {
			return true
		}
	}
	return false
}

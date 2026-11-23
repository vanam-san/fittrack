package units

const kgPerPound = 0.45359237

func WeightInKG(value float64, unit string) float64 {
	if unit == "lb" {
		return value * kgPerPound
	}
	return value
}

func WeightFromKG(value float64, unit string) float64 {
	if unit == "lb" {
		return value / kgPerPound
	}
	return value
}

package math

func AddToAverage[TNumber Number](currentAverage TNumber, value TNumber, size TNumber) TNumber {
	return (size*currentAverage + value) / (size + 1)
}

package domain

// CalculateLuhn return the check number
func CalculateLuhn(number int64) int64 {
	checkNumber := checksum(number)

	if checkNumber == 0 {
		return 0
	}
	return int64(10) - checkNumber
}

// Valid проверка что число - валидное по алгоритму Луна
func IsLuhnValid(number int64) bool {
	ten := int64(10)

	return (number%ten+checksum(number/ten))%ten == int64(0)
}

func checksum(number int64) int64 {
	var luhn int64
	ten := int64(10)
	for i := int64(0); number > int64(0); i++ {
		cur := number % ten

		if i%int64(2) == int64(0) { // even
			cur = cur * int64(2)
			if cur > int64(9) {
				cur = cur%ten + cur/ten
			}
		}

		luhn += cur
		number = number / 10
	}
	return luhn % 10
}

package domain

import "fmt"

type Money int64

func (m Money) IsPositive() bool {
	return m > 0
}

func (m Money) String() string {
	value := int64(m)
	sign := ""
	if value < 0 {
		sign = "-"
		value = -value
	}
	return fmt.Sprintf("%sR$ %d,%02d", sign, value/100, value%100)
}

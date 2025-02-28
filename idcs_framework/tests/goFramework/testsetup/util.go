package testsetup

import (
	"goFramework/utils"
	"math"
)

// Random payload generation for jwt token
func Rand_token_payload_gen() string {
	tid := utils.GenerateString(12)
	return tid
}

func SearchSlice(elems []string, v string) bool {
	for _, s := range elems {
		if v == s {
			return true
		}
	}
	return false
}

func RoundFloat(val float64, precision uint) float64 {
	ratio := math.Pow(10, float64(precision))
	return math.Round(val*ratio) / ratio
}

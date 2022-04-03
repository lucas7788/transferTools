package utils

import (
	"fmt"
	"math/big"
	"testing"
)

func TestParseParams(t *testing.T) {

	val := ToIntByPrecise("0.00000000000000000001", 18)
	fmt.Println(val)
	sum := new(big.Int)
	sum = sum.Add(sum, val)
	fmt.Println(ToStringByPrecise(sum, 18))
}

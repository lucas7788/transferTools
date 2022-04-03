package utils

import (
	"math/big"
	"strings"
)


func ToStringByPrecise(bigNum *big.Int, decimals uint64) string {
	result := ""
	destStr := bigNum.String()
	destLen := uint64(len(destStr))
	if decimals >= destLen { // add "0.000..." at former of destStr
		var i uint64 = 0
		prefix := "0."
		for ; i < decimals-destLen; i++ {
			prefix += "0"
		}
		if destLen > 12 {
			destStr = destStr[:12]
		}
		result = prefix + destStr
	} else { // add "."
		pointIndex := destLen - decimals
		if len(destStr[pointIndex:]) > 12 {
			destStr = destStr[:pointIndex+12]
		}
		result = destStr[0:pointIndex] + "." + destStr[pointIndex:]
	}
	// delete no need "0" at last of result
	i := len(result) - 1
	for ; i >= 0; i-- {
		if result[i] != '0' {
			break
		}
	}
	result = result[:i+1]
	// delete "." at last of result
	if result[len(result)-1] == '.' {
		result = result[:len(result)-1]
	}
	return result
}

func ToIntByPrecise(str string, decimals uint64) *big.Int {
	result := new(big.Int)
	splits := strings.Split(str, ".")
	if len(splits) == 1 { // doesn't contain "."
		var i uint64 = 0
		for ; i < decimals; i++ {
			str += "0"
		}
		intValue, ok := new(big.Int).SetString(str, 10)
		if ok {
			result.Set(intValue)
		}
	} else if len(splits) == 2 {
		value := new(big.Int)
		ok := false
		floatLen := uint64(len(splits[1]))
		if floatLen <= decimals { // add "0" at last of str
			parseString := strings.Replace(str, ".", "", 1)
			var i uint64 = 0
			for ; i < decimals-floatLen; i++ {
				parseString += "0"
			}
			value, ok = value.SetString(parseString, 10)
		} else { // remove redundant digits after "."
			splits[1] = splits[1][:decimals]
			parseString := splits[0] + splits[1]
			value, ok = value.SetString(parseString, 10)
		}
		if ok {
			result.Set(value)
		}
	}

	return result
}

func ToIntFromBool(d bool) int {
	if d == true {
		return 1
	} else {
		return 0
	}
}

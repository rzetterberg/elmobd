package elmobd

import (
	"fmt"
	"strconv"
)

func BytesToUint64(bytes []byte) (uint64, error) {
	var res uint64

	amount := len(bytes)

	if amount > 8 {
		return 0, fmt.Errorf("Got more than 8 bytes")
	}

	for i := range bytes {
		curr := bytes[amount-(i+1)]

		res |= uint64(curr) << uint(i*8)
	}

	return res, nil
}

func HexLitsToBytes(literals []string) ([]byte, error) {
	var result []byte

	for i := range literals {
		if len(literals[i]) != 2 {
			continue
		}

		curr, err := strconv.ParseUint(
			literals[i],
			16,
			8,
		)

		if err != nil {
			return nil, err
		}

		result = append(result, uint8(curr))
	}

	return result, nil
}

func ByteToBits(input byte) [8]bool {
	var result [8]bool

	for i := range result {
		result[7-i] = (input & 1) == 1
		input = input >> 1
	}

	return result
}

func BytesToBits(inputs []byte) []bool {
	var result []bool

	for i := range inputs {
		curr := ByteToBits(inputs[i])

		result = append(
			result,
			curr[:]...,
		)
	}

	return result
}

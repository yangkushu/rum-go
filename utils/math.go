package utils

import (
	"errors"
	"math/big"
)

func intToBigInt(n int) *big.Int {
	return big.NewInt(int64(n))
}

func SafeIntAdd(a int, b int) (int, error) {
	result := intToBigInt(a).Add(intToBigInt(a), intToBigInt(b))
	if !result.IsInt64() {
		return 0, errors.New("integer overflow")
	}
	return int(result.Int64()), nil
}

func SafeIntSub(a int, b int) (int, error) {
	result := intToBigInt(a).Sub(intToBigInt(a), intToBigInt(b))
	if !result.IsInt64() {
		return 0, errors.New("integer overflow")
	}
	return int(result.Int64()), nil
}

func SafeIntMul(a int, b int) (int, error) {
	result := intToBigInt(a).Mul(intToBigInt(a), intToBigInt(b))
	if !result.IsInt64() {
		return 0, errors.New("integer overflow")
	}
	return int(result.Int64()), nil
}

func SafeIntDiv(a int, b int) (int, error) {
	if b == 0 {
		return 0, errors.New("division by zero")
	}
	result := intToBigInt(a).Div(intToBigInt(a), intToBigInt(b))
	return int(result.Int64()), nil
}

func SafeFloat64Add(x, y float64) float64 {
	// 初始化big.Float并设置值
	fX := new(big.Float).SetFloat64(x)
	fY := new(big.Float).SetFloat64(y)

	// 执行加法
	result := new(big.Float).Add(fX, fY)

	// 将big.Float转换成float64
	outcome, _ := result.Float64()
	return outcome
}

func SafeFloat64Sub(x, y float64) float64 {
	fX := new(big.Float).SetFloat64(x)
	fY := new(big.Float).SetFloat64(y)

	result := new(big.Float).Sub(fX, fY)

	outcome, _ := result.Float64()
	return outcome
}

func SafeFloat64Mul(x, y float64) float64 {
	fX := new(big.Float).SetFloat64(x)
	fY := new(big.Float).SetFloat64(y)

	result := new(big.Float).Mul(fX, fY)

	outcome, _ := result.Float64()
	return outcome
}

func SafeFloat64Div(x, y float64) float64 {
	fX := new(big.Float).SetFloat64(x)
	fY := new(big.Float).SetFloat64(y)

	result := new(big.Float).Quo(fX, fY)

	outcome, _ := result.Float64()
	return outcome
}

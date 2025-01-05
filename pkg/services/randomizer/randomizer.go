package randomizer

import (
	"eclipse/internal/base"
	"eclipse/internal/logger"
	"math"
	"math/big"
	"math/rand"
	"time"
)

func GetRandomValueWithPrecision(minValue, maxValue float64, minPrecision, maxPrecision int, decimals float64) (float64, string) {
	value := minValue + rand.Float64()*(maxValue-minValue)
	precision := base.GetRandomPrecision(minPrecision, maxPrecision)
	roundedValue := base.RoundFloat(value, precision)

	multiplier := new(big.Float).SetFloat64(math.Pow(10, decimals))

	bf := new(big.Float).SetFloat64(roundedValue)

	weiFloat := new(big.Float).Mul(bf, multiplier)

	weiInt := new(big.Int)
	weiFloat.Int(weiInt)

	return roundedValue, weiInt.String()
}

func RandomDelay(min, max float64, inMinutes bool) {
	delayRange := max - min
	randomDelay := min + rand.Float64()*delayRange

	var delayDuration time.Duration
	var unitStr string

	if inMinutes {
		delayDuration = time.Duration(randomDelay * float64(time.Minute))
		unitStr = "минут"
	} else {
		delayDuration = time.Duration(randomDelay * float64(time.Second))
		unitStr = "секунд"
	}

	logger.Info("Ожидание выполнения: %.2f %s\n", randomDelay, unitStr)
	time.Sleep(delayDuration)
}

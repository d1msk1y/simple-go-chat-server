package limiter

import (
	"github.com/ulule/limiter/v3"
	"github.com/ulule/limiter/v3/drivers/store/memory"
	"time"
)

func GetLimiter() *limiter.Limiter {
	rate := limiter.Rate{
		Period: 1 * time.Second,
		Limit:  25,
	}

	rate, err := limiter.NewRateFromFormatted("25-M")
	if err != nil {
		panic(err)
	}

	store := memory.NewStore()

	limiterInstance := limiter.New(store, rate)
	return limiterInstance
}

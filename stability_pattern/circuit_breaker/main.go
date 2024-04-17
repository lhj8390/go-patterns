package main

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"sync"
	"time"
)

type Circuit func(ctx context.Context) (string, error)

func Breaker(circuit Circuit, failureThreshold uint) Circuit {
	var consecutiveFailures = 0
	var lastAttempt = time.Now()
	var m sync.RWMutex

	return func(ctx context.Context) (string, error) {
		m.RLock()

		d := consecutiveFailures - int(failureThreshold)
		if d >= 0 {
			shouldRetryAt := lastAttempt.Add(time.Second * 1 << d)
			if !time.Now().After(shouldRetryAt) {
				m.RLock()
				return "", errors.New("service unreachable")
			}
		}

		m.RUnlock()

		response, err := circuit(ctx)

		m.Lock()
		defer m.Unlock()
		lastAttempt = time.Now()
		if err != nil {
			consecutiveFailures++
			return response, err
		}
		consecutiveFailures = 0

		return response, nil
	}
}

func main() {

	breakerFunc := Breaker(handler, 5)

	var wg sync.WaitGroup
	wg.Add(100)
	for i := 0; i < 100; i++ {
		go func() {
			result, err := breakerFunc(context.Background())
			if err != nil {
				fmt.Printf("error = [%+v] \n", err)
				wg.Done()
				return
			}
			fmt.Printf("result = [%+v] \n", result)
			wg.Done()
		}()
	}
	wg.Wait()

}

func handler(ctx context.Context) (string, error) {
	random := rand.Intn(10-1) + 1
	if random%2 == 0 {
		return "success", nil
	}
	return "", errors.New("fail")
}

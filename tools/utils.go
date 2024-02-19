package tools

import "time"

func Retry(retryLimit int, sleepTime time.Duration, function func() error) error {
	var err error
	for i := 1; i <= retryLimit; i++ {
		err = function()
		if err == nil {
			return nil
		}
		time.Sleep(sleepTime)
	}
	return err
}

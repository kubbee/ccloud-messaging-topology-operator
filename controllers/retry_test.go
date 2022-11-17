package controllers

import (
	"errors"
	"testing"
	"time"
)

func TestRetry(t *testing.T) {

	for a := 1; a <= 3; a++ {

		if err := f(); err != nil {

			time.Sleep(time.Duration(10*a) * time.Second)

			if a == 3 {
				t.Log(err, "error to read crentials from cluster")
			}

		} else {
			break
		}
	}
}

func f() error {
	return errors.New("error")
}

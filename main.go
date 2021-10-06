package main

import (
	"time"

	"github.com/tsuzu/maiphone/pkg/interphone"
)

func main() {
	_, err := interphone.NewInterphone("10.20.40.123:14000", "10.20.40.123:16000", "37DE4683-CEBC-4A33-83E3-D2FCF5C4A54C")

	if err != nil {
		panic(err)
	}

	time.Sleep(30 * time.Second)
}

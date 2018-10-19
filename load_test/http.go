package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	vegeta "github.com/tsenart/vegeta/lib"
)

func main() {
	rps := flag.Int("rps", 100, "Request per Second")
	duration := flag.Duration("dur", 5*time.Second, "How long the test going to be performed")
	varianceNum := flag.Int("vars", 1000000, "Total num of variance on the request")

	flag.Parse()

	targets := []vegeta.Target{}
	for i := 0; i < *varianceNum; i++ {
		targets = append(targets, vegeta.Target{
			Method: "GET",
			URL:    fmt.Sprintf("http://localhost:8080/inquiry/%d", i),
		})
	}

	rate := uint64(*rps)
	targeter := vegeta.NewStaticTargeter(targets...)
	attacker := vegeta.NewAttacker()

	enc := vegeta.NewEncoder(os.Stdout)

	for res := range attacker.Attack(targeter, rate, *duration, "Inquiry Test") {
		enc.Encode(res)
	}
}

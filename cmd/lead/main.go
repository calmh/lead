package main

import (
	"flag"
	"fmt"
	"log"
	"time"

	"github.com/calmh/lead"
)

func main() {
	brightness := flag.Float64("brightness", -1, "Set brightness")
	red := flag.Float64("red", -1, "Set red")
	green := flag.Float64("green", -1, "Set green")
	blue := flag.Float64("blue", -1, "Set blue")
	flag.Parse()

	cs, err := lead.Discover("172.16.32.0/24")
	if err != nil {
		log.Fatal(err)
	}

	for _, c := range cs {
		if *brightness >= 0 {
			if err := c.SetBrightness(*brightness); err != nil {
				fmt.Println(err)
			}
		}
		if *red >= 0 {
			c.SetRGB(*red, *green, *blue)
		}
	}
}

func fade(from, to float64, steps int, during time.Duration, c *lead.Controller) {
	step := (to - from) / float64(steps)
	v := from
	for i := 0; i < steps; i++ {
		c.SetBrightness(v)
		v += step
		time.Sleep(during / time.Duration(steps))
	}
}

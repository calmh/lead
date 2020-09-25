package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/alecthomas/kingpin"
	"github.com/calmh/lead"
)

var (
	brightness optInt
	color      rgb
	discover   = ""
	controller = ""
)

var sunrise = slide{
	fromBrightness: 1,
	toBrightness:   32,
	fromRGB:        rgb{red: 255, green: 32, blue: 0},
	toRGB:          rgb{red: 255, green: 192, blue: 32},
}

type slide struct {
	fromBrightness int
	toBrightness   int
	fromRGB        rgb
	toRGB          rgb
}

func (s *slide) brightness(p float64) int {
	d := float64(s.toBrightness-s.fromBrightness) * p
	return int(float64(s.fromBrightness) + d)
}

func (s *slide) color(p float64) rgb {
	var res rgb
	dr := float64(s.toRGB.red-s.fromRGB.red) * p
	res.red = int(float64(s.fromRGB.red) + dr)
	dg := float64(s.toRGB.green-s.fromRGB.green) * p
	res.green = int(float64(s.fromRGB.green) + dg)
	db := float64(s.toRGB.blue-s.fromRGB.blue) * p
	res.blue = int(float64(s.fromRGB.blue) + db)
	return res
}

func main() {
	argNetwork := kingpin.Arg("network", "Network (i.e., 172.16.32.0/24) to probe").Required().String()
	duration := kingpin.Arg("duration", "Duration time of sunrise").Default("30m").Duration()
	kingpin.Parse()

	tcs, err := lead.Discover(*argNetwork)
	if err != nil {
		fmt.Println("Discovering controllers:", err)
		os.Exit(1)
	}

	var wg sync.WaitGroup
	for _, c := range tcs {
		wg.Add(1)
		go func(c *lead.Controller) {
			defer wg.Done()
			fmt.Println(c, "init")
			for i := 0; i < 5; i++ {
				if err := c.SetOn(true); err != nil {
					fmt.Printf("Turning on %s: %v\n", c.Address(), err)
				}
				time.Sleep(100 * time.Millisecond)
				if err := c.SetBrightness(sunrise.brightness(0)); err != nil {
					fmt.Printf("Setting brightness on %s: %v\n", c.Address(), err)
				}
				time.Sleep(100 * time.Millisecond)
				color := sunrise.color(0)
				if err := c.SetRGB(color.red, color.green, color.blue); err != nil {
					fmt.Printf("Setting RGB on %s: %v\n", c.Address(), err)
				}
				time.Sleep(100 * time.Millisecond)
			}
		}(c)
	}
	wg.Wait()

	var prevBrightness int
	var prevColor rgb
	for b := 1; b <= 100; b++ {
		time.Sleep(*duration / 100)
		p := float64(b) / 100
		color := sunrise.color(p)
		brightness := sunrise.brightness(p)
		fmt.Println("color", color, "brightness", brightness)

		var wg sync.WaitGroup
		for _, c := range tcs {
			wg.Add(1)
			go func(c *lead.Controller) {
				defer wg.Done()
				if color != prevColor {
					if err := c.SetRGB(color.red, color.green, color.blue); err != nil {
						fmt.Printf("Setting brightness on %s: %v\n", c.Address(), err)
					}
					time.Sleep(100 * time.Millisecond)
				}
				if brightness != prevBrightness {
					if err := c.SetBrightness(brightness); err != nil {
						fmt.Printf("Setting brightness on %s: %v\n", c.Address(), err)
					}
				}
			}(c)
		}
		wg.Wait()

		prevColor = color
		prevBrightness = brightness
	}
}

type rgb struct {
	red, green, blue int
	isSet            bool
}

func (v *rgb) Set(rgb string) error {
	fields := strings.Split(rgb, ",")
	if len(fields) != 3 {
		return fmt.Errorf("cannot parse as R,G,B")
	}

	var err error
	v.red, err = strconv.Atoi(fields[0])
	if err != nil {
		return err
	}
	v.green, err = strconv.Atoi(fields[1])
	if err != nil {
		return err
	}
	v.blue, err = strconv.Atoi(fields[2])
	if err != nil {
		return err
	}

	v.isSet = true
	return nil
}

func (v *rgb) String() string {
	if !v.isSet {
		return ""
	}
	return fmt.Sprintf("%d,%d,%d", v.red, v.green, v.blue)
}

type optInt struct {
	val   int
	isSet bool
}

func (v *optInt) Set(s string) error {
	var err error
	v.val, err = strconv.Atoi(s)
	if err != nil {
		return err
	}

	v.isSet = true
	return nil
}

func (v *optInt) String() string {
	if !v.isSet {
		return ""
	}
	return strconv.Itoa(v.val)
}

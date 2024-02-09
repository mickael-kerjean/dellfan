package main

import (
	"bytes"
	"errors"
	"fmt"
	"log/slog"
	"math"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

//    ^
//    |   manual        automatic
//    |<------------> <-------------
//    |              |
// s2 |              x
//    |             /|
//    |            / |
// s1 |          x/  |
//    |         /|   |
//    |        / |   |
// s0 |------x/  |   |
//    |      |   |   |
//    |      |   |   |
//    -------x---x---x------------>
//           t0  t1  t2

var (
	t = []float64{
		40, // t0
		60, // t1
		80, // t2
	}
	s = []float64{
		5,  // s0
		15, // s1
		50, // s2
	}
)

func main() {
	var currentSpeed float64
	for {
		temperature, err := ipmiGetTemp()
		if err != nil {
			slog.Error(fmt.Sprintf("cannot get temperature %s", err.Error()))
			time.Sleep(5 * time.Second)
			continue
		}
		manual, speed := tempToSpeed(float64(temperature))
		if math.Abs(speed-currentSpeed) <= 1 { // do nothing if not worth it
			time.Sleep(1 * time.Second)
			continue
		} else if currentSpeed == 100 && temperature > int(t[2]+t[1])/2 { // prevent auto/manual bounce
			time.Sleep(1 * time.Second)
			continue
		}

		if err = ipmiFancontrol(manual, fmt.Sprintf("0x%x", int(speed))); err != nil {
			time.Sleep(5 * time.Second)
			slog.Error(fmt.Sprintf("error in ipmi: %s", err.Error()))
			continue
		}
		slog.Info(fmt.Sprintf("temp=%d manual=%v speed=%d", int(temperature), manual, int(speed)))
		currentSpeed = speed
		time.Sleep(1 * time.Second)
	}
}

func tempToSpeed(temperature float64) (manual bool, speed float64) {
	if temperature < t[0] {
		return true, s[0]
	} else if temperature < t[1] {
		var (
			a = (s[1] - s[0]) / (t[1] - t[0])
			b = s[0] - a*t[0]
		)
		return true, a*temperature + b
	} else if temperature < t[2] {
		var (
			a = (s[2] - s[1]) / (t[2] - t[1])
			b = s[1] - a*t[1]
		)
		return true, a*temperature + b
	} else {
		return false, 100
	}
}

func ipmiGetTemp() (int, error) {
	b := new(bytes.Buffer)
	cmd := createCommand("sdr", "get", "Temp")
	cmd.Stdout = b
	if err := cmd.Run(); err != nil {
		return 0, err
	}
	pref := " Sensor Reading        : "
	for _, line := range strings.Split(b.String(), "\n") {
		if strings.HasPrefix(line, pref) == false {
			continue
		}
		arr := strings.SplitN(strings.TrimPrefix(line, pref), " ", 2)
		if len(arr) < 1 {
			continue
		}
		return strconv.Atoi(string(arr[0]))
	}
	return 0, errors.New("not found")
}

func ipmiFancontrol(isManual bool, speed string) (err error) {
	if isManual {
		err = createCommand("raw", "0x30", "0x30", "0x01", "0x00").Run()
	} else {
		err = createCommand("raw", "0x30", "0x30", "0x01", "0x01").Run()
	}
	if err != nil {
		return err
	}
	if isManual == false {
		return nil
	}
	if err = createCommand("raw", "0x30", "0x30", "0x02", "0xff", speed).Run(); err != nil {
		return err
	}
	return nil
}

func createCommand(args ...string) *exec.Cmd {
	cmd := exec.Command(
		"ipmitool",
		// append(
		// 	[]string{"-I", "lanplus", "-H", "192.168.68.101", "-U", "root", "-P", "calvin"},
		// 	args...,
		// )...,
		args...,
	)
	return cmd
}

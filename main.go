package main

import (
	"bytes"
	"errors"
	"log/slog"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

var (
	IDLE_SPEED      = "0x08"
	TEMP_TRIGGER    = 40
	TEMP_HISTERESIS = 3
	CURRENT_STATE   = CURRENT_STATE_UNKNOWN
)

const (
	CURRENT_STATE_UNKNOWN int = iota
	CURRENT_STATE_MANUAL
	CURRENT_STATE_AUTOMATIC
)

func main() {
	for {
		temp, err := getTemperature()
		if err != nil {
			time.Sleep(5 * time.Second)
			continue
		}
		if CURRENT_STATE == CURRENT_STATE_AUTOMATIC {
			fanManualMode(temp <= (TEMP_TRIGGER - TEMP_HISTERESIS))
		} else {
			fanManualMode(temp <= TEMP_TRIGGER)
		}
		time.Sleep(2 * time.Second)
	}
}

func fanManualMode(manualMode bool) error {
	switch CURRENT_STATE {
	case CURRENT_STATE_AUTOMATIC:
		if manualMode == false {
			return nil
		}
		CURRENT_STATE = CURRENT_STATE_UNKNOWN
		if err := fanManualMode(manualMode); err != nil {
			return err
		}
	case CURRENT_STATE_MANUAL:
		if manualMode {
			return nil
		}
		CURRENT_STATE = CURRENT_STATE_UNKNOWN
		if err := fanManualMode(manualMode); err != nil {
			return err
		}
	case CURRENT_STATE_UNKNOWN:
		var cmd *exec.Cmd
		if manualMode {
			cmd = exec.Command(
				"ipmitool", "raw",
				"0x30", "0x30",
				"0x01", "0x00",
			)
		} else {
			cmd = exec.Command(
				"ipmitool", "raw",
				"0x30", "0x30",
				"0x01", "0x01",
			)
		}
		if err := cmd.Run(); err != nil {
			return err
		}

		if !manualMode {
			slog.Info("state change - mode=automatic")
			CURRENT_STATE = CURRENT_STATE_AUTOMATIC
			return nil
		}
		if err := exec.Command(
			"ipmitool", "raw",
			"0x30", "0x30",
			"0x02", "0xff", IDLE_SPEED,
		).Run(); err != nil {
			return err
		}
		slog.Info("state change - mode=manual")
		CURRENT_STATE = CURRENT_STATE_MANUAL
		return nil

	}
	return errors.New("Unknown state")
}

func getTemperature() (int, error) {
	b := new(bytes.Buffer)
	cmd := exec.Command(
		"ipmitool",
		"sdr", "get", "Temp",
	)
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

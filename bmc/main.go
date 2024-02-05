package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gebn/bmc"
	"github.com/gebn/bmc/pkg/ipmi"

	. "github.com/mickael-kerjean/dellfan-shutup/ipmi-commands"
)

var (
	IDLE_SPEED      = byte(0x08)
	TEMP_TRIGGER    = 40
	TEMP_HISTERESIS = 3
)

var (
	IPMI_HOST = "192.168.68.101"
	// IPMI_HOST     = "127.0.0.1"
	IPMI_USERNAME = "root"
	IPMI_PASSWORD = "calvin"
	CURRENT_STATE = CURRENT_STATE_UNKNOWN
)

const (
	CURRENT_STATE_UNKNOWN int = iota
	CURRENT_STATE_MANUAL
	CURRENT_STATE_AUTOMATIC
)

func main() {
	slog.Info("- start dellfan")
	machine, err := bmc.Dial(context.Background(), IPMI_HOST)
	if err != nil {
		slog.Error(fmt.Sprintf("cannot send command: %s", err.Error()))
		os.Exit(1)
		return
	}
	slog.Info("- get system guid")
	if _, err = machine.GetSystemGUID(context.Background()); err != nil {
		slog.Error(fmt.Sprintf("cannot get system guid: %s", err.Error()))
		os.Exit(1)
		return
	}
	slog.Info("- start session")
	sess, err := machine.NewSession(context.Background(), &bmc.SessionOpts{
		Username:          IPMI_USERNAME,
		Password:          []byte(IPMI_PASSWORD),
		MaxPrivilegeLevel: ipmi.PrivilegeLevelAdministrator,
	})
	if err != nil {
		slog.Error(fmt.Sprintf("cannot establish session: %s", err.Error()))
		return
	}
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigs
		slog.Info("cleanup before exit")
		machine.Close()
		sess.Close(context.Background())
		os.Exit(0)
	}()

	// STEP2: elevate our session so we can manipulate fan
	slog.Info("- elevate our session privilege")
	if err := bmc.ValidateResponse(sess.SendCommand(context.Background(), &ipmi.SetSessionPrivilegeLevelCmd{
		Req: ipmi.SetSessionPrivilegeLevelReq{
			PrivilegeLevel: ipmi.PrivilegeLevelAdministrator,
		},
	})); err != nil {
		slog.Error(fmt.Sprintf("cannot establish session: %s", err.Error()))
		return
	}

	// STEP3: get sensor information
	slog.Info("- get sensor info")
	repo, err := bmc.RetrieveSDRRepository(context.Background(), sess)
	if err != nil {
		slog.Error(fmt.Sprintf("cannot get sdr repo: %s", err.Error()))
		os.Exit(1)
		return
	}
	fsr := repo[0]

	// STEP4: hot loop
	for {
		reader, err := bmc.NewSensorReader(fsr)
		if err != nil {
			time.Sleep(10 * time.Second)
			slog.Error(fmt.Sprintf("cannot get sensor reader %s", err.Error()))
			continue
		}
		reading, err := reader.Read(context.Background(), sess)
		if err != nil {
			time.Sleep(10 * time.Second)
			slog.Error(fmt.Sprintf("cannot read sensor data %s", err.Error()))
			continue
		}

		if CURRENT_STATE == CURRENT_STATE_AUTOMATIC {
			FanManualMode(sess, int(reading) <= (TEMP_TRIGGER-TEMP_HISTERESIS))
		} else {
			FanManualMode(sess, int(reading) <= TEMP_TRIGGER)
		}
		time.Sleep(1 * time.Second)
	}
}

func FanManualMode(sess bmc.Session, manualMode bool) error {
	switch CURRENT_STATE {
	case CURRENT_STATE_AUTOMATIC:
		if manualMode == false {
			return nil
		}
		CURRENT_STATE = CURRENT_STATE_UNKNOWN
		if err := FanManualMode(sess, manualMode); err != nil {
			return err
		}
	case CURRENT_STATE_MANUAL:
		if manualMode {
			return nil
		}
		CURRENT_STATE = CURRENT_STATE_UNKNOWN
		if err := FanManualMode(sess, manualMode); err != nil {
			return err
		}
	case CURRENT_STATE_UNKNOWN:
		if err := bmc.ValidateResponse(sess.SendCommand(
			context.Background(),
			&FanModeCommand{
				Req: FanModeRequest{
					Manual: manualMode,
				},
			},
		)); err != nil {
			return err
		}
		if !manualMode {
			slog.Info("state change - mode=automatic")
			CURRENT_STATE = CURRENT_STATE_AUTOMATIC
			return nil
		}
		if err := bmc.ValidateResponse(sess.SendCommand(
			context.Background(),
			&FanSpeedCommand{
				Req: FanSpeedRequest{
					Speed: IDLE_SPEED,
				},
			},
		)); err != nil {
			return err
		}
		slog.Info("state change - mode=manual")
		CURRENT_STATE = CURRENT_STATE_MANUAL
		return nil

	}
	return errors.New("Unknown state")
}

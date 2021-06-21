package main

import (
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/pflag"
	"gitlab.com/gomidi/midi"
	"gitlab.com/gomidi/rtmididrv"
)

type config struct {
	list   bool
	device string
}

func main() {
	err := run()
	if err != nil {
		log.Fatal(err)
	}
	log.Infof("Exiting cleanly.")
}

func initconfig() *config {
	cfg := &config{}
	pflag.BoolVar(&cfg.list, "list", cfg.list, "List input MIDI devices")
	pflag.StringVar(&cfg.device, "device", cfg.device, "Input MIDI device")
	return cfg
}

func list(driver midi.Driver) error {
	inputs, err := driver.Ins()
	if err != nil {
		return err
	}
	for _, in := range inputs {
		log.Infof("MIDI port #%d: %s", in.Number(), in.String())
	}
	return nil
}

func mididevice(device string) (int, string) {
	n, err := strconv.Atoi(device)
	if err == nil {
		return n, ""
	}
	if len(device) == 0 {
		return 0, ""
	}
	return -1, device
}

func callback(data []byte, delta int64) {
	log.Infof("delta %d data %v", delta, data)
}

func run() error {
	cfg := initconfig()
	pflag.Parse()

	driver, err := rtmididrv.New()
	if err != nil {
		return err
	}

	if cfg.list {
		return list(driver)
	}

	port, device := mididevice(cfg.device)

	log.Infof("Using MIDI device #%d (%s)", port, device)

	input, err := midi.OpenIn(driver, port, device)
	if err != nil {
		return fmt.Errorf("open MIDI device: %w", err)
	}

	err = input.SetListener(callback)
	if err != nil {
		return fmt.Errorf("register callback: %w", err)
	}

	log.Infof("Listening for MIDI data...")

	signals := make(chan os.Signal)
	signal.Notify(signals, syscall.SIGINT)
	<-signals

	log.Infof("Closing MIDI port...")

	err = input.Close()
	if err != nil {
		return err
	}

	return nil
}

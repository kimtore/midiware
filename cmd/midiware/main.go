package main

import (
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"github.com/ambientsound/midiware/pkg/translator"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/pflag"
	"gitlab.com/gomidi/midi"
	"gitlab.com/gomidi/midi/reader"
	"gitlab.com/gomidi/midi/writer"
	"gitlab.com/gomidi/rtmididrv"
)

const outputName = "midiware"

type config struct {
	list   bool
	device string
	debug  bool
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
	pflag.BoolVar(&cfg.debug, "debug", cfg.debug, "Print debugging info for in/out data")
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

	if cfg.debug {
		log.SetLevel(log.TraceLevel)
	}

	driver, err := rtmididrv.New()
	if err != nil {
		return err
	}

	if cfg.list {
		return list(driver)
	}

	output, err := driver.OpenVirtualOut(outputName)
	if err != nil {
		return fmt.Errorf("open virtual output MIDI device: %w", err)
	}

	defer output.Close()

	wr := writer.New(output)
	trans := translator.New(wr)

	port, device := mididevice(cfg.device)

	log.Infof("Using MIDI device #%d (%s)", port, device)

	input, err := midi.OpenIn(driver, port, device)
	if err != nil {
		return fmt.Errorf("open MIDI device: %w", err)
	}

	defer input.StopListening()
	defer input.Close()

	//err = input.SetListener(callback)
	rd := reader.New(reader.NoLogger(), reader.Each(trans.Process))
	err = rd.ListenTo(input)
	if err != nil {
		return fmt.Errorf("listen to MIDI channel: %w", err)
	}

	log.Infof("Listening for MIDI data...")

	signals := make(chan os.Signal)
	signal.Notify(signals, syscall.SIGINT)
	sig := <-signals

	log.Infof("Caught signal %d %s", sig, sig)

	return nil
}

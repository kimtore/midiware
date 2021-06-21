package translator

import (
	log "github.com/sirupsen/logrus"
	"gitlab.com/gomidi/midi"
	"gitlab.com/gomidi/midi/reader"
)

type Translator struct {
	w midi.Writer
}

func New(w midi.Writer) *Translator {
	return &Translator{
		w: w,
	}
}

func (t *Translator) Process(pos *reader.Position, msg midi.Message) {
	log.Infof("MIDI input: %s", msg)
	t.w.Write(msg)
}

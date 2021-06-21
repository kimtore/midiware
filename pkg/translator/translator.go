package translator

import (
	log "github.com/sirupsen/logrus"
	"gitlab.com/gomidi/midi"
	"gitlab.com/gomidi/midi/midimessage/channel"
	"gitlab.com/gomidi/midi/reader"
)

type Translator struct {
	w        midi.Writer
	notes    map[uint8]NoteMode
	cc       map[uint8]CcMode
	encoders map[uint8]uint8
	toggles  map[uint8]bool
}

type Note interface {
	midi.Message
	Key() uint8
}

type Cc interface {
	midi.Message
	Controller() uint8
	Value() uint8
}

type NoteMode int

type CcMode int

const (
	NoteDefault NoteMode = iota
	NoteSilence
	NoteToggle

	CcDefault CcMode = iota
	CcEncoder
)

const encoderThreshold = 64

const maxValue = 127

func New(w midi.Writer) *Translator {
	return &Translator{
		w: w,
		notes: map[uint8]NoteMode{
			// Touch-sensitive encoder knobs
			9:  NoteSilence,
			10: NoteSilence,
			0:  NoteSilence,
			1:  NoteSilence,
			2:  NoteSilence,
			3:  NoteSilence,
			4:  NoteSilence,
			5:  NoteSilence,
			6:  NoteSilence,
			7:  NoteSilence,
			8:  NoteSilence,

			// Top 8x4 grid used for toggles
			68: NoteToggle,
			69: NoteToggle,
			70: NoteToggle,
			71: NoteToggle,
			72: NoteToggle,
			73: NoteToggle,
			74: NoteToggle,
			75: NoteToggle,
			76: NoteToggle,
			77: NoteToggle,
			78: NoteToggle,
			79: NoteToggle,
			80: NoteToggle,
			81: NoteToggle,
			82: NoteToggle,
			83: NoteToggle,
			84: NoteToggle,
			85: NoteToggle,
			86: NoteToggle,
			87: NoteToggle,
			88: NoteToggle,
			89: NoteToggle,
			90: NoteToggle,
			91: NoteToggle,
			92: NoteToggle,
			93: NoteToggle,
			94: NoteToggle,
			95: NoteToggle,
			96: NoteToggle,
			97: NoteToggle,
			98: NoteToggle,
			99: NoteToggle,
		},
		cc: map[uint8]CcMode{
			14: CcEncoder,
			15: CcEncoder,
			71: CcEncoder,
			72: CcEncoder,
			73: CcEncoder,
			74: CcEncoder,
			75: CcEncoder,
			76: CcEncoder,
			77: CcEncoder,
			78: CcEncoder,
			79: CcEncoder,
		},
		toggles:  map[uint8]bool{},
		encoders: map[uint8]uint8{},
	}
}

func (t *Translator) TranslateNote(note Note) midi.Message {
	key := note.Key()
	mode := t.notes[key]
	switch mode {
	default:
		fallthrough
	case NoteDefault:
		return note
	case NoteSilence:
		return nil
	case NoteToggle:
		_, ok := note.(channel.NoteOn)
		if !ok {
			// Toggle only on NoteOn
			return nil
		}
		t.toggles[key] = !t.toggles[key]
		log.Debugf("Setting on/off toggle %d to %v", key, t.toggles[key])
		if t.toggles[key] {
			return channel.Channel0.NoteOn(key, maxValue)
		} else {
			return channel.Channel0.NoteOff(key)
		}
	}
}

func (t *Translator) TranslateEncoder(cc Cc) midi.Message {
	ctrl := cc.Controller()
	mode := t.cc[ctrl]
	switch mode {
	default:
		fallthrough
	case CcDefault:
		return cc
	case CcEncoder:
		value := cc.Value()
		if value < encoderThreshold && t.encoders[ctrl] < maxValue {
			t.encoders[ctrl]++
		} else if value > encoderThreshold && t.encoders[ctrl] > 0 {
			t.encoders[ctrl]--
		}
		log.Debugf("Setting encoder %d to %d", ctrl, t.encoders[ctrl])
		return channel.Channel0.ControlChange(ctrl, t.encoders[ctrl])
	}
}

func (t *Translator) Process(pos *reader.Position, msg midi.Message) {
	var out midi.Message

	log.Tracef("MIDI input: %s", msg)

	switch m := msg.(type) {
	case channel.NoteOn:
		out = t.TranslateNote(m)
	case channel.NoteOff:
		out = t.TranslateNote(m)
	case channel.NoteOffVelocity:
		out = t.TranslateNote(m)
	case channel.ControlChange:
		out = t.TranslateEncoder(m)
	default:
		// pass-through unknown messages
		out = msg
	}

	if out == nil {
		log.Tracef("No output")
		return
	}

	log.Tracef("MIDI output: %s", out)

	err := t.w.Write(out)
	if err != nil {
		log.Errorf("Error writing MIDI: %s", err)
	}
}

# midiware

MIDI middleware written in Go.

It gives you the opportunity to remap MIDI note or cc messages as follows:

- notes can be silenced completely, or push to toggle on/off
- cc messages can be interpreted as rotary encoders

The current state of this program works for the Ableton Push 2.

Configuration must be done in [source code](pkg/translator/translator.go).

Released under the MIT license.

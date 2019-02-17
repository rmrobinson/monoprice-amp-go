package monoprice

import (
	"errors"
	"fmt"
)

var (
	// ErrUnsupportedRange is returned if the requested parameter exceeds the supported range of inputs.
	ErrUnsupportedRange = errors.New("unsupported range")
)

// Zone represents a single output on a receiver.
type Zone struct {
	a  amplifier
	id int

	state *State
}

// newZone creates a new zone using the specified amplifier, ID and state string.
// This is not public to ensure that the amplifier implementation properly queries
// for the relevant data to create the zone with.
func newZone(a amplifier, id int, line string) (*Zone, error) {
	z := &Zone{
		a:  a,
		id: id,
	}

	err := z.update(line)
	return z, err
}

// ID returns the control ID of this zone.
func (z *Zone) ID() string {
	return fmt.Sprintf("%d%d", z.a.ID(), z.id)
}

// Refresh queries the amplifier for the current state of each field.
func (z *Zone) Refresh() error {
	cmd := fmt.Sprintf("%s%s\r", queryRequestPrefix, z.ID())

	// Synchronize writes to the amp to avoid reading back interleaved output
	z.a.lock()
	defer z.a.unlock()

	err := z.a.execute(cmd)
	if err != nil {
		return err
	}

	line, err := z.a.read()
	if err != nil {
		return err
	}

	return z.update(line)
}

// State returns a cached state of this zone.
// This will be in sync if all changes are made via this amp instance,
// however changes made by the wall controllers will not be reflected here.
func (z *Zone) State() *State {
	return z.state
}

// SetPower applies the supplied isOn state to the zone.
func (z *Zone) SetPower(on bool) error {
	powerValue := 0
	if on {
		powerValue = 1
	}

	err := z.set(power, powerValue)
	if err != nil {
		return err
	}

	z.state.IsOn = on
	return nil
}

// SetMute applies the supplied isMuteOn state to the zone.
func (z *Zone) SetMute(on bool) error {
	muteValue := 0
	if on {
		muteValue = 1
	}

	err := z.set(mute, muteValue)
	if err != nil {
		return err
	}

	z.state.IsMuteOn = on
	return nil
}

// SetVolume applies the supplied volume level to the zone.
func (z *Zone) SetVolume(level int) error {
	if level < 0 || level > 38 {
		return ErrUnsupportedRange
	}

	err := z.set(volume, level)
	if err != nil {
		return err
	}

	z.state.Volume = level
	return nil
}

// SetTreble applies the supplied treble level to the zone.
func (z *Zone) SetTreble(level int) error {
	if level < 0 || level > 14 {
		return ErrUnsupportedRange
	}

	err := z.set(treble, level)
	if err != nil {
		return err
	}

	z.state.Volume = level
	return nil
}

// SetBass applies the supplied bass level to the zone.
func (z *Zone) SetBass(level int) error {
	if level < 0 || level > 14 {
		return ErrUnsupportedRange
	}

	err := z.set(bass, level)
	if err != nil {
		return err
	}

	z.state.Volume = level
	return nil
}

// SetBalance applies the supplied balance level to the zone.
func (z *Zone) SetBalance(level int) error {
	if level < 0 || level > 38 {
		return ErrUnsupportedRange
	}

	err := z.set(balance, level)
	if err != nil {
		return err
	}

	z.state.Volume = level
	return nil
}

// SetSourceChannel changes the source input used by the zone.
func (z *Zone) SetSourceChannel(channelID int) error {
	if channelID < 1 || channelID > 6 {
		return ErrUnsupportedRange
	}

	err := z.set(sourceChannel, channelID)
	if err != nil {
		return err
	}

	z.state.SourceChannelID = channelID
	return nil
}

// set applies the request action to the zone.
func (z *Zone) set(ac actionCode, actionValue int) error {
	// Synchronize writes to the amp to avoid reading back interleaved output
	z.a.lock()
	defer z.a.unlock()

	cmd := fmt.Sprintf("%s%s%s%02d\r", controlRequestPrefix, z.ID(), ac, actionValue)
	return z.a.execute(cmd)
}

// update takes the supplied read line and updates the state of the zone.
func (z *Zone) update(line string) error {
	state := &State{}

	err := state.parse(line)
	if err != nil {
		return err
	}

	z.state = state
	return err
}

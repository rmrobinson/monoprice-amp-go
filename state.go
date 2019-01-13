package monoprice

import (
	"errors"
	"strconv"
	"strings"
)

var (
	// ErrInvalidCommandCode is returned if an invalid command code is detected
	ErrInvalidCommandCode = errors.New("invalid command code")
	// ErrInvalidInput is detected if the command string format doesn't match as expected
	ErrInvalidInput       = errors.New("invalid command string")
)

// State contains the current state of a given zone, as retrieved from the controller.
type State struct {
	IsOn              bool
	IsMuteOn          bool
	IsDoNotDisturbOn  bool
	IsPAOn            bool
	Volume            int
	Treble            int
	Bass              int
	Balance           int
	SourceChannelID   int
	IsKeypadConnected bool
}

// parse takes a line and converts it into a state struct.
// It takes either a full status line returned by inquiry command structure (1)
// or a single-field status line returned by inquiry command structure (2)
// This will update the existing state struct with any included fields.
// A single-field status line will not clear all the other fields.
func (s *State) parse(line string) error {
	// The output seems to end with two carriage returns and then one line feed (vs. a single CR)
	// It also starts with #> on OS X and just > on Linux, as documented. Clean it up before using.
	line = strings.TrimSuffix(line, "\r\r\n")
	line = strings.TrimPrefix(line, "#>")
	line = strings.TrimPrefix(line, ">")

	var err error
	if len(line) == query1ResponseLength {
		s.IsPAOn, err = s.parseBool(line[2:4])
		if err != nil {
			return err
		}

		s.IsOn, err = s.parseBool(line[4:6])
		if err != nil {
			return err
		}

		s.IsMuteOn, err = s.parseBool(line[6:8])
		if err != nil {
			return err
		}

		s.IsDoNotDisturbOn, err = s.parseBool(line[8:10])
		if err != nil {
			return err
		}

		s.Volume, err = s.parseInt(line[10:12])
		if err != nil {
			return err
		}

		s.Treble, err = s.parseInt(line[12:14])
		if err != nil {
			return err
		}

		s.Bass, err = s.parseInt(line[14:16])
		if err != nil {
			return err
		}

		s.Balance, err = s.parseInt(line[16:18])
		if err != nil {
			return err
		}

		s.SourceChannelID, err = s.parseInt(line[18:20])
		if err != nil {
			return err
		}

		s.IsKeypadConnected, err = s.parseBool(line[20:22])
		if err != nil {
			return err
		}

		return nil
	} else if len(line) == query2ResponseLength {
		switch line[2:4] {
		case string(pa):
			s.IsPAOn, err = s.parseBool(line[4:6])
			return err
		case string(power):
			s.IsOn, err = s.parseBool(line[4:6])
			return err
		case string(mute):
			s.IsMuteOn, err = s.parseBool(line[4:6])
			return err
		case string(doNotDisturb):
			s.IsDoNotDisturbOn, err = s.parseBool(line[4:6])
			return err
		case string(volume):
			s.Volume, err = s.parseInt(line[4:6])
			return err
		case string(treble):
			s.Treble, err = s.parseInt(line[4:6])
			return err
		case string(bass):
			s.Bass, err = s.parseInt(line[4:6])
			return err
		case string(balance):
			s.Balance, err = s.parseInt(line[4:6])
			return err
		case string(sourceChannel):
			s.SourceChannelID, err = s.parseInt(line[4:6])
			return err
		case string(keypadConnectingStatus):
			s.IsKeypadConnected, err = s.parseBool(line[4:6])
			return err
		default:
			return ErrInvalidCommandCode
		}
	}

	return ErrInvalidInput
}

func (s *State) parseBool(data string) (bool, error) {
	val, err := strconv.ParseInt(data, 10, 0)
	if err != nil {
		return false, nil
	}

	return val == 1, nil
}

func (s *State) parseInt(data string) (int, error) {
	val, err := strconv.ParseInt(data, 10, 32)
	if err != nil {
		return -1, nil
	}

	return int(val), nil
}

package monoprice

import (
	"errors"
	"testing"
)

var (
	errExec = errors.New("sample exec error")
	errRead = errors.New("sample read error")
)

var zoneRefreshTests = []struct {
	data      string
	execErr   error
	readErr   error
	result    State
	resultErr error
}{
	{"#>1100010000120707070101\r\r\n",
		nil,
		nil,
		State{
			IsPAOn:            false,
			IsOn:              true,
			IsDoNotDisturbOn:  false,
			IsMuteOn:          false,
			Volume:            12,
			Treble:            7,
			Bass:              7,
			Balance:           7,
			SourceChannelID:   1,
			IsKeypadConnected: true,
		},
		nil,
	},
	{"#>1100010000120707070101\r\r\n",
		errExec,
		nil,
		State{
			IsPAOn:            false,
			IsOn:              false,
			IsDoNotDisturbOn:  false,
			IsMuteOn:          false,
			Volume:            0,
			Treble:            0,
			Bass:              0,
			Balance:           0,
			SourceChannelID:   0,
			IsKeypadConnected: false,
		},
		errExec,
	},
	{"#>1100010000120707070101\r\r\n",
		nil,
		errRead,
		State{
			IsPAOn:            false,
			IsOn:              false,
			IsDoNotDisturbOn:  false,
			IsMuteOn:          false,
			Volume:            0,
			Treble:            0,
			Bass:              0,
			Balance:           0,
			SourceChannelID:   0,
			IsKeypadConnected: false,
		},
		errRead,
	},
	{"invalid data",
		nil,
		nil,
		State{
			IsPAOn:            false,
			IsOn:              false,
			IsDoNotDisturbOn:  false,
			IsMuteOn:          false,
			Volume:            0,
			Treble:            0,
			Bass:              0,
			Balance:           0,
			SourceChannelID:   0,
			IsKeypadConnected: false,
		},
		ErrInvalidInput,
	},
}

func TestZoneRefresh(t *testing.T) {
	amp := &testAmp{}
	for _, tt := range zoneRefreshTests {
		t.Run(tt.data, func(t *testing.T) {
			amp.readErr = tt.readErr
			amp.execErr = tt.execErr
			amp.data = tt.data

			z, _ := newZone(amp, 1, "")
			err := z.Refresh()
			if err != tt.resultErr {
				t.Errorf("got %s, expected %s", err, tt.resultErr)
			}
			if err != nil {
				return
			}

			if *z.state != tt.result {
				t.Errorf("got %v, want %v", z.state, tt.result)
			}
		})
	}
}

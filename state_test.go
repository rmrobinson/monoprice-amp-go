package monoprice

import "testing"

var stateTests = []struct {
	input  string
	result State
	err    error
}{
	{"#>1100010000120707070101\r\r\n",
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
	{"#>1100010000120708070101\r\r\n",
		State{
			IsPAOn:            false,
			IsOn:              true,
			IsDoNotDisturbOn:  false,
			IsMuteOn:          false,
			Volume:            12,
			Treble:            7,
			Bass:              8,
			Balance:           7,
			SourceChannelID:   1,
			IsKeypadConnected: true,
		},
		nil,
	}, {"#>11VO14\r\r\n",
		State{
			IsPAOn:            false,
			IsOn:              false,
			IsDoNotDisturbOn:  false,
			IsMuteOn:          false,
			Volume:            14,
			Treble:            0,
			Bass:              0,
			Balance:           0,
			SourceChannelID:   0,
			IsKeypadConnected: false,
		},
		nil,
	},
	{"#>11HI14\r\r\n",
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
		ErrInvalidCommandCode,
	},
	{">11\r\r\n",
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

func TestStateParser(t *testing.T) {
	for _, tt := range stateTests {
		t.Run(tt.input, func(t *testing.T) {
			var s State
			err := s.parse(tt.input)
			if err != tt.err {
				t.Errorf("got %s, expected %s", err, tt.err)
			} else if s != tt.result {
				t.Errorf("got %v, want %v", s, tt.result)
			}
		})
	}
}

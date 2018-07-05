package monoprice

import (
	"bufio"
	"errors"
	"fmt"
	"github.com/tarm/serial"
	"log"
	"strings"
)

const (
	controlRequestPrefix = "<"
	queryRequestPrefix   = "?"
	query1ResponseLength = 22
	query2ResponseLength = 6
)

type actionCode string

const (
	power                  actionCode = "PR"
	mute                              = "MU"
	doNotDisturb                      = "DT"
	volume                            = "VO"
	treble                            = "TR"
	bass                              = "BS"
	balance                           = "BL"
	sourceChannel                     = "CH"
	keypadConnectingStatus            = "LS"
	pa                                = "PA"
)

// amplifier defines the required methods for an implementation of an amplifier.
// Currently only used for testing.
type amplifier interface {
	ID() int
	execute(string) error
	read() (string, error)
}

// SerialAmplifier is an implementation of the Monoprice amplifier backed by a serial port.
type SerialAmplifier struct {
	zones map[int]*Zone
	id    int

	port *serial.Port
}

// NewSerialAmplifier creates a new serial amplifier using the supplied serial port.
// If the amplifier cannot be queried (i.e. if the port is not ready) an error will be returned.
func NewSerialAmplifier(port *serial.Port) (*SerialAmplifier, error) {
	ret := &SerialAmplifier{
		zones: map[int]*Zone{},
		port:  port,
		id:    1,
	}

	cmd := fmt.Sprintf("%s%d0\r", queryRequestPrefix, ret.id)

	err := ret.execute(cmd)
	if err != nil {
		return nil, err
	}

	for i := 1; i <= 6; i++ {
		line, err := ret.read()
		if err != nil {
			log.Printf("Error reading line: %s\n", err.Error())
		}

		z, err := newZone(ret, i, line)
		if err != nil {
			log.Printf("Error creating zone: %s\n", err.Error())
			continue
		}

		ret.zones[i] = z
	}

	return ret, nil
}

// ID returns the ID (1-3) of this amplifier.
func (a *SerialAmplifier) ID() int {
	return a.id
}

// Zone retrieves the cached state of the specified zone.
// If the underlying zone may have changed (using a wall controller, for example),
// then refresh should be called on the returned zone before using the data.
func (a *SerialAmplifier) Zone(id int) *Zone {
	if zone, ok := a.zones[id]; ok {
		return zone
	}

	return nil
}

// execute handles the logic of writing to the serial port and reading back the echoed command.
func (a *SerialAmplifier) execute(command string) error {
	wroteCount, err := a.port.Write([]byte(command))
	if err != nil {
		return err
	}

	// Read back the echoed command
	reader := bufio.NewReader(a.port)

	read, err := reader.ReadString('\n')
	if err != nil {
		return err
	}

	read = strings.TrimSuffix(read, "\n")
	// Commands seem to be read back with a 'commented out' version of it.
	read = strings.TrimPrefix(read, "#")

	if len(read) != wroteCount {
		log.Printf("read '%x', wrote '%x'\n", read, command)
		return errors.New("read back different length than wrote")
	} else if read != command {
		log.Printf("read '%x', wrote '%x'\n", read, command)
		return errors.New("read back different string than command")
	}

	return nil
}

// read retrieves the next line available on the serial port.
// It is the caller's responsibility to know how many times it may be necessary to call
// based upon the previously sent command; this will block if there is nothing to read.
func (a *SerialAmplifier) read() (string, error) {
	reader := bufio.NewReader(a.port)

	// Read the response to the command
	read, err := reader.ReadString('\n')

	return read, err
}

package monoprice

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"strings"
	"sync"

	"github.com/tarm/serial"
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

// ErrTimeout is returned if the amplifier doesn't respond in time. The serial port must be opened
// with a non-zero ReadTimeout for this to be detected; otherwise a silent amplifier blocks forever.
var ErrTimeout = errors.New("timed out waiting for amplifier")

// serialPort is the subset of *serial.Port used by SerialAmplifier.
type serialPort interface {
	io.ReadWriter
	Flush() error
}

// timeoutReader turns the zero-byte, no-error read that a serial port with a ReadTimeout
// returns on expiry into ErrTimeout, which bufio would otherwise retry (and eventually
// report as io.ErrNoProgress).
type timeoutReader struct {
	r io.Reader
}

func (t timeoutReader) Read(p []byte) (int, error) {
	n, err := t.r.Read(p)
	if n == 0 && err == nil {
		return 0, ErrTimeout
	}
	return n, err
}

// amplifier defines the required methods for an implementation of an amplifier.
// Currently only used for testing.
type amplifier interface {
	ID() int
	execute(string) error
	read() (string, error)

	lock()
	unlock()
}

// SerialAmplifier is an implementation of the Monoprice amplifier backed by a serial port.
type SerialAmplifier struct {
	zones map[int]*Zone
	id    int

	port serialPort
	// reader is shared across all reads; a per-call reader would discard any bytes it had
	// buffered beyond the first line, corrupting the next response.
	reader   *bufio.Reader
	portLock sync.Mutex
}

// NewSerialAmplifier creates a new serial amplifier using the supplied serial port.
// If the amplifier cannot be queried (i.e. if the port is not ready) an error will be returned.
// The port should be opened with a ReadTimeout (see serial.Config) so a stalled amplifier
// results in ErrTimeout rather than a hang.
func NewSerialAmplifier(port *serial.Port) (*SerialAmplifier, error) {
	return newSerialAmplifier(port)
}

func newSerialAmplifier(port serialPort) (*SerialAmplifier, error) {
	ret := &SerialAmplifier{
		zones:  map[int]*Zone{},
		port:   port,
		reader: bufio.NewReader(timeoutReader{port}),
		id:     1,
	}

	err := ret.setup()
	if err != nil {
		return nil, err
	}

	return ret, nil
}

func (a *SerialAmplifier) lock() {
	a.portLock.Lock()
}

func (a *SerialAmplifier) unlock() {
	a.portLock.Unlock()
}

func (a *SerialAmplifier) setup() error {
	cmd := fmt.Sprintf("%s%d0\r", queryRequestPrefix, a.id)
	err := a.execute(cmd)
	if err != nil {
		return err
	}

	for i := 1; i <= 6; i++ {
		line, err := a.read()
		if err != nil {
			return err
		}

		z, err := newZone(a, i, line)
		if err != nil {
			continue
		}

		a.zones[i] = z
	}

	return nil
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

// Reset is used to clear the underlying serial port and reset things to a good state.
// This may be used if errors are detected on the port to clear any oddities being read.
// Any existing zone references will be invalidated and should not be used anymore.
func (a *SerialAmplifier) Reset() error {
	a.lock()
	defer a.unlock()

	err := a.port.Flush()
	if err != nil {
		return err
	}

	a.reader.Reset(timeoutReader{a.port})
	a.zones = map[int]*Zone{}
	return a.setup()
}

// execute handles the logic of writing to the serial port and reading back the echoed command.
func (a *SerialAmplifier) execute(command string) error {
	wroteCount, err := a.port.Write([]byte(command))
	if err != nil {
		return err
	}

	// Read back the echoed command
	read, err := a.readLine()
	if err != nil {
		return err
	}

	read = strings.TrimSuffix(read, "\n")
	// Commands seem to be read back with a 'commented out' version of it.
	read = strings.TrimPrefix(read, "#")

	if len(read) != wroteCount {
		return errors.New("read back different length than wrote")
	} else if read != command {
		return errors.New("read back different string than command")
	}

	return nil
}

// read retrieves the next line available on the serial port.
// It is the caller's responsibility to know how many times it may be necessary to call
// based upon the previously sent command; this will block if there is nothing to read.
func (a *SerialAmplifier) read() (string, error) {
	return a.readLine()
}

// readLine reads up to the next line feed. On failure any partial line and stale buffered
// input are discarded so the next command starts from a clean stream.
func (a *SerialAmplifier) readLine() (string, error) {
	line, err := a.reader.ReadString('\n')
	if err != nil {
		a.reader.Reset(timeoutReader{a.port})
		_ = a.port.Flush()
		return "", err
	}
	return line, nil
}

func newReader(port io.Reader) *bufio.Reader {
	return bufio.NewReader(timeoutReader{port})
}

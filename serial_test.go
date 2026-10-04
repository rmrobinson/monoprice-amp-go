package monoprice

import (
	"bytes"
	"io"
	"testing"
)

// fakePort serves canned reads; once exhausted it behaves like a serial port with a
// ReadTimeout (zero bytes, no error).
type fakePort struct {
	in      bytes.Buffer
	written bytes.Buffer
	flushed int
}

func (f *fakePort) Read(p []byte) (int, error)  { return f.in.Read(p) }
func (f *fakePort) Write(p []byte) (int, error) { return f.written.Write(p) }
func (f *fakePort) Flush() error                { f.flushed++; return nil }

// emptyIsTimeout makes bytes.Buffer's io.EOF look like a serial read timeout.
type timeoutPort struct{ *fakePort }

func (t timeoutPort) Read(p []byte) (int, error) {
	n, err := t.fakePort.Read(p)
	if err == io.EOF {
		return 0, nil
	}
	return n, err
}

func TestReadsShareBufferedBytes(t *testing.T) {
	fp := &fakePort{}
	// Two lines arrive in one chunk, as happens when the amp answers a whole-amp query.
	fp.in.WriteString("#?10\r\n#>1100010000120707070101\r\r\n#>1200010000120707070102\r\r\n")
	a := &SerialAmplifier{id: 1, port: timeoutPort{fp}}
	a.reader = newReader(a.port)

	if err := a.execute("?10\r"); err != nil {
		t.Fatal(err)
	}
	for i, want := range []string{"#>1100010000120707070101\r\r\n", "#>1200010000120707070102\r\r\n"} {
		got, err := a.read()
		if err != nil || got != want {
			t.Fatalf("line %d: got %q, %v; want %q", i, got, err, want)
		}
	}
}

func TestReadTimeout(t *testing.T) {
	fp := &fakePort{}
	a := &SerialAmplifier{id: 1, port: timeoutPort{fp}}
	a.reader = newReader(a.port)

	if _, err := a.read(); err != ErrTimeout {
		t.Fatalf("got %v, want ErrTimeout", err)
	}
	if fp.flushed != 1 {
		t.Fatalf("port flushed %d times after timeout, want 1", fp.flushed)
	}
}

func TestSettersUpdateOnlyTheirField(t *testing.T) {
	z := &Zone{a: &testAmp{}, id: 1, state: &State{Volume: 12}}
	if err := z.SetTreble(3); err != nil {
		t.Fatal(err)
	}
	if err := z.SetBass(4); err != nil {
		t.Fatal(err)
	}
	if err := z.SetBalance(20); err != nil {
		t.Fatal(err)
	}
	want := State{Volume: 12, Treble: 3, Bass: 4, Balance: 20}
	if *z.state != want {
		t.Fatalf("got %+v, want %+v", *z.state, want)
	}
	if z.SetBalance(21) != ErrUnsupportedRange {
		t.Fatal("balance 21 accepted")
	}
}

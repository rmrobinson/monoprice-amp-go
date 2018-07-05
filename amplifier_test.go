package monoprice

type testAmp struct {
	execErr error
	readErr error
	data    string
}

func (t *testAmp) ID() int {
	return 512
}
func (t *testAmp) execute(cmd string) error {
	return t.execErr
}
func (t *testAmp) read() (string, error) {
	return t.data, t.readErr
}

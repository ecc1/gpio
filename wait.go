package gpio

import (
	"fmt"
	"time"

	"golang.org/x/sys/unix"
)

// TimeoutError indicates that a Wait operation on a GPIO input pin timed out.
type TimeoutError struct {
	pin     *Pin
	timeout time.Duration
}

func (t TimeoutError) Error() string {
	return fmt.Sprintf("gpio%d.Wait timeout after %v", t.pin.number, t.timeout)
}

// This must be long enough to read the entire value file (0 or 1 and newline).
var valueBuf = make([]byte, 4)

// Wait waits with the given timeout for a GPIO input pin to become active.
func (pin *Pin) Wait(timeout time.Duration) error {
	fd, err := unix.Open(pin.value, unix.O_NONBLOCK|unix.O_RDONLY, 0)
	if err != nil {
		return err
	}
	defer func() { _ = unix.Close(fd) }()
	_, err = unix.Read(fd, valueBuf)
	// Return immediately if the value is already active.
	if err != nil || valueBuf[0] == '1' {
		return err
	}
	fds := []unix.PollFd{{Fd: int32(fd), Events: unix.POLLPRI}}
	n, err := unix.Poll(fds, int(timeout/time.Millisecond))
	if err != nil {
		return err
	}
	switch n {
	case 1:
		return nil
	case 0:
		return TimeoutError{pin: pin, timeout: timeout}
	default:
		return fmt.Errorf("gpio%d.Select returned %d", pin.number, n)
	}
}

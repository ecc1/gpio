package gpio

import (
	"fmt"
	"os"
	"syscall"
	"time"
)

type TimeoutError struct {
	pin     *Pin
	timeout time.Duration
}

func (t TimeoutError) Error() string {
	return fmt.Sprintf("gpio%d.Wait timeout after %v", t.pin.number, t.timeout)
}

func (pin *Pin) Wait(timeout time.Duration) error {
	f, err := os.Open(pin.value)
	if err != nil {
		return err
	}
	defer f.Close()
	fd := f.Fd()
	nfds := int(fd) + 1
	var eset syscall.FdSet
	eset.Bits[fd/64] = 1 << (fd % 64)
	t := (*syscall.Timeval)(nil)
	ns := timeout.Nanoseconds()
	if ns != -1 {
		tv := syscall.NsecToTimeval(ns)
		t = &tv
	}
	var buf [4]byte
	f.Read(buf[:]) // prevent Select from returning immediately
	n, err := syscall.Select(nfds, nil, nil, &eset, t)
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

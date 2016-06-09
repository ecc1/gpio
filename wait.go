package gpio

import (
	"fmt"
	"os"
	"syscall"
)

func (pin *Pin) Wait() error {
	f, err := os.Open(pin.value)
	if err != nil {
		return err
	}
	defer f.Close()
	fd := f.Fd()
	nfds := int(fd) + 1
	var eset syscall.FdSet
	eset.Bits[fd/64] = 1 << (fd % 64)
	var buf [4]byte
	f.Read(buf[:]) // prevent Select from returning immediately
	n, err := syscall.Select(nfds, nil, nil, &eset, nil)
	if err != nil {
		return err
	}
	if n != 1 {
		return fmt.Errorf("gpio%d.Select returned %d", pin.number, n)
	}
	return nil
}

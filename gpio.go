package gpio

import (
	"fmt"
	"io/ioutil"
	"os"
	"time"
)

type InputPin interface {
	Read() (bool, error)
	Wait(time.Duration) error
}

type OutputPin interface {
	Write(bool) error
}

type Pin struct {
	number int
	dir    string
	value  string
}

func Input(pinNumber int, edge string, activeLow bool) (InputPin, error) {
	pin, err := newPin(pinNumber, activeLow)
	if err != nil {
		return nil, err
	}
	err = writeFile(fmt.Sprintf("%s/direction", pin.dir), "in")
	if err != nil {
		return nil, err
	}
	err = writeFile(fmt.Sprintf("%s/edge", pin.dir), edge)
	if err != nil {
		return nil, err
	}
	return pin, nil
}

func Output(pinNumber int, activeLow bool) (OutputPin, error) {
	pin, err := newPin(pinNumber, activeLow)
	if err != nil {
		return nil, err
	}
	err = writeFile(fmt.Sprintf("%s/direction", pin.dir), "out")
	if err != nil {
		return nil, err
	}
	return pin, nil
}

func (pin *Pin) Read() (bool, error) {
	return readBoolFile(pin.value)
}

func fileExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.Mode().IsRegular()
}

func directoryExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.Mode().IsDir()
}

func pinDirectory(pinNumber int) (dir string, err error) {
	const gpioDir = "/sys/class/gpio"
	dir = fmt.Sprintf("%s/gpio%d", gpioDir, pinNumber)
	if !directoryExists(dir) {
		err = writeFile(fmt.Sprintf("%s/export", gpioDir), fmt.Sprintf("%d", pinNumber))
		if err != nil {
			return
		}
	}
	if !directoryExists(dir) {
		err = fmt.Errorf("failed to export GPIO directory %s", dir)
	}
	return
}

func newPin(pinNumber int, activeLow bool) (*Pin, error) {
	dir, err := pinDirectory(pinNumber)
	if err != nil {
		return nil, err
	}
	value := fmt.Sprintf("%s/value", dir)
	if !fileExists(value) {
		return nil, fmt.Errorf("%s does not exist", value)
	}
	err = writeBoolFile(fmt.Sprintf("%s/active_low", dir), activeLow)
	if err != nil {
		return nil, err
	}
	return &Pin{number: pinNumber, dir: dir, value: value}, nil
}

func readFile(file string) (string, error) {
	v, err := ioutil.ReadFile(file)
	return string(v), err
}

func readBoolFile(file string) (bool, error) {
	v, err := readFile(file)
	if err != nil {
		return false, err
	}
	// compare without trailing '\n'
	s := v[:len(v)-1]
	switch s {
	case "0":
		return false, nil
	case "1":
		return true, nil
	default:
		return false, fmt.Errorf("read %s from %s instead of boolean value", s, file)
	}
}

func (pin *Pin) Write(value bool) error {
	return writeBoolFile(pin.value, value)
}

func writeFile(file string, contents string) error {
	return ioutil.WriteFile(file, []byte(contents), 0644)
}

func writeBoolFile(file string, value bool) error {
	if value {
		return writeFile(file, "1")
	} else {
		return writeFile(file, "0")
	}
}

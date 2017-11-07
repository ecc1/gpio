package gpio

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"time"
)

type (
	// InputPin is the interface satisfied by GPIO input pins.
	InputPin interface {
		Read() (bool, error)
	}

	// InterruptPin is the interface satisfied by GPIO interrupt pins.
	InterruptPin interface {
		InputPin
		Wait(time.Duration) error
	}

	// OutputPin is the interface satisfied by GPIO output pins.
	OutputPin interface {
		Write(bool) error
	}

	// Pin represents a GPIO pin.
	Pin struct {
		number int
		dir    string
		value  string
	}
)

// Input initializes a GPIO input pin with the given pin number.
func Input(pinNumber int, activeLow bool) (InputPin, error) {
	pin, err := newPin(pinNumber, activeLow)
	if err != nil {
		return nil, err
	}
	err = writeFile(path.Join(pin.dir, "direction"), "in")
	return pin, err
}

// Interrupt initializes a GPIO interrupt pin with the given pin number.
// The edge parameter must be "rising", "falling", or "both".
func Interrupt(pinNumber int, activeLow bool, edge string) (InterruptPin, error) {
	pin, err := newPin(pinNumber, activeLow)
	if err != nil {
		return nil, err
	}
	err = writeFile(path.Join(pin.dir, "direction"), "in")
	if err != nil {
		return pin, err
	}
	err = writeFile(path.Join(pin.dir, "edge"), edge)
	return pin, err
}

var gpioDirection = map[bool]string{true: "high", false: "low"}

// Output initializes a GPIO output pin with the given pin number
// and initial logical value.
func Output(pinNumber int, activeLow bool, initialValue bool) (OutputPin, error) {
	pin, err := newPin(pinNumber, activeLow)
	if err != nil {
		return nil, err
	}
	// Set direction based on initial *logical* value.
	direction := gpioDirection[initialValue != activeLow]
	err = writeFile(path.Join(pin.dir, "direction"), direction)
	return pin, err
}

func (pin *Pin) Read() (bool, error) {
	return readBoolFile(pin.value)
}

func fileExists(path string) (bool, error) {
	return existsWithPredicate(path, func(info os.FileInfo) bool {
		return info.Mode().IsRegular()
	})
}

func directoryExists(path string) (bool, error) {
	return existsWithPredicate(path, func(info os.FileInfo) bool {
		return info.Mode().IsDir()
	})
}

func existsWithPredicate(path string, predicate func(os.FileInfo) bool) (bool, error) {
	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}
	return predicate(info), nil
}

func pinDirectory(pinNumber int) (string, error) {
	const gpioDir = "/sys/class/gpio/"
	dir := path.Join(gpioDir, fmt.Sprintf("gpio%d/", pinNumber))
	tried := false
	for {
		exists, err := directoryExists(dir)
		if err != nil || exists {
			return dir, err
		}
		if tried {
			return dir, fmt.Errorf("failed to export GPIO directory %s", dir)
		}
		err = writeFile(path.Join(gpioDir, "export"), fmt.Sprintf("%d", pinNumber))
		if err != nil {
			return dir, err
		}
		tried = true
		// Give udev rules a chance to execute on newly-created gpio%d directory.
		time.Sleep(time.Second)
	}
}

func newPin(pinNumber int, activeLow bool) (*Pin, error) {
	dir, err := pinDirectory(pinNumber)
	if err != nil {
		return nil, err
	}
	value := path.Join(dir, "value")
	exists, err := fileExists(value)
	if err != nil || !exists {
		return nil, err
	}
	err = writeBoolFile(path.Join(dir, "active_low"), activeLow)
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
	var b string
	switch value {
	case true:
		b = "1"
	case false:
		b = "0"
	}
	return writeFile(file, b)
}

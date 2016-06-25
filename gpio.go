package gpio

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
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
	err = writeFile(path.Join(pin.dir, "direction"), "in")
	if err != nil {
		return nil, err
	}
	err = writeFile(path.Join(pin.dir, "edge"), edge)
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
	err = writeFile(path.Join(pin.dir, "direction"), "out")
	if err != nil {
		return nil, err
	}
	return pin, nil
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
	if value {
		return writeFile(file, "1")
	} else {
		return writeFile(file, "0")
	}
}

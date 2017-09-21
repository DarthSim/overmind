package term

import (
	"os"
	"unsafe"

	"golang.org/x/sys/unix"
)

// Size represents terminal size
type Size struct {
	Rows    uint16
	Cols    uint16
	Xpixels uint16
	Ypixels uint16
}

// Params contains terminal state and size
type Params struct {
	termios unix.Termios
	ws      Size
}

func ioctl(fd, request, argp uintptr) error {
	if _, _, err := unix.Syscall6(unix.SYS_IOCTL, fd, request, argp, 0, 0, 0); err != 0 {
		return err
	}
	return nil
}

func getTermios(f *os.File) (t unix.Termios, err error) {
	err = ioctl(f.Fd(), ioctlReadTermios, uintptr(unsafe.Pointer(&t)))
	return
}

func setTermios(f *os.File, t unix.Termios) error {
	return ioctl(f.Fd(), ioctlWriteTermios, uintptr(unsafe.Pointer(&t)))
}

// GetSize returns terminal size
func GetSize(f *os.File) (ws Size, err error) {
	err = ioctl(f.Fd(), unix.TIOCGWINSZ, uintptr(unsafe.Pointer(&ws)))
	return
}

// SetSize sets new terminsl size
func SetSize(f *os.File, ws Size) error {
	return ioctl(f.Fd(), unix.TIOCSWINSZ, uintptr(unsafe.Pointer(&ws)))
}

// GetParams returns terminal params
func GetParams(f *os.File) (p Params, err error) {
	if p.termios, err = getTermios(f); err != nil {
		panic(err)
	}
	if p.ws, err = GetSize(f); err != nil {
		panic(err)
	}
	return
}

// SetParams applies provided params to terminal
func SetParams(f *os.File, p Params) (err error) {
	if err = setTermios(f, p.termios); err != nil {
		panic(err)
	}
	if err = SetSize(f, p.ws); err != nil {
		panic(err)
	}
	return
}

// MakeRaw makes terminal raw
func MakeRaw(f *os.File) error {
	termios, err := getTermios(f)
	if err != nil {
		return err
	}

	termios.Iflag &^= unix.IGNBRK | unix.BRKINT | unix.PARMRK | unix.ISTRIP | unix.INLCR | unix.IGNCR | unix.ICRNL | unix.IXON
	termios.Oflag &^= unix.OPOST
	termios.Cflag &^= unix.CSIZE | unix.PARENB
	termios.Cflag |= unix.CS8
	termios.Lflag &^= unix.ECHO | unix.ECHONL | unix.ICANON | unix.ISIG | unix.IEXTEN
	termios.Cc[unix.VMIN] = 1
	termios.Cc[unix.VTIME] = 0

	return setTermios(f, termios)
}

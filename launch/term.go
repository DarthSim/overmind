package launch

import (
	"os"
	"syscall"
	"unsafe"

	"github.com/pkg/term"
	"github.com/pkg/term/termios"
)

type winsize struct {
	wsRow    uint16
	wsCol    uint16
	wsXpixel uint16
	wsYpixel uint16
}

type termParams struct {
	ttyOrig syscall.Termios
	ws      winsize
}

func ioctl(fd, request, argp uintptr) error {
	if _, _, err := syscall.Syscall(syscall.SYS_IOCTL, fd, request, argp); err != 0 {
		return err
	}
	return nil
}

func getTermParams(f *os.File) (p termParams, err error) {
	if err = termios.Tcgetattr(f.Fd(), &p.ttyOrig); err != nil {
		return
	}
	if err = ioctl(f.Fd(), syscall.TIOCGWINSZ, uintptr(unsafe.Pointer(&p.ws))); err != nil {
		return
	}
	return
}

func applyTermParams(f *os.File, p termParams) (err error) {
	if err = termios.Tcsetattr(f.Fd(), termios.TCSANOW, &p.ttyOrig); err != nil {
		return
	}
	if err = ioctl(f.Fd(), syscall.TIOCSWINSZ, uintptr(unsafe.Pointer(&p.ws))); err != nil {
		return
	}
	return
}

func rawTerm() (t *term.Term, err error) {
	if t, err = term.Open("/dev/tty"); err != nil {
		return
	}

	err = t.SetRaw()

	return
}

func closeTerm(t *term.Term) {
	t.Restore()
	t.Close()
}

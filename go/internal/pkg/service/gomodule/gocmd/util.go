package gocmd

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/alessio/shellescape"
)

func formatArgs(args []string) string {
	var sb strings.Builder
	sb.WriteString(shellescape.Quote(args[0]))
	for i := 1; i < len(args); i++ {
		arg := args[i]
		sb.WriteByte(' ')
		sb.WriteString(shellescape.Quote(arg))
	}
	return sb.String()
}

func fdSeekToStart(fd *os.File) error {
	offset, err := fd.Seek(0, io.SeekStart)
	if err != nil {
		return err
	}
	if offset != 0 {
		return fmt.Errorf("(*os.File).Seek(0, io.SeekStart) returned unexpected non-zero offset %d", offset)
	}
	return nil
}

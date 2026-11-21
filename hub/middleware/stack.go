package middleware

import (
	"bytes"
	"fmt"
	"os"
	"runtime"
	"strings"
)

const (
	dunno     = "???"
	centerDot = "·"
	dot       = "."
	slash     = "/"
)

// stack returns a nicely formatted stack frame, skipping skip frames.
func stack(skip int) []byte {
	return stackFrom(skip+1, os.ReadFile)
}

// stackFrom is stack with the source files read through readFile.
func stackFrom(skip int, readFile func(string) ([]byte, error)) []byte {
	buf := new(bytes.Buffer) // the returned data
	// As we loop, we open files and read them. These variables record the currently
	// loaded file.
	var lines [][]byte
	var lastFile string
	for i := skip; ; i++ { // Skip the expected number of frames
		pc, file, line, ok := runtime.Caller(i)
		if !ok {
			break
		}
		// Print this much at least.  If we can't find the source, it won't show.
		fmt.Fprintf(buf, "%s:%d (0x%x)\n", file, line, pc)
		if file != lastFile {
			data, err := readFile(file)
			if err != nil {
				continue
			}
			lines = bytes.Split(data, []byte{'\n'})
			lastFile = file
		}
		fmt.Fprintf(buf, "\t%s: %s\n", function(pc), source(lines, line))
	}
	return buf.Bytes()
}

// source returns a space-trimmed slice of the n'th line.
func source(lines [][]byte, n int) []byte {
	n-- // in stack trace, lines are 1-indexed but our array is 0-indexed
	if n < 0 || n >= len(lines) {
		return []byte(dunno)
	}
	return bytes.TrimSpace(lines[n])
}

// function returns, if possible, the name of the function containing the PC.
func function(pc uintptr) []byte {
	fn := runtime.FuncForPC(pc)
	if fn == nil {
		return []byte(dunno)
	}
	return []byte(shortName(fn.Name()))
}

// shortName drops the package path of a function name and replaces its
// center dots: runtime/debug.*T·ptrmethod becomes *T.ptrmethod. The package
// path might contain dots (e.g. code.google.com/...), so the path prefix is
// eliminated first.
func shortName(name string) string {
	if lastSlash := strings.LastIndex(name, slash); lastSlash >= 0 {
		name = name[lastSlash+1:]
	}
	if period := strings.Index(name, dot); period >= 0 {
		name = name[period+1:]
	}
	return strings.ReplaceAll(name, centerDot, dot)
}

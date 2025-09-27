/*
 * Copyright (c) 2025 Eric Faurot <eric.faurot@plakar.io>
 *
 * Permission to use, copy, modify, and distribute this software for any
 * purpose with or without fee is hereby granted, provided that the above
 * copyright notice and this permission notice appear in all copies.
 *
 * THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES
 * WITH REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF
 * MERCHANTABILITY AND FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR
 * ANY SPECIAL, DIRECT, INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES
 * WHATSOEVER RESULTING FROM LOSS OF USE, DATA OR PROFITS, WHETHER IN AN
 * ACTION OF CONTRACT, NEGLIGENCE OR OTHER TORTIOUS ACTION, ARISING OUT OF
 * OR IN CONNECTION WITH THE USE OR PERFORMANCE OF THIS SOFTWARE.
 */

package logging

import (
	"bytes"
	"fmt"
	"io"
	"os"
)

type simple struct {
	level        LogLevel
	addEndOfLine bool
	hasPrefix    bool
	prefix       string
	debug        io.Writer
	info         io.Writer
	warn         io.Writer
	error        io.Writer
	fatal        io.Writer
}

func NewDefaultLogger() Logger {
	return &simple{
		level:        INFO,
		addEndOfLine: true,
		debug:        os.Stdout,
		info:         os.Stdout,
		warn:         os.Stderr,
		error:        os.Stderr,
		fatal:        os.Stderr,
	}
}

func NewWriterLogger(writer io.Writer, addEndOfLine bool) Logger {
	return &simple{
		level:        INFO,
		addEndOfLine: addEndOfLine,
		debug:        writer,
		info:         writer,
		warn:         writer,
		error:        writer,
		fatal:        writer,
	}
}

func (l *simple) doLog(writer io.Writer, levelPrefix, format string, args ...any) {
	var buf bytes.Buffer
	buf.WriteString(levelPrefix)
	if l.hasPrefix {
		//buf.WriteString("[")
		buf.WriteString(l.prefix)
		buf.WriteString(": ")
	}
	if len(args) == 0 {
		buf.WriteString(format)
	} else {
		fmt.Fprintf(&buf, format, args...)
	}
	if l.addEndOfLine {
		buf.WriteByte('\n')
	}
	writer.Write(buf.Bytes())
}

func (l *simple) WithPrefix(prefix string, args ...any) Logger {
	if len(args) > 0 {
		prefix = fmt.Sprintf(prefix, args...)
	}
	var r simple = *l
	r.prefix = l.prefix + prefix
	if r.prefix != "" {
		r.hasPrefix = true
	}
	return &r
}

func (l *simple) SetLevel(level LogLevel) {
	l.level = level
}

func (l *simple) Debug(format string, args ...any) {
	if l.level >= DEBUG {
		l.doLog(l.debug, "DEBUG: ", format, args...)
	}
}

func (l *simple) Info(format string, args ...any) {
	if l.level >= INFO {
		l.doLog(l.info, "INFO: ", format, args...)
	}
}

func (l *simple) Warn(format string, args ...any) {
	if l.level >= WARN {
		l.doLog(l.warn, "WARN: ", format, args...)
	}
}

func (l *simple) Error(format string, args ...any) {
	if l.level >= ERROR {
		l.doLog(l.error, "ERROR: ", format, args...)
	}
}

func (l *simple) Fatal(format string, args ...any) {
	if l.level >= FATAL {
		l.doLog(l.fatal, "FATAL: ", format, args...)
		os.Exit(1)
	}
}

func (l *simple) Write(data []byte) (int, error) {
	if l.level >= INFO {
		var buf bytes.Buffer
		buf.WriteString("STDOUT: ")
		if l.hasPrefix {
			buf.WriteString(l.prefix)
			buf.WriteByte(' ')
		}
		buf.WriteString(l.prefix)
		buf.Write(data)
		l.info.Write(buf.Bytes())
	}
	return len(data), nil
}

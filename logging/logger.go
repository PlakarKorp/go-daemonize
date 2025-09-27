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
	"io"
	"log/syslog"
	"os"
	"time"
)

type logger struct {
	level       LogLevel
	formatter   LogFormatter
	addLineFeed bool
	debug       io.Writer
	info        io.Writer
	warn        io.Writer
	error       io.Writer
	fatal       io.Writer
}

func (l *logger) doLog(w io.Writer, level LogLevel, message string, args ...any) {
	if l.level < level {
		return
	}
	entry := LogEntry{
		Timestamp: time.Now(),
		Level:     level,
		Message:   message,
		Args:      args,
	}

	var buf bytes.Buffer
	l.formatter.FormatLog(&buf, &entry)
	if l.addLineFeed {
		buf.WriteRune('\n')
	}
	w.Write(buf.Bytes())
}

func (l *logger) Debug(format string, args ...any) {
	l.doLog(l.debug, DEBUG, format, args...)
}

func (l *logger) Info(format string, args ...any) {
	l.doLog(l.info, INFO, format, args...)
}

func (l *logger) Warn(format string, args ...any) {
	l.doLog(l.warn, WARN, format, args...)
}

func (l *logger) Error(format string, args ...any) {
	l.doLog(l.error, ERROR, format, args...)
}

func (l *logger) Fatal(format string, args ...any) {
	l.doLog(l.fatal, FATAL, format, args...)
	os.Exit(1)
}

func (l *logger) WithLevel(level LogLevel) Logger {
	var new logger
	new = *l
	new.level = level
	return &new
}

func (l *logger) WithFormatter(formatter LogFormatter) Logger {
	var new logger
	new = *l
	new.formatter = formatter
	return &new
}

func NewDefaultLogger() Logger {
	return &logger{
		level:       INFO,
		formatter:   &DefaultFormatter{},
		addLineFeed: true,
		debug:       os.Stdout,
		info:        os.Stdout,
		warn:        os.Stderr,
		error:       os.Stderr,
		fatal:       os.Stderr,
	}
}

func NewWriterLogger(writer io.Writer, addLineFeed bool) Logger {
	return &logger{
		level:       INFO,
		addLineFeed: addLineFeed,
		formatter:   &DefaultFormatter{},
		debug:       writer,
		info:        writer,
		warn:        writer,
		error:       writer,
		fatal:       writer,
	}
}

func NewSyslogLogger(priority syslog.Priority, tag string) (Logger, error) {
	writer, err := syslog.New(priority, tag)
	if err != nil {
		return nil, err
	}
	return NewWriterLogger(writer, false), nil
}

func NewFileLogger(path string) (Logger, error) {
	writer, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return nil, err
	}
	return NewWriterLogger(writer, true), nil
}

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
	"context"
	"fmt"
	"io"
	"log/syslog"
	"os"
)

const (
	NONE = iota
	FATAL
	ERROR
	WARN
	INFO
	DEBUG
	ALL
)

type Logger struct {
	level        int
	addEndOfLine bool
	hasPrefix    bool
	prefix       string
	debug        io.Writer
	info         io.Writer
	warn         io.Writer
	error        io.Writer
	fatal        io.Writer
}

func NewDefault() *Logger {
	return &Logger{
		level:        INFO,
		addEndOfLine: true,
		debug:        os.Stdout,
		info:         os.Stdout,
		warn:         os.Stderr,
		error:        os.Stderr,
		fatal:        os.Stderr,
	}
}

func NewWithWriter(writer io.Writer, addEndOfLine bool) *Logger {
	return &Logger{
		level:        INFO,
		addEndOfLine: addEndOfLine,
		debug:        writer,
		info:         writer,
		warn:         writer,
		error:        writer,
		fatal:        writer,
	}
}

func NewFile(path string) (*Logger, error) {
	writer, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return nil, err
	}
	return NewWithWriter(writer, true), nil
}

func NewSyslog(priority syslog.Priority, tag string) (*Logger, error) {
	writer, err := syslog.New(priority, tag)
	if err != nil {
		return nil, err
	}
	return NewWithWriter(writer, false), nil
}

func (l *Logger) doLog(writer io.Writer, levelPrefix, format string, args ...any) {
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

func (l *Logger) WithPrefix(prefix string, args ...any) *Logger {
	if len(args) > 0 {
		prefix = fmt.Sprintf(prefix, args...)
	}
	var r Logger = *l
	r.prefix = l.prefix + prefix
	if r.prefix != "" {
		r.hasPrefix = true
	}
	return &r
}

func (l *Logger) SetLevel(level int) {
	l.level = level
}

func (l *Logger) Debug(format string, args ...any) {
	if l.level >= DEBUG {
		l.doLog(l.debug, "DEBUG: ", format, args...)
	}
}

func (l *Logger) Info(format string, args ...any) {
	if l.level >= INFO {
		l.doLog(l.info, "INFO: ", format, args...)
	}
}

func (l *Logger) Warn(format string, args ...any) {
	if l.level >= WARN {
		l.doLog(l.warn, "WARN: ", format, args...)
	}
}

func (l *Logger) Error(format string, args ...any) {
	if l.level >= ERROR {
		l.doLog(l.error, "ERROR: ", format, args...)
	}
}

func (l *Logger) Fatal(format string, args ...any) {
	if l.level >= FATAL {
		l.doLog(l.fatal, "FATAL: ", format, args...)
		os.Exit(1)
	}
}

func (l *Logger) Write(data []byte) (int, error) {
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

var defaultLogger *Logger = NewDefault()

func SetDefaultLogger(logger *Logger) {
	defaultLogger = logger
}

func GetDefaultLogger() *Logger {
	return defaultLogger
}

func SetLevel(level int) {
	defaultLogger.SetLevel(level)
}

func Debug(format string, args ...any) {
	defaultLogger.Debug(format, args...)
}

func Info(format string, args ...any) {
	defaultLogger.Info(format, args...)
}

func Warn(format string, args ...any) {
	defaultLogger.Warn(format, args...)
}

func Error(format string, args ...any) {
	defaultLogger.Error(format, args...)
}

func Fatal(format string, args ...any) {
	defaultLogger.Fatal(format, args...)
}

/* Contextualized logger */

type key int

const loggerKey key = 0

func GetLogger(ctx context.Context) *Logger {
	return ctx.Value(loggerKey).(*Logger)
}

func WithLogger(ctx context.Context, logger *Logger) context.Context {
	return context.WithValue(ctx, loggerKey, logger)
}

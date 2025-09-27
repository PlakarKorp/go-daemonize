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
	"context"
)

type LogLevel int

const (
	NONE LogLevel = iota
	FATAL
	ERROR
	WARN
	INFO
	DEBUG
	ALL
)

type Logger interface {
	SetLevel(level LogLevel)
	Debug(format string, args ...any)
	Info(format string, args ...any)
	Warn(format string, args ...any)
	Error(format string, args ...any)
	Fatal(format string, args ...any)
}

var defaultLogger Logger = NewDefaultLogger()

func SetDefaultLogger(logger Logger) {
	defaultLogger = logger
}

func GetDefaultLogger() Logger {
	return defaultLogger
}

func SetLevel(level LogLevel) {
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

func GetLogger(ctx context.Context) Logger {
	return ctx.Value(loggerKey).(Logger)
}

func WithLogger(ctx context.Context, logger Logger) context.Context {
	return context.WithValue(ctx, loggerKey, logger)
}

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

package daemonize

import (
	"flag"
	"fmt"
	"io"
	"log/syslog"
	"os"
	"path/filepath"
	"sync"
	"syscall"

	"github.com/PlakarKorp/go-daemonize/logging"
)

type Configuration interface {
	Parse(rd io.Reader) error
}

type Daemon struct {
	name    string
	version string
	logTag  string
	config  Configuration
	isDebug bool

	wg       sync.WaitGroup
	services map[string]Service
}

type Option func(*Daemon)

func NewDaemon(opts ...Option) *Daemon {
	d := &Daemon{}

	for _, opt := range opts {
		opt(d)
	}
	d.setUp()
	return d
}

func WithName(name string) Option {
	return func(d *Daemon) { d.name = name }
}

func WithVersion(version string) Option {
	return func(d *Daemon) { d.version = version }
}

func WithLogTag(tag string) Option {
	return func(d *Daemon) { d.logTag = tag }
}

func WithConfiguration(config Configuration) Option {
	return func(d *Daemon) { d.config = config }
}

func (daemon *Daemon) IsDebugMode() bool {
	return daemon.isDebug
}

func (daemon *Daemon) setUp() {
	var opt_version bool
	var opt_configFile string
	var opt_foreground bool
	var opt_logFile string

	// Parse cmdline parameters
	flag.StringVar(&opt_configFile, "config", "", "configuration file")
	flag.BoolVar(&opt_version, "version", false, "show version")
	flag.BoolVar(&daemon.isDebug, "debug", false, "debug mode")
	flag.BoolVar(&opt_foreground, "foreground", false, "run in foreground")
	flag.StringVar(&opt_logFile, "log", "", "log file")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [OPTIONS]\n", flag.CommandLine.Name())
		fmt.Fprintf(os.Stderr, "\nOPTIONS:\n")
		flag.PrintDefaults()
	}
	flag.Parse()

	if daemon.name == "" {
		daemon.name = filepath.Base(os.Args[0])
	}
	if daemon.logTag == "" {
		daemon.logTag = daemon.name
	}

	if opt_version {
		fmt.Println(daemon.name, "Version", daemon.version)
		os.Exit(0)
	}

	// Read configuration
	if daemon.config != nil {
		if opt_configFile == "" {
			logging.Fatal("no configuration file specified")
		}
		fp, err := os.Open(opt_configFile)
		if err != nil {
			logging.Fatal("failed to open configuration file: %v", err)
		}
		defer fp.Close()

		if err := daemon.config.Parse(fp); err != nil {
			logging.Fatal("failed to parse config file: %v", err)
		}
	}

	// Do fork+exec if needed
	if !opt_foreground && os.Getenv("REEXEC") == "" {
		pid, err := daemon.doDaemonize(os.Args)
		if err != nil {
			logging.Fatal("failed to rexec: %v", err)
		}
		fmt.Println("started with pid", pid)
		os.Exit(0)
	}

	// Setup logging
	if opt_logFile != "" {
		logger, err := logging.NewFileLogger(opt_logFile)
		if err != nil {
			logging.Fatal("cannot open log file: %v", err)
		}
		logging.SetDefaultLogger(logger)
	} else if !opt_foreground {
		logger, err := logging.NewSyslogLogger(syslog.LOG_INFO|syslog.LOG_USER, daemon.logTag)
		if err != nil {
			logging.Fatal("cannot open syslog: %v", err)
		}
		logging.SetDefaultLogger(logger)
	}
}

func (daemon *Daemon) doDaemonize(argv []string) (int, error) {
	binary, err := os.Executable()
	if err != nil {
		return -1, err
	}

	procAttr := syscall.ProcAttr{
		Sys: &syscall.SysProcAttr{
			Setsid: true,
		},
	}
	procAttr.Files = []uintptr{
		uintptr(syscall.Stdin),
		uintptr(syscall.Stdout),
		uintptr(syscall.Stderr),
	}
	procAttr.Env = append(os.Environ(),
		"REEXEC=1",
	)

	pid, err := syscall.ForkExec(binary, argv, &procAttr)
	if err != nil {
		return -1, err
	}
	return pid, nil
}

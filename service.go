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
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/PlakarKorp/go-daemonize/logging"
)

type key int

const serviceProviderKey key = 0

func GetServiceProvider(ctx context.Context) ServiceProvider {
	return ctx.Value(serviceProviderKey).(ServiceProvider)
}

func GetService(ctx context.Context, name string) Service {
	return GetServiceProvider(ctx).GetService(name)
}

type ServiceProvider interface {
	GetService(string) Service
}

type Service interface {
	Start(ctx context.Context) error
	Shutdown(ctx context.Context) error
}

func (daemon *Daemon) Run(ctx context.Context) {

	ctx = context.WithValue(ctx, serviceProviderKey, daemon)

	logging.GetLogger(ctx).Info("starting")
	for name, service := range daemon.services {
		daemon.startService(ctx, name, service)
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	sig := <-quit
	logging.GetLogger(ctx).Info("got signal %v", sig)
	logging.GetLogger(ctx).Info("shutting down...")
	for name, service := range daemon.services {
		daemon.stopService(ctx, name, service, 5*time.Second)
	}

	daemon.wg.Wait()
	logging.GetLogger(ctx).Info("exiting")
}

func (daemon *Daemon) AddService(name string, svc Service) {
	if daemon.services == nil {
		daemon.services = make(map[string]Service)
	}
	daemon.services[name] = svc
}

func (daemon *Daemon) GetService(name string) Service {
	return daemon.services[name]
}

func (daemon *Daemon) startService(ctx context.Context, name string, svc Service) {
	daemon.wg.Add(1)
	go func() {
		logging.GetLogger(ctx).Info("starting service %s", name)
		err := svc.Start(ctx)
		if err != nil {
			logging.GetLogger(ctx).Warn("service %s failed: %v", name, err)
		}
		logging.GetLogger(ctx).Info("service %s stopped", name)
		daemon.wg.Done()
	}()
}

func (daemon *Daemon) stopService(ctx context.Context, name string, svc Service, timeout time.Duration) {
	go func() {
		logging.GetLogger(ctx).Info("stopping service %s", name)
		sctx, release := context.WithTimeout(context.Background(), timeout)
		defer release()
		if err := svc.Shutdown(sctx); err != nil {
			logging.GetLogger(ctx).Warn("shutdon error for servive %s: %v", name, err)
		}
	}()
}

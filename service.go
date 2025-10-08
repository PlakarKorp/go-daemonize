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
	"errors"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/PlakarKorp/go-daemonize/logging"
)

type ServiceProvider interface {
	GetService(string) Service
}

type Service interface {
	Run(*ServiceController, context.Context) error
}

type ServiceStatus string

const (
	ServiceDown     ServiceStatus = "down"
	ServiceStarting ServiceStatus = "starting"
	ServiceUp       ServiceStatus = "up"
	ServiceStopping ServiceStatus = "stopping"
)

type ServiceController struct {
	name    string
	service Service
	ctx     context.Context
	status  ServiceStatus
	mu      sync.Mutex
	stop    context.CancelCauseFunc
}

var Stopped = errors.New("stopped")

func (daemon *Daemon) Run(ctx context.Context) {
	var wg sync.WaitGroup

	logger := logging.GetLogger(ctx)

	ctx = context.WithValue(ctx, serviceProviderKey, daemon)

	ok := true
	for _, ctrl := range daemon.services {
		if err := ctrl.startService(ctx, &wg); err != nil {
			logger.Error("failed to start service %s: %v", err)
			ok = false
			break
		}
	}

	if ok {
		quit := make(chan os.Signal, 1)
		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

		sig := <-quit
		logger.Info("got signal %v", sig)
		logger.Info("shutting down...")
	}

	for _, ctrl := range daemon.services {
		ctrl.stopService()
	}

	wg.Wait()
	logging.GetLogger(ctx).Info("exiting")
}

func (daemon *Daemon) AddService(name string, service Service) {
	if daemon.services == nil {
		daemon.services = make(map[string]*ServiceController)
	}
	daemon.services[name] = &ServiceController{
		name:    name,
		service: service,
		status:  ServiceDown,
		mu:      sync.Mutex{},
	}
}

func (daemon *Daemon) GetService(name string) Service {
	ctrl, found := daemon.services[name]
	if !found {
		return nil
	}
	return ctrl.service
}

func (ctrl *ServiceController) startService(ctx context.Context, wg *sync.WaitGroup) error {
	ctrl.ctx = ctx
	ctrl.mu.Lock()
	if ctrl.status != ServiceDown {
		err := fmt.Errorf("service is %s", ctrl.status)
		ctrl.mu.Unlock()
		return err
	}

	logging.GetLogger(ctx).Info("%s: starting...", ctrl.name)
	ctrl.status = ServiceStarting
	ctrl.mu.Unlock()

	go func() {
		serviceCtx, cancel := context.WithCancelCause(ctx)
		ctrl.stop = cancel
		err := ctrl.service.Run(ctrl, serviceCtx)
		ctrl.stop = nil
		if err != nil {
			logging.GetLogger(ctx).Warn("service %s returned error: %v", ctrl.name, err)
		}
		ctrl.mu.Lock()
		ctrl.status = ServiceStopping
		logging.GetLogger(ctx).Info("%s: stopped", ctrl.name)
		ctrl.mu.Unlock()
		wg.Done()
	}()

	wg.Add(1)
	return nil
}

func (ctrl *ServiceController) stopService() {
	if ctrl.stop != nil {
		ctrl.stop(Stopped)
		ctrl.stop = nil
	}
}

func (ctrl *ServiceController) Up() {
	ctrl.mu.Lock()
	ctrl.status = ServiceUp
	logging.GetLogger(ctrl.ctx).Info("%s: up", ctrl.name)
	ctrl.mu.Unlock()
}

func (ctrl *ServiceController) Stopping() {
	ctrl.mu.Lock()
	ctrl.status = ServiceStopping
	logging.GetLogger(ctrl.ctx).Info("%s: stopping", ctrl.name)
	ctrl.mu.Unlock()
}

func (ctrl *ServiceController) Run(ctx context.Context) error {
	ctrl.Up()
	<-ctx.Done()
	if context.Cause(ctx) == Stopped {
		return nil
	}
	return ctx.Err()
}

type key int

const serviceProviderKey key = 0

func GetServiceProvider(ctx context.Context) ServiceProvider {
	return ctx.Value(serviceProviderKey).(ServiceProvider)
}

func GetService(ctx context.Context, name string) Service {
	return GetServiceProvider(ctx).GetService(name)
}

// Package delegate provides the ability to make function calls between
// different domain packages when an import is not possible.
package db

import (
	"context"
	"fmt"

	"github.com/zacksfF/Build-A-Go-Apps-That-Scale-On-Google-Cloud/Observability/logging"
)

// These types are just for documentation so we know what keys go
// where in the map.
type (
	domain string
	action string
)

// Delegate manages the set of functions to be called by domain
// packages when an import is not possible.
type Delegate struct {
	log   *logging.Logger
	funcs map[domain]map[action][]Func
}

// New constructs a delegate for indirect api access.
func New(log *logging.Logger) *Delegate {
	return &Delegate{
		log:   log,
		funcs: make(map[domain]map[action][]Func),
	}
}

// Register adds a function to be called for a specified domain and action.
func (d *Delegate) Register(domainType string, actionType string, fn Func) {
	aMap, ok := d.funcs[domain(domainType)]
	if !ok {
		aMap = make(map[action][]Func)
		d.funcs[domain(domainType)] = aMap
	}

	funcs := aMap[action(actionType)]
	funcs = append(funcs, fn)
	aMap[action(actionType)] = funcs
}

// Call executes all functions registered for the specified domain and
// action. These functions are executed synchronously on the G making the call.
func (d *Delegate) Call(ctx context.Context, data Data) error {
	d.log.Info(ctx, "delegate call", "status", "started", "domain", data.Domain, "action", data.Action, "params", data.RawParams)
	defer d.log.Info(ctx, "delegate call", "status", "completed")

	if dMap, ok := d.funcs[domain(data.Domain)]; ok {
		if funcs, ok := dMap[action(data.Action)]; ok {
			for _, fn := range funcs {
				d.log.Info(ctx, "delegate call", "status", "sending")

				if err := fn(ctx, data); err != nil {
					d.log.Error(ctx, "delegate call", "msg", err)
				}
			}
		}
	}

	return nil
}

// Func represents a function that is registered and called by the system.
type Func func(context.Context, Data) error

// Data represents an event between domains.
type Data struct {
	Domain    string
	Action    string
	RawParams []byte
}

// String implements the Stringer interface.
func (d Data) String() string {
	return fmt.Sprintf(
		"Event{Domain:%#v, Action:%#v, RawParams:%#v}",
		d.Domain, d.Action, string(d.RawParams),
	)
}

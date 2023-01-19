package EventStore

import (
	"context"
	"fmt"
)

const (
	aggregateStartVersion                = -1 // used for EventStoreDB
	aggregateAppliedEventsInitialCap     = 10
	aggregateUncommittedEventsInitialCap = 10
)

type HandleCommand interface {
	HandleCommand(ctx context.Context, command Command) error
}

// Apply process Aggregate Event
type Apply interface {
	Apply(event Event) error
}

type When interface {
	When(event Event) error
}

// Load create Aggregate state from Event's.
type Load interface {
	Load(events []Event) error
}

type Aggregate interface {
	When
	AggregateRoot
}

// AggregateRoot contains all methods of Aggregate
type AggregateRoot interface {
	GetUncommittedEvents() []Event
	GetID() string
	SetID(string)
	GetVersion() int64
	ClearUncommittedEvents()
	SetType(aggregateType AggregateType)
	SetVersion(version int64)
	GetType() AggregateType
	RaiseEvent(event Event) error
	String() string
	GetAppliedEvents() []Event
	Test(t TestFunc, event Event)
	Load
	Apply
	When
}

type AggregateBase struct {
	AggregateRoot
	UncommitedEvents []Event
	AppliedEvents    []Event
	ID               string
	Version          int64
	Type             AggregateType
}

func (a *AggregateBase) GetUncommittedEvents() []Event {
	return a.UncommitedEvents
}

func (a *AggregateBase) GetID() string {
	return a.ID
}

func (a *AggregateBase) SetID(id string) {
	a.ID = id
}

func (a *AggregateBase) GetVersion() int64 {
	return a.Version
}
func (a *AggregateBase) SetVersion(version int64) {
	a.Version = version
}

func (a *AggregateBase) ClearUncommittedEvents() {
	a.UncommitedEvents = make([]Event, 0, aggregateUncommittedEventsInitialCap)
}

func (a *AggregateBase) SetType(aggregateType AggregateType) {
	a.Type = aggregateType
}

func (a *AggregateBase) GetType() AggregateType {
	return a.Type
}

type TestFunc func(event Event)

func (a *AggregateBase) Test(t TestFunc, event Event) {
	t(event)
}

func (a *AggregateBase) Apply(event Event) error {
	event.SetAggregateType(a.GetType())
	if err := a.When(event); err != nil {
		return err
	}
	a.Version++
	event.SetVersion(a.GetVersion())
	a.UncommitedEvents = append(a.UncommitedEvents, event)
	return nil
}

func (a *AggregateBase) Load(events []Event) error {

	for _, evt := range events {
		if err := a.When(evt); err != nil {
			return err
		}

		a.AppliedEvents = append(a.AppliedEvents, evt)
		a.Version++
	}
	return nil
}

// RaiseEvent function is used to replay event occured to update current aggregate,
// should be only used with events from eventstore
func (a *AggregateBase) RaiseEvent(event Event) error {
	event.SetAggregateType(a.GetType())

	if err := a.When(event); err != nil {
		return err
	}
	a.AppliedEvents = append(a.AppliedEvents, event)
	a.Version = event.GetVersion()
	return nil
}

func (a *AggregateBase) String() string {
	return fmt.Sprintf("ID: {%s}, Version: {%v}, Type: {%v}, AppliedEvents: {%v}, UncommittedEvents: {%v}",
		a.GetID(),
		a.GetVersion(),
		a.GetType(),
		len(a.GetAppliedEvents()),
		len(a.GetUncommittedEvents()),
	)
}

func (a *AggregateBase) GetAppliedEvents() []Event {
	return a.AppliedEvents
}

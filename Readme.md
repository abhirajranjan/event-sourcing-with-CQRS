# event-store
event-store is a golang implementation of event store used in event sourcing architecture to store and load [aggregates](https://www.eventstore.com/event-sourcing#Write-model) from [eventstoredb](https://www.eventstore.com/event-sourcing#Event-database-event-store)

## Features
* load aggregate directly from the aggregate store
* embed aggregate struct and perform domain logic and call apply to apply aggregate

## Basic Usage
aggregateStore handles all write tasks providing the aggregate logic

```golang
import {
    "fmt"

    // golang client for eventstoredb
    "github.com/EventStore/EventStore-Client-Go/esdb"
	es "github.com/abhirajranjan/eventstore/pkg/eventstore"
}

func main {
    // eventstoredb config
    setting, err := esdb.ParseConnectionString("esdb://localhost:2113?tls=false")
	if err != nil {
		panic(err)
	}
    // eventstoredb client instance
	conn, err := esdb.NewClient(setting)
	if err != nil {
		panic(err)
	}
	defer conn.Close()
	fmt.Println("connection created")
	store := es.NewAggregateStore(logger, conn)
}
```

## Events

events are entity that describe certain thing that has happened. Due to the fact that they happened in past makes them immutable.

Abstract Event is defined as es.Event

```golang
type event struct {
	EventID     string
	EventType   string
	Data        []byte
	Timestamp   time.Time
	Version     int64
	Metadata    []byte
	AggregateID string
	AggregateType
}
```
user generating new event have to set these feilds. 

Data field is marshaled json byte object that has domain fields related to Aggregate Type

### NewEventFromRecordedEvent

```golang
func NewEventFromRecordedEvent(revent *esdb.RecordedEvent) *event
```

NewEventFromRecordedEvent is use to load events already occured in past.
Internally this function is used in stores to load event from eventstoredb.

### EventFromEventData

```golang
func EventFromEventData(recordedEvent esdb.RecordedEvent) (event, error)
```

EventFromEventData is used to generate es.event from eventstoredb event type in deserialized form. used in stores to get events and call "when" method on aggregate.


### NewEventFromEventData

```golang
func NewEventFromEventData(eventd esdb.EventData) event
```

NewEventFromEventData is use to generate event from eventData.

### NewBaseEvent

```golang
NewBaseEvent(aggregate Aggregate, eventType string) event
```
used to generate new event for aggregate.
it generates new EventID and set Event ID, EventType, EventVersion as aggregateID, aggregateType, AggregateVersion respectively.
new timestamp is also initalized as current time.Time object.

## Aggregate

```golang
type Aggregate struct {
	es.AggregateBase
	newID int64
    name string
}
```

create a custom aggregate struct to embed AggregateBase with additional domain fields to get aggregate methods

### when function

embedding AggregateBase with custom fields to store the current state.

"When" method needs to be implemented in new embedded struct to maintain current state and  apply domain logic event applied to it

Note: "When" method should not be directly called in the code, instead call "Apply" method on the aggregate to apply an event which internally calls "When" in addition to maintaining all the uncommited events

```golang
type Data struct {
	Item  string
	Value int64
}

type Aggregate struct {
	es.AggregateBase
    // embed Data having domain fields
	Data Data
}

// When method takes event of type es.Event and return error
func (a *Aggregate) When(event es.Event) error {
	switch event.GetEventType() {
	case "setItem":
		var d Data
		if err := json.Unmarshal(event.GetData(), &d); err != nil {
			return fmt.Errorf("unmarshling error in setItem")
		}
		a.Data.Item = d.Item
	case "setValue":
		var d Data
		if err := json.Unmarshal(event.GetData(), &d); err != nil {
			return fmt.Errorf("unmarshling error in setValue")
		}
		a.Data.Value = d.Value
	default:
		return fmt.Errorf("event type %v not defined", event.GetEventType())
	}
	fmt.Printf("(item) %s %d\n", a.Data.Item, a.Data.Value)
	return nil
}
```

### create new Aggregate

initate new aggregate. aggregateRoot is an interface that provides "When" function for "apply" method.

```golang
// initiate aggregate
initalaggregate := &Aggregate{}
// set the AggregateRoot interface to itself so that it can call embedded "When" function
initalaggregate.AggregateRoot = initalaggregate
// set inital version to -1 
// version automatically increments by 1 so the event commited with version 0
initalaggregate.SetVersion(-1)
// set aggregate type to event type matched in "When" function
initalaggregate.SetType("setItem")
```

### load current state of aggregate

AggregateStore "load" method is used to load current state of aggregate into aggregate object.

```golang
// create new aggregate
initialaggregate := &Aggregate{}
initialaggregate.AggregateRoot = initialaggregate
initialaggregate.SetType("setItem")

store := es.NewAggregateStore(logger, conn)

// load events sequentially into initalaggregate to get current state
store.Load(context.TODO(), &initialaggregate)
```


### apply method

"Apply" event is used to apply the event to current aggregate.
it internally calls aggregate's "When" method to maintain the current state and if no error was returned it appends event to uncommited events.

```golang
// create new event from base event
event := es.NewBaseEvent(initalaggregate, "setItem")
data, err := json.Marshal(Data{Item: "gate", Value: 1})
if err != nil {
	panic(err)
}

event.SetData(data)
fmt.Println("id: ", initalaggregate.GetID())

if err := initalaggregate.Apply(&event); err != nil {
	panic(err)
}
```

## AggregateStore

aggregate store gives functions used to to easily restore and check aggregate's current state in eventstoredb

### Load function

```golang
func (as *aggregateStore) Load(ctx context.Context, aggregate Aggregate) error
```

get the current state of aggregate based on aggregateID from database by getting events sequentially and calling aggregate "when" method to perform domain logic

for easy debugging, applied events are stored in aggregate to easily provide tracking of events

### Save function

```golang
func (as *aggregateStore) Save(ctx context.Context, aggregate Aggregate) error
```

Save function is used to Save all uncommited events in aggregate to database to persist changes

Save function should be called after all domain checks and processing of event in current state of aggregate

### Exists function

```golang
func (as *aggregateStore) Exists(ctx context.Context, streamID string) error
```

Exists checks if the current aggregateID is present in eventstore or not.

if given aggregateID does not exists, it returns esdb.ErrNoStreamFound
if given aggregateID exists in eventstore, it returns nil
if any other error occured, it returns wrapped error
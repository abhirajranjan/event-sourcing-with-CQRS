package EventStore

import (
	"encoding/json"
	"time"

	esdb "github.com/EventStore/EventStore-Client-Go/esdb"
	uuid "github.com/satori/go.uuid"
)

type Event interface {
	ToEventData() esdb.EventData
	GetEventID() string
	SetEventID(id string) error
	GetEventType() string
	SetEventType(etype string) error
	GetData() []byte
	GetJsonData(data interface{}) error
	SetData(data []byte) error
	GetTimestamp() time.Time
	SetTimestamp(timestamp time.Time) error
	GetVersion() int64
	SetVersion(version int64) error
	GetMetadata() []byte
	GetJsonMetadata(data interface{}) error
	SetMetadata(data []byte) error
	GetAggregateID() string
	SetAggregateID(id string) error
	GetAggregateType() AggregateType
	SetAggregateType(newtype AggregateType) error
}

type AggregateType string

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

func NewEventFromRecordedEvent(revent *esdb.RecordedEvent) *event {
	return &event{
		EventID:     revent.EventID.String(),
		EventType:   revent.EventType,
		Data:        revent.Data,
		Timestamp:   revent.CreatedDate,
		AggregateID: revent.StreamID,
		Version:     int64(revent.EventNumber),
		Metadata:    revent.UserMetadata,
	}
}

func EventFromEventData(recordedEvent esdb.RecordedEvent) (event, error) {
	return event{
		EventID:     recordedEvent.EventID.String(),
		EventType:   recordedEvent.EventType,
		Data:        recordedEvent.Data,
		Timestamp:   recordedEvent.CreatedDate,
		AggregateID: recordedEvent.StreamID,
		Version:     int64(recordedEvent.Position.Commit),
		Metadata:    nil,
	}, nil
}

func NewEventFromEventData(eventd esdb.EventData) event {
	return event{
		EventID:   eventd.EventID.String(),
		EventType: eventd.EventType,
		Data:      eventd.Data,
		Metadata:  eventd.Metadata,
	}
}

// NewBaseEvent new base Event constructor with configured EventID, Aggregate properties and Timestamp.
func NewBaseEvent(aggregate Aggregate, eventType string) event {
	return event{
		EventID:       uuid.NewV4().String(),
		AggregateType: aggregate.GetType(),
		AggregateID:   aggregate.GetID(),
		Version:       aggregate.GetVersion(),
		EventType:     eventType,
		Timestamp:     time.Now().UTC(),
	}
}

func (e *event) ToEventData() esdb.EventData {
	return esdb.EventData{
		EventType:   e.EventType,
		ContentType: esdb.JsonContentType,
		Data:        e.Data,
		Metadata:    e.Metadata,
	}
}

func (e *event) GetEventID() string {
	return e.EventID
}
func (e *event) SetEventID(id string) error {
	e.EventID = id
	return nil
}

func (e *event) GetEventType() string {
	return e.EventType
}

func (e *event) SetEventType(etype string) error {
	e.EventType = etype
	return nil
}

func (e *event) GetData() []byte {
	return e.Data
}

func (e *event) GetJsonData(data interface{}) error {
	return json.Unmarshal(e.Data, data)
}

func (e *event) SetData(data []byte) error {
	e.Data = data
	return nil
}

func (e *event) GetTimestamp() time.Time {
	return e.Timestamp
}

func (e *event) SetTimestamp(timestamp time.Time) error {
	e.Timestamp = timestamp
	return nil
}

func (e *event) GetVersion() int64 {
	return e.Version
}

func (e *event) SetVersion(version int64) error {
	e.Version = version
	return nil
}

func (e *event) GetMetadata() []byte {
	return e.Metadata
}

func (e *event) GetJsonMetadata(data interface{}) error {
	return json.Unmarshal(e.Metadata, data)
}

func (e *event) SetMetadata(data []byte) error {
	e.Metadata = data
	return nil
}

func (e *event) GetAggregateID() string {
	return e.AggregateID
}

func (e *event) SetAggregateID(id string) error {
	e.AggregateID = id
	return nil
}

func (e *event) GetAggregateType() AggregateType {
	return e.AggregateType
}

func (e *event) SetAggregateType(newtype AggregateType) error {
	e.AggregateType = newtype
	return nil
}

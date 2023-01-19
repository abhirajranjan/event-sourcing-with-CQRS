package EventStore

import (
	"context"
	"io"
	"math"

	esdb "github.com/EventStore/EventStore-Client-Go/esdb"
	"github.com/pkg/errors"
)

const count = math.MaxInt64

// AggregateStore is responsible for loading and saving aggregates.
type AggregateStore interface {
	// Load function loads the most recent version of an aggregate
	Load(ctx context.Context, aggregate Aggregate) error

	// Save saves the uncommitted events for an aggregate.
	Save(ctx context.Context, aggregate Aggregate) error

	// Exists check aggregate exists by id.
	Exists(ctx context.Context, streamID string) error
}

type aggregateStore struct {
	log Logger
	db  *esdb.Client
}

func NewAggregateStore(logger Logger, db *esdb.Client) AggregateStore {
	return &aggregateStore{log: logger, db: db}
}

// load series of events to aggregate to get current state
func (as *aggregateStore) Load(ctx context.Context, aggregate Aggregate) error {
	stream, err := as.db.ReadStream(ctx, aggregate.GetID(), esdb.ReadStreamOptions{}, count)
	if err != nil {
		as.log.Error(errors.Wrap(err, "db.ReadStream"))
		return errors.Wrap(err, "db.ReadStream")
	}
	defer stream.Close()

	for {
		resolvedevent, err := stream.Recv()
		if errors.Is(err, esdb.ErrStreamNotFound) {
			as.log.Error(errors.Wrap(err, "Stream.Recv"))
			return errors.Wrap(err, "Stream.Recv")
		}
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			as.log.Error(errors.Wrap(err, "Stream.Recv"))
			return errors.Wrap(err, "Stream.Recv")
		}
		event := NewEventFromRecordedEvent(resolvedevent.Event)
		if err := aggregate.RaiseEvent(event); err != nil {
			as.log.Error(errors.Wrap(err, "Raise Event"))
			return errors.Wrap(err, "Raise Event")
		}
	}
	return nil
}

// save all uncommited events of aggregate to eventstore
func (as *aggregateStore) Save(ctx context.Context, aggregate Aggregate) error {
	if len(aggregate.GetUncommittedEvents()) == 0 {
		as.log.Debugf("(Save) [no uncommittedEvents] len: {%d}", len(aggregate.GetUncommittedEvents()))
		return nil
	}

	eventData := make([]esdb.EventData, 0, len(aggregate.GetUncommittedEvents()))
	for _, evt := range aggregate.GetUncommittedEvents() {
		eventData = append(eventData, evt.ToEventData())
	}

	var expectedRevision esdb.ExpectedRevision
	if aggregate.GetVersion() == 0 {
		expectedRevision = esdb.NoStream{}
	} else {
		readstreamopts := esdb.ReadStreamOptions{Direction: esdb.Backwards, From: esdb.End{}}
		readstream, err := as.db.ReadStream(ctx, aggregate.GetID(), readstreamopts, 1)

		if err != nil {
			as.log.Error(errors.Wrap(err, "db.ReadStream"))
			return errors.Wrap(err, "db.ReadStream")
		}
		defer readstream.Close()

		lastevent, err := readstream.Recv()
		if err != nil {
			as.log.Error(errors.Wrap(err, "Stream.Recv"))
			return errors.Wrap(err, "Stream.Recv")
		}
		expectedRevision = esdb.Revision(lastevent.OriginalEvent().EventNumber)
	}

	streamopts := esdb.AppendToStreamOptions{ExpectedRevision: expectedRevision}
	stream, err := as.db.AppendToStream(ctx, aggregate.GetID(), streamopts, eventData...)
	if err != nil {
		as.log.Error(errors.Wrap(err, "db.Append"))
		return errors.Wrap(err, "db.Append")
	}
	as.log.Debugf("(Save) AppendToStream: ", stream)
	aggregate.ClearUncommittedEvents()
	return nil
}

func (as *aggregateStore) Exists(ctx context.Context, streamID string) error {
	readopts := esdb.ReadStreamOptions{Direction: esdb.Backwards, From: esdb.Revision(1)}
	stream, err := as.db.ReadStream(ctx, streamID, readopts, 1)
	if err != nil {
		as.log.Error(errors.Wrap(err, "db.ReadStream"))
		return errors.Wrap(err, "db.ReadStream")
	}
	_, err = stream.Recv()
	if errors.Is(err, esdb.ErrStreamNotFound) {
		as.log.Error(errors.Wrap(err, "stream.Recv"))
		return errors.Wrap(err, "stream.Recv")
	}
	if errors.Is(err, io.EOF) {
		return nil
	}
	if err != nil {
		as.log.Error(errors.Wrap(err, "stream.Recv"))
		return errors.Wrap(err, "stream.Recv")
	}
	return nil
}

package store

import (
	"context"
	"fmt"
	"github.com/gocql/gocql"
	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel/trace"
	"os"
	"report_service/domain"
	"time"
)

const (
	DATABASE_CASSANDRA = "events"
	COLLECTION_EVENT   = "events"
)

type EventCassandraStore struct {
	session *gocql.Session
	logger  *logrus.Logger
	tracer  trace.Tracer
}

func New(logger *logrus.Logger, tracer trace.Tracer) (*EventCassandraStore, error) {
	db := os.Getenv("EVENT_DB")

	cluster := gocql.NewCluster(db)
	cluster.Keyspace = "system"
	session, err := cluster.CreateSession()
	if err != nil {
		logger.Println(err)
		return nil, err
	}

	err = session.Query(
		fmt.Sprintf(`CREATE KEYSPACE IF NOT EXISTS %s
					WITH replication = {
						'class' : 'SimpleStrategy',
						'replication_factor' : %d
					}`, DATABASE_CASSANDRA, 1)).Exec()
	if err != nil {
		logger.Println(err)
	}
	session.Close()

	cluster.Keyspace = DATABASE_CASSANDRA
	cluster.Consistency = gocql.One
	session, err = cluster.CreateSession()

	if err != nil {
		logger.Println(err)
		return nil, err
	}

	return &EventCassandraStore{
		session: session,
		logger:  logger,
		tracer:  tracer,
	}, nil
}

func (store *EventCassandraStore) CloseSession() {
	store.session.Close()
}

func (store *EventCassandraStore) CreateTables() {
	err := store.session.Query(
		fmt.Sprintf(`CREATE TABLE IF NOT EXISTS %s (id UUID, event_type text, timestamp int, timespent int, PRIMARY KEY ((id), event_type, timestamp))`, COLLECTION_EVENT)).Exec()

	if err != nil {
		store.logger.Printf("CASSANDRA CREATE TABLE ERR: %s", err.Error())
	}
}

func (store *EventCassandraStore) CreateEvent(ctx context.Context, event *domain.Event) (*domain.Event, error) {
	ctx, span := store.tracer.Start(ctx, "EventStore.CreateEvent")
	defer span.End()

	store.logger.Infoln("EventCassandra.CreateEvent : reached CreateEvent in store")

	tweetID, err := gocql.ParseUUID(event.TweetID)
	insert := fmt.Sprintf("INSERT INTO %s (id, event_type, timestamp, timespent) VALUES (?, ?, ?, ?)", COLLECTION_EVENT)

	err = store.session.Query(
		insert, tweetID, event.Type, event.Timestamp, event.Timespent).Exec()
	if err != nil {
		store.logger.Errorf("Error in EventCassandra.CreateEvent : %s", err)
		return nil, err
	}
	return event, nil
}

func (store *EventCassandraStore) GetTimespentMonthlyEvents(ctx context.Context, event *domain.Event) (int64, error) {
	ctx, span := store.tracer.Start(ctx, "EventStore.GetTimespentMonthlyEvents")
	defer span.End()

	store.logger.Infoln("EventCassandra.CreateEvent : reached CreateEvent in store")

	thisT := time.Unix(event.Timestamp, 0)
	localTime := time.Date(thisT.Year(), thisT.Month(), thisT.Day(), thisT.Hour()+1, 0, 0, 0, time.Local)

	firstMonthDate := time.Date(localTime.Year(), localTime.Month(), 1, 0, 0, 0, 0, time.Local)
	lastMonthDate := time.Date(localTime.Year(), localTime.Month()+1, 1, 0, 0, 0, 0, time.Local)

	insert := fmt.Sprintf("SELECT COUNT(timespent), SUM(timespent) FROM %s WHERE id = ? "+
		"AND event_type = ? AND timestamp >= ? AND timestamp < ?",
		COLLECTION_EVENT)
	scanner := store.session.Query(
		insert, event.TweetID, event.Type, firstMonthDate.Unix(), lastMonthDate.Unix()).Iter().Scanner()

	var timeSum int64
	var entries int64
	for scanner.Next() {
		err := scanner.Scan(&entries, &timeSum)
		if err != nil {
			store.logger.Errorf("Error in getting timespent ==> EventStore.GetTimespentMonthlyEvents: %s", err.Error())
			return 0, err
		}
	}

	if entries == 0 {
		return 0, nil
	}
	return timeSum / entries, nil
}

func (store *EventCassandraStore) GetTimespentDailyEvents(ctx context.Context, event *domain.Event) (int64, error) {
	ctx, span := store.tracer.Start(ctx, "EventStore.GetTimespentDailyEvents")
	defer span.End()

	store.logger.Infoln("EventCassandra.CreateEvent : reached CreateEvent in store")

	thisT := time.Unix(event.Timestamp, 0)
	localTime := time.Date(thisT.Year(), thisT.Month(), thisT.Day(), thisT.Hour()+1, 0, 0, 0, time.Local)
	startOfDay := time.Date(localTime.Year(), localTime.Month(), localTime.Day(), 0, 0, 0, 0, time.Local)
	endOfDay := startOfDay.Add(24 * time.Hour)

	insert := fmt.Sprintf("SELECT COUNT(timespent), SUM(timespent) FROM %s WHERE id = ? "+
		"AND event_type = ? AND timestamp >= ? AND timestamp < ?", COLLECTION_EVENT)

	scanner := store.session.Query(
		insert, event.TweetID, event.Type, startOfDay.Unix(), endOfDay.Unix()).Iter().Scanner()

	var timeSum int64
	var entries int64
	for scanner.Next() {
		err := scanner.Scan(&entries, &timeSum)
		if err != nil {
			store.logger.Errorf("Error in getting timespent ==> EventStore.GetTimespentMonthlyEvents: %s", err.Error())
			return 0, err
		}
	}

	if entries == 0 {
		return 0, nil
	}
	return timeSum / entries, nil
}

package domain

import (
	"context"
)

type EventStore interface {
	CreateEvent(context.Context, *Event) (*Event, error)
	GetTimespentDailyEvents(ctx context.Context, event *Event) (int64, error)
	GetTimespentMonthlyEvents(ctx context.Context, event *Event) (int64, error)
}

package domain

import (
	"context"
	events "github.com/zjalicf/twitter-clone-common/common/saga/create_event"
)

type ReportStore interface {
	CreateReport(context.Context, *events.Event, int64, int64) (*events.Event, error)
	GetReportForAd(ctx context.Context, tweetID string, reportType string, date int64) (*Report, error)
}

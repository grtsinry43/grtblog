package backup

import (
	"context"
	"errors"
	"time"
)

var (
	ErrNotFound      = errors.New("backup not found")
	ErrBackupRunning = errors.New("another backup is already running")
	ErrInvalidTicket = errors.New("invalid or expired download ticket")
)

type Repository interface {
	Create(context.Context, *Record) error
	Update(context.Context, *Record) error
	Get(context.Context, string) (*Record, error)
	List(context.Context) ([]Record, error)
	Delete(context.Context, string) error
	MarkInterrupted(context.Context) error
	CreateTicket(context.Context, DownloadTicket) error
	ResolveTicket(context.Context, string) (*Record, error)
	DeleteExpiredTickets(context.Context) error
	GetSchedule(context.Context) (*Schedule, error)
	SaveSchedule(context.Context, *Schedule) error
	TryClaimSchedule(context.Context, time.Time) (bool, *Schedule, error)
	SetPinned(context.Context, string, bool) error
}

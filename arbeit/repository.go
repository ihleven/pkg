package arbeit

import (
	"time"

	"github.com/ihleven/pkg/kalender"
)

type Repository interface {
	// RetrieveArbeitsjahr(year int, accountID int) (*Arbeitsjahr, error)
	// RetrieveArbeitsmonat(year, month, accountID int) (*Arbeitsmonat, error)
	ListArbeitstage(accountID, year, month, week int) ([]Arbeitstag, error)
	ReadArbeitstag(int, time.Time) (*Arbeitstag, error)
	// UpdateArbeitstag(int, *Arbeitstag) error

	ListZeitspannen(account int, datum time.Time) ([]Zeitspanne, error)
	UpsertZeitspanne(account int, datum time.Time, z *Zeitspanne) error
	DeleteZeitspanne(account int, datum time.Time, nr int) error

	Close()

	SelectArbeitsmonate(int, int, int) ([]Arbeitsmonat, error)
	RetrieveArbeitsjahre(account, jahr int) ([]Arbeitsjahr, error)

	UpsertKalendertag(k kalender.Tag) error
	// UpsertArbeitstag(a *Arbeitstag) error
	SaveArbeitstag(account int, datum time.Time, job string, a Arbeitstag) error
	SetupArbeitsjahr(int, string, int, *time.Time, *time.Time) error
	SetupArbeitsmonat(account int, job string, year, month int) error

	ListUrlaube(account, jahr, nr int) ([]Urlaub, error)
}

// var Repo Repository

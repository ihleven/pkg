package repository

import (
	"fmt"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/pkg/errors"
)

const (
	name = "cloud11"
	user = "matt"
	pwrd = "dummy"
	host = "localhost"
	port = 5432
)

// NewPostgresRepository ...
func NewPostgresRepository() (*Repository, error) {

	dbinfo := fmt.Sprintf("host=localhost port=5432 user=%s password=%s dbname=%s sslmode=disable", user, pwrd, name)
	db, err := sqlx.Open("postgres", dbinfo)
	if err != nil {
		return nil, errors.Wrapf(err, "Could not get database handle: '%s'", dbinfo)
	}

	err = db.Ping()
	if err != nil {
		return nil, errors.Wrapf(err, "Could not verify database connection.")
	}
	return &Repository{db}, nil
}

// Repository ...
type Repository struct {
	DB *sqlx.DB
}

// Close ...
func (r Repository) Close() {
	fmt.Println("closing repository")
	r.DB.Close()
}

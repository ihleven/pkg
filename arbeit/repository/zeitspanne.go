package repository

import (
	"fmt"
	"time"

	"github.com/ihleven/errors"
	"github.com/ihleven/pkg/arbeit"
	pq "github.com/lib/pq"
)

func (r *Repository) ListZeitspannen(account int, date time.Time) ([]arbeit.Zeitspanne, error) {

	query := `SELECT nr, status, start, ende, dauer
		        FROM c11_zeitspanne
		       WHERE account=$1 AND datum=$2
			ORDER BY nr`

	zeitspannen := []arbeit.Zeitspanne{}
	err := r.DB.Select(&zeitspannen, query, account, date)

	return zeitspannen, err
}

func (r *Repository) UpsertZeitspanne(account int, datum time.Time, z *arbeit.Zeitspanne) error {

	stmt := `
		INSERT INTO c11_zeitspanne 
		        	(account,datum,nr,status,start,ende,dauer)
			 VALUES ($1,$2,$3,$4,$5,$6,$7)`

	_, err := r.DB.Exec(stmt, account, time.Time(datum), z.Nr, z.Status, z.Start, z.Ende, z.Dauer)
	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok && pqErr.Code != "23505" { //"23505": "unique_violation",
			return errors.Wrap(err, "Could not insert zeitspanne %d, %s, %d, %s, %s, %s, %s", account, datum, z.Nr, z.Status, z.Start, z.Ende, z.Dauer)
		}
		fmt.Println("-----------------", err)
	}

	stmt = `
		UPDATE c11_zeitspanne 
	   	   SET status=$1,start=$2,ende=$3,dauer=$4
	 	 WHERE account=$5 AND datum=$6 AND nr=$7
	`
	fmt.Println("usert:", z.Status)
	_, err = r.DB.Exec(stmt, z.Status, z.Start, z.Ende, z.Dauer, account, time.Time(datum), z.Nr)
	if err != nil {
		return errors.Wrap(err, "Could not update zeitspanne %d, %s, %d, %s, %s, %s, %s", account, datum, z.Nr, z.Status, z.Start, z.Ende, z.Dauer)
	}
	return nil
}

func (r Repository) DeleteZeitspanne(account int, datum time.Time, nr int) error {

	stmt := `DELETE FROM c11_zeitspanne WHERE account=$1 AND datum=$2 AND nr=$3`
	_, err := r.DB.Exec(stmt, account, datum, nr)
	if err != nil {
		return errors.Wrap(err, "Could not delete go_zeitspanne %s %d", datum, nr)
	}
	return nil
}

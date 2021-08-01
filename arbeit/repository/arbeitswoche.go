package repository

import (
	"github.com/ihleven/pkg/arbeit"

	"github.com/pkg/errors"

	_ "github.com/lib/pq"
)

func (r Repository) RetrieveArbeitstage(year, month, week int, accountID int) (a []arbeit.Arbeitstag, err error) {

	aa := []arbeit.Arbeitstag{}

	query := `
		SELECT a.id, status, kategorie, soll, beginn, ende, brutto, pausen, netto, differenz 
		  FROM go_arbeitstag a, kalendertag k  
		 WHERE a.tag_id=k.id
	`
	if month != 0 {
		query += "AND k.jahr_id=$1 AND k.monat=$2"
		err = r.DB.Select(&aa, query, year, month)
	}
	if week != 0 {
		query += "AND k.kw_jahr=$1 AND k.kw_jahr=$2"
		err = r.DB.Select(&aa, query, year, week)
	}
	if err != nil {
		err = errors.Wrapf(err, "Could not Select  arbeitstage %v", aa)
	}

	return
}

package repository

import (
	"fmt"

	"github.com/ihleven/errors"
	"github.com/ihleven/pkg/kalender"
	pq "github.com/lib/pq"
)

func (r *Repository) UpsertKalendertag(k kalender.Tag) error {
	// fmt.Println("UpsertKalendertag")

	stmt := `
		INSERT INTO c11_datum (id,datum,jahr,monat,tag,jahrtag,kw_jahr,kw,kw_tag,feiertag)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)
	`
	_, err := r.DB.Exec(stmt, k.ID, k.Datum, k.Jahr, k.Monat, k.Tag, k.Jahrtag, k.KwJahr, k.KwNr, k.KwTag, k.Feiertag)
	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok && pqErr.Code != "23505" { //"23505": "unique_violation",
			return errors.Wrap(err, "Could not insert kalendertag %d", k.ID)
		}
	}

	stmt = `
		UPDATE c11_datum 
	   	   SET datum=$2,jahr=$3,monat=$4,tag=$5,jahrtag=$6,kw_jahr=$7,kw=$8,kw_tag=$9,feiertag=$10
	 	 WHERE id=$1
	`
	n, err := r.DB.Exec(stmt, k.ID, k.Datum, k.Jahr, k.Monat, k.Tag, k.Jahrtag, k.KwJahr, k.KwNr, k.KwTag, k.Feiertag)
	fmt.Println("UpsertKalendertag", n, err)
	if err != nil {
		return errors.Wrap(err, "Could not update kalendertag %d", k.ID)
	}
	return nil
}

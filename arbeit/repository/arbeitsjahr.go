package repository

import (
	"fmt"
	"time"

	"github.com/ihleven/pkg/arbeit"
	pq "github.com/lib/pq"

	"github.com/ihleven/errors"
)

func (r *Repository) SetupArbeitsjahr(account int, job string, year int, von, bis *time.Time) error {

	stmt := `
		INSERT INTO c11_arbeitsjahr (account, job, jahr, von, bis) 
		VALUES ($1, $2, $3, $4, $5)
	`
	n, err := r.DB.Exec(stmt, account, job, year, von, bis)
	if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "23505" { //"23505": "unique_violation"
		return nil
	}
	if err != nil {
		return errors.Wrap(err, "Could not setup c11_arbeitsjahr")
	}
	fmt.Printf("Successfully set up Arbeitsjahr: %d %s %d -> %d rows affected", account, job, year, n)
	return nil
}

func (r *Repository) SaveArbeitsjahr(a arbeit.Arbeitsjahr) error {

	err := r.InsertArbeitsjahr(a)
	if pqErr, ok := errors.Cause(err).(*pq.Error); ok && pqErr.Code == "23505" { //"23505": "unique_violation"
		err = r.UpdateArbeitsjahr(a)
	}
	if err != nil {
		return errors.Wrap(err, "Could not save arbeitsjahr")
	}
	return nil
}

func (r *Repository) InsertArbeitsjahr(a arbeit.Arbeitsjahr) error {

	stmt := `
		INSERT INTO c11_arbeitsjahr (
			account, job, jahr, 
			von, bis, ARBTG , K, B, H, D, ZA, soll, ist, diff, saldo, zeiterfassung, uvorj, uansp, usond, ugepl, urest, uausz
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21, $22)
	`
	_, err := r.DB.Exec(stmt, a.Account, a.Job, a.Jahr,
		a.Von, a.Bis, a.Arbeitstage, a.Krankheitstage, a.Bürotage, a.Homeoffice, a.Dienstreise, a.Zeitausgleich,
		a.Soll, a.Ist, a.Diff, a.Saldo, a.Zeiterfassung, a.Urlaub.Vorjahr, a.Urlaub.Anspruch, a.Urlaub.Sonderurlaub, a.Urlaub.Geplant, a.Urlaub.Rest, a.Urlaub.Auszahlung,
	)
	if err != nil {
		return errors.Wrap(err, "Could not insert c11_arbeitsjahr")
	}
	return nil
}

func (r *Repository) UpdateArbeitsjahr(a arbeit.Arbeitsjahr) error {

	stmt := `
		UPDATE c11_arbeitsjahr 
		   SET von=$4, bis=$5, ARBTG=$6, K=$7, B=$8, H=$9, D=$10, ZA=$11, 
		       soll=$12, ist=$13, diff=$14, saldo=$15, zeiterfassung=$16,
			   uvorj=$17, uansp=$18, usond=$19, ugepl=$20, urest=$21, uausz=$22 
		 WHERE account=$1 AND job=$2 AND jahr=$3
	`
	_, err := r.DB.Exec(stmt, a.Account, a.Job, a.Jahr,
		a.Von, a.Bis, a.Arbeitstage, a.Krankheitstage, a.Bürotage, a.Homeoffice, a.Dienstreise, a.Zeitausgleich,
		a.Soll, a.Ist, a.Diff, a.Saldo, a.Zeiterfassung,
		a.Urlaub.Vorjahr, a.Urlaub.Anspruch, a.Urlaub.Sonderurlaub, a.Urlaub.Geplant, a.Urlaub.Rest, a.Urlaub.Auszahlung,
	)
	if err != nil {
		return errors.Wrap(err, "Could not update c11_arbeitsjahr %v", a)
	}
	return nil
}

// type Arbeitsjahr struct {
// 	ID             int
// 	UrlaubVorjahr  sql.NullFloat64
// 	UrlaubAnspruch sql.NullFloat64
// 	UrlaubTage     sql.NullFloat64
// 	UrlaubGeplant  sql.NullFloat64
// 	UrlaubRest     sql.NullFloat64
// 	Soll           sql.NullFloat64
// 	Ist            sql.NullFloat64
// 	Differenz      sql.NullFloat64
// 	// TageArbeit            sql.NullFloat64
// 	// TageKrank             sql.NullFloat64
// 	// tageFreizeitausgleich sql.NullFloat64
// 	// tageBuero             sql.NullFloat64
// 	// tageDienstreise       sql.NullFloat64
// 	// tageHomeoffice        sql.NullFloat64
// 	// tageFrei              sql.NullFloat64
// 	// jahrID                sql.NullInt64
// 	// userID                sql.NullInt64
// 	// Monate                []Arbeitsmonat
// }

func (r *Repository) RetrieveArbeitsjahre(account, jahr int) ([]arbeit.Arbeitsjahr, error) {

	arbeitsjahre := make([]arbeit.Arbeitsjahr, 0)

	query := `
		SELECT account, job, jahr, von, bis, ARBTG , K, B, H, D, ZA, soll, ist, diff, saldo, zeiterfassung, uvorj, uansp, usond, ugepl, urest, uausz		
		  FROM c11_arbeitsjahr 
		 WHERE account=$1
	`
	if jahr != 0 {
		query += " AND jahr=$2 ORDER BY job"
	} else {
		query += " AND jahr>$2 ORDER BY jahr, von"
	}

	rows, err := r.DB.Query(query, account, jahr)
	if err != nil {
		return nil, errors.Wrap(err, "Could not query for rows")
	}
	defer rows.Close()

	a := arbeit.Arbeitsjahr{}

	for rows.Next() {
		err := rows.Scan(
			&a.Account,
			&a.Job,
			&a.Jahr,
			&a.Von,
			&a.Bis,
			&a.Arbeitstage,
			&a.Krankheitstage,
			&a.Bürotage,
			&a.Homeoffice,
			&a.Dienstreise,
			&a.Zeitausgleich,
			&a.Soll,
			&a.Ist,
			&a.Diff,
			&a.Saldo,
			&a.Zeiterfassung,
			&a.Urlaub.Vorjahr,
			&a.Urlaub.Anspruch,
			// &a.Urlaub.Tage,
			&a.Urlaub.Sonderurlaub,
			&a.Urlaub.Geplant,
			&a.Urlaub.Rest,
			&a.Urlaub.Auszahlung,
		)
		if err != nil {
			return nil, errors.Wrap(err, "Could not scan for row (RetrieveArbeitsjahre)")
		}
		arbeitsjahre = append(arbeitsjahre, a)
	}
	err = rows.Err()
	if err != nil {
		return nil, errors.Wrap(err, "rows error (RetrieveArbeitsjahre)")
	}
	return arbeitsjahre, nil
}

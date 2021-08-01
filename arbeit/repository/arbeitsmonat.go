package repository

import (
	"fmt"

	"github.com/ihleven/errors"
	"github.com/ihleven/pkg/arbeit"

	_ "github.com/lib/pq"
	pq "github.com/lib/pq"
)

func (r *Repository) SetupArbeitsmonat(account int, job string, year, month int) error {

	stmt := `
		INSERT INTO c11_arbeitsmonat (account, job, jahr, monat) 
		VALUES ($1, $2, $3, $4)
	`
	n, err := r.DB.Exec(stmt, account, job, year, month)
	if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "23505" { //"23505": "unique_violation"
		return nil
	}
	if err != nil {
		return errors.Wrap(err, "Could not setup c11_arbeitsmonat")
	}
	fmt.Printf("Successfully set up Arbeitsmonat: %d %s %d -> %d rows affected", account, job, year, n)
	return nil
}

func (r *Repository) SelectArbeitsmonate(accountID int, year, month int) ([]arbeit.Arbeitsmonat, error) {

	arbeitsmonate := []arbeit.Arbeitsmonat{}
	var err error

	query := `
		SELECT monat, a, k, u, soll, ist, diff, saldo, zeiterfassung
		  FROM c11_arbeitsmonat m 
		 WHERE m.account=$1 AND m.jahr=$2 
	`
	if month != 0 {
		query += "AND m.monat=$3"
		err = r.DB.Select(&arbeitsmonate, query, accountID, year, month)
	} else {
		query += "ORDER BY monat"
		err = r.DB.Select(&arbeitsmonate, query, accountID, year)
	}
	if err != nil {
		return nil, errors.Wrap(err, "Could not Select Arbeitsmonate for (%d, %d, %d)", accountID, year, month)
	}
	return arbeitsmonate, nil
}

// func (r Repository) RetrieveArbeitsmonat(year, month int, accountID int) (*arbeit.Arbeitsmonat, error) {
// 	fmt.Println("RetrieveArbeitsmonat", year, month)

// 	arbeitsmonat := []arbeit.Arbeitstag{}

// 	query := `
// 		SELECT k.*, a.*
// 		  FROM kalendertag k LEFT OUTER JOIN go_arbeitstag a
// 		    ON k.id=a.tag_id
// 		 WHERE k.jahr_id=$1 AND k.monat=$2
// 	`
// 	query = `
// 		SELECT a.id, status, kategorie, krankmeldung, urlaubstage, soll, beginn, ende, brutto, pausen, extra, netto, differenz,
// 				k.jahr_id, k.monat, k.tag, k.datum, k.feiertag, k.kw_jahr, k.kw_nr , k.kw_tag, k.jahrtag, k.ordinal
// 		  FROM go_arbeitstag a, kalendertag k
// 		 WHERE a.tag_id=k.id
// 		   AND k.jahr_id=$1 AND k.monat=$2
// 	  `
// 	err := r.DB.Select(&arbeitsmonat, query, year, month)

// 	if err != nil {
// 		err = errors.Wrap(err, "Could not Select  arbeitstage %v", arbeitsmonat)
// 	}
// 	//fmt.Printf("monat: %v\n", arbeitsmonat)
// 	return nil, nil
// }

// func (r Repository) RetrieveJahresArbeitsmonate(year, accountID int) ([]arbeit.Arbeitsmonat, error) {
// 	fmt.Println("RetrieveJahresArbeitsmonat", year)

// 	arbeitsmonat := []arbeit.Arbeitstag{}

// 	query := `
// 		SELECT k.*, a.*
// 		  FROM kalendertag k LEFT OUTER JOIN go_arbeitstag a
// 		    ON k.id=a.tag_id
// 		 WHERE k.jahr_id=$1 AND k.monat=$2
// 	`
// 	query = `
// 		SELECT a.id, status, kategorie, krankmeldung, urlaubstage, soll, beginn, ende, brutto, pausen, extra, netto, differenz,
// 				k.jahr_id, k.monat, k.tag, k.datum, k.feiertag, k.kw_jahr, k.kw_nr , k.kw_tag, k.jahrtag, k.ordinal
// 		  FROM go_arbeitstag a, kalendertag k
// 		 WHERE a.tag_id=k.id
// 		   AND k.jahr_id=$1 AND k.monat=$2
// 	  `
// 	err := r.DB.Select(&arbeitsmonat, query, year)

// 	if err != nil {
// 		err = errors.Wrap(err, "Could not Select  arbeitstage %v", arbeitsmonat)
// 	}
// 	//fmt.Printf("monat: %v\n", arbeitsmonat)
// 	return nil, nil
// }

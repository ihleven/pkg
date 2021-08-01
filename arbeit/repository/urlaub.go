package repository

import (
	"github.com/ihleven/errors"
	"github.com/ihleven/pkg/arbeit"
)

func (r *Repository) ListUrlaube(account, jahr, nr int) ([]arbeit.Urlaub, error) {

	urlaube := []arbeit.Urlaub{}

	query := `
		SELECT account, job, jahr, nr, von, bis, num_urlaub, num_ausgl, num_sonder, grund, beantragt, genehmigt, kommentar
		  FROM c11_urlaub a
		 WHERE account=$1 AND jahr=$2
	`
	err := r.DB.Select(&urlaube, query, account, jahr)
	if err != nil {
		return nil, errors.Wrap(err, "Could not select urlaube for ($d, $d)", account, jahr)
	}
	return urlaube, nil
}

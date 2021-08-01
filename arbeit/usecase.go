package arbeit

import (
	"fmt"
	"time"

	"github.com/ihleven/errors"
	"github.com/ihleven/pkg/kalender"
)

func NewUsecase(repo Repository) *Usecase {
	return &Usecase{repo}
}

type Usecase struct {
	repo Repository
}

func (uc *Usecase) ListArbeitsjahre(account int) (Arbeitsjahre, error) {

	arbeitsjahre, err := uc.repo.RetrieveArbeitsjahre(account, 0)
	if err != nil {
		return nil, err
	}
	return arbeitsjahre, nil
}

func (uc *Usecase) Arbeitsjahr(account, year int) (*Arbeitsjahr, error) {

	arbeitsjahre, err := uc.repo.RetrieveArbeitsjahre(account, year)
	if err != nil {
		fmt.Printf("err: %+v\n", err)
		return nil, err
	}
	if len(arbeitsjahre) == 0 {
		return nil, errors.NewWithCode(errors.NotFound, "Arbeitsjahr %d for account %d not found", year, account)
	}
	arbeitsjahr := arbeitsjahre[0]

	arbeitsjahr.Monate, err = uc.repo.SelectArbeitsmonate(account, year, 0)
	if err != nil {
		return nil, err
	}
	arbeitsjahr.Urlaube, err = uc.repo.ListUrlaube(account, year, 0)
	if err != nil {
		return nil, err
	}
	return &arbeitsjahr, nil
}

func (uc *Usecase) SetupArbeitsjahr(year, account int) (*Arbeitsjahr, error) {

	err := uc.repo.SetupArbeitsjahr(account, "IC", year, nil, nil)
	if err != nil {
		fmt.Println(err)
		// return nil, err
	}
	for month := 1; month <= 12; month++ {
		err := uc.repo.SetupArbeitsmonat(account, "IC", year, month)
		if err != nil {
			fmt.Println(err)
			// return nil, err
		}
	}
	for _, k := range kalender.ListKalendertage(year) {

		err := uc.repo.UpsertKalendertag(k)
		if err != nil {
			return nil, err
		}
		arbeitstag := Arbeitstag{
			Account: 1, Datum: Date(k.Datum), Job: "IC", Status: "A", Kategorie: "-", Soll: 8,
			Kommentar: "testkommentar",
		}
		if k.KwTag > 5 {
			arbeitstag.Status = "-"
			arbeitstag.Soll = 0
		}
		if k.Feiertag != "" {
			arbeitstag.Status = "F"
		}
		err = uc.repo.SaveArbeitstag(1, k.Datum, "IC", arbeitstag)
		if err != nil {
			fmt.Println(err)
		}
	}

	arbeitsjahre, err := uc.repo.RetrieveArbeitsjahre(account, year)
	if err != nil {
		return nil, err
	}
	if len(arbeitsjahre) == 0 {
		return nil, errors.NewWithCode(errors.NotFound, "Arbeitsjahr %d for account %d not found", year, account)
	}
	arbeitsjahr := arbeitsjahre[0]

	arbeitsjahr.Monate, err = uc.repo.SelectArbeitsmonate(account, year, 0)
	return &arbeitsjahr, nil
}

func (uc *Usecase) GetArbeitsmonat(year int, month int, account int) (*Arbeitsmonat, error) {

	am, err := uc.repo.SelectArbeitsmonate(account, year, month)

	if len(am) == 0 {
		return nil, nil
	}
	m := am[0]

	m.Tage, err = uc.repo.ListArbeitstage(account, year, month, 0)
	if err != nil {
		return nil, err
	}
	//return &Arbeitsmonat{m}, nil

	return &m, err
}

/// ARBEITSTAG ///
func (uc *Usecase) GetArbeitstag(year, month, day int, accountID int) (*Arbeitstag, error) {

	datum := time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.UTC)
	at, err := uc.repo.ReadArbeitstag(accountID, datum)
	if err != nil {
		//return &Arbeitstag{}, nil
		return nil, errors.Wrap(err, "Could not read Arbeitstag: %d/%d/%d, %d", year, month, day, accountID)
	}
	return at, nil
}

func (uc *Usecase) UpdateArbeitstag(year, month, day int, accountID int, arbeitstag *Arbeitstag) (*Arbeitstag, error) {

	// id := ((year*100+month)*100+day)*1000 + accountID
	datum := time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.UTC)
	fmt.Println("usecase update arbeitstag", year, month, day, accountID, arbeitstag, datum)

	arbeitstag.Pausen, _ = uc.UpdateZeitspannen(accountID, datum, arbeitstag.Zeitspannen)
	arbeitstag.Extra = 0

	// arbeitstagDB, err := Repo.ReadArbeitstag(id)
	// if err != nil {
	// 	fmt.Println("read at error:", err)
	// 	return errors.Wrapf(err, "Could not retrieve Arbeitstag %s%s%s", year, month, day)
	// }
	if arbeitstag.Start != nil && arbeitstag.Ende != nil {
		arbeitstag.Brutto = arbeitstag.Ende.Sub(*arbeitstag.Start).Hours()
		arbeitstag.Netto = arbeitstag.Brutto - arbeitstag.Pausen + arbeitstag.Extra
		arbeitstag.Differenz = arbeitstag.Soll - arbeitstag.Netto
	}

	err := uc.repo.SaveArbeitstag(accountID, datum, "IC", *arbeitstag)
	if err != nil {
		return nil, errors.Wrap(err, "Could not update Arbeitstag %v", datum)
	}
	fmt.Println("sucess update arbeitstag", datum)
	return nil, nil
}

func (uc *Usecase) UpdateZeitspannen(account int, datum time.Time, zeitspannen []Zeitspanne) (float64, error) {

	pausen := 0.0

	// Zeitspannen in der DB loeschen, deren Nr. es nicht mehr gibt
	dbZeitspannen, err := uc.repo.ListZeitspannen(account, datum)
	if err != nil {
		return 0.0, err
	}
	for _, dbZeitspanne := range dbZeitspannen {
		if !IsContained(zeitspannen, dbZeitspanne) {
			uc.repo.DeleteZeitspanne(account, datum, dbZeitspanne.Nr)
		}
	}
	// Insert oder Update Zeitspannen
	for _, zeitspanne := range zeitspannen {
		dauer := zeitspanne.Ende.Sub(*zeitspanne.Start).Hours()

		zeitspanne.Dauer = dauer
		pausen += dauer
		fmt.Println("Dauer: ", dauer, zeitspanne)
		err := uc.repo.UpsertZeitspanne(account, datum, &zeitspanne)
		if err != nil {
			fmt.Println("error upsert:", err)
			return 0.0, err
		}
	}
	return pausen, nil
}

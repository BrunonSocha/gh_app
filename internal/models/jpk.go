package models

import (
	"database/sql"
	"encoding/xml"
	"errors"
	"math"
	"time"
)

type JPKModel struct {
	DB *sql.DB
}

type JPKMetadata struct {
	Id int
	ConfirmedAt *time.Time
	UPO *string
	Rok int
	Miesiac int
}

type JPK struct {
	XMLName xml.Name `xml:JPK`
	XMLTypes string `xml:"xmlns:etd,attr"`
	XMLSchema string `xml:"xmlns:xsi,attr"`
	XMLPattern string `xml:"xmlns,attr"`
	Naglowek NaglowekJPK `xml:"Naglowek"`
	Podmiot1 Podmiot1 `xml:"Podmiot1"`
	Deklaracja Deklaracja `xml:"Deklaracja"`
	Ewidencja Ewidencja `xml:"Ewidencja"`
}

type NaglowekJPK struct {
	KodFormularza KodFormularza `xml:"KodFormularza"`
	WariantFormularza int `xml:"WariantFormularza"`
	DataWytworzeniaJPK string `xml:"DataWytworzeniaJPK"`
	NazwaSystemu string `xml:"NazwaSystemu"`
	CelZlozenia CelZlozenia `xml:"CelZlozenia"`
	KodUrzedu int `xml:"KodUrzedu"`
	Rok int `xml:"Rok"`
	Miesiac int `xml:"Miesiac"`
}

type CelZlozenia struct {
	Poz string `xml:"poz,attr"`
	Cel int `xml:",chardata"`
}

type KodFormularza struct {
	KodSystemowy string `xml:"kodSystemowy,attr"`
	WersjaSchemy string `xml:"wersjaSchemy,attr"`
	Kod string `xml:",chardata"`
}

type Podmiot1 struct {
	Rola string `xml:"rola,attr"`
	OsobaNiefizyczna OsobaNiefizyczna `xml:"OsobaNiefizyczna"`
}

type OsobaNiefizyczna struct {
	NIP string `xml:"NIP"`
	PelnaNazwa string `xml:"PelnaNazwa"`
	Email string `xml:"Email"`
	Telefon string `xml:"Telefon"`
}

type Deklaracja struct {
	Naglowek NaglowekDekl `xml:"Naglowek"`
	PozycjeSzczegolowe PozycjeSzczegolowe `xml:"PozycjeSzczegolowe"`
	Pouczenia int `xml:"Pouczenia"`
}

type NaglowekDekl struct {
	KodFormularzaDekl KodFormularzaDekl `xml:"KodFormularzaDekl"`
	WariantFormularzaDekl int `xml:"WariantFormularzaDekl"`
}

type KodFormularzaDekl struct {
	KodSystemowy string `xml:"kodSystemowy,attr"`
	KodPodatku string `xml:"kodPodatku,attr"`
	RodzajZobowiazania string `xml:"rodzajZobowiazania,attr"`
	WersjaSchemy string `xml:"wersjaSchemy,attr"`
	Kod string `xml:",chardata"`
}

type PozycjeSzczegolowe struct {
	P_37 int `xml:"P_37"`
	P_38 int `xml:"P_38"`
	P_39 int `xml:"P_39"`
	P_42 int `xml:"P_42"`
	P_43 int `xml:"P_43"`
	P_48 int `xml:"P_48"`
	P_51 int `xml:"P_51"`
	P_53 int `xml:"P_53"`
	P_62 int `xml:"P_62"`
	P_68 int `xml:"P_68"`
	P_69 int `xml:"P_69"`
}

type Ewidencja struct {
	SprzedazWiersz []SprzedazWiersz `xml:"SprzedazWiersz"`
	SprzedazCtrl SprzedazCtrl `xml:"SprzedazCtrl"`
	ZakupWiersz []ZakupWiersz `xml:"ZakupWiersz"`
	ZakupCtrl ZakupCtrl `xml:"ZakupCtrl"`
}

type SprzedazWiersz struct {
	LpSprzedazy int `xml:"LpSprzedazy"`
	KodKrajuNadaniaTIN string `xml:"KodKrajuNadaniaTIN"`
	NrKontrahenta string `xml:"NrKontrahenta"`
	NazwaKontrahenta string `xml:"NazwaKontrahenta"`
	DowodSprzedazy string `xml:"DowodSprzedazy"`
	DataWystawienia string `xml:"DataWystawienia"`
	K_19 float64 `xml:"K_19"`
	K_20 float64 `xml:"K_20"`
}

type SprzedazCtrl struct {
	LiczbaWierszySprzedazy int `xml:"LiczbaWierszySprzedazy"`
	PodatekNalezny float64 `xml:"PodatekNalezny"`
}

type ZakupWiersz struct {
	LpZakupu int `xml:"LpZakupu"`
	KodKrajuNadaniaTIN string `xml:"KodKrajuNadaniaTIN"`
	NrDostawcy string `xml:"NrDostawcy"`
	NazwaDostawcy string `xml:"NazwaDostawcy"`
	DowodZakupu string `xml:"DowodZakupu"`
	DataZakupu string `xml:"DataZakupu"`
	K_42 float64 `xml:"K_42"`
	K_43 float64 `xml:"K_43"`
}

type ZakupCtrl struct {
	LiczbaWierszyZakupow int `xml:"LiczbaWierszyZakupow"`
	PodatekNaliczony float64 `xml:"PodatekNaliczony"`
}

func (m *JPKModel) NewJpk(inv []*Invoice) (*JPK, error) {
	var podatekNaliczony float64 = 0
	var podatekNalezny float64 = 0
	var podstawaSprzedazy float64 = 0
	var podstawaZakupu float64 = 0
	var sprzedazWiersz []SprzedazWiersz
	var zakupWiersz []ZakupWiersz
	var companyName string
	var poprzedniVat int
	now := time.Now()
	previousMonth := now.AddDate(0, -1, 0)
	previousPeriod := previousMonth.AddDate(0, -1, 0)
	err := m.DB.QueryRow("SELECT vat FROM JpkFiles WHERE year = @p1 AND month = @p2 AND confirmed_at IS NOT NULL", time.Now().Year(), previousPeriod.Month()).Scan(&poprzedniVat)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			poprzedniVat = 0
		} else {
			return nil, err
		}
	}
	saleCount := 0
	purcCount := 0
	for _, i := range inv {
		nameRow := m.DB.QueryRow("SELECT nazwa FROM Companies WHERE nip = @p1", i.Nip)
		err := nameRow.Scan(&companyName)
		if err != nil {
			return nil, err
		}
		switch i.Inv_type {
		case SaleInvoice:
			saleCount++
			podatekNalezny += i.Podatek
			podstawaSprzedazy += i.Netto
			sprzedazWiersz = append(sprzedazWiersz, SprzedazWiersz{LpSprzedazy: saleCount, KodKrajuNadaniaTIN: "PL", NrKontrahenta: i.Nip, NazwaKontrahenta: companyName, DowodSprzedazy: i.Nr_faktury, DataWystawienia: string(i.Data.Format("2006-01-02")), K_19: i.Netto, K_20: i.Podatek})
		case PurchaseInvoice:
			purcCount++
			podatekNaliczony += i.Podatek
			podstawaZakupu += i.Netto
			zakupWiersz = append(zakupWiersz, ZakupWiersz{LpZakupu: purcCount, KodKrajuNadaniaTIN: "PL", NrDostawcy: i.Nip, NazwaDostawcy: companyName, DowodZakupu: i.Nr_faktury, DataZakupu: string(i.Data.Format("2006-01-02")), K_42: i.Netto, K_43: i.Podatek})
		}
	}
	var p_51 int
	var p_53 int
	if podatekNalezny > (podatekNaliczony+float64(poprzedniVat)) {
		p_51 = int(math.Round(podatekNalezny) - (math.Round(podatekNaliczony) + float64(poprzedniVat)))
		p_53 = 0
	} else {
		p_51 = 0
		p_53 = int(math.Round(podatekNaliczony) + float64(poprzedniVat) - math.Round(podatekNalezny))
	}
	

	jpk := &JPK{
		XMLTypes: "http://crd.gov.pl/xml/schematy/dziedzinowe/mf/2021/06/08/eD/DefinicjeTypy/", 
		XMLSchema: "http://www.w3.org/2001/XMLSchema-instance",
		XMLPattern: "http://crd.gov.pl/wzor/2021/12/27/11148/",
		Naglowek: NaglowekJPK{
			KodFormularza: KodFormularza {
				KodSystemowy: "JPK_V7M (2)",
				WersjaSchemy: "1-0E",
				Kod: "JPK_VAT",
			},
			WariantFormularza: 2,
			DataWytworzeniaJPK: string(time.Now().Format(time.RFC3339Nano)),
			NazwaSystemu: "Formularz uproszczony",
			CelZlozenia: CelZlozenia{
				Poz: "P_7",
				Cel: 1,
			},
			KodUrzedu: 1210,
			Rok: time.Now().Year(),
			Miesiac: int(previousMonth.Month()),
		},
		Podmiot1: Podmiot1{
			Rola: "Podatnik",
			OsobaNiefizyczna: OsobaNiefizyczna{
				NIP: "6793194113",
				PelnaNazwa: "Grey House sp. z o.o.",
				Email: "info@greyhouse.es",
				Telefon: "608415900",
			},
		},
		Deklaracja: Deklaracja {
			Naglowek: NaglowekDekl{
				KodFormularzaDekl: KodFormularzaDekl {
					KodSystemowy: "VAT-7 (22)",
					KodPodatku: "VAT",
					RodzajZobowiazania: "Z",
					WersjaSchemy: "1-0E",
					Kod: "VAT-7",
				},
				WariantFormularzaDekl: 22,
			},
			PozycjeSzczegolowe: PozycjeSzczegolowe{
				P_37: int(math.Round(podstawaSprzedazy)),
				P_38: int(math.Round(podatekNalezny)),
				P_39: poprzedniVat,
				P_42: int(math.Round(podstawaZakupu)),
				P_43: int(math.Round(podatekNaliczony)),
				P_48: poprzedniVat + int(math.Round(podatekNaliczony)),
				P_51: p_51,
				P_53: p_53,
				P_62: p_53,
				P_68: 0,
				P_69: 0,
			},
			Pouczenia: 1,
		},
		Ewidencja: Ewidencja{
			SprzedazWiersz: sprzedazWiersz,
			SprzedazCtrl: SprzedazCtrl{
				LiczbaWierszySprzedazy: saleCount,
				PodatekNalezny: podatekNalezny,
			},
			ZakupWiersz: zakupWiersz,
			ZakupCtrl: ZakupCtrl{
				LiczbaWierszyZakupow: purcCount,
				PodatekNaliczony: podatekNaliczony,
			},
		},
	}
	return jpk, nil
}

func (m *JPKModel) InsertDB(jpk *JPK, jpk_data string) (int, error) {
	// implement insertion to DB after creation
	stmt := "INSERT INTO JpkFiles(year, month, xml_content, generated_at) OUTPUT Inserted.id VALUES(@p1, @p2, @p3, @p4)"
	var resId int
	err := m.DB.QueryRow(stmt, jpk.Naglowek.Rok, jpk.Naglowek.Miesiac, jpk_data, time.Now()).Scan(&resId)
	if err != nil {
		return 0, err
	}
	return resId, nil
}

func (m *JPKModel) Get(id int) (*JPK, *JPKMetadata, error){
	stmt := "SELECT xml_content, id, confirmed_at, upo_reference_number FROM JpkFiles WHERE id = @p1"
	row := m.DB.QueryRow(stmt, id)
	var byteArray []byte
	jpkmetadata := &JPKMetadata{}
	err := row.Scan(&byteArray, &jpkmetadata.Id, &jpkmetadata.ConfirmedAt, &jpkmetadata.UPO)
	if err != nil || len(byteArray) == 0{
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil, ErrNoRecord
		} else {
			return nil, nil, err
		}
	}

	jpk := &JPK{}
	err = xml.Unmarshal(byteArray, jpk)
	if err != nil {
		return nil, nil, err
	}
	return jpk, jpkmetadata, nil
}

func (m *JPKModel) GetAll() ([]*JPKMetadata, error) {
	stmt := "SELECT id, confirmed_at, upo_reference_number, year, month FROM JpkFiles"
	rows, err := m.DB.Query(stmt)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	jpkfiles := []*JPKMetadata{}
	for rows.Next() {
		jpkdata := &JPKMetadata{}
		err = rows.Scan(&jpkdata.Id, &jpkdata.ConfirmedAt, &jpkdata.UPO, &jpkdata.Rok, &jpkdata.Miesiac)
		if err != nil {
			return nil, err
		}
		jpkfiles = append(jpkfiles, jpkdata)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}
	return jpkfiles, nil
}

func (m *JPKModel) Confirm(id int) (error) {
	stmt := "UPDATE JpkFiles SET confirmed_at = @p1 WHERE id = @p2"
	rows, err := m.DB.Exec(stmt, time.Now(), id)
	if err != nil {
		return err
	}
	rowsAff, err := rows.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAff == 0 {
		return ErrNoRecord
	}
	return nil
}

func (m *JPKModel) Delete(id int) (error) {
	stmt := "DELETE FROM JpkFiles WHERE id = @p1 AND confirmed_at IS NULL"
	row, err := m.DB.Exec(stmt, id)
	if err != nil {
		return err
	}
	rowsAff, err := row.RowsAffected()
	if err != nil {
		return err
	} 

	if rowsAff == 0 {
		return ErrNoRecord
	}
	return nil
}

func (m *JPKModel) GetContent(id int) ([]byte, error) {
	stmt := "SELECT xml_content FROM JpkFiles WHERE id = @p1"
	var content []byte
	err := m.DB.QueryRow(stmt, id).Scan(&content)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNoRecord
		}
		return nil, err
	}
	if len(content) == 0 {
		return nil, errors.New("No file content.")
	}

	return content, nil
}

package schema

type Airport struct {
	AirportCode string `db:"airport_code"`
	AirportName string `db:"airport_name"`
	City        string `db:"city"`
	Country     string `db:"country"`
}

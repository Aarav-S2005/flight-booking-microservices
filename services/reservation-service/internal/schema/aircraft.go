package schema

type Aircraft struct {
	AircraftType string `db:"aircraft_type" json:"aircraft_type"`
	TotalSeats int `db:"total_seats" json:"total_seats"`
	ColumnsAvailable []string `db:"columns_available" json:"columns_available"`
	TotalRows int `db:"total_rows" json:"total_rows"`
}

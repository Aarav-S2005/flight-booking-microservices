package booking

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrInsufficientSeatsLeft = errors.New("insufficient number of seats left")
	ErrBookingNotFound       = errors.New("booking not found")
)

type Repository struct {
	db *pgxpool.Pool
}

type FlightRecord struct {
	FlightID     uuid.UUID
	TotalSeats   int
	AircraftType string
}

type SeatUpdate struct {
	FlightID  uuid.UUID
	SeatsLeft int
	Version   uint64
}

func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

func (repo *Repository) bookTicketAndSave(ctx context.Context, reqBody BookTicketDTO, flightIDs []uuid.UUID, bookingUserID uuid.UUID) (uuid.UUID, []SeatUpdate, error) {
	tx, err := repo.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return uuid.Nil, nil, err
	}
	defer tx.Rollback(ctx)

	// acquiring lock on flightIDs that user wants to book
	updatedSeats, err := lockAndDecreaseSeats(ctx, tx, flightIDs, len(reqBody.PassengerDetails))
	if err != nil {
		return uuid.Nil, nil, err
	}

	// creating a booking
	bookingID, err := createBooking(ctx, tx, bookingUserID, reqBody.Email, reqBody.Phone, reqBody.TotalFare)
	if err != nil {
		return uuid.Nil, nil, err
	}

	// inserting passengers
	err = insertAllPassengers(ctx, tx, reqBody.PassengerDetails, bookingID)
	if err != nil {
		return uuid.Nil, nil, err
	}

	err = insertFlightSegments(ctx, tx, flightIDs, bookingID)
	if err != nil {
		return uuid.Nil, nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return uuid.Nil, nil, err
	}
	return bookingID, updatedSeats, nil
}

func (repo *Repository) findAllMissingIDs(ctx context.Context, flightIDs []uuid.UUID) ([]uuid.UUID, error) {
	rows, err := repo.db.Query(ctx, `
		SELECT id
		FROM unnest($1::uuid[]) AS id
		WHERE NOT EXISTS (
			SELECT 1 FROM flights f WHERE f.flight_id = id
		)
	`, flightIDs)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	missing := make([]uuid.UUID, 0)
	for rows.Next() {
		var id uuid.UUID
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		missing = append(missing, id)
	}
	return missing, rows.Err()
}

func (repo *Repository) existsByFlightID(ctx context.Context, flightID uuid.UUID) (bool, error) {
	var exists bool
	err := repo.db.QueryRow(
		ctx,
		"SELECT EXISTS (SELECT 1 FROM flights WHERE id = $1)",
		flightID,
	).Scan(&exists)
	if err != nil {
		return false, err
	}
	return exists, nil
}

func (repo *Repository) addSingleFlight(ctx context.Context, flight FlightRecord) error {
	_, err := repo.db.Exec(ctx, `
		insert 
		into flights(flight_id, aircraft_type, seats_left, total_seats, version) 
		values ($1, $2, $3, $4, 0) 
		on conflict (flight_id) do nothing
		`, flight.FlightID, flight.AircraftType, flight.TotalSeats, flight.TotalSeats)
	return err
}

func (repo *Repository) getAllFlightByBookingUserIDGroupedByBookingID(ctx context.Context, bookingUserID uuid.UUID) (map[uuid.UUID][]uuid.UUID, error) {
	rows, err := repo.db.Query(ctx,
		`select b.booking_id, fs.flight_id 
			from flight_segments fs
			join bookings b
			on fs.booking_id = b.booking_id
			where booking_user_id = $1 
			order by b.booking_id, fs.segment_order`,
		bookingUserID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	flightIDs := make(map[uuid.UUID][]uuid.UUID)
	count := 0
	for rows.Next() {
		var flightID uuid.UUID
		var bookingID uuid.UUID
		if err := rows.Scan(&bookingID, &flightID); err != nil {
			return nil, err
		}
		flightIDs[bookingID] = append(flightIDs[bookingID], flightID)
		count++
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if count == 0 {
		return nil, ErrBookingNotFound
	}

	return flightIDs, nil
}

func (repo *Repository) getAllPassengerByBookingUserIDGroupedByBookingID(ctx context.Context, bookingUserID uuid.UUID) (map[uuid.UUID][]PassengerDetails, error) {
	rows, err := repo.db.Query(ctx, `
		select b.booking_id, first_name, last_name, age, gender, passport_number
		from passengers p
		join bookings b
		on p.booking_id = b.booking_id
		where b.booking_user_id = $1
		`,
		bookingUserID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	passengerDetails := make(map[uuid.UUID][]PassengerDetails)

	for rows.Next() {
		var bookingID uuid.UUID
		var firstName, lastName, gender, passportNumber string
		var age int
		if err := rows.Scan(&bookingID, &firstName, &lastName, &age, &gender, &passportNumber); err != nil {
			return nil, err
		}
		passengerDetails[bookingID] = append(passengerDetails[bookingID], PassengerDetails{
			FirstName:      firstName,
			LastName:       lastName,
			Gender:         gender,
			PassportNumber: passportNumber,
			Age:            age,
		})
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return passengerDetails, nil
}

func (repo *Repository) getAllTotalFaresByBookingUserIDGroupedByBookingID(ctx context.Context, bookingUserID uuid.UUID) (map[uuid.UUID]int, error) {
	rows, err := repo.db.Query(ctx,
		`select booking_id, total_fare 
		from bookings 
		where booking_user_id = $1
		`,
		bookingUserID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	totalFares := make(map[uuid.UUID]int)
	for rows.Next() {
		var bookingID uuid.UUID
		var totalFare int
		if err := rows.Scan(&bookingID, &totalFare); err != nil {
			return nil, err
		}
		totalFares[bookingID] = totalFare
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return totalFares, nil
}

// Booking Transaction Breakdown

func lockAndDecreaseSeats(ctx context.Context, tx pgx.Tx, flightIDs []uuid.UUID, passengerCount int) ([]SeatUpdate, error) {
	rows, err := tx.Query(ctx, `
        WITH locked AS (
            SELECT flight_id FROM flights
            WHERE flight_id = ANY($1)
            ORDER BY flight_id
            FOR UPDATE
        )
        UPDATE flights f
        SET seats_left = f.seats_left - $2, version = version + 1
        FROM locked l
        WHERE f.flight_id = l.flight_id AND f.seats_left >= $2
        RETURNING f.flight_id, f.seats_left, f.version
    `, flightIDs, passengerCount)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	updated := make([]SeatUpdate, 0, len(flightIDs))
	for rows.Next() {
		var u SeatUpdate
		if err := rows.Scan(&u.FlightID, &u.SeatsLeft, &u.Version); err != nil {
			return nil, err
		}
		updated = append(updated, u)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if len(updated) != len(flightIDs) {
		return nil, ErrInsufficientSeatsLeft
	}
	return updated, nil
}

func createBooking(ctx context.Context, tx pgx.Tx, bookingUserID uuid.UUID, email, phone string, totalFare int) (uuid.UUID, error) {
	var bookingID uuid.UUID
	err := tx.QueryRow(ctx, "insert into bookings(booking_user_id, email, phone, total_fare) values ($1, $2, $3, $4) returning booking_id", bookingUserID, email, phone, totalFare).Scan(&bookingID)
	if err != nil {
		return uuid.Nil, err
	}
	return bookingID, nil
}

func insertAllPassengers(ctx context.Context, tx pgx.Tx, passengerDetails []PassengerDetails, bookingID uuid.UUID) error {
	passengersRows := make([][]any, 0, len(passengerDetails))
	for _, p := range passengerDetails {
		passengersRows = append(passengersRows, []any{
			bookingID,
			p.FirstName,
			p.LastName,
			p.Age,
			p.Gender,
			p.PassportNumber,
		})
	}
	_, err := tx.CopyFrom(
		ctx,
		pgx.Identifier{"passenger"},
		[]string{
			"booking_id",
			"first_name",
			"last_name",
			"age",
			"gender",
			"passport_number",
		},
		pgx.CopyFromRows(passengersRows),
	)
	if err != nil {
		return err
	}
	return nil
}

func insertFlightSegments(ctx context.Context, tx pgx.Tx, flightIDs []uuid.UUID, bookingID uuid.UUID) error {
	rows := make([][]any, 0, len(flightIDs))
	for i, id := range flightIDs {
		rows = append(rows, []any{bookingID, id, i + 1})
	}
	_, err := tx.CopyFrom(ctx,
		pgx.Identifier{"flight_segments"},
		[]string{"booking_id", "flight_id", "segment_order"},
		pgx.CopyFromRows(rows),
	)
	return err
}

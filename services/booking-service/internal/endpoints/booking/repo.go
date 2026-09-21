package booking

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrInsufficientSeatsLeft = errors.New("insufficient number of seats left")
	ErrBookingNotFound       = errors.New("booking not found")
	ErrStatusNotConfirmed    = errors.New("status not confirmed")
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

func (repo *Repository) bookTicketAndSave(ctx context.Context, reqBody BookTicketDTO, flightIDs []uuid.UUID, bookingUserID uuid.UUID) (uuid.UUID, []SeatUpdate, []uuid.UUID, error) {
	tx, err := repo.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return uuid.Nil, nil, nil, err
	}
	defer tx.Rollback(ctx)

	// acquiring lock on flightIDs that user wants to book
	updatedSeats, err := lockAndDecreaseSeats(ctx, tx, flightIDs, len(reqBody.PassengerDetails))
	if err != nil {
		return uuid.Nil, nil, nil, err
	}

	// creating a booking
	bookingID, err := createBooking(ctx, tx, bookingUserID, reqBody.Email, reqBody.Phone, reqBody.TotalFare)
	if err != nil {
		return uuid.Nil, nil, nil, err
	}

	// inserting passengers
	passengerIDs, err := insertAllPassengers(ctx, tx, reqBody.PassengerDetails, bookingID)
	if err != nil {
		return uuid.Nil, nil, nil, err
	}

	err = insertFlightSegments(ctx, tx, flightIDs, bookingID)
	if err != nil {
		return uuid.Nil, nil, nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return uuid.Nil, nil, nil, err
	}
	return bookingID, updatedSeats, passengerIDs, nil
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
		"SELECT EXISTS (SELECT 1 FROM flights WHERE flight_id = $1)",
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

func (repo *Repository) getTotalFareByBookingIDAndUserID(ctx context.Context, bookingID uuid.UUID, userID uuid.UUID) (int, error) {
	var totalFare int
	err := repo.db.QueryRow(ctx, "select total_fare from bookings where booking_id = $1 and booking_user_id = $2", bookingID, userID).Scan(&totalFare)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return 0, ErrBookingNotFound
		}
		return 0, err
	}
	return totalFare, nil

}

func (repo *Repository) getBookingCreationTime(ctx context.Context, userID, bookingID uuid.UUID) (time.Time, error) {
	var creationTime time.Time
	err := repo.db.QueryRow(ctx, "select created_at from bookings where booking_id = $1 and booking_user_id = $2", bookingID, userID).Scan(&creationTime)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return time.Time{}, ErrBookingNotFound
		}
		return time.Time{}, err
	}
	return creationTime, nil
}

func (repo *Repository) updateStatusByBookingID(ctx context.Context, bookingID uuid.UUID, status string) error {
	_, err := repo.db.Exec(ctx, "update bookings set status = $1 where booking_id = $1", status, bookingID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrBookingNotFound
		}
		return err
	}
	return nil
}

func (repo *Repository) getPassengersIDbyBookingID(ctx context.Context, bookingID uuid.UUID) ([]uuid.UUID, error) {
	var passengerIDs []uuid.UUID
	rows, err := repo.db.Query(ctx, "select passenger_id from passengers where booking_id = $1", bookingID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrBookingNotFound
		}
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var passengerID uuid.UUID
		if err := rows.Scan(&passengerID); err != nil {
			return nil, err
		}
		passengerIDs = append(passengerIDs, passengerID)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return passengerIDs, nil
}

func (repo *Repository) getFlightIDsBytBookingID(ctx context.Context, bookingID uuid.UUID) ([]uuid.UUID, error) {
	var flightIDs []uuid.UUID
	rows, err := repo.db.Query(ctx, "select flight_id from flight_segments where booking_id = $1 order by segment_order", bookingID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrBookingNotFound
		}
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var flightID uuid.UUID
		if err := rows.Scan(&flightID); err != nil {
			return nil, err
		}
		flightIDs = append(flightIDs, flightID)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return flightIDs, nil
}

func (repo *Repository) checkStatusByBookingID(ctx context.Context, bookingID, userID uuid.UUID) (string, error) {
	var status string
	err := repo.db.QueryRow(ctx, "select status from bookings where booking_id = $1 and booking_user_id = $2", bookingID, userID).Scan(&status)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", ErrBookingNotFound
		}
		return "", err
	}
	return status, nil
}

func (repo *Repository) increaseFlightSeatsByBookingID(ctx context.Context, bookingID uuid.UUID) ([]SeatUpdate, error) {
	tx, err := repo.db.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)
	const query = `
		WITH passenger_count AS (
			SELECT COUNT(*)::int AS count
			FROM passengers
			WHERE booking_id = $1
		),
		updated_flights AS (
			UPDATE flights f
			SET
				seats_left = f.seats_left + pc.count,
				version = f.version + 1
			FROM flight_segments fs
			CROSS JOIN passenger_count pc
			WHERE fs.booking_id = $1
			  AND f.flight_id = fs.flight_id
			RETURNING
				f.flight_id,
				f.seats_left,
				f.version
		)
		SELECT
			flight_id,
			seats_left,
			version
		FROM updated_flights
		ORDER BY flight_id
	`
	rows, err := tx.Query(ctx, query, bookingID)
	if err != nil {
		return nil, fmt.Errorf("increase flight seats for booking %s: %w", bookingID, err)
	}
	defer rows.Close()

	updates := make([]SeatUpdate, 0)

	for rows.Next() {
		var update SeatUpdate

		if err := rows.Scan(
			&update.FlightID,
			&update.SeatsLeft,
			&update.Version,
		); err != nil {
			return nil, fmt.Errorf("scan seat update: %w", err)
		}
		updates = append(updates, update)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate seat updates: %w", err)
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit seat updates: %w", err)
	}

	return updates, nil
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

func insertAllPassengers(ctx context.Context, tx pgx.Tx, passengerDetails []PassengerDetails, bookingID uuid.UUID) ([]uuid.UUID, error) {
	if len(passengerDetails) == 0 {
		return []uuid.UUID{}, nil
	}

	values := make([]string, 0, len(passengerDetails))
	args := make([]any, 0, len(passengerDetails)*6)

	for i, p := range passengerDetails {
		offset := i * 6

		values = append(values, fmt.Sprintf(
			"($%d, $%d, $%d, $%d, $%d, $%d)",
			offset+1,
			offset+2,
			offset+3,
			offset+4,
			offset+5,
			offset+6,
		))

		args = append(args,
			bookingID,
			p.FirstName,
			p.LastName,
			p.Age,
			p.Gender,
			p.PassportNumber,
		)
	}

	query := fmt.Sprintf(`
        INSERT INTO passengers (
            booking_id,
            first_name,
            last_name,
            age,
            gender,
            passport_number
        )
        VALUES %s
        RETURNING passenger_id
    `, strings.Join(values, ", "))

	rows, err := tx.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	passengerIDs := make([]uuid.UUID, 0, len(passengerDetails))

	for rows.Next() {
		var id uuid.UUID

		if err := rows.Scan(&id); err != nil {
			return nil, err
		}

		passengerIDs = append(passengerIDs, id)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return passengerIDs, nil
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

package database

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/Aarav-S2005/flight-booking-microservices/shared/utility"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrExternal  = errors.New("external error")
	ErrNotFound  = errors.New("not found")
	ErrSeatTaken = errors.New("seat already taken")
	ErrRetryable = errors.New("transient conflict, retry")
)

const (
	pgUniqueViolation   = "23505"
	pgSerializationFail = "40001"
	pgDeadlockDetected  = "40P01"
)

type Repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

func (r *Repository) ExistsByBookingID(ctx context.Context, bookingID uuid.UUID) (bool, error) {
	var found bool
	err := r.db.QueryRow(ctx, "select exists (select 1 from reservations where booking_id = $1)", bookingID).Scan(&found)
	if err != nil {
		return false, err
	}
	return found, nil
}

func (r *Repository) InsertReservationWithoutSeatReservation(ctx context.Context, bookingID, userID uuid.UUID, passengers, flightIDs []string) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	passengerUUIDs, err := utility.ToUUIDs(passengers)
	if err != nil {
		return ErrExternal
	}

	flightUUIDs, err := utility.ToUUIDs(flightIDs)
	if err != nil {
		return ErrExternal
	}
	var reservationID uuid.UUID
	err = tx.QueryRow(ctx, "insert into reservations (user_id, booking_id, passenger_count) values ($1, $2, $3) on conflict (booking_id) do nothing returning reservation_id", userID, bookingID, len(passengers)).Scan(&reservationID)
	if err != nil {
		return err
	}
	_, err = tx.Exec(ctx, `
        INSERT INTO reservation_flights (
            reservation_id,
            flight_id,
            segment_number
        )
        SELECT
            $1,
            flight_id,
            row_number() OVER ()
        FROM unnest($2::uuid[]) AS flight_id
        ON CONFLICT (reservation_id, flight_id) DO NOTHING
    `, reservationID, flightUUIDs)
	if err != nil {
		return err
	}
	_, err = tx.Exec(ctx, `
        INSERT INTO seat_allocation (
            reservation_id,
            flight_id,
            column_allocated,
            seat_number,
            passenger_id
        )
        SELECT
            $1,
            f.flight_id,
            NULL,
            NULL,
            p.passenger_id
        FROM unnest($2::uuid[]) AS f(flight_id)
        CROSS JOIN unnest($3::uuid[]) AS p(passenger_id)
        ON CONFLICT (
            reservation_id,
            flight_id,
            passenger_id
        ) DO NOTHING
    `, reservationID, flightUUIDs, passengerUUIDs)
	if err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func (r *Repository) GetSeatsLeftByAircraftType(ctx context.Context, aircraftType string) (int, error) {
	var totalSeats int
	err := r.db.QueryRow(ctx, "select total_seats from aircraft where aircraft_type = $1", aircraftType).Scan(&totalSeats)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return 0, ErrNotFound
		}
		return 0, err
	}
	return totalSeats, nil
}

// TRANSACTION FOR SEAT RESERVATION

type ReservationTxModel struct {
	ID             uuid.UUID
	UserID         uuid.UUID
	PassengerCount int
}

type FlightSegmentTxModel struct {
	FlightID      uuid.UUID
	AircraftType  string
	DepartureTime time.Time
	Columns       []string
	TotalRows     int
}

type SeatAllocationTxModel struct {
	PassengerID uuid.UUID
	Column      *string
	SeatNumber  *int
}

type SeatTx interface {
	LockReservationByBooking(ctx context.Context, bookingID uuid.UUID) (*ReservationTxModel, error)

	ListFlightSegments(ctx context.Context, reservationID uuid.UUID) ([]FlightSegmentTxModel, error)

	LockAllocations(ctx context.Context, reservationID, flightID uuid.UUID) ([]SeatAllocationTxModel, error)

	AssignSeat(ctx context.Context, reservationID, flightID, passengerID uuid.UUID, column string, number int) error
}

type seatTx struct {
	tx pgx.Tx
}

func (r *Repository) WithTx(ctx context.Context, fn func(tx SeatTx) error) error {
	tx, err := r.db.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.ReadCommitted})
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback(context.WithoutCancel(ctx)) }()

	if err := fn(&seatTx{tx: tx}); err != nil {
		return err
	}
	if err := tx.Commit(ctx); err != nil {
		return translateErr(err)
	}
	return nil
}

func (t *seatTx) LockReservationByBooking(ctx context.Context, bookingID uuid.UUID) (*ReservationTxModel, error) {
	var res ReservationTxModel
	if err := t.tx.QueryRow(ctx, `
		SELECT reservation_id, user_id, passenger_count
		FROM reservations
		WHERE booking_id = $1
		FOR UPDATE`,
		bookingID).Scan(&res.ID, &res.UserID, &res.PassengerCount); err != nil {
		return nil, translateErr(err)
	}
	return &res, nil
}

func (t *seatTx) ListFlightSegments(ctx context.Context, reservationID uuid.UUID) ([]FlightSegmentTxModel, error) {
	rows, err := t.tx.Query(ctx, `
		SELECT f.flight_id,
		       f.aircraft_type,
		       f.departure_time,
		       a.columns_available,
		       a.total_rows
		FROM reservation_flights rf
		JOIN flights  f ON f.flight_id     = rf.flight_id
		JOIN aircraft a ON a.aircraft_type = f.aircraft_type
		WHERE rf.reservation_id = $1
		ORDER BY rf.segment_number`,
		reservationID)
	if err != nil {
		return nil, translateErr(err)
	}
	defer rows.Close()

	var out []FlightSegmentTxModel
	for rows.Next() {
		var s FlightSegmentTxModel
		if err := rows.Scan(&s.FlightID, &s.AircraftType, &s.DepartureTime, &s.Columns, &s.TotalRows); err != nil {
			return nil, translateErr(err)
		}
		out = append(out, s)
	}
	if err := rows.Err(); err != nil {
		return nil, translateErr(err)
	}
	return out, nil
}

func (t *seatTx) LockAllocations(ctx context.Context, reservationID, flightID uuid.UUID) ([]SeatAllocationTxModel, error) {
	rows, err := t.tx.Query(ctx, `
		SELECT passenger_id, column_allocated, seat_number
		FROM seat_allocation
		WHERE reservation_id = $1 AND flight_id = $2
		ORDER BY passenger_id
		FOR UPDATE`,
		reservationID, flightID)
	if err != nil {
		return nil, translateErr(err)
	}
	defer rows.Close()

	var out []SeatAllocationTxModel
	for rows.Next() {
		var a SeatAllocationTxModel
		if err := rows.Scan(&a.PassengerID, &a.Column, &a.SeatNumber); err != nil {
			return nil, translateErr(err)
		}
		out = append(out, a)
	}
	if err := rows.Err(); err != nil {
		return nil, translateErr(err)
	}
	return out, nil
}

func (t *seatTx) AssignSeat(ctx context.Context, reservationID, flightID, passengerID uuid.UUID, column string, number int) error {
	tag, err := t.tx.Exec(ctx, `
		UPDATE seat_allocation
		SET column_allocated = $1,
		    seat_number      = $2
		WHERE reservation_id = $3
		  AND flight_id      = $4
		  AND passenger_id   = $5
		  AND column_allocated IS NULL
		  AND seat_number      IS NULL`,
		column, number, reservationID, flightID, passengerID)
	if err != nil {
		return translateErr(err)
	}
	if tag.RowsAffected() != 1 {
		return fmt.Errorf("assign seat: expected 1 row updated, got %d", tag.RowsAffected())
	}
	return nil
}

func translateErr(err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, pgx.ErrNoRows) {
		return ErrNotFound
	}

	if pgErr, ok := errors.AsType[*pgconn.PgError](err); ok {
		switch pgErr.Code {
		case pgUniqueViolation:
			return fmt.Errorf("%w: %s", ErrSeatTaken, pgErr.Detail)

		case pgSerializationFail, pgDeadlockDetected:
			return fmt.Errorf("%w: %s", ErrRetryable, pgErr.Message)
		}
	}
	return err
}

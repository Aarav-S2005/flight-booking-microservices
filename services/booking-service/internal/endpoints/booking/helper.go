package booking

import (
	"errors"
	"fmt"
	"net/mail"
	"strings"

	"github.com/google/uuid"
)

func ValidateBookTicketRequestDTO(reqBody BookTicketDTO) error {
	if strings.TrimSpace(reqBody.Email) == "" {
		return errors.New("email is required")
	}
	if _, err := mail.ParseAddress(reqBody.Email); err != nil {
		return errors.New("invalid email")
	}
	if strings.TrimSpace(reqBody.Phone) == "" {
		return errors.New("phone is required")
	}
	if len(reqBody.Phone) != 10 {
		return errors.New("phone length must be 10")
	}
	if reqBody.TotalFare <= 0 {
		return errors.New("total_fare must be greater than 0")
	}
	if len(reqBody.PassengerDetails) == 0 {
		return errors.New("at least one passenger is required")
	}
	if len(reqBody.FlightSegments) == 0 {
		return errors.New("at least one flight segment is required")
	}
	for i, passenger := range reqBody.PassengerDetails {
		if strings.TrimSpace(passenger.FirstName) == "" {
			return fmt.Errorf("passenger_details[%d].first_name is required", i)
		}
		if strings.TrimSpace(passenger.LastName) == "" {
			return fmt.Errorf("passenger_details[%d].last_name is required", i)
		}
		if passenger.Age < 0 || passenger.Age > 120 {
			return fmt.Errorf("passenger_details[%d].age must be between 0 and 120", i)
		}
		passenger.Gender = strings.ToUpper(strings.TrimSpace(passenger.Gender))
		switch passenger.Gender {
		case "Male", "Female", "Other":
		default:
			return fmt.Errorf("passenger_details[%d].gender must be one of M, F, or O", i)
		}
		if strings.TrimSpace(passenger.PassportNumber) == "" {
			return fmt.Errorf("passenger_details[%d].passport_number is required", i)
		}
	}
	set := map[int]bool{}
	for i, segment := range reqBody.FlightSegments {
		if strings.TrimSpace(segment.FlightID) == "" {
			return fmt.Errorf("flight_segments[%d].flight_id is required", i)
		}
		if segment.SegmentOrder <= 0 {
			return fmt.Errorf("flight_segments[%d].segment_order must be greater than 0", i)
		}
		if _, ok := set[segment.SegmentOrder]; ok {
			return fmt.Errorf("duplicate segment_order: %d", segment.SegmentOrder)
		}
		set[segment.SegmentOrder] = true
	}
	return nil
}

func computeUniqueFlightIDs(grouped map[uuid.UUID][]uuid.UUID) []uuid.UUID {
	seen := make(map[uuid.UUID]struct{})

	var flightIDs []uuid.UUID

	for _, ids := range grouped {
		for _, id := range ids {
			if _, exists := seen[id]; exists {
				continue
			}

			seen[id] = struct{}{}
			flightIDs = append(flightIDs, id)
		}
	}

	return flightIDs
}

func UUIDsToStrings(ids []uuid.UUID) []string {
	if len(ids) == 0 {
		return []string{}
	}

	result := make([]string, len(ids))
	for i, id := range ids {
		result[i] = id.String()
	}

	return result
}

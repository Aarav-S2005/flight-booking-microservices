package booking

import "github.com/google/uuid"

func ValidateBookTicketRequestDTO(reqBody BookTicketDTO) error {
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

package services

import (
	"bookcabin-test/internal/core/domain"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
)

func criteriaHash(c domain.SearchCriteria) string {
	key := fmt.Sprintf("%s_%s_%s_%s", c.Origin, c.Destination, c.DepartureDate, c.CabinClass)
	hasher := sha256.New()
	hasher.Write([]byte(key))
	return hex.EncodeToString(hasher.Sum(nil))
}

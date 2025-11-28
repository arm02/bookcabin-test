package providers

import (
	"bookcabin-test/internal/core/domain"
	"time"

	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

var p = message.NewPrinter(language.Make("id"))

var (
	LocationWIB, _  = time.LoadLocation("Asia/Jakarta")
	LocationWITA, _ = time.LoadLocation("Asia/Makassar")
	LocationWIT, _  = time.LoadLocation("Asia/Jayapura")
)

type ProviderInterface interface {
	Search(criteria domain.SearchCriteria) ([]domain.UnifiedFlight, error)
	Name() string
}

func CalculateDuration(start, end time.Time) int {
	return int(end.Sub(start).Minutes())
}

func FormatIDR(v float64) string {
	amount := int64(v)
	return p.Sprintf("Rp%d", amount)
}

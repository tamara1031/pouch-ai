package middleware

import (
	"fmt"
	"pouch-ai/internal/domain"
	"pouch-ai/internal/service"
)

// DefaultReservationCost is the cost we reserve if we can't estimate it.
const DefaultReservationCost = 0.01

func NewUsageTrackingMiddleware(keyService *service.KeyService) domain.Middleware {
	return domain.MiddlewareFunc(func(req *domain.Request, next domain.Handler) (*domain.Response, error) {
		var reservedAmount float64

		// 1. Reserve Budget
		if req.Key != nil && !req.Key.IsMock {
			// Estimate Cost
			if req.Provider != nil {
				usage, err := req.Provider.EstimateUsage(req.Model, req.RawBody)
				if err == nil && usage != nil && usage.TotalCost > 0 {
					reservedAmount = usage.TotalCost
				}
			}

			// Fallback if estimation failed or returned 0
			if reservedAmount <= 0 {
				reservedAmount = DefaultReservationCost
			}

			if err := keyService.ReserveUsage(req.Context, req.Key.ID, reservedAmount); err != nil {
				if err == domain.ErrBudgetExceeded {
					return nil, fmt.Errorf("budget limit exceeded")
				}
				return nil, fmt.Errorf("failed to reserve budget: %w", err)
			}
		}

		// 2. Execute Request
		resp, err := next.Handle(req)

		// 3. Complete Usage (Commit or Rollback)
		if req.Key != nil && !req.Key.IsMock {
			var realCost float64
			if err == nil && resp != nil {
				realCost = resp.TotalCost
			}

			// CompleteUsage handles adjustment (realCost - reservedAmount).
			// If request failed (realCost=0), it refunds the reservation.
			_ = keyService.CompleteUsage(req.Context, req.Key.ID, realCost, reservedAmount)
		}

		return resp, err
	})
}

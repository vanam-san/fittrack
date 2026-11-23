package desktop

import (
	"context"

	"github.com/vanam-san/fittrack/internal/analytics"
)

type AnalyticsService struct{ analytics *analytics.Service }

func NewAnalyticsService(service *analytics.Service) *AnalyticsService {
	return &AnalyticsService{analytics: service}
}

func (s *AnalyticsService) Report(ctx context.Context, from, to, grouping string) (analytics.Report, error) {
	result, err := s.analytics.Report(ctx, from, to, grouping)
	return result, analytics.UserError(err)
}

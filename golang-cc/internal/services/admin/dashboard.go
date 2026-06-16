package admin

import (
	"context"

	"github.com/rafa/golang-cc/internal/domain"
)

func (s *Service) Dashboard(ctx context.Context) (domain.Dashboard, error) {
	return s.repository.Dashboard(ctx)
}

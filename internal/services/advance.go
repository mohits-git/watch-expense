package services

import (
	"context"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/mohits-git/watch-expense/internal/domain"
	"github.com/mohits-git/watch-expense/internal/ports"
	"github.com/mohits-git/watch-expense/internal/utils/apperr"
	"github.com/mohits-git/watch-expense/internal/utils/authctx"
	"github.com/mohits-git/watch-expense/internal/utils/validator"
)

type AdvanceService interface {
	CreateAdvance(ctx context.Context, advance domain.Advance) (string, error)
	GetAdvanceByID(ctx context.Context, advanceID string) (domain.Advance, error)
	UpdateAdvance(ctx context.Context, advance domain.Advance) error
	UpdateAdvanceStatus(ctx context.Context, advanceID string, status domain.RequestStatus) error
	GetAllAdvances(ctx context.Context, filterOptions domain.AdvancesFilterOptions) ([]domain.Advance, int, error)
	GetAdvanceSummary(ctx context.Context) (domain.AdvanceSummary, error)
}

type advanceService struct {
	advanceRepo ports.AdvanceRepository
}

func NewAdvanceService(advanceRepo ports.AdvanceRepository) AdvanceService {
	return &advanceService{
		advanceRepo: advanceRepo,
	}
}

func (s *advanceService) CreateAdvance(ctx context.Context, advance domain.Advance) (string, error) {
	claims, ok := authctx.UserClaimsFromCtx(ctx)
	if !ok {
		return "", apperr.NewAppError(apperr.ErrUnauthorized, "unauthorized", nil)
	}

	if claims.Role != domain.Employee {
		return "", apperr.NewAppError(apperr.ErrForbidden, "only employees can create advances", nil)
	}

	advance.UserID = claims.UserID

	advance.Status = domain.Pending

	if !validator.ValidateAdvanceCreation(advance) {
		return "", apperr.NewAppError(apperr.ErrInvalid, "invalid advance data", nil)
	}

	advance.ID = uuid.New().String()
	advance.CreatedAt = time.Now().Unix()
	advance.UpdatedAt = time.Now().Unix()

	return s.advanceRepo.SaveAdvance(ctx, advance)
}

func (s *advanceService) UpdateAdvance(ctx context.Context, advance domain.Advance) error {
	claims, ok := authctx.UserClaimsFromCtx(ctx)
	if !ok {
		return apperr.NewAppError(apperr.ErrUnauthorized, "unauthorized", nil)
	}

	if claims.Role != domain.Employee {
		return apperr.NewAppError(apperr.ErrForbidden, "only employees can update advances", nil)
	}

	if !validator.ValidateAdvanceUpdate(advance) {
		return apperr.NewAppError(apperr.ErrInvalid, "invalid advance data", nil)
	}

	existingAdvance, err := s.advanceRepo.FindAdvanceById(ctx, advance.ID)
	if err != nil {
		return err
	}

	if existingAdvance.UserID != claims.UserID {
		return apperr.NewAppError(apperr.ErrForbidden, "you can only update your own advances", nil)
	}

	if existingAdvance.Status != domain.Pending {
		return apperr.NewAppError(apperr.ErrForbidden, "cannot update advance that is not pending", nil)
	}

	advance.Status = domain.Pending
	advance.UserID = claims.UserID

	advance.UpdatedAt = time.Now().Unix()
	advance.CreatedAt = existingAdvance.CreatedAt
	advance.ApprovedBy = existingAdvance.ApprovedBy
	advance.ApprovedAt = existingAdvance.ApprovedAt
	advance.ReviewedBy = existingAdvance.ReviewedBy
	advance.ReviewedAt = existingAdvance.ReviewedAt

	return s.advanceRepo.UpdateAdvance(ctx, advance)
}

func (s *advanceService) UpdateAdvanceStatus(ctx context.Context, advanceID string, status domain.RequestStatus) error {
	claims, ok := authctx.UserClaimsFromCtx(ctx)
	if !ok {
		return apperr.NewAppError(apperr.ErrUnauthorized, "unauthorized", nil)
	}

	if claims.Role != domain.Admin {
		return apperr.NewAppError(apperr.ErrForbidden, "only admin can update advance status", nil)
	}

	if !validator.ValidateUUID(advanceID) {
		return apperr.NewAppError(apperr.ErrInvalid, "invalid advance ID", nil)
	}

	if !validator.ValidateAdvanceStatus(status) || status == domain.Pending {
		return apperr.NewAppError(apperr.ErrInvalid, "invalid status", nil)
	}

	advance, err := s.advanceRepo.FindAdvanceById(ctx, advanceID)
	if err != nil {
		return err
	}

	reviewerID := claims.UserID

	advance.Status = status

	if status == domain.Approved {
		advance.ApprovedBy = reviewerID
		advance.ApprovedAt = time.Now().Unix()
	}

	if status == domain.Reviewed {
		advance.ReviewedBy = reviewerID
		advance.ReviewedAt = time.Now().Unix()
	}

	advance.UpdatedAt = time.Now().Unix()

	return s.advanceRepo.UpdateAdvance(ctx, advance)
}

func (s *advanceService) GetAdvanceByID(ctx context.Context, advanceID string) (domain.Advance, error) {
	if !validator.ValidateUUID(advanceID) {
		return domain.Advance{}, apperr.NewAppError(apperr.ErrInvalid, "invalid advance ID", nil)
	}

	claims, ok := authctx.UserClaimsFromCtx(ctx)
	if !ok {
		return domain.Advance{}, apperr.NewAppError(apperr.ErrUnauthorized, "unauthorized", nil)
	}

	advance, err := s.advanceRepo.FindAdvanceById(ctx, advanceID)
	if err != nil {
		return domain.Advance{}, err
	}

	if claims.Role != domain.Admin && advance.UserID != claims.UserID {
		return domain.Advance{}, apperr.NewAppError(apperr.ErrForbidden, "access denied", nil)
	}

	return advance, nil
}

func (s *advanceService) GetAllAdvances(ctx context.Context, filterOptions domain.AdvancesFilterOptions) ([]domain.Advance, int, error) {
	claims, ok := authctx.UserClaimsFromCtx(ctx)
	if !ok {
		return nil, 0, apperr.NewAppError(apperr.ErrUnauthorized, "unauthorized", nil)
	}

	if claims.Role != domain.Admin {
		filterOptions.UserID = claims.UserID
	}
	return s.advanceRepo.FindAllAdvances(ctx, filterOptions)
}

func (s *advanceService) GetAdvanceSummary(ctx context.Context) (domain.AdvanceSummary, error) {
	claims, ok := authctx.UserClaimsFromCtx(ctx)
	if !ok {
		return domain.AdvanceSummary{}, apperr.NewAppError(apperr.ErrUnauthorized, "unauthorized", nil)
	}

	userID := claims.UserID
	if claims.Role == domain.Admin {
		userID = ""
	}

	type result struct {
		value float64
		err   error
	}

	approvedChan := make(chan result, 1)
	reconciledChan := make(chan result, 1)
	pendingChan := make(chan result, 1)
	rejectedChan := make(chan result, 1)

	var wg sync.WaitGroup
	wg.Add(4)

	go func() {
		defer wg.Done()
		sum, err := s.advanceRepo.GetAdvanceSumByStatus(ctx, userID, domain.Approved)
		approvedChan <- result{value: sum, err: err}
	}()

	go func() {
		defer wg.Done()
		sum, err := s.advanceRepo.GetReconciledAdvancesSum(ctx, userID)
		reconciledChan <- result{value: sum, err: err}
	}()

	go func() {
		defer wg.Done()
		sum, err := s.advanceRepo.GetAdvanceSumByStatus(ctx, userID, domain.Pending)
		pendingChan <- result{value: sum, err: err}
	}()

	go func() {
		defer wg.Done()
		sum, err := s.advanceRepo.GetAdvanceSumByStatus(ctx, userID, domain.Rejected)
		rejectedChan <- result{value: sum, err: err}
	}()

	wg.Wait()
	close(approvedChan)
	close(reconciledChan)
	close(pendingChan)
	close(rejectedChan)

	approvedResult := <-approvedChan
	if approvedResult.err != nil {
		return domain.AdvanceSummary{}, approvedResult.err
	}

	reconciledResult := <-reconciledChan
	if reconciledResult.err != nil {
		return domain.AdvanceSummary{}, reconciledResult.err
	}

	pendingResult := <-pendingChan
	if pendingResult.err != nil {
		return domain.AdvanceSummary{}, pendingResult.err
	}

	rejectedResult := <-rejectedChan
	if rejectedResult.err != nil {
		return domain.AdvanceSummary{}, rejectedResult.err
	}

	return domain.AdvanceSummary{
		Approved:   approvedResult.value,
		Reconciled: reconciledResult.value,
		Pending:    pendingResult.value,
		Rejected:   rejectedResult.value,
	}, nil
}

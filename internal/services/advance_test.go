package services

import (
	"context"
	"testing"

	"github.com/mohits-git/watch-expense/internal/domain"
	"github.com/mohits-git/watch-expense/internal/utils/apperr"
	"github.com/mohits-git/watch-expense/internal/utils/authctx"
	mockrepository "github.com/mohits-git/watch-expense/tests/mock_repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func Test_services_NewAdvanceService(t *testing.T) {
	advanceRepo := mockrepository.NewMockAdvanceRepository()
	advanceService := NewAdvanceService(advanceRepo)
	require.NotNil(t, advanceService, "NewAdvanceService() returned nil")
}

func Test_services_AdvanceService_CreateAdvance(t *testing.T) {
	employeeID := "650e8400-e29b-41d4-a716-446655440000"
	type args struct {
		ctx     context.Context
		advance domain.Advance
	}
	tests := []struct {
		name        string
		args        args
		getMockRepo func() *mockrepository.AdvanceRepository
		wantErr     bool
		errCode     apperr.AppErrorCode
	}{
		{
			name: "create advance successfully",
			args: args{
				ctx: authctx.WithUserClaims(context.Background(), &authctx.UserClaims{
					UserID: employeeID,
					Role:   domain.Employee,
				}),
				advance: domain.Advance{
					Amount:      5000.00,
					Purpose:     "Business trip",
					Description: "Travel advance for conference",
				},
			},
			getMockRepo: func() *mockrepository.AdvanceRepository {
				advanceRepo := mockrepository.NewMockAdvanceRepository()
				advanceRepo.On("SaveAdvance", mock.Anything, mock.AnythingOfType("domain.Advance")).Return("new-advance-id", nil)
				return advanceRepo
			},
			wantErr: false,
		},
		{
			name: "unauthorized - no user claims",
			args: args{
				ctx: context.Background(),
				advance: domain.Advance{
					Amount:      5000.00,
					Purpose:     "Business trip",
					Description: "Travel advance",
				},
			},
			getMockRepo: func() *mockrepository.AdvanceRepository {
				advanceRepo := mockrepository.NewMockAdvanceRepository()
				return advanceRepo
			},
			wantErr: true,
			errCode: apperr.ErrUnauthorized,
		},
		{
			name: "forbidden - user is admin",
			args: args{
				ctx: authctx.WithUserClaims(context.Background(), &authctx.UserClaims{
					UserID: employeeID,
					Role:   domain.Admin,
				}),
				advance: domain.Advance{
					Amount:      5000.00,
					Purpose:     "Business trip",
					Description: "Travel advance",
				},
			},
			getMockRepo: func() *mockrepository.AdvanceRepository {
				advanceRepo := mockrepository.NewMockAdvanceRepository()
				return advanceRepo
			},
			wantErr: true,
			errCode: apperr.ErrForbidden,
		},
		{
			name: "invalid advance data - zero amount",
			args: args{
				ctx: authctx.WithUserClaims(context.Background(), &authctx.UserClaims{
					UserID: employeeID,
					Role:   domain.Employee,
				}),
				advance: domain.Advance{
					Amount:      0,
					Purpose:     "Business trip",
					Description: "Travel advance",
				},
			},
			getMockRepo: func() *mockrepository.AdvanceRepository {
				advanceRepo := mockrepository.NewMockAdvanceRepository()
				return advanceRepo
			},
			wantErr: true,
			errCode: apperr.ErrInvalid,
		},
		{
			name: "invalid advance data - empty purpose",
			args: args{
				ctx: authctx.WithUserClaims(context.Background(), &authctx.UserClaims{
					UserID: employeeID,
					Role:   domain.Employee,
				}),
				advance: domain.Advance{
					Amount:      5000.00,
					Purpose:     "",
					Description: "Travel advance",
				},
			},
			getMockRepo: func() *mockrepository.AdvanceRepository {
				advanceRepo := mockrepository.NewMockAdvanceRepository()
				return advanceRepo
			},
			wantErr: true,
			errCode: apperr.ErrInvalid,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			advanceRepo := tt.getMockRepo()
			advanceService := NewAdvanceService(advanceRepo)

			result, err := advanceService.CreateAdvance(tt.args.ctx, tt.args.advance)

			if tt.wantErr {
				assert.Error(t, err, "CreateAdvance() should return error")
				appErr, ok := err.(*apperr.AppError)
				assert.True(t, ok, "Error should be of type *apperr.AppError")
				assert.Equal(t, tt.errCode, appErr.Code, "Error code should match expected")
			} else {
				assert.NoError(t, err, "CreateAdvance() should not return error")
				assert.NotEmpty(t, result, "CreateAdvance() should return advance ID")
			}

			advanceRepo.AssertExpectations(t)
		})
	}
}

func Test_services_AdvanceService_UpdateAdvance(t *testing.T) {
	validAdvanceID := "550e8400-e29b-41d4-a716-446655440000"
	employeeID := "650e8400-e29b-41d4-a716-446655440000"

	type args struct {
		ctx     context.Context
		advance domain.Advance
	}
	tests := []struct {
		name        string
		args        args
		getMockRepo func() *mockrepository.AdvanceRepository
		wantErr     bool
		errCode     apperr.AppErrorCode
	}{
		{
			name: "update advance successfully",
			args: args{
				ctx: authctx.WithUserClaims(context.Background(), &authctx.UserClaims{
					UserID: employeeID,
					Role:   domain.Employee,
				}),
				advance: domain.Advance{
					ID:          validAdvanceID,
					Amount:      6000.00,
					Purpose:     "Updated purpose",
					Description: "Updated description",
				},
			},
			getMockRepo: func() *mockrepository.AdvanceRepository {
				advanceRepo := mockrepository.NewMockAdvanceRepository()
				advanceRepo.On("FindAdvanceById", mock.Anything, validAdvanceID).Return(domain.Advance{
					ID:        validAdvanceID,
					UserID:    employeeID,
					Status:    domain.Pending,
					CreatedAt: 1234567890,
				}, nil)
				advanceRepo.On("UpdateAdvance", mock.Anything, mock.AnythingOfType("domain.Advance")).Return(nil)
				return advanceRepo
			},
			wantErr: false,
		},
		{
			name: "unauthorized - no user claims",
			args: args{
				ctx: context.Background(),
				advance: domain.Advance{
					ID:          validAdvanceID,
					Amount:      6000.00,
					Purpose:     "Updated purpose",
					Description: "Updated description",
				},
			},
			getMockRepo: func() *mockrepository.AdvanceRepository {
				advanceRepo := mockrepository.NewMockAdvanceRepository()
				return advanceRepo
			},
			wantErr: true,
			errCode: apperr.ErrUnauthorized,
		},
		{
			name: "forbidden - user is admin",
			args: args{
				ctx: authctx.WithUserClaims(context.Background(), &authctx.UserClaims{
					UserID: employeeID,
					Role:   domain.Admin,
				}),
				advance: domain.Advance{
					ID:          validAdvanceID,
					Amount:      6000.00,
					Purpose:     "Updated purpose",
					Description: "Updated description",
				},
			},
			getMockRepo: func() *mockrepository.AdvanceRepository {
				advanceRepo := mockrepository.NewMockAdvanceRepository()
				return advanceRepo
			},
			wantErr: true,
			errCode: apperr.ErrForbidden,
		},
		{
			name: "invalid advance ID",
			args: args{
				ctx: authctx.WithUserClaims(context.Background(), &authctx.UserClaims{
					UserID: employeeID,
					Role:   domain.Employee,
				}),
				advance: domain.Advance{
					ID:          "invalid-id",
					Amount:      6000.00,
					Purpose:     "Updated purpose",
					Description: "Updated description",
				},
			},
			getMockRepo: func() *mockrepository.AdvanceRepository {
				advanceRepo := mockrepository.NewMockAdvanceRepository()
				return advanceRepo
			},
			wantErr: true,
			errCode: apperr.ErrInvalid,
		},
		{
			name: "forbidden - updating another user's advance",
			args: args{
				ctx: authctx.WithUserClaims(context.Background(), &authctx.UserClaims{
					UserID: employeeID,
					Role:   domain.Employee,
				}),
				advance: domain.Advance{
					ID:          validAdvanceID,
					Amount:      6000.00,
					Purpose:     "Updated purpose",
					Description: "Updated description",
				},
			},
			getMockRepo: func() *mockrepository.AdvanceRepository {
				advanceRepo := mockrepository.NewMockAdvanceRepository()
				advanceRepo.On("FindAdvanceById", mock.Anything, validAdvanceID).Return(domain.Advance{
					ID:     validAdvanceID,
					UserID: "other-employee-id",
					Status: domain.Pending,
				}, nil)
				return advanceRepo
			},
			wantErr: true,
			errCode: apperr.ErrForbidden,
		},
		{
			name: "forbidden - advance is not pending",
			args: args{
				ctx: authctx.WithUserClaims(context.Background(), &authctx.UserClaims{
					UserID: employeeID,
					Role:   domain.Employee,
				}),
				advance: domain.Advance{
					ID:          validAdvanceID,
					Amount:      6000.00,
					Purpose:     "Updated purpose",
					Description: "Updated description",
				},
			},
			getMockRepo: func() *mockrepository.AdvanceRepository {
				advanceRepo := mockrepository.NewMockAdvanceRepository()
				advanceRepo.On("FindAdvanceById", mock.Anything, validAdvanceID).Return(domain.Advance{
					ID:     validAdvanceID,
					UserID: employeeID,
					Status: domain.Approved,
				}, nil)
				return advanceRepo
			},
			wantErr: true,
			errCode: apperr.ErrForbidden,
		},
		{
			name: "invalid advance data - zero amount",
			args: args{
				ctx: authctx.WithUserClaims(context.Background(), &authctx.UserClaims{
					UserID: employeeID,
					Role:   domain.Employee,
				}),
				advance: domain.Advance{
					ID:          validAdvanceID,
					Amount:      0,
					Purpose:     "Updated purpose",
					Description: "Updated description",
				},
			},
			getMockRepo: func() *mockrepository.AdvanceRepository {
				advanceRepo := mockrepository.NewMockAdvanceRepository()
				return advanceRepo
			},
			wantErr: true,
			errCode: apperr.ErrInvalid,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			advanceRepo := tt.getMockRepo()
			advanceService := NewAdvanceService(advanceRepo)

			err := advanceService.UpdateAdvance(tt.args.ctx, tt.args.advance)

			if tt.wantErr {
				assert.Error(t, err, "UpdateAdvance() should return error")
				appErr, ok := err.(*apperr.AppError)
				assert.True(t, ok, "Error should be of type *apperr.AppError")
				assert.Equal(t, tt.errCode, appErr.Code, "Error code should match expected")
			} else {
				assert.NoError(t, err, "UpdateAdvance() should not return error")
			}

			advanceRepo.AssertExpectations(t)
		})
	}
}

func Test_services_AdvanceService_UpdateAdvanceStatus(t *testing.T) {
	validAdvanceID := "550e8400-e29b-41d4-a716-446655440000"
	adminID := "750e8400-e29b-41d4-a716-446655440000"

	type args struct {
		ctx       context.Context
		advanceID string
		status    domain.RequestStatus
	}
	tests := []struct {
		name        string
		args        args
		getMockRepo func() *mockrepository.AdvanceRepository
		wantErr     bool
		errCode     apperr.AppErrorCode
	}{
		{
			name: "update advance status to approved successfully",
			args: args{
				ctx: authctx.WithUserClaims(context.Background(), &authctx.UserClaims{
					UserID: adminID,
					Role:   domain.Admin,
				}),
				advanceID: validAdvanceID,
				status:    domain.Approved,
			},
			getMockRepo: func() *mockrepository.AdvanceRepository {
				advanceRepo := mockrepository.NewMockAdvanceRepository()
				advanceRepo.On("FindAdvanceById", mock.Anything, validAdvanceID).Return(domain.Advance{
					ID:     validAdvanceID,
					Status: domain.Pending,
				}, nil)
				advanceRepo.On("UpdateAdvance", mock.Anything, mock.AnythingOfType("domain.Advance")).Return(nil)
				return advanceRepo
			},
			wantErr: false,
		},
		{
			name: "update advance status to rejected successfully",
			args: args{
				ctx: authctx.WithUserClaims(context.Background(), &authctx.UserClaims{
					UserID: adminID,
					Role:   domain.Admin,
				}),
				advanceID: validAdvanceID,
				status:    domain.Rejected,
			},
			getMockRepo: func() *mockrepository.AdvanceRepository {
				advanceRepo := mockrepository.NewMockAdvanceRepository()
				advanceRepo.On("FindAdvanceById", mock.Anything, validAdvanceID).Return(domain.Advance{
					ID:     validAdvanceID,
					Status: domain.Pending,
				}, nil)
				advanceRepo.On("UpdateAdvance", mock.Anything, mock.AnythingOfType("domain.Advance")).Return(nil)
				return advanceRepo
			},
			wantErr: false,
		},
		{
			name: "update advance status to reviewed successfully",
			args: args{
				ctx: authctx.WithUserClaims(context.Background(), &authctx.UserClaims{
					UserID: adminID,
					Role:   domain.Admin,
				}),
				advanceID: validAdvanceID,
				status:    domain.Reviewed,
			},
			getMockRepo: func() *mockrepository.AdvanceRepository {
				advanceRepo := mockrepository.NewMockAdvanceRepository()
				advanceRepo.On("FindAdvanceById", mock.Anything, validAdvanceID).Return(domain.Advance{
					ID:     validAdvanceID,
					Status: domain.Approved,
				}, nil)
				advanceRepo.On("UpdateAdvance", mock.Anything, mock.AnythingOfType("domain.Advance")).Return(nil)
				return advanceRepo
			},
			wantErr: false,
		},
		{
			name: "unauthorized - no user claims",
			args: args{
				ctx:       context.Background(),
				advanceID: validAdvanceID,
				status:    domain.Approved,
			},
			getMockRepo: func() *mockrepository.AdvanceRepository {
				advanceRepo := mockrepository.NewMockAdvanceRepository()
				return advanceRepo
			},
			wantErr: true,
			errCode: apperr.ErrUnauthorized,
		},
		{
			name: "forbidden - user is employee",
			args: args{
				ctx: authctx.WithUserClaims(context.Background(), &authctx.UserClaims{
					UserID: adminID,
					Role:   domain.Employee,
				}),
				advanceID: validAdvanceID,
				status:    domain.Approved,
			},
			getMockRepo: func() *mockrepository.AdvanceRepository {
				advanceRepo := mockrepository.NewMockAdvanceRepository()
				return advanceRepo
			},
			wantErr: true,
			errCode: apperr.ErrForbidden,
		},
		{
			name: "invalid advance ID",
			args: args{
				ctx: authctx.WithUserClaims(context.Background(), &authctx.UserClaims{
					UserID: adminID,
					Role:   domain.Admin,
				}),
				advanceID: "invalid-id",
				status:    domain.Approved,
			},
			getMockRepo: func() *mockrepository.AdvanceRepository {
				advanceRepo := mockrepository.NewMockAdvanceRepository()
				return advanceRepo
			},
			wantErr: true,
			errCode: apperr.ErrInvalid,
		},
		{
			name: "invalid status - pending not allowed",
			args: args{
				ctx: authctx.WithUserClaims(context.Background(), &authctx.UserClaims{
					UserID: adminID,
					Role:   domain.Admin,
				}),
				advanceID: validAdvanceID,
				status:    domain.Pending,
			},
			getMockRepo: func() *mockrepository.AdvanceRepository {
				advanceRepo := mockrepository.NewMockAdvanceRepository()
				return advanceRepo
			},
			wantErr: true,
			errCode: apperr.ErrInvalid,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			advanceRepo := tt.getMockRepo()
			advanceService := NewAdvanceService(advanceRepo)

			err := advanceService.UpdateAdvanceStatus(tt.args.ctx, tt.args.advanceID, tt.args.status)

			if tt.wantErr {
				assert.Error(t, err, "UpdateAdvanceStatus() should return error")
				appErr, ok := err.(*apperr.AppError)
				assert.True(t, ok, "Error should be of type *apperr.AppError")
				assert.Equal(t, tt.errCode, appErr.Code, "Error code should match expected")
			} else {
				assert.NoError(t, err, "UpdateAdvanceStatus() should not return error")
			}

			advanceRepo.AssertExpectations(t)
		})
	}
}

func Test_services_AdvanceService_GetAdvanceByID(t *testing.T) {
	validAdvanceID := "550e8400-e29b-41d4-a716-446655440000"
	employeeID := "650e8400-e29b-41d4-a716-446655440000"

	type args struct {
		ctx       context.Context
		advanceID string
	}
	tests := []struct {
		name        string
		args        args
		getMockRepo func() *mockrepository.AdvanceRepository
		want        domain.Advance
		wantErr     bool
		errCode     apperr.AppErrorCode
	}{
		{
			name: "get advance by ID successfully - admin",
			args: args{
				ctx: authctx.WithUserClaims(context.Background(), &authctx.UserClaims{
					UserID: "admin-id",
					Role:   domain.Admin,
				}),
				advanceID: validAdvanceID,
			},
			getMockRepo: func() *mockrepository.AdvanceRepository {
				advanceRepo := mockrepository.NewMockAdvanceRepository()
				advanceRepo.On("FindAdvanceById", mock.Anything, validAdvanceID).Return(domain.Advance{
					ID:          validAdvanceID,
					UserID:      employeeID,
					Amount:      5000.00,
					Purpose:     "Business trip",
					Description: "Travel advance",
					Status:      domain.Pending,
				}, nil)
				return advanceRepo
			},
			want: domain.Advance{
				ID:          validAdvanceID,
				UserID:      employeeID,
				Amount:      5000.00,
				Purpose:     "Business trip",
				Description: "Travel advance",
				Status:      domain.Pending,
			},
			wantErr: false,
		},
		{
			name: "get advance by ID successfully - own advance",
			args: args{
				ctx: authctx.WithUserClaims(context.Background(), &authctx.UserClaims{
					UserID: employeeID,
					Role:   domain.Employee,
				}),
				advanceID: validAdvanceID,
			},
			getMockRepo: func() *mockrepository.AdvanceRepository {
				advanceRepo := mockrepository.NewMockAdvanceRepository()
				advanceRepo.On("FindAdvanceById", mock.Anything, validAdvanceID).Return(domain.Advance{
					ID:          validAdvanceID,
					UserID:      employeeID,
					Amount:      5000.00,
					Purpose:     "Business trip",
					Description: "Travel advance",
					Status:      domain.Pending,
				}, nil)
				return advanceRepo
			},
			want: domain.Advance{
				ID:          validAdvanceID,
				UserID:      employeeID,
				Amount:      5000.00,
				Purpose:     "Business trip",
				Description: "Travel advance",
				Status:      domain.Pending,
			},
			wantErr: false,
		},
		{
			name: "invalid advance ID",
			args: args{
				ctx: authctx.WithUserClaims(context.Background(), &authctx.UserClaims{
					UserID: "admin-id",
					Role:   domain.Admin,
				}),
				advanceID: "invalid-id",
			},
			getMockRepo: func() *mockrepository.AdvanceRepository {
				advanceRepo := mockrepository.NewMockAdvanceRepository()
				return advanceRepo
			},
			wantErr: true,
			errCode: apperr.ErrInvalid,
		},
		{
			name: "unauthorized - no user claims",
			args: args{
				ctx:       context.Background(),
				advanceID: validAdvanceID,
			},
			getMockRepo: func() *mockrepository.AdvanceRepository {
				advanceRepo := mockrepository.NewMockAdvanceRepository()
				return advanceRepo
			},
			wantErr: true,
			errCode: apperr.ErrUnauthorized,
		},
		{
			name: "forbidden - accessing another user's advance",
			args: args{
				ctx: authctx.WithUserClaims(context.Background(), &authctx.UserClaims{
					UserID: "other-employee-id",
					Role:   domain.Employee,
				}),
				advanceID: validAdvanceID,
			},
			getMockRepo: func() *mockrepository.AdvanceRepository {
				advanceRepo := mockrepository.NewMockAdvanceRepository()
				advanceRepo.On("FindAdvanceById", mock.Anything, validAdvanceID).Return(domain.Advance{
					ID:     validAdvanceID,
					UserID: employeeID,
				}, nil)
				return advanceRepo
			},
			wantErr: true,
			errCode: apperr.ErrForbidden,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			advanceRepo := tt.getMockRepo()
			advanceService := NewAdvanceService(advanceRepo)

			result, err := advanceService.GetAdvanceByID(tt.args.ctx, tt.args.advanceID)

			if tt.wantErr {
				assert.Error(t, err, "GetAdvanceByID() should return error")
				appErr, ok := err.(*apperr.AppError)
				assert.True(t, ok, "Error should be of type *apperr.AppError")
				assert.Equal(t, tt.errCode, appErr.Code, "Error code should match expected")
			} else {
				assert.NoError(t, err, "GetAdvanceByID() should not return error")
				assert.Equal(t, tt.want, result, "GetAdvanceByID() returned incorrect advance")
			}

			advanceRepo.AssertExpectations(t)
		})
	}
}

func Test_services_AdvanceService_GetAllAdvances(t *testing.T) {
	employeeID1 := "650e8400-e29b-41d4-a716-446655440000"
  employeeID2 := "650e8400-e29b-41d4-a716-446655440001"
	advanceID1 := "550e8400-e29b-41d4-a716-446655440000"
	advanceID2 := "550e8400-e29b-41d4-a716-446655440001"

	type args struct {
		ctx           context.Context
		filterOptions domain.AdvancesFilterOptions
	}
	tests := []struct {
		name        string
		args        args
		getMockRepo func() *mockrepository.AdvanceRepository
		wantList    []domain.Advance
		wantTotal   int
		wantErr     bool
		errCode     apperr.AppErrorCode
	}{
		{
			name: "get all advances successfully - admin",
			args: args{
				ctx: authctx.WithUserClaims(context.Background(), &authctx.UserClaims{
					UserID: "admin-id",
					Role:   domain.Admin,
				}),
				filterOptions: domain.AdvancesFilterOptions{
					Page:  1,
					Limit: 10,
				},
			},
			getMockRepo: func() *mockrepository.AdvanceRepository {
				advanceRepo := mockrepository.NewMockAdvanceRepository()
				advances := []domain.Advance{
					{ID: advanceID1, UserID: employeeID1, Amount: 5000.00},
					{ID: advanceID2, UserID: employeeID2, Amount: 3000.00},
				}
				advanceRepo.On("FindAllAdvances", mock.Anything, mock.AnythingOfType("domain.AdvancesFilterOptions")).
					Return(advances, 2, nil)
				return advanceRepo
			},
			wantList: []domain.Advance{
					{ID: advanceID1, UserID: employeeID1, Amount: 5000.00},
					{ID: advanceID2, UserID: employeeID2, Amount: 3000.00},
			},
			wantTotal: 2,
			wantErr:   false,
		},
		{
			name: "get advances for employee - filtered by userID",
			args: args{
				ctx: authctx.WithUserClaims(context.Background(), &authctx.UserClaims{
					UserID: employeeID1,
					Role:   domain.Employee,
				}),
				filterOptions: domain.AdvancesFilterOptions{
					Page:  1,
					Limit: 10,
				},
			},
			getMockRepo: func() *mockrepository.AdvanceRepository {
				advanceRepo := mockrepository.NewMockAdvanceRepository()
				advances := []domain.Advance{
					{ID: advanceID1, UserID: employeeID1, Amount: 5000.00},
				}
				// Employee requests should have their UserID automatically set
				advanceRepo.On("FindAllAdvances", mock.Anything, mock.MatchedBy(func(filter domain.AdvancesFilterOptions) bool {
					return filter.UserID == employeeID1
				})).Return(advances, 1, nil)
				return advanceRepo
			},
			wantList: []domain.Advance{
				{ID: advanceID1, UserID: employeeID1, Amount: 5000.00},
			},
			wantTotal: 1,
			wantErr:   false,
		},
		{
			name: "unauthorized - no user claims",
			args: args{
				ctx: context.Background(),
				filterOptions: domain.AdvancesFilterOptions{
					Page:  1,
					Limit: 10,
				},
			},
			getMockRepo: func() *mockrepository.AdvanceRepository {
				advanceRepo := mockrepository.NewMockAdvanceRepository()
				return advanceRepo
			},
			wantErr: true,
			errCode: apperr.ErrUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			advanceRepo := tt.getMockRepo()
			advanceService := NewAdvanceService(advanceRepo)

			resultList, resultTotal, err := advanceService.GetAllAdvances(tt.args.ctx, tt.args.filterOptions)

			if tt.wantErr {
				assert.Error(t, err, "GetAllAdvances() should return error")
				appErr, ok := err.(*apperr.AppError)
				assert.True(t, ok, "Error should be of type *apperr.AppError")
				assert.Equal(t, tt.errCode, appErr.Code, "Error code should match expected")
			} else {
				assert.NoError(t, err, "GetAllAdvances() should not return error")
				assert.Equal(t, tt.wantList, resultList, "GetAllAdvances() returned incorrect advances")
				assert.Equal(t, tt.wantTotal, resultTotal, "GetAllAdvances() returned incorrect total")
			}

			advanceRepo.AssertExpectations(t)
		})
	}
}

func Test_services_AdvanceService_GetAdvanceSummary(t *testing.T) {
	employeeID := "650e8400-e29b-41d4-a716-446655440000"

	tests := []struct {
		name        string
		ctx         context.Context
		getMockRepo func() *mockrepository.AdvanceRepository
		want        domain.AdvanceSummary
		wantErr     bool
		errCode     apperr.AppErrorCode
	}{
		{
			name: "get advance summary successfully - admin",
			ctx: authctx.WithUserClaims(context.Background(), &authctx.UserClaims{
				UserID: "admin-id",
				Role:   domain.Admin,
			}),
			getMockRepo: func() *mockrepository.AdvanceRepository {
				advanceRepo := mockrepository.NewMockAdvanceRepository()
				advanceRepo.On("GetAdvanceSumByStatus", mock.Anything, "", domain.Approved).Return(15000.00, nil)
				advanceRepo.On("GetReconciledAdvancesSum", mock.Anything, "").Return(8000.00, nil)
				advanceRepo.On("GetAdvanceSumByStatus", mock.Anything, "", domain.Pending).Return(5000.00, nil)
				advanceRepo.On("GetAdvanceSumByStatus", mock.Anything, "", domain.Rejected).Return(2000.00, nil)
				return advanceRepo
			},
			want: domain.AdvanceSummary{
				Approved:   15000.00,
				Reconciled: 8000.00,
				Pending:    5000.00,
				Rejected:   2000.00,
			},
			wantErr: false,
		},
		{
			name: "get advance summary successfully - employee",
			ctx: authctx.WithUserClaims(context.Background(), &authctx.UserClaims{
				UserID: employeeID,
				Role:   domain.Employee,
			}),
			getMockRepo: func() *mockrepository.AdvanceRepository {
				advanceRepo := mockrepository.NewMockAdvanceRepository()
				advanceRepo.On("GetAdvanceSumByStatus", mock.Anything, employeeID, domain.Approved).Return(8000.00, nil)
				advanceRepo.On("GetReconciledAdvancesSum", mock.Anything, employeeID).Return(5000.00, nil)
				advanceRepo.On("GetAdvanceSumByStatus", mock.Anything, employeeID, domain.Pending).Return(2000.00, nil)
				advanceRepo.On("GetAdvanceSumByStatus", mock.Anything, employeeID, domain.Rejected).Return(1000.00, nil)
				return advanceRepo
			},
			want: domain.AdvanceSummary{
				Approved:   8000.00,
				Reconciled: 5000.00,
				Pending:    2000.00,
				Rejected:   1000.00,
			},
			wantErr: false,
		},
		{
			name: "unauthorized - no user claims",
			ctx:  context.Background(),
			getMockRepo: func() *mockrepository.AdvanceRepository {
				advanceRepo := mockrepository.NewMockAdvanceRepository()
				return advanceRepo
			},
			wantErr: true,
			errCode: apperr.ErrUnauthorized,
		},
		{
			name: "error from repository - approved sum",
			ctx: authctx.WithUserClaims(context.Background(), &authctx.UserClaims{
				UserID: employeeID,
				Role:   domain.Employee,
			}),
			getMockRepo: func() *mockrepository.AdvanceRepository {
				advanceRepo := mockrepository.NewMockAdvanceRepository()
				advanceRepo.On("GetAdvanceSumByStatus", mock.Anything, employeeID, domain.Approved).
					Return(0.0, apperr.NewAppError(apperr.ErrInternal, "database error", nil))
				advanceRepo.On("GetReconciledAdvancesSum", mock.Anything, employeeID).Return(5000.00, nil)
				advanceRepo.On("GetAdvanceSumByStatus", mock.Anything, employeeID, domain.Pending).Return(2000.00, nil)
				advanceRepo.On("GetAdvanceSumByStatus", mock.Anything, employeeID, domain.Rejected).Return(1000.00, nil)
				return advanceRepo
			},
			wantErr: true,
			errCode: apperr.ErrInternal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			advanceRepo := tt.getMockRepo()
			advanceService := NewAdvanceService(advanceRepo)

			result, err := advanceService.GetAdvanceSummary(tt.ctx)

			if tt.wantErr {
				assert.Error(t, err, "GetAdvanceSummary() should return error")
				appErr, ok := err.(*apperr.AppError)
				assert.True(t, ok, "Error should be of type *apperr.AppError")
				assert.Equal(t, tt.errCode, appErr.Code, "Error code should match expected")
			} else {
				assert.NoError(t, err, "GetAdvanceSummary() should not return error")
				assert.Equal(t, tt.want, result, "GetAdvanceSummary() returned incorrect summary")
			}

			advanceRepo.AssertExpectations(t)
		})
	}
}

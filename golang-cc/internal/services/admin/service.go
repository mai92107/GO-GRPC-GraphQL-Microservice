package admin

import (
	"context"

	"github.com/rafa/golang-cc/internal/domain"
)

type Repository interface {
	Dashboard(context.Context) (domain.Dashboard, error)
	ListUsers(context.Context) ([]domain.AdminUser, error)
	ListTelegramBindings(context.Context) ([]domain.TelegramBinding, error)
	CreateTelegramBinding(context.Context, int64, string) error
	DeleteTelegramBinding(context.Context, int64) error
	UpdateUserStatus(context.Context, string, string, string) error
	UserEmail(context.Context, string) (string, error)
	ListInvitings(context.Context) ([]domain.Invitation, error)
	DeleteInvitation(context.Context, string) error
	ListBanks(context.Context) ([]domain.Bank, error)
	CreateBank(context.Context, string, domain.BankInput) error
	UpdateBank(context.Context, string, domain.BankInput) error
	DeleteBank(context.Context, string) error
	ListNetworks(context.Context) ([]string, error)
	ListCards(context.Context) ([]domain.Card, error)
	GetCardInfo(context.Context, string) (domain.CardInfo, error)
	CreateCard(context.Context, string, domain.CardInput) error
	UpdateCard(context.Context, string, domain.CardInput) error
	DeleteCard(context.Context, string) error
	ListActivities(context.Context, ActivityFilters) ([]domain.ActivitySummary, error)
	GetActivity(context.Context, string) (domain.ActivityFlow, error)
	CreateActivity(context.Context, domain.ActivityFlow) error
	UpdateActivity(context.Context, domain.ActivityFlow) error
	PublishActivity(context.Context, string, string, string) error
	SetActivityStatus(context.Context, string, bool) error
	DeleteActivity(context.Context, string) error
	ListRequirementTypes(context.Context) ([]domain.RequirementTypeOption, error)
	ActivityRequirementOptions(context.Context, RequirementOptionFilters) (domain.RequirementOptionSet, error)
	ListRegions(context.Context) ([]domain.LookupItem, error)
	CreateRegion(context.Context, domain.LookupItem) error
	UpdateRegion(context.Context, domain.LookupItem) error
	DeleteRegion(context.Context, string) error
	ListUserQualifications(context.Context) ([]domain.LookupItem, error)
	CreateUserQualification(context.Context, domain.LookupItem) error
	UpdateUserQualification(context.Context, domain.LookupItem) error
	DeleteUserQualification(context.Context, string) error
	ListCategories(context.Context) ([]domain.Category, error)
	CreateCategory(context.Context, string, string) error
	UpdateCategory(context.Context, string, string, bool) error
	DeleteCategory(context.Context, string) error
	ListRewardUnits(context.Context) ([]domain.RewardUnit, error)
	CreateRewardUnit(context.Context, string, domain.RewardUnitInput) error
	UpdateRewardUnit(context.Context, string, domain.RewardUnitInput) error
	DeleteRewardUnit(context.Context, string) error
	ListPaymentMethods(context.Context) ([]domain.PaymentMethod, error)
	CreatePaymentMethod(context.Context, domain.PaymentMethod) error
	UpdatePaymentMethod(context.Context, string, string, bool) error
	DeletePaymentMethod(context.Context, string) error
	ListMerchants(context.Context) ([]domain.Merchant, error)
	CreateMerchant(context.Context, domain.Merchant) (string, error)
	UpdateMerchant(context.Context, domain.Merchant) error
	DeleteMerchant(context.Context, string) error
}

type AuthService interface {
	CreateInvitation(context.Context, string, string) error
	RequestPasswordReset(context.Context, string) error
}

type Service struct {
	repository Repository
	auth       AuthService
}

func New(repository Repository, auth AuthService) *Service {
	return &Service{repository: repository, auth: auth}
}

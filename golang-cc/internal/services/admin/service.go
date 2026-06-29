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
	ListCardNetworks(context.Context) ([]string, error)
	ListCardProducts(context.Context) ([]domain.CardProduct, error)
	GetCardProduct(context.Context, string, bool) (domain.CardProduct, error)
	CreateCardProduct(context.Context, string, domain.CardProductInput) error
	UpdateCardProduct(context.Context, string, domain.CardProductInput) error
	DeleteCardProduct(context.Context, string) error
	ListActivities(context.Context) ([]domain.Activity, error)
	CreateActivity(context.Context, domain.Activity) error
	UpdateActivity(context.Context, domain.Activity) error
	DeleteActivity(context.Context, string) error
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
	PublishComponentVersion(context.Context, string, domain.RewardComponentVersionInput) (string, error)
	PublishConditionVersion(context.Context, string, domain.RewardConditionVersionInput) (string, error)
	PublishCapVersion(context.Context, string, domain.RewardCapVersionInput) (string, error)
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

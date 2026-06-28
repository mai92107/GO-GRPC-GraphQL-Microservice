package member

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/rafa/golang-cc/internal/domain"
	memberrepo "github.com/rafa/golang-cc/internal/repositories/member"
	"github.com/rafa/golang-cc/internal/services/recommendations"
)

type Service struct {
	repository   *memberrepo.Repository
	transactions *memberrepo.TransactionRepository
}

func New(repository *memberrepo.Repository, transactions *memberrepo.TransactionRepository) *Service {
	return &Service{repository: repository, transactions: transactions}
}

type CardInput struct {
	CardProductID string
	Nickname      string
	LastFour      string
	IsActive      bool
	StatementDay  *int
	PaymentDueDay *int
	AccountTier   string
	CardNetworkID string
}

type PreferenceInput struct {
	RewardUnitID string
	Weight       string
}

type RecommendationInput struct {
	AmountMinor  int64
	CategoryID   string
	MerchantID   string
	MerchantName string
	Date         recommendations.LocalDate
}

type TransactionInput struct {
	CardID          string
	AmountMinor     int64
	CategoryID      string
	MerchantName    string
	MerchantID      string
	PaymentMethodID string
	TransactionDate recommendations.LocalDate
	Note            string
}

type Transaction struct {
	ID                string
	UserID            string
	CardID            string
	AmountMinor       int64
	CategoryID        string
	MerchantName      string
	MerchantID        string
	PaymentMethodID   string
	PaymentMethodName string
	TransactionDate   recommendations.LocalDate
	Note              string
	Allocations       []recommendations.RuleEvaluation
}

var ErrTransactionNotFound = memberrepo.ErrNotFound

func (s *Service) Catalog(ctx context.Context) ([]domain.CatalogCard, error) {
	return s.repository.Catalog(ctx)
}

func (s *Service) CatalogCard(ctx context.Context, id string) (domain.CatalogCard, error) {
	return s.repository.CatalogCard(ctx, id)
}

func (s *Service) ListCards(ctx context.Context, userID string) ([]domain.MemberCard, error) {
	return s.repository.ListCards(ctx, userID)
}

func (s *Service) GetCard(ctx context.Context, userID, id string) (domain.MemberCard, error) {
	return s.repository.GetCard(ctx, userID, id)
}

func (s *Service) CreateCard(ctx context.Context, userID string, input CardInput) (string, error) {
	if input.CardProductID == "" {
		return "", domain.ErrInvalidInput
	}
	id := newID()
	err := s.repository.CreateCard(ctx, id, userID, input.CardProductID, memberrepo.CardWrite{
		Nickname: input.Nickname, LastFour: input.LastFour, IsActive: input.IsActive,
		StatementDay: input.StatementDay, PaymentDueDay: input.PaymentDueDay,
		AccountTier:   input.AccountTier,
		CardNetworkID: input.CardNetworkID,
	})
	return id, err
}

func (s *Service) UpdateCard(ctx context.Context, userID, id string, input CardInput) error {
	return s.repository.UpdateCard(ctx, id, userID, memberrepo.CardWrite{
		Nickname: input.Nickname, LastFour: input.LastFour, IsActive: input.IsActive,
		StatementDay: input.StatementDay, PaymentDueDay: input.PaymentDueDay,
		AccountTier:   input.AccountTier,
		CardNetworkID: input.CardNetworkID,
	})
}

func (s *Service) DeleteCard(ctx context.Context, userID, id string) error {
	return s.repository.DeleteCard(ctx, id, userID)
}

func (s *Service) RewardUnits(ctx context.Context) ([]domain.MemberRewardUnit, error) {
	return s.repository.RewardUnits(ctx)
}

func (s *Service) Categories(ctx context.Context) ([]domain.MemberCategory, error) {
	return s.repository.Categories(ctx)
}

func (s *Service) PaymentMethods(ctx context.Context) ([]domain.MemberPaymentMethod, error) {
	return s.repository.PaymentMethods(ctx, "")
}

func (s *Service) UserPaymentMethods(ctx context.Context, userID string) ([]domain.MemberPaymentMethod, error) {
	return s.repository.PaymentMethods(ctx, userID)
}

func (s *Service) UpdatePaymentMethods(ctx context.Context, userID string, codes []string) error {
	if userID == "" {
		return domain.ErrInvalidInput
	}
	return s.repository.UpdatePaymentMethods(ctx, userID, codes)
}

func (s *Service) RewardOverview(ctx context.Context, userID, memberCardID string, at time.Time) (domain.MemberCardRewardOverview, error) {
	if userID == "" || memberCardID == "" || at.IsZero() {
		return domain.MemberCardRewardOverview{}, domain.ErrInvalidInput
	}
	return s.repository.RewardOverview(ctx, userID, memberCardID, at)
}

func (s *Service) SetQualificationStatus(ctx context.Context, userID, memberCardID, planID string, qualified bool, effectiveFrom time.Time) error {
	if userID == "" || memberCardID == "" || planID == "" || effectiveFrom.IsZero() {
		return domain.ErrInvalidInput
	}
	return s.repository.SetQualificationStatus(ctx, userID, memberCardID, planID, qualified, effectiveFrom)
}

func (s *Service) Merchants(ctx context.Context, category string) ([]domain.MemberMerchant, error) {
	return s.repository.Merchants(ctx, category)
}

func (s *Service) Preferences(ctx context.Context, userID string) ([]domain.RewardPreference, error) {
	return s.repository.Preferences(ctx, userID)
}

func (s *Service) UpdatePreferences(ctx context.Context, userID string, input []PreferenceInput) error {
	items := make([]memberrepo.PreferenceWrite, 0, len(input))
	for _, item := range input {
		value, err := recommendations.ParseDecimal(item.Weight)
		if err != nil || value.Sign() < 0 || item.RewardUnitID == "" {
			return domain.ErrInvalidInput
		}
		items = append(items, memberrepo.PreferenceWrite{RewardUnitID: item.RewardUnitID, Weight: item.Weight})
	}
	return s.repository.UpdatePreferences(ctx, userID, items)
}

func (s *Service) Recommend(ctx context.Context, userID string, input RecommendationInput) (recommendations.Result, error) {
	return s.transactions.Recommend(ctx, recommendations.ID(userID), input.AmountMinor, input.CategoryID, input.MerchantID, input.MerchantName, input.Date)
}

func (s *Service) ListTransactions(ctx context.Context, userID, id string) ([]domain.TransactionSummary, error) {
	return s.transactions.List(ctx, userID, id)
}

func (s *Service) RewardCalculations(ctx context.Context, userID, transactionID string) ([]domain.RewardCalculationSnapshot, error) {
	if userID == "" || transactionID == "" {
		return nil, domain.ErrInvalidInput
	}
	return s.repository.RewardCalculations(ctx, userID, transactionID)
}

func (s *Service) CreateTransaction(ctx context.Context, userID string, input TransactionInput) (Transaction, error) {
	if err := validateTransaction(userID, input); err != nil {
		return Transaction{}, err
	}
	result, err := s.transactions.Create(ctx, recommendations.ID(userID), repositoryTransactionInput(input))
	return serviceTransaction(result), err
}

func (s *Service) UpdateTransaction(ctx context.Context, userID, id string, input TransactionInput) (Transaction, error) {
	if id == "" {
		return Transaction{}, domain.ErrInvalidInput
	}
	if err := validateTransaction(userID, input); err != nil {
		return Transaction{}, err
	}
	result, err := s.transactions.Update(ctx, recommendations.ID(userID), recommendations.ID(id), repositoryTransactionInput(input))
	return serviceTransaction(result), err
}

func (s *Service) DeleteTransaction(ctx context.Context, userID, id string) error {
	return s.transactions.Delete(ctx, recommendations.ID(userID), recommendations.ID(id))
}

func IsTransactionNotFound(err error) bool {
	return errors.Is(err, ErrTransactionNotFound)
}

func repositoryTransactionInput(input TransactionInput) memberrepo.WriteInput {
	return memberrepo.WriteInput{
		CardID: recommendations.ID(input.CardID), AmountMinor: input.AmountMinor,
		CategoryID: input.CategoryID, MerchantID: input.MerchantID, MerchantName: input.MerchantName, PaymentMethodID: input.PaymentMethodID,
		TransactionDate: input.TransactionDate, Note: input.Note,
	}
}

func serviceTransaction(input memberrepo.Transaction) Transaction {
	return Transaction{
		ID: string(input.ID), UserID: string(input.UserID), CardID: string(input.CardID),
		AmountMinor: input.AmountMinor, CategoryID: input.CategoryID, MerchantID: input.MerchantID, MerchantName: input.MerchantName, PaymentMethodID: input.PaymentMethodID, PaymentMethodName: input.PaymentMethodName,
		TransactionDate: input.TransactionDate, Note: input.Note, Allocations: input.Allocations,
	}
}

func validateTransaction(userID string, input TransactionInput) error {
	if userID == "" || input.CardID == "" || input.AmountMinor <= 0 ||
		input.TransactionDate.Time.IsZero() || strings.TrimSpace(input.CategoryID) == "" ||
		strings.TrimSpace(input.PaymentMethodID) == "" {
		return domain.ErrInvalidInput
	}
	return nil
}

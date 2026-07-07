package member

import (
	"context"
	"errors"
	"testing"

	"github.com/rafa/golang-cc/internal/domain"
)

func TestCreateCardRejectsMissingCatalogCard(t *testing.T) {
	service := &Service{}
	_, err := service.CreateCard(context.Background(), "member-id", CardInput{})
	if !errors.Is(err, domain.ErrInvalidInput) {
		t.Fatalf("expected invalid input, got %v", err)
	}
}

func TestCreateCardRejectsInvalidCreditLimit(t *testing.T) {
	service := &Service{}
	_, err := service.CreateCard(context.Background(), "member-id", CardInput{CardID: "card-id", CreditLimit: "0"})
	if !errors.Is(err, domain.ErrInvalidInput) {
		t.Fatalf("expected invalid input, got %v", err)
	}
}

func TestUpdateCardRejectsInvalidCreditLimit(t *testing.T) {
	service := &Service{}
	err := service.UpdateCard(context.Background(), "member-id", "member-card-id", CardInput{CreditLimit: ""})
	if !errors.Is(err, domain.ErrInvalidInput) {
		t.Fatalf("expected invalid input, got %v", err)
	}
}

func TestUpdatePreferencesRejectsInvalidWeight(t *testing.T) {
	service := &Service{}
	err := service.UpdatePreferences(context.Background(), "member-id", []PreferenceInput{
		{RewardUnitID: "unit-id", Weight: "-1"},
	})
	if !errors.Is(err, domain.ErrInvalidInput) {
		t.Fatalf("expected invalid input, got %v", err)
	}
}

func TestCreateTransactionRejectsInvalidInputWithoutDatabase(t *testing.T) {
	service := &Service{}
	_, err := service.CreateTransaction(context.Background(), "member-id", TransactionInput{})
	if !errors.Is(err, domain.ErrInvalidInput) {
		t.Fatalf("expected invalid input, got %v", err)
	}
}

package admin

import (
	"time"

	"github.com/rafa/golang-cc/internal/domain"
)

type dashboardResponse struct {
	Members          int `json:"members"`
	Banks            int `json:"banks"`
	Cards            int `json:"cards"`
	ActiveActivities int `json:"active_activities"`
}

type userResponse struct {
	ID          string    `json:"id"`
	Email       string    `json:"email"`
	DisplayName string    `json:"display_name"`
	Role        string    `json:"role"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
}

type telegramBindingResponse struct {
	ChatID      int64     `json:"chat_id"`
	UserID      string    `json:"user_id"`
	Email       string    `json:"email"`
	DisplayName string    `json:"display_name"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type invitationResponse struct {
	ID         string     `json:"id"`
	Email      string     `json:"email"`
	ExpiresAt  time.Time  `json:"expires_at"`
	AcceptedAt *time.Time `json:"accepted_at"`
	CreatedAt  time.Time  `json:"created_at"`
}

type bankResponse struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	Code       string `json:"code"`
	WebsiteURL string `json:"website_url"`
	IsActive   bool   `json:"is_active"`
}

type cardProductActivityResponse struct {
	ID         string                    `json:"id"`
	Name       string                    `json:"name"`
	StartDate  string                    `json:"start_date"`
	EndDate    string                    `json:"end_date"`
	IsActive   bool                      `json:"is_active"`
	SourceURL  string                    `json:"source_url"`
	VerifiedAt *string                   `json:"verified_at"`
	Benefits   []activityBenefitResponse `json:"benefits"`
}

type cardProductResponse struct {
	ID           string                        `json:"id"`
	BankID       string                        `json:"bank_id"`
	BankName     string                        `json:"bank_name"`
	Name         string                        `json:"name"`
	IsActive     bool                          `json:"is_active"`
	AccountTiers []string                      `json:"account_tiers"`
	Activities   []cardProductActivityResponse `json:"activities"`
}

type activityResponse struct {
	ID                string                    `json:"id"`
	CardProductID     string                    `json:"card_product_id"`
	Name              string                    `json:"name"`
	StartDate         string                    `json:"start_date"`
	EndDate           string                    `json:"end_date"`
	IsActive          bool                      `json:"is_active"`
	SourceURL         string                    `json:"source_url"`
	VerifiedAt        *string                   `json:"verified_at"`
	SharedMonthlyCaps map[string]string         `json:"shared_monthly_caps"`
	Benefits          []activityBenefitResponse `json:"benefits"`
}
type activityBenefitResponse struct {
	ID                   string   `json:"id"`
	RewardUnitID         string   `json:"reward_unit_id"`
	Name                 string   `json:"name"`
	Rate                 string   `json:"rate"`
	MonthlyCap           *string  `json:"monthly_cap"`
	StackGroup           string   `json:"stack_group"`
	Priority             int      `json:"priority"`
	RequiredAccountTiers []string `json:"required_account_tiers"`
	ActionRequired       string   `json:"action_required"`
	ActionMessage        string   `json:"action_message"`
	PaymentMethods       []string `json:"payment_methods"`
	CategoryCodes        []string `json:"category_codes"`
	MerchantCodes        []string `json:"merchant_codes"`
}

type categoryResponse struct {
	Code     string `json:"code"`
	Name     string `json:"name"`
	IsActive bool   `json:"is_active"`
}

func mapDashboard(value domain.Dashboard) dashboardResponse {
	return dashboardResponse(value)
}

func mapUsers(values []domain.AdminUser) []userResponse {
	result := make([]userResponse, 0, len(values))
	for _, value := range values {
		result = append(result, userResponse(value))
	}
	return result
}

func mapTelegramBindings(values []domain.TelegramBinding) []telegramBindingResponse {
	result := make([]telegramBindingResponse, 0, len(values))
	for _, value := range values {
		result = append(result, telegramBindingResponse(value))
	}
	return result
}

func mapInvitations(values []domain.Invitation) []invitationResponse {
	result := make([]invitationResponse, 0, len(values))
	for _, value := range values {
		result = append(result, invitationResponse(value))
	}
	return result
}

func mapBanks(values []domain.Bank) []bankResponse {
	result := make([]bankResponse, 0, len(values))
	for _, value := range values {
		result = append(result, bankResponse(value))
	}
	return result
}

func mapCardProducts(values []domain.CardProduct) []cardProductResponse {
	result := make([]cardProductResponse, 0, len(values))
	for _, value := range values {
		item := cardProductResponse{ID: value.ID, BankID: value.BankID, BankName: value.BankName, Name: value.Name, IsActive: value.IsActive, AccountTiers: value.AccountTiers, Activities: []cardProductActivityResponse{}}
		for _, activity := range value.Activities {
			item.Activities = append(item.Activities, mapCardActivity(activity))
		}
		result = append(result, item)
	}
	return result
}

func mapActivities(values []domain.Activity) []activityResponse {
	result := make([]activityResponse, 0, len(values))
	for _, value := range values {
		result = append(result, mapActivity(value))
	}
	return result
}

func mapBenefits(values []domain.ActivityBenefit) []activityBenefitResponse {
	out := make([]activityBenefitResponse, 0, len(values))
	for _, b := range values {
		out = append(out, activityBenefitResponse{ID: b.ID, RewardUnitID: b.RewardUnitID, Name: b.Name, Rate: b.Rate, MonthlyCap: b.MonthlyCap, StackGroup: b.StackGroup, Priority: b.Priority, RequiredAccountTiers: b.RequiredAccountTiers, ActionRequired: b.ActionRequired, ActionMessage: b.ActionMessage, PaymentMethods: b.PaymentMethods, CategoryCodes: b.CategoryCodes, MerchantCodes: b.MerchantCodes})
	}
	return out
}
func mapCardActivity(a domain.CardProductActivity) cardProductActivityResponse {
	return cardProductActivityResponse{ID: a.ID, Name: a.Name, StartDate: a.StartDate, EndDate: a.EndDate, IsActive: a.IsActive, SourceURL: a.SourceURL, VerifiedAt: a.VerifiedAt, Benefits: mapBenefits(a.Benefits)}
}
func mapActivity(a domain.Activity) activityResponse {
	return activityResponse{ID: a.ID, CardProductID: a.CardProductID, Name: a.Name, StartDate: a.StartDate, EndDate: a.EndDate, IsActive: a.IsActive, SourceURL: a.SourceURL, VerifiedAt: a.VerifiedAt, SharedMonthlyCaps: a.SharedMonthlyCaps, Benefits: mapBenefits(a.Benefits)}
}

func mapCategories(values []domain.Category) []categoryResponse {
	result := make([]categoryResponse, 0, len(values))
	for _, value := range values {
		result = append(result, categoryResponse(value))
	}
	return result
}

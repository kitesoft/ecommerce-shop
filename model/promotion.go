package model

import (
	"encoding/json"
	"io"
	"time"

	"github.com/dankobgd/ecommerce-shop/utils/locale"
	"github.com/nicksnyder/go-i18n/v2/i18n"
)

// error msgs
var (
	msgInvalidPromotion           = &i18n.Message{ID: "model.promotion.validate.app_error", Other: "invalid promotion data"}
	msgValidatePromotionPromoCode = &i18n.Message{ID: "model.promotion.validate.promo_code.app_error", Other: "invalid promo code"}
	msgValidatePromotionType      = &i18n.Message{ID: "model.promotion.validate.type.app_error", Other: "invalid promotion type"}
	msgValidatePromotionAmount    = &i18n.Message{ID: "model.promotion.validate.amount.app_error", Other: "invalid promotion amount value"}
	msgValidatePromotionStartsAt  = &i18n.Message{ID: "model.promotion.validate.starts_at.app_error", Other: "invalid promotion starts_at timestamp"}
	msgValidatePromotionEndsAt    = &i18n.Message{ID: "model.promotion.validate.ends_at.app_error", Other: "invalid promotion ends_at timestamp"}
)

// Promotion is the promotion model (discount for order)
type Promotion struct {
	TotalRecordsCount
	PromoCode   string    `json:"promo_code" db:"promo_code"`
	Type        string    `json:"type" db:"type"`
	Amount      int       `json:"amount" db:"amount"`
	Description string    `json:"description" db:"description"`
	StartsAt    time.Time `json:"starts_at" db:"starts_at"`
	EndsAt      time.Time `json:"ends_at" db:"ends_at"`
}

// PromotionDetail is is the promotion association
type PromotionDetail struct {
	UserID    int64  `json:"user_id" db:"user_id"`
	PromoCode string `json:"promo_code" db:"promo_code"`
}

// Validate validates the category and returns an error if it doesn't pass criteria
func (p *Promotion) Validate() *AppErr {
	var errs ValidationErrors
	l := locale.GetUserLocalizer("en")

	if p.PromoCode == "" {
		errs.Add(Invalid("promotion.promo_code", l, msgValidatePromotionPromoCode))
	}
	if p.Type == "" {
		errs.Add(Invalid("promotion.type", l, msgValidatePromotionType))
	}
	if p.Amount == 0 {
		errs.Add(Invalid("promotion.amount", l, msgValidatePromotionAmount))
	}
	if p.StartsAt.IsZero() {
		errs.Add(Invalid("promotion.starts_at", l, msgValidatePromotionStartsAt))
	}
	if p.EndsAt.IsZero() || p.EndsAt.Before(p.StartsAt) {
		errs.Add(Invalid("promotion.ends_at", l, msgValidatePromotionEndsAt))
	}

	if !errs.IsZero() {
		return NewValidationError("Promotion", msgInvalidPromotion, "", errs)
	}
	return nil
}

// PromotionPatch is the category patch model
type PromotionPatch struct {
	PromoCode   *string    `json:"promo_code"`
	Type        *string    `json:"type"`
	Amount      *int       `json:"amount"`
	Description *string    `json:"description"`
	StartsAt    *time.Time `json:"starts_at"`
	EndsAt      *time.Time `json:"ends_at"`
}

// Patch patches the category fields that are provided
func (p *Promotion) Patch(patch *PromotionPatch) {
	if patch.PromoCode != nil {
		p.PromoCode = *patch.PromoCode
	}
	if patch.Type != nil {
		p.Type = *patch.Type
	}
	if patch.Amount != nil {
		p.Amount = *patch.Amount
	}
	if patch.Description != nil {
		p.Description = *patch.Description
	}
	if patch.StartsAt != nil {
		p.StartsAt = *patch.StartsAt
	}
	if patch.EndsAt != nil {
		p.EndsAt = *patch.EndsAt
	}
}

// PromotionPatchFromJSON decodes the input and returns the PromotionPatch
func PromotionPatchFromJSON(data io.Reader) (*PromotionPatch, error) {
	var patch *PromotionPatch
	err := json.NewDecoder(data).Decode(&patch)
	return patch, err
}

// PromotionFromJSON decodes the input and returns the Promotion
func PromotionFromJSON(data io.Reader) (*Promotion, error) {
	var p *Promotion
	err := json.NewDecoder(data).Decode(&p)
	return p, err
}

// ToJSON converts Category to json string
func (p *Promotion) ToJSON() string {
	b, _ := json.Marshal(p)
	return string(b)
}

// IsActive checks if the promotion is currently active
func (p *Promotion) IsActive(t time.Time) bool {
	return t.After(p.StartsAt) && t.Before(p.EndsAt)
}
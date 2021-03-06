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
	msgInvalidReview = &i18n.Message{ID: "model.review.validate.app_error", Other: "invalid review data"}

	msgValidateReviewID        = &i18n.Message{ID: "model.review.validate.id.app_error", Other: "invalid  review id"}
	msgValidateReviewUserID    = &i18n.Message{ID: "model.review.validate.user_id.app_error", Other: "invalid review user id"}
	msgValidateReviewProductID = &i18n.Message{ID: "model.review.validate.product_id.app_error", Other: "invalid review product id"}
	msgValidateReviewRating    = &i18n.Message{ID: "model.review.validate.rating.app_error", Other: "invalid review rating"}
	msgValidateReviewTitle     = &i18n.Message{ID: "model.review.validate.title.app_error", Other: "invalid review title"}
	msgValidateReviewComment   = &i18n.Message{ID: "model.review.validate.comment.app_error", Other: "invalid review comment"}
	msgValidateReviewCrAt      = &i18n.Message{ID: "model.review.validate.created_at.app_error", Other: "invalid review created_at timestamp"}
	msgValidateReviewUpAt      = &i18n.Message{ID: "model.review.validate.updated_at.app_error", Other: "invalid review updated_at timestamp"}
)

// ProductReview is the review model
type ProductReview struct {
	ID        int64     `json:"id" db:"id"`
	UserID    int64     `json:"user_id" db:"user_id"`
	ProductID int64     `json:"product_id" db:"product_id"`
	Rating    int       `json:"rating" db:"rating"`
	Title     string    `json:"title" db:"title"`
	Comment   string    `json:"comment" db:"comment"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
	User      *User     `json:"user"`
}

// ProductReviewPatch is the patch for review
type ProductReviewPatch struct {
	Rating  *int    `json:"rating,omitempty"`
	Title   *string `json:"title,omitempty"`
	Comment *string `json:"comment,omitempty"`
}

// Patch patches the product review
func (rev *ProductReview) Patch(patch *ProductReviewPatch) {
	if patch.Rating != nil {
		rev.Rating = *patch.Rating
	}
	if patch.Title != nil {
		rev.Title = *patch.Title
	}
	if patch.Comment != nil {
		rev.Comment = *patch.Comment
	}
}

// ReviewFromJSON decodes the input and returns the Review
func ReviewFromJSON(data io.Reader) (*ProductReview, error) {
	var rev *ProductReview
	err := json.NewDecoder(data).Decode(&rev)
	return rev, err
}

// ReviewPatchFromJSON decodes the input and returns the ReviewPatch
func ReviewPatchFromJSON(data io.Reader) (*ProductReviewPatch, error) {
	var p *ProductReviewPatch
	err := json.NewDecoder(data).Decode(&p)
	return p, err
}

// PreSave will fill timestamps
func (rev *ProductReview) PreSave() {
	rev.CreatedAt = time.Now()
	rev.UpdatedAt = rev.CreatedAt
}

// PreUpdate sets the update timestamp
func (rev *ProductReview) PreUpdate() {
	rev.UpdatedAt = time.Now()
}

// Validate validates the review and returns an error if it doesn't pass criteria
func (rev *ProductReview) Validate() *AppErr {
	var errs ValidationErrors
	l := locale.GetUserLocalizer("en")

	if rev.ID != 0 {
		errs.Add(Invalid("id", l, msgValidateReviewID))
	}
	if rev.Rating < 0 || rev.Rating > 5 {
		errs.Add(Invalid("rating", l, msgValidateReviewRating))
	}
	if rev.Title == "" {
		errs.Add(Invalid("title", l, msgValidateReviewTitle))
	}
	if rev.Comment == "" {
		errs.Add(Invalid("comment", l, msgValidateReviewComment))
	}
	if rev.CreatedAt.IsZero() {
		errs.Add(Invalid("created_at", l, msgValidateReviewCrAt))
	}
	if rev.UpdatedAt.IsZero() {
		errs.Add(Invalid("updated_at", l, msgValidateReviewUpAt))
	}

	if !errs.IsZero() {
		return NewValidationError("Review", msgInvalidReview, "", errs)
	}
	return nil
}

// Validate validates the review and returns an error if it doesn't pass criteria
func (patch *ProductReviewPatch) Validate() *AppErr {
	var errs ValidationErrors
	l := locale.GetUserLocalizer("en")

	if patch.Rating != nil && *patch.Rating < 0 || *patch.Rating > 5 {
		errs.Add(Invalid("rating", l, msgValidateReviewRating))
	}
	if patch.Title != nil && *patch.Title == "" {
		errs.Add(Invalid("title", l, msgValidateReviewTitle))
	}
	if patch.Comment != nil && *patch.Comment == "" {
		errs.Add(Invalid("comment", l, msgValidateReviewComment))
	}

	if !errs.IsZero() {
		return NewValidationError("Review", msgInvalidReview, "", errs)
	}
	return nil
}

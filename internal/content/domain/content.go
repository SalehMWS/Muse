package domain

import (
	"strings"
	"time"

	"github.com/google/uuid"
)

const (
	MaxTitleLength   = 200
	MaxCaptionLength = 2200
	MaxTags          = 30
	MaxTagLength     = 50
)

type Status string

const (
	StatusDraft    Status = "draft"
	StatusArchived Status = "archived"
)

func ValidStatus(s Status) bool {
	return s == StatusDraft || s == StatusArchived
}

type ContentType string

const (
	TypeImage    ContentType = "image"
	TypeVideo    ContentType = "video"
	TypeCarousel ContentType = "carousel"
	TypeReel     ContentType = "reel"
	TypeStory    ContentType = "story"
)

func ValidContentType(t ContentType) bool {
	switch t {
	case TypeImage, TypeVideo, TypeCarousel, TypeReel, TypeStory:
		return true
	default:
		return false
	}
}

type Visibility string

const (
	VisibilityPublic  Visibility = "public"
	VisibilityPrivate Visibility = "private"
)

func ValidVisibility(v Visibility) bool {
	return v == VisibilityPublic || v == VisibilityPrivate
}

type Content struct {
	ID          uuid.UUID
	UserID      uuid.UUID
	Title       string
	Caption     string
	Status      Status
	Language    string
	ContentType ContentType
	Visibility  Visibility
	Tags        []string
	PublishedAt *time.Time
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type NewContentInput struct {
	Title       string
	Caption     string
	Language    string
	ContentType string
	Visibility  string
	Tags        []string
}

func NewContent(userID uuid.UUID, in NewContentInput) (Content, error) {
	title := strings.TrimSpace(in.Title)
	language := defaultString(strings.TrimSpace(in.Language), "en")
	contentType := defaultString(strings.TrimSpace(in.ContentType), string(TypeImage))
	visibility := defaultString(strings.TrimSpace(in.Visibility), string(VisibilityPrivate))

	if err := validateFields(title, in.Caption, contentType, visibility); err != nil {
		return Content{}, err
	}

	tags, err := NormalizeTags(in.Tags)
	if err != nil {
		return Content{}, err
	}

	now := time.Now()
	return Content{
		ID:          uuid.New(),
		UserID:      userID,
		Title:       title,
		Caption:     in.Caption,
		Status:      StatusDraft,
		Language:    language,
		ContentType: ContentType(contentType),
		Visibility:  Visibility(visibility),
		Tags:        tags,
		CreatedAt:   now,
		UpdatedAt:   now,
	}, nil
}

type UpdateContentInput struct {
	Title       *string
	Caption     *string
	Status      *string
	Language    *string
	ContentType *string
	Visibility  *string
	Tags        *[]string
}

func (c *Content) Apply(in UpdateContentInput) error {
	if in.Title != nil {
		c.Title = strings.TrimSpace(*in.Title)
	}
	if in.Caption != nil {
		c.Caption = *in.Caption
	}
	if in.Language != nil {
		c.Language = defaultString(strings.TrimSpace(*in.Language), "en")
	}
	if in.ContentType != nil {
		c.ContentType = ContentType(strings.TrimSpace(*in.ContentType))
	}
	if in.Visibility != nil {
		c.Visibility = Visibility(strings.TrimSpace(*in.Visibility))
	}
	if in.Status != nil {
		status := Status(strings.TrimSpace(*in.Status))
		if !ValidStatus(status) {
			return ErrInvalidStatus
		}
		c.Status = status
	}
	if in.Tags != nil {
		tags, err := NormalizeTags(*in.Tags)
		if err != nil {
			return err
		}
		c.Tags = tags
	}
	return validateFields(c.Title, c.Caption, string(c.ContentType), string(c.Visibility))
}

func (c *Content) Archive() {
	c.Status = StatusArchived
}

func (c Content) Duplicate() Content {
	now := time.Now()
	dup := c
	dup.ID = uuid.New()
	dup.Status = StatusDraft
	dup.Title = "Copy of " + c.Title
	dup.Tags = append([]string(nil), c.Tags...)
	dup.PublishedAt = nil
	dup.CreatedAt = now
	dup.UpdatedAt = now
	return dup
}

func NormalizeTags(raw []string) ([]string, error) {
	seen := make(map[string]struct{}, len(raw))
	out := make([]string, 0, len(raw))
	for _, tag := range raw {
		tag = strings.ToLower(strings.TrimSpace(tag))
		tag = strings.TrimPrefix(tag, "#")
		if tag == "" {
			continue
		}
		if len(tag) > MaxTagLength {
			return nil, ErrTagTooLong
		}
		if _, ok := seen[tag]; ok {
			continue
		}
		seen[tag] = struct{}{}
		out = append(out, tag)
	}
	if len(out) > MaxTags {
		return nil, ErrTooManyTags
	}
	return out, nil
}

func validateFields(title, caption, contentType, visibility string) error {
	if len(title) > MaxTitleLength {
		return ErrTitleTooLong
	}
	if len(caption) > MaxCaptionLength {
		return ErrCaptionTooLong
	}
	if !ValidContentType(ContentType(contentType)) {
		return ErrInvalidType
	}
	if !ValidVisibility(Visibility(visibility)) {
		return ErrInvalidVisibility
	}
	return nil
}

func defaultString(value, fallback string) string {
	if value == "" {
		return fallback
	}
	return value
}

package domain_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/SalehMWS/Muse/internal/content/domain"
)

func TestNewContent_Defaults(t *testing.T) {
	content, err := domain.NewContent(uuid.New(), domain.NewContentInput{Title: "  Hello  "})
	if err != nil {
		t.Fatalf("NewContent() unexpected error: %v", err)
	}
	if content.Title != "Hello" {
		t.Fatalf("Title = %q, want trimmed Hello", content.Title)
	}
	if content.Status != domain.StatusDraft {
		t.Fatalf("Status = %q, want draft", content.Status)
	}
	if content.Language != "en" || content.ContentType != domain.TypeImage || content.Visibility != domain.VisibilityPrivate {
		t.Fatalf("defaults wrong: %+v", content)
	}
	if content.ID == uuid.Nil {
		t.Fatal("ID not generated")
	}
}

func TestNewContent_Validation(t *testing.T) {
	tests := []struct {
		name string
		in   domain.NewContentInput
		want error
	}{
		{"title too long", domain.NewContentInput{Title: strings.Repeat("a", domain.MaxTitleLength+1)}, domain.ErrTitleTooLong},
		{"caption too long", domain.NewContentInput{Caption: strings.Repeat("a", domain.MaxCaptionLength+1)}, domain.ErrCaptionTooLong},
		{"invalid type", domain.NewContentInput{ContentType: "hologram"}, domain.ErrInvalidType},
		{"invalid visibility", domain.NewContentInput{Visibility: "secret"}, domain.ErrInvalidVisibility},
		{"tag too long", domain.NewContentInput{Tags: []string{strings.Repeat("t", domain.MaxTagLength+1)}}, domain.ErrTagTooLong},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := domain.NewContent(uuid.New(), tt.in); !errors.Is(err, tt.want) {
				t.Fatalf("NewContent() error = %v, want %v", err, tt.want)
			}
		})
	}
}

func TestNormalizeTags(t *testing.T) {
	tags, err := domain.NormalizeTags([]string{"#Foo", " Bar ", "foo", "", "BAR"})
	if err != nil {
		t.Fatalf("NormalizeTags() unexpected error: %v", err)
	}
	want := []string{"foo", "bar"}
	if len(tags) != len(want) {
		t.Fatalf("NormalizeTags() = %v, want %v", tags, want)
	}
	for i := range want {
		if tags[i] != want[i] {
			t.Fatalf("NormalizeTags()[%d] = %q, want %q", i, tags[i], want[i])
		}
	}

	many := make([]string, domain.MaxTags+1)
	for i := range many {
		many[i] = uuid.NewString()[:8]
	}
	if _, err := domain.NormalizeTags(many); !errors.Is(err, domain.ErrTooManyTags) {
		t.Fatalf("NormalizeTags() too many error = %v, want %v", err, domain.ErrTooManyTags)
	}
}

func TestContent_Apply(t *testing.T) {
	content, _ := domain.NewContent(uuid.New(), domain.NewContentInput{Title: "Original"})

	newTitle := "Updated"
	archived := string(domain.StatusArchived)
	if err := content.Apply(domain.UpdateContentInput{Title: &newTitle, Status: &archived}); err != nil {
		t.Fatalf("Apply() unexpected error: %v", err)
	}
	if content.Title != "Updated" || content.Status != domain.StatusArchived {
		t.Fatalf("Apply() did not update: %+v", content)
	}

	bad := "published"
	if err := content.Apply(domain.UpdateContentInput{Status: &bad}); !errors.Is(err, domain.ErrInvalidStatus) {
		t.Fatalf("Apply() invalid status error = %v, want %v", err, domain.ErrInvalidStatus)
	}
}

func TestContent_Duplicate(t *testing.T) {
	original, _ := domain.NewContent(uuid.New(), domain.NewContentInput{Title: "Launch", Tags: []string{"promo"}})
	original.Archive()

	dup := original.Duplicate()
	if dup.ID == original.ID {
		t.Fatal("Duplicate() reused ID")
	}
	if dup.Status != domain.StatusDraft {
		t.Fatalf("Duplicate() status = %q, want draft", dup.Status)
	}
	if dup.Title != "Copy of Launch" {
		t.Fatalf("Duplicate() title = %q, want Copy of Launch", dup.Title)
	}
	if len(dup.Tags) != 1 || dup.Tags[0] != "promo" {
		t.Fatalf("Duplicate() tags = %v, want [promo]", dup.Tags)
	}
}

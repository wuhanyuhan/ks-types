package kstypes

import (
	"reflect"
	"testing"
)

func TestSharedWidgetSchemas_AllRegistered(t *testing.T) {
	t.Parallel()
	expected := map[string]reflect.Type{
		"ks://widgets/list-actions@v1":   reflect.TypeOf(WidgetListActionsV1{}),
		"ks://widgets/diff-review@v1":    reflect.TypeOf(WidgetDiffReviewV1{}),
		"ks://widgets/timeline@v1":       reflect.TypeOf(WidgetTimelineV1{}),
		"ks://widgets/card-grid@v1":      reflect.TypeOf(WidgetCardGridV1{}),
		"ks://widgets/image-variants@v1": reflect.TypeOf(WidgetImageVariantsV1{}),
	}
	if len(SharedWidgetSchemas) != len(expected) {
		t.Fatalf("schema count: got %d, want %d", len(SharedWidgetSchemas), len(expected))
	}
	for uri, want := range expected {
		got, ok := SharedWidgetSchemas[uri]
		if !ok {
			t.Errorf("missing schema: %s", uri)
			continue
		}
		if got != want {
			t.Errorf("schema %s: got %v, want %v", uri, got, want)
		}
	}
}

type widgetValidator interface {
	Validate() error
}

func TestSharedWidgetSchemas_AllImplementValidator(t *testing.T) {
	t.Parallel()
	validatorType := reflect.TypeOf((*widgetValidator)(nil)).Elem()
	for uri, schemaType := range SharedWidgetSchemas {
		ptrType := reflect.PointerTo(schemaType)
		if !schemaType.Implements(validatorType) && !ptrType.Implements(validatorType) {
			t.Errorf("widget %s schema must implement Validate() error", uri)
		}
	}
}

func TestDeprecatedWidgets_EmptyAtV1(t *testing.T) {
	t.Parallel()
	if len(DeprecatedWidgets) != 0 {
		t.Errorf("DeprecatedWidgets should be empty at v1, got %d entries", len(DeprecatedWidgets))
	}
}

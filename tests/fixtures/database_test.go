package fixtures

import (
	"testing"

	"github.com/kainuguru/kainuguru-api/pkg/normalize"
)

func TestFixtureNormalizedProductName_NormalizesLithuanianText(t *testing.T) {
	normalizer := normalize.NewLithuanianNormalizer()
	got := fixtureNormalizedProductName(normalizer, "SŪRIS LIETUVIŠKAS 1kg")
	want := "Sūris Lietuviškas 1 kg"
	if got != want {
		t.Fatalf("expected %q, got %q", want, got)
	}
}

func TestFixtureNormalizedProductName_FallsBackWhenEmpty(t *testing.T) {
	normalizer := normalize.NewLithuanianNormalizer()
	input := "!!!"
	got := fixtureNormalizedProductName(normalizer, input)
	if got != input {
		t.Fatalf("expected fallback to original %q, got %q", input, got)
	}
}

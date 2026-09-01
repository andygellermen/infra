package app

import "testing"

func TestNewComposesAllFeatureModules(t *testing.T) {
	t.Parallel()

	application := New(nil)
	if application.Analysis == nil || application.Knowledge == nil || application.Rules == nil || application.Presentation == nil || application.Questions == nil {
		t.Fatalf("incomplete application composition: %+v", application)
	}
}

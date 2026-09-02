package experience

import (
	"encoding/json"
	"os"
	"reflect"
	"testing"

	"github.com/andygellermann/infra/apps/sprach-a-lyzer/internal/analysis"
	"github.com/andygellermann/infra/apps/sprach-a-lyzer/internal/questions"
)

func TestMVPExperienceGoldenV01(t *testing.T) {
	t.Parallel()
	data, err := os.ReadFile("../../data/golden/sprach-a-lyzer_mvp-experience_v0.1.json")
	if err != nil {
		t.Fatal(err)
	}
	var suite struct {
		Version  string `json:"suite_version"`
		Contract string `json:"experience_contract"`
		Cases    []struct {
			ID                   string   `json:"id"`
			Request              Request  `json:"request"`
			Patterns             []string `json:"expected_patterns"`
			AssessableDimensions []string `json:"expected_assessable_dimensions"`
			Alternatives         int      `json:"expected_alternatives"`
		} `json:"cases"`
	}
	if err := json.Unmarshal(data, &suite); err != nil {
		t.Fatal(err)
	}
	if suite.Version != "0.1" || suite.Contract != ContractVersion || len(suite.Cases) != 3 {
		t.Fatalf("suite envelope=%+v", suite)
	}
	analyzer := analysis.NewDefault()
	service := New(analyzer, questions.New(analyzer))
	for _, testCase := range suite.Cases {
		t.Run(testCase.ID, func(t *testing.T) {
			result, err := service.Analyze(testCase.Request)
			if err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(result.Core.Patterns, testCase.Patterns) || len(result.Alternatives) != testCase.Alternatives || len(result.SuggestedQuestions) != 5 {
				t.Fatalf("result patterns=%v alternatives=%d questions=%d", result.Core.Patterns, len(result.Alternatives), len(result.SuggestedQuestions))
			}
			assessable := []string{}
			for _, dimension := range result.Dimensions {
				if dimension.Score != nil {
					assessable = append(assessable, string(dimension.ID))
				}
			}
			if !reflect.DeepEqual(assessable, testCase.AssessableDimensions) {
				t.Fatalf("assessable dimensions=%v; want %v", assessable, testCase.AssessableDimensions)
			}
			if result.Privacy.RawTextStored || result.Privacy.AnalysisStored || result.Privacy.ExternalTransfer || result.Privacy.AIUsed {
				t.Fatalf("privacy=%+v", result.Privacy)
			}
		})
	}
}

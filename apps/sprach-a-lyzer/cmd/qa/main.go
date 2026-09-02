// Command qa exposes the deterministic Question/Answer MVP for local smoke tests.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/andygellermann/infra/apps/sprach-a-lyzer/internal/questions"
)

func main() {
	questionID := flag.String("question", "", "canonical question ID, for example CQ007")
	answer := flag.String("answer", "", "answer to analyze")
	profile := flag.String("profile", "PRIVATE", "PRIVATE or CORPORATE presentation profile")
	selectQuestions := flag.Bool("select", false, "offer an adaptive question set instead of analyzing an answer")
	renderQuestion := flag.Bool("render", false, "render a profile-specific question variant")
	action := flag.String("action", "DEFAULT", "render action: DEFAULT, SIMPLIFY or REPHRASE")
	deepOptIn := flag.Bool("deep-reflection-opt-in", false, "explicitly opt in to a private Deep Reflective rephrase")
	gaps := flag.String("gaps", "", "comma-separated construct gaps used by adaptive selection")
	limit := flag.Int("limit", 8, "number of offered questions (5 to 8)")
	flag.Parse()

	service := questions.NewDefault()
	var output any
	var err error
	if *renderQuestion {
		output, err = service.Render(questions.RenderRequest{
			QuestionID: *questionID, Profile: *profile, Action: *action, DeepReflectionOptIn: *deepOptIn,
		})
	} else if *selectQuestions {
		output, err = service.Select(questions.SelectionRequest{
			Profile: *profile, ConstructGaps: splitList(*gaps), Limit: *limit,
		})
	} else {
		output, err = service.Analyze(questions.AnswerRequest{
			QuestionID: *questionID, Answer: *answer, Profile: *profile,
		})
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetEscapeHTML(false)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(output); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func splitList(value string) []string {
	if strings.TrimSpace(value) == "" {
		return nil
	}
	parts := strings.Split(value, ",")
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		if trimmed := strings.TrimSpace(part); trimmed != "" {
			result = append(result, strings.ToUpper(trimmed))
		}
	}
	return result
}

package rules

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"regexp"
	"slices"

	"github.com/andygellermann/infra/apps/sprach-a-lyzer/internal/dimension"
	"github.com/andygellermann/infra/apps/sprach-a-lyzer/internal/policy"
)

const ContractVersion = "0.3"

var ruleKeyPattern = regexp.MustCompile(`^R-[A-Z0-9]+(?:-[A-Z0-9]+)*$`)
var uuidPattern = regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[1-8][0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$`)

type Condition struct {
	Op            string          `json:"op,omitempty"`
	Children      []Condition     `json:"children,omitempty"`
	Child         *Condition      `json:"child,omitempty"`
	Field         string          `json:"field,omitempty"`
	Operator      string          `json:"operator,omitempty"`
	Value         json.RawMessage `json:"value,omitempty"`
	CaseSensitive bool            `json:"case_sensitive,omitempty"`
}

type Action struct {
	Type          policy.RuleActionType `json:"type"`
	Dimension     dimension.ID          `json:"dimension,omitempty"`
	Value         *float64              `json:"value,omitempty"`
	Confidence    *float64              `json:"confidence,omitempty"`
	Factor        *float64              `json:"factor,omitempty"`
	ReasonKey     string                `json:"reason_key,omitempty"`
	Key           string                `json:"key,omitempty"`
	Lexeme        string                `json:"lexeme,omitempty"`
	Sense         string                `json:"sense,omitempty"`
	Tokens        []string              `json:"tokens,omitempty"`
	SemanticScore *bool                 `json:"semantic_score,omitempty"`
	MessageKey    string                `json:"message_key,omitempty"`
}

type Definition struct {
	ContractVersion    string    `json:"contract_version"`
	ID                 string    `json:"id"`
	Key                string    `json:"key"`
	Name               string    `json:"name"`
	Description        string    `json:"description"`
	Priority           int       `json:"priority"`
	Enabled            bool      `json:"enabled"`
	Scope              string    `json:"scope"`
	Status             string    `json:"status"`
	Version            int       `json:"version"`
	EvidenceClass      string    `json:"evidence_class,omitempty"`
	SourceKeys         []string  `json:"source_keys,omitempty"`
	Condition          Condition `json:"condition"`
	Actions            []Action  `json:"actions"`
	ConfidenceModifier float64   `json:"confidence_modifier"`
	StopProcessing     bool      `json:"stop_processing"`
}

func DecodeDefinition(data []byte) (Definition, error) {
	var definition Definition
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&definition); err != nil {
		return Definition{}, fmt.Errorf("decode rule: %w", err)
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		return Definition{}, fmt.Errorf("decode rule: trailing JSON value")
	}
	if err := definition.Validate(); err != nil {
		return Definition{}, err
	}
	return definition, nil
}

func (d Definition) Validate() error {
	if d.ContractVersion != ContractVersion {
		return fmt.Errorf("rule %q contract_version = %q; want %q", d.Key, d.ContractVersion, ContractVersion)
	}
	if !uuidPattern.MatchString(d.ID) || !ruleKeyPattern.MatchString(d.Key) || d.Name == "" {
		return fmt.Errorf("rule %q must contain a canonical id, key and name", d.Key)
	}
	if d.Priority < 0 || d.Priority > 10000 || d.Version < 1 || d.ConfidenceModifier < 0 || d.ConfidenceModifier > 1 {
		return fmt.Errorf("rule %s contains an out-of-range version, priority or confidence modifier", d.Key)
	}
	if !slices.Contains([]string{"TEXT", "QUESTION_ANSWER", "SPOKEN_DICTATION"}, d.Scope) {
		return fmt.Errorf("rule %s has unsupported scope %q", d.Key, d.Scope)
	}
	if !slices.Contains([]string{"DRAFT", "TESTING", "APPROVED", "PRODUCTION", "ARCHIVED"}, d.Status) {
		return fmt.Errorf("rule %s has unsupported status %q", d.Key, d.Status)
	}
	if d.EvidenceClass != "" && !slices.Contains([]string{"A", "B", "C", "D", "E"}, d.EvidenceClass) {
		return fmt.Errorf("rule %s has unsupported evidence class %q", d.Key, d.EvidenceClass)
	}
	if !uniqueStrings(d.SourceKeys) {
		return fmt.Errorf("rule %s contains duplicate source keys", d.Key)
	}
	if err := validateCondition(d.Condition, 0); err != nil {
		return fmt.Errorf("rule %s condition: %w", d.Key, err)
	}
	if len(d.Actions) == 0 {
		return fmt.Errorf("rule %s must contain at least one action", d.Key)
	}
	for index, action := range d.Actions {
		if err := validateAction(action); err != nil {
			return fmt.Errorf("rule %s action %d: %w", d.Key, index, err)
		}
	}
	return nil
}

func validateCondition(condition Condition, depth int) error {
	if depth > 16 {
		return fmt.Errorf("condition depth exceeds 16")
	}
	switch condition.Op {
	case "AND", "OR":
		if len(condition.Children) == 0 || condition.Child != nil || condition.Field != "" || condition.Operator != "" || len(condition.Value) != 0 || condition.CaseSensitive {
			return fmt.Errorf("%s requires children only", condition.Op)
		}
		for _, child := range condition.Children {
			if err := validateCondition(child, depth+1); err != nil {
				return err
			}
		}
		return nil
	case "NOT":
		if condition.Child == nil || len(condition.Children) != 0 || condition.Field != "" || condition.Operator != "" || len(condition.Value) != 0 || condition.CaseSensitive {
			return fmt.Errorf("NOT requires one child only")
		}
		return validateCondition(*condition.Child, depth+1)
	case "":
	default:
		return fmt.Errorf("unsupported operation %q", condition.Op)
	}
	if len(condition.Children) != 0 || condition.Child != nil {
		return fmt.Errorf("predicate must not contain child conditions")
	}
	fields := []string{"selected_sense", "tokens", "phrase", "context", "input_mode", "target_type", "expectation_source", "discourse_relation", "pattern", "proposition_feature"}
	operators := []string{"EQUALS", "NOT_EQUALS", "CONTAINS", "CONTAINS_ANY", "CONTAINS_ALL", "MATCHES", "EXISTS", "NOT_EXISTS"}
	if !slices.Contains(fields, condition.Field) || !slices.Contains(operators, condition.Operator) {
		return fmt.Errorf("unsupported predicate %q/%q", condition.Field, condition.Operator)
	}
	hasValue := len(condition.Value) > 0
	needsValue := condition.Operator != "EXISTS" && condition.Operator != "NOT_EXISTS"
	if hasValue != needsValue {
		return fmt.Errorf("operator %s value presence is invalid", condition.Operator)
	}
	return nil
}

func validateAction(action Action) error {
	if !slices.Contains(policy.RuleActionTypes(), action.Type) {
		return fmt.Errorf("unregistered action type %q", action.Type)
	}
	allowedFields := map[policy.RuleActionType][]string{
		policy.AddContribution:      {"dimension", "value", "confidence", "reason_key"},
		policy.MultiplyContribution: {"dimension", "factor", "reason_key"},
		policy.CapMin:               {"dimension", "value", "reason_key"},
		policy.CapMax:               {"dimension", "value", "reason_key"},
		policy.SetValue:             {"dimension", "value", "reason_key"},
		policy.Invert:               {"dimension", "reason_key"},
		policy.Suppress:             {"dimension", "reason_key"},
		policy.MarkNonAssessable:    {"dimension", "reason_key"},
		policy.AddPattern:           {"key"},
		policy.AddExplanation:       {"key"},
		policy.AddReflectionPrompt:  {"key"},
		policy.SelectSense:          {"lexeme", "sense"},
		policy.AddResonanceHint:     {"tokens", "semantic_score", "message_key"},
		policy.StopRuleChain:        {},
	}
	for field := range populatedActionFields(action) {
		if !slices.Contains(allowedFields[action.Type], field) {
			return fmt.Errorf("%s does not allow field %s", action.Type, field)
		}
	}
	canonicalDimension := slices.Contains(dimension.All(), action.Dimension)
	switch action.Type {
	case policy.AddContribution, policy.CapMin, policy.CapMax, policy.SetValue:
		if !canonicalDimension || action.Value == nil || action.ReasonKey == "" || *action.Value < -100 || *action.Value > 100 {
			return fmt.Errorf("%s requires canonical dimension, bounded value and reason_key", action.Type)
		}
	case policy.MultiplyContribution:
		if !canonicalDimension || action.Factor == nil || action.ReasonKey == "" || *action.Factor < 0 || *action.Factor > 10 {
			return fmt.Errorf("%s requires canonical dimension, bounded factor and reason_key", action.Type)
		}
	case policy.Invert, policy.Suppress, policy.MarkNonAssessable:
		if !canonicalDimension || action.ReasonKey == "" {
			return fmt.Errorf("%s requires canonical dimension and reason_key", action.Type)
		}
	case policy.AddPattern, policy.AddExplanation, policy.AddReflectionPrompt:
		if action.Key == "" {
			return fmt.Errorf("%s requires key", action.Type)
		}
	case policy.SelectSense:
		if action.Lexeme == "" || action.Sense == "" {
			return fmt.Errorf("SELECT_SENSE requires lexeme and sense")
		}
	case policy.AddResonanceHint:
		if len(action.Tokens) < 2 || !uniqueStrings(action.Tokens) || action.SemanticScore == nil || *action.SemanticScore {
			return fmt.Errorf("ADD_RESONANCE_HINT requires two tokens and semantic_score=false")
		}
	case policy.StopRuleChain:
	}
	if action.Confidence != nil && (*action.Confidence < 0 || *action.Confidence > 1) {
		return fmt.Errorf("confidence must be between 0 and 1")
	}
	return nil
}

func populatedActionFields(action Action) map[string]bool {
	result := make(map[string]bool)
	if action.Dimension != "" {
		result["dimension"] = true
	}
	if action.Value != nil {
		result["value"] = true
	}
	if action.Confidence != nil {
		result["confidence"] = true
	}
	if action.Factor != nil {
		result["factor"] = true
	}
	if action.ReasonKey != "" {
		result["reason_key"] = true
	}
	if action.Key != "" {
		result["key"] = true
	}
	if action.Lexeme != "" {
		result["lexeme"] = true
	}
	if action.Sense != "" {
		result["sense"] = true
	}
	if len(action.Tokens) != 0 {
		result["tokens"] = true
	}
	if action.SemanticScore != nil {
		result["semantic_score"] = true
	}
	if action.MessageKey != "" {
		result["message_key"] = true
	}
	return result
}

func uniqueStrings(values []string) bool {
	seen := make(map[string]bool, len(values))
	for _, value := range values {
		if value == "" || seen[value] {
			return false
		}
		seen[value] = true
	}
	return true
}

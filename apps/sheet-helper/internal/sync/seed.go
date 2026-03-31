package sync

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/andygellermann/infra/apps/sheet-helper/internal/model"
	"github.com/andygellermann/infra/apps/sheet-helper/internal/storage"
)

type SeedData struct {
	Routes    []model.Route      `json:"routes"`
	VCards    []model.VCardEntry `json:"vcards"`
	Texts     []model.TextEntry  `json:"texts"`
	ListItems []model.ListItem   `json:"list_items"`
}

func SeedFromFile(ctx context.Context, store *storage.Store, path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read seed file: %w", err)
	}

	var seed SeedData
	if err := json.Unmarshal(data, &seed); err != nil {
		return fmt.Errorf("decode seed json: %w", err)
	}

	if err := store.ReplaceAll(ctx, seed.Routes, seed.VCards, seed.Texts, seed.ListItems); err != nil {
		return fmt.Errorf("replace data: %w", err)
	}
	return nil
}

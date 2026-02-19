package utils

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"
)

const positionsFilePath = "/app/pdf/positions.json"

type PositionEntry struct {
	FieldKey string  `json:"fieldKey"`
	X        float64 `json:"x"`
	Y        float64 `json:"y"`
	FontSize float64 `json:"fontSize"`
	Page     int     `json:"page"`
}

type PositionsDB map[string][]PositionEntry

var posMu sync.RWMutex

func LoadPositions() (PositionsDB, error) {
	posMu.RLock()
	defer posMu.RUnlock()

	data, err := os.ReadFile(positionsFilePath)
	if err != nil {
		if os.IsNotExist(err) {
			return PositionsDB{}, nil
		}
		return nil, fmt.Errorf("read positions: %w", err)
	}
	var db PositionsDB
	if err := json.Unmarshal(data, &db); err != nil {
		return nil, fmt.Errorf("parse positions: %w", err)
	}
	return db, nil
}

func SavePositions(db PositionsDB) error {
	posMu.Lock()
	defer posMu.Unlock()

	data, err := json.MarshalIndent(db, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal positions: %w", err)
	}
	return os.WriteFile(positionsFilePath, data, 0644)
}

func GetStampsForRealm(realm string, c ContractData) ([]TextStamp, error) {
	db, err := LoadPositions()
	if err != nil {
		return nil, err
	}
	entries, ok := db[realm]
	if !ok || len(entries) == 0 {
		return nil, fmt.Errorf("no positions configured for realm %q", realm)
	}
	return entriesToStamps(entries, c), nil
}

func entriesToStamps(entries []PositionEntry, c ContractData) []TextStamp {
	stamps := make([]TextStamp, 0, len(entries))
	for _, e := range entries {
		text := fieldValue(e.FieldKey, c)
		stamps = append(stamps, TextStamp{
			Text:     text,
			X:        e.X,
			Y:        e.Y,
			FontSize: e.FontSize,
			Page:     e.Page,
		})
	}
	return stamps
}

func fieldValue(key string, c ContractData) string {
	switch key {
	case "NumID": return fmt.Sprintf("%d", c.NumID)
	case "CreatedAt": return c.CreatedAt.Format("02.01.2006")
	case "ItemAmount": return fmt.Sprintf("%g", c.ItemAmount)
	case "PaymentMethod": return c.PaymentMethod
	case "Realm": return c.Realm
	case "Currency": return c.Currency
	case "CurrencyAmount": return fmt.Sprintf("%g", c.CurrencyAmount)
	case "UserName": return c.UserName
	case "UserPhone": return c.UserPhone
	case "UserEmail": return c.UserEmail
	case "Package": return c.Package
	case "Item": return c.Item
	case "Insurance": return fmt.Sprintf("%t", c.Insurance)
	default: return ""
	}
}

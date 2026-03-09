package project

import (
	"encoding/json"

	"github.com/CDFalcon/ccmux/internal/migration"
)

var migrations = migration.NewRegistry()

func init() {
	migrations.Register(1, migrateV1ToV2)
}

func migrateV1ToV2(data []byte) ([]byte, error) {
	var raw map[string]interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, err
	}
	raw["version"] = 2
	return json.Marshal(raw)
}

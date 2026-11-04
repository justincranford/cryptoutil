// Copyright (c) 2025-2026 Justin Cranford.

package lifecycle

import json "encoding/json"

func jsonDecodeVerifyCopy(input any, destination any) error {
	raw, err := json.Marshal(input)
	if err != nil {
		return err
	}

	return json.Unmarshal(raw, destination)
}

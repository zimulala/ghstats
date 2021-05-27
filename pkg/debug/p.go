// Copyright 2021 ghstats Project Authors. Licensed under MIT.

package debug

import "encoding/json"

func PrettyFormat(input interface{}) string {
	output, err := json.MarshalIndent(input, "", "  ")
	if err != nil {
		panic(err)
	}
	return string(output)
}

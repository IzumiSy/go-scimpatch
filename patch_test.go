package scimpatch

import (
	"encoding/json"
	"testing"
)

func TestNewPatch(t *testing.T) {
	t.Run("Simple replace case", func(t *testing.T) {
		src := `{
			"name": "seiya"
		}`

		op := `[{
			"op": "replace",
			"path": "name",
  		"value": "akahane"
		}]`

		expectResult := `{
			"name": "akahane"
		}`

		var jsonifiedOp []Operation
		if err := json.Unmarshal(([]byte)(op), &jsonifiedOp); err != nil {
			t.FailNow()
			return
		}

		patch := NewPatch(jsonifiedOp)
		result, err := patch(src)
		if err != nil {
			t.FailNow()
			return
		}

		if result != expectResult {
			t.Errorf("Patched result is not equal to the expected result \n%s", result)
		}
	})
}

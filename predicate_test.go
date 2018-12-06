package scimpatch

import (
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestEvaluatePredicate(t *testing.T) {
	schema := &Schema{}
	err := json.Unmarshal([]byte(UserSchemaJson), &schema)
	assert.Nil(t, err)

	for _, test := range []struct {
		name       string
		filterText string
		data       Complex
		expect     bool
	}{
		{
			"eq",
			"userName eq \"david\"",
			Complex{"userName": "david"},
			true,
		},
		{
			"eq",
			"name.familyName eq \"qiu\"",
			Complex{"name": map[string]interface{}{"familyName": "Qiu"}},
			true,
		},
		{
			"ne",
			"name.familyName ne \"qiu\"",
			Complex{"name": map[string]interface{}{"familyName": "Qiu"}},
			false,
		},
		{
			"sw",
			"userName sw \"D\"",
			Complex{"userName": "david"},
			true,
		},
		{
			"ew",
			"userName ew \"D\"",
			Complex{"userName": "david"},
			true,
		},
		{
			"co",
			"userName co \"A\"",
			Complex{"userName": "david"},
			true,
		},
		{
			"gt",
			"meta.created gt \"2017-01-01\"",
			Complex{"meta": map[string]interface{}{"created": "2017-01-01"}},
			false,
		},
		{
			"ge",
			"meta.created ge \"2017-01-01\"",
			Complex{"meta": map[string]interface{}{"created": "2017-01-01"}},
			true,
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			filter, err := NewFilter(test.filterText)
			require.Nil(t, err)
			require.NotNil(t, filter)

			result := newPredicate(filter, schema).evaluate(test.data)
			assert.Equal(t, test.expect, result)
		})
	}
}

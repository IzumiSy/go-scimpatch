package scimpatch

import (
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestEvaluatePredicate(t *testing.T) {
	schema := &Schema{}
	err := json.Unmarshal([]byte(TestUserSchemaJson), &schema)
	assert.Nil(t, err)

	for _, test := range []struct {
		filterText string
		data       Complex
		expect     bool
	}{
		{
			"userName eq \"david\"",
			Complex{"userName": "david"},
			true,
		},
		{
			"name.familyName eq \"qiu\"",
			Complex{"name": map[string]interface{}{"familyName": "Qiu"}},
			true,
		},
		{
			"name.familyName ne \"qiu\"",
			Complex{"name": map[string]interface{}{"familyName": "Qiu"}},
			false,
		},
		{
			"userName sw \"D\"",
			Complex{"userName": "david"},
			true,
		},
		{
			"userName ew \"D\"",
			Complex{"userName": "david"},
			true,
		},
		{
			"userName co \"A\"",
			Complex{"userName": "david"},
			true,
		},
		{
			"meta.created gt \"2017-01-01\"",
			Complex{"meta": map[string]interface{}{"created": "2017-01-01"}},
			false,
		},
		{
			"meta.created ge \"2017-01-01\"",
			Complex{"meta": map[string]interface{}{"created": "2017-01-01"}},
			true,
		},
	} {
		filter, err := NewFilter(test.filterText)
		require.Nil(t, err)
		require.NotNil(t, filter)

		result := newPredicate(filter, schema).evaluate(test.data)
		assert.Equal(t, test.expect, result)
	}
}

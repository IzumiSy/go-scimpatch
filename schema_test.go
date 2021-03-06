package scimpatch

import (
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestSchema_GetAttribute(t *testing.T) {
	schema := &Schema{}
	err := json.Unmarshal([]byte(UserSchemaJson), &schema)
	assert.Nil(t, err)

	for _, test := range []struct {
		pathText  string
		assertion func(attr *Attribute)
	}{
		{
			"schemas",
			func(attr *Attribute) {
				require.NotNil(t, attr)
				assert.Equal(t, "schemas", attr.Assist.FullPath)
			},
		},
		{
			"ID",
			func(attr *Attribute) {
				require.NotNil(t, attr)
				assert.Equal(t, "id", attr.Assist.FullPath)
			},
		},
		{
			"meta.Created",
			func(attr *Attribute) {
				require.NotNil(t, attr)
				assert.Equal(t, "meta.created", attr.Assist.FullPath)
			},
		},
		{
			"Name.familyName",
			func(attr *Attribute) {
				require.NotNil(t, attr)
				assert.Equal(t, "urn:ietf:params:scim:schemas:core:2.0:User:name.familyName", attr.Assist.FullPath)
			},
		},
		{
			"urn:ietf:params:scim:schemas:core:2.0:User:emails",
			func(attr *Attribute) {
				require.NotNil(t, attr)
				assert.Equal(t, "urn:ietf:params:scim:schemas:core:2.0:User:emails", attr.Assist.FullPath)
			},
		},
		{
			"urn:ietf:params:scim:schemas:core:2.0:User:groups.value",
			func(attr *Attribute) {
				require.NotNil(t, attr)
				assert.Equal(t, "urn:ietf:params:scim:schemas:core:2.0:User:groups.value", attr.Assist.FullPath)
			},
		},
		{
			"groups[type eq \"direct\"].value",
			func(attr *Attribute) {
				require.NotNil(t, attr)
				assert.Equal(t, "urn:ietf:params:scim:schemas:core:2.0:User:groups.value", attr.Assist.FullPath)
			},
		},
	} {
		t.Run(test.pathText, func(t *testing.T) {
			p, err := NewPath(test.pathText)
			require.Nil(t, err)
			test.assertion(schema.GetAttribute(p, true))
		})
	}
}

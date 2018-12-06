package scimpatch

import (
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"reflect"
	"testing"
)

func TestApplyPatchUsers(t *testing.T) {
	schema := &Schema{}
	err := json.Unmarshal([]byte(UserSchemaJson), &schema)
	assert.Nil(t, err)

	for _, test := range []struct {
		patch     Patch
		assertion func(r *Resource, err error)
	}{
		{
			// add: simple path
			Patch{Op: Add, Path: "userName", Value: "foo"},
			func(r *Resource, err error) {
				assert.Nil(t, err)
				assert.Equal(t, "foo", r.GetData()["userName"])
			},
		},
		{
			// add: duplex path
			Patch{Op: Add, Path: "name.familyName", Value: "foo"},
			func(r *Resource, err error) {
				assert.Nil(t, err)
				assert.Equal(t, "foo", r.GetData()["name"].(map[string]interface{})["familyName"])
			},
		},
		{
			// add: AzureAD style
			Patch{Op: "Replace", Path: "displayName", Value: []interface{}{
				map[string]interface{}{"$ref": nil, "value": "hoge太郎"},
			}},
			func(r *Resource, err error) {
				assert.Nil(t, err)
				assert.Equal(t, "hoge太郎", r.GetData()["displayName"])
			},
		},
		{
			// add: AzureAD style for boolean
			Patch{Op: "Replace", Path: "active", Value: []interface{}{
				map[string]interface{}{"$ref": nil, "value": "True"},
			}},
			func(r *Resource, err error) {
				assert.Nil(t, err)
				assert.Equal(t, true, r.GetData()["active"])
			},
		},
		{
			// add: implicit path
			Patch{Op: Add, Path: "", Value: map[string]interface{}{"userName": "foo", "externalId": "bar"}},
			func(r *Resource, err error) {
				assert.Nil(t, err)
				assert.Equal(t, "foo", r.GetData()["userName"])
				assert.Equal(t, "bar", r.GetData()["externalId"])
			},
		},
		{
			// add: multivalued
			Patch{Op: Add, Path: "emails", Value: map[string]interface{}{"value": "foo@bar.com"}},
			func(r *Resource, err error) {
				assert.Nil(t, err)
				emailsVal := reflect.ValueOf(r.GetData()["emails"])
				if emailsVal.Kind() == reflect.Interface {
					emailsVal = emailsVal.Elem()
				}
				assert.Equal(t, 3, emailsVal.Len())
				assert.True(t, reflect.DeepEqual(emailsVal.Index(2).Interface(), map[string]interface{}{"value": "foo@bar.com"}))
			},
		},
		{
			// add: AzureAD style for multivalued
			Patch{Op: "Add", Path: "emails", Value: []interface{}{
				map[string]interface{}{"$ref": nil, "value": "foo@bar.com"},
			}},
			func(r *Resource, err error) {
				assert.Nil(t, err)
				emailsVal := reflect.ValueOf(r.GetData()["emails"])
				if emailsVal.Kind() == reflect.Interface {
					emailsVal = emailsVal.Elem()
				}
				assert.Equal(t, 3, emailsVal.Len())
				assert.True(t, reflect.DeepEqual(emailsVal.Index(2).Interface(), map[string]interface{}{
					"$ref": nil, "value": "foo@bar.com",
				}))
			},
		},
		{
			// add: duplex multivalued
			Patch{Op: Add, Path: "emails.value", Value: "foo@bar.com"},
			func(r *Resource, err error) {
				assert.Nil(t, err)
				emailsVal := reflect.ValueOf(r.GetData()["emails"])
				if emailsVal.Kind() == reflect.Interface {
					emailsVal = emailsVal.Elem()
				}
				assert.Equal(t, 2, emailsVal.Len())
				assert.Equal(t, "foo@bar.com", emailsVal.Index(0).Elem().MapIndex(reflect.ValueOf("value")).Interface())
				assert.Equal(t, "foo@bar.com", emailsVal.Index(1).Elem().MapIndex(reflect.ValueOf("value")).Interface())
			},
		},
		{
			// add: duplex multivalued with filter
			Patch{Op: Add, Path: "emails[type eq \"work\"].value", Value: "foo@bar.com"},
			func(r *Resource, err error) {
				assert.Nil(t, err)
				emailsVal := reflect.ValueOf(r.GetData()["emails"])
				if emailsVal.Kind() == reflect.Interface {
					emailsVal = emailsVal.Elem()
				}
				assert.Equal(t, 2, emailsVal.Len())
				assert.Equal(t, "foo@bar.com", emailsVal.Index(0).Elem().MapIndex(reflect.ValueOf("value")).Interface())
				assert.NotEqual(t, "foo@bar.com", emailsVal.Index(1).Elem().MapIndex(reflect.ValueOf("value")).Interface())
			},
		},
		{
			// replace: simple path
			Patch{Op: Replace, Path: "userName", Value: "foo"},
			func(r *Resource, err error) {
				assert.Nil(t, err)
				assert.Equal(t, "foo", r.GetData()["userName"])
			},
		},
		{
			// replace: duplex path
			Patch{Op: Replace, Path: "name.familyName", Value: "foo"},
			func(r *Resource, err error) {
				assert.Nil(t, err)
				assert.Equal(t, "foo", r.GetData()["name"].(map[string]interface{})["familyName"])
			},
		},
		{
			// replace : duplex multivalued
			Patch{Op: Add, Path: "emails.value", Value: "foo@bar.com"},
			func(r *Resource, err error) {
				assert.Nil(t, err)
				emailsVal := reflect.ValueOf(r.GetData()["emails"])
				if emailsVal.Kind() == reflect.Interface {
					emailsVal = emailsVal.Elem()
				}
				assert.Equal(t, 2, emailsVal.Len())
				assert.Equal(t, "foo@bar.com", emailsVal.Index(0).Elem().MapIndex(reflect.ValueOf("value")).Interface())
				assert.Equal(t, "foo@bar.com", emailsVal.Index(1).Elem().MapIndex(reflect.ValueOf("value")).Interface())
			},
		},
		{
			// replace: duplex multivalued with filter
			Patch{Op: Add, Path: "emails[type eq \"work\"].value", Value: "foo@bar.com"},
			func(r *Resource, err error) {
				assert.Nil(t, err)
				emailsVal := reflect.ValueOf(r.GetData()["emails"])
				if emailsVal.Kind() == reflect.Interface {
					emailsVal = emailsVal.Elem()
				}
				assert.Equal(t, 2, emailsVal.Len())
				assert.Equal(t, "foo@bar.com", emailsVal.Index(0).Elem().MapIndex(reflect.ValueOf("value")).Interface())
				assert.NotEqual(t, "foo@bar.com", emailsVal.Index(1).Elem().MapIndex(reflect.ValueOf("value")).Interface())
			},
		},
		{
			// remove: simple path
			Patch{Op: Remove, Path: "userName"},
			func(r *Resource, err error) {
				assert.Nil(t, err)
				assert.Nil(t, r.GetData()["userName"])
			},
		},
		{
			// remove: duplex path
			Patch{Op: Remove, Path: "name.familyName"},
			func(r *Resource, err error) {
				assert.Nil(t, err)
				assert.Nil(t, r.GetData()["name"].(map[string]interface{})["familyName"])
			},
		},
		{
			// remove: multivalued
			Patch{Op: Remove, Path: "emails"},
			func(r *Resource, err error) {
				assert.Nil(t, err)
				assert.Nil(t, r.GetData()["emails"])
			},
		},
		{
			// remove: duplex multivalued
			Patch{Op: Remove, Path: "emails.value"},
			func(r *Resource, err error) {
				assert.Nil(t, err)
				emailsVal := reflect.ValueOf(r.GetData()["emails"])
				if emailsVal.Kind() == reflect.Interface {
					emailsVal = emailsVal.Elem()
				}
				assert.Equal(t, 2, emailsVal.Len())
				assert.False(t, emailsVal.Index(0).Elem().MapIndex(reflect.ValueOf("value")).IsValid())
				assert.False(t, emailsVal.Index(1).Elem().MapIndex(reflect.ValueOf("value")).IsValid())
			},
		},
		{
			// replace: duplex multivalued with filter
			Patch{Op: Remove, Path: "emails[type eq \"work\"].value"},
			func(r *Resource, err error) {
				assert.Nil(t, err)
				emailsVal := reflect.ValueOf(r.GetData()["emails"])
				if emailsVal.Kind() == reflect.Interface {
					emailsVal = emailsVal.Elem()
				}
				assert.False(t, emailsVal.Index(0).Elem().MapIndex(reflect.ValueOf("value")).IsValid())
				assert.True(t, emailsVal.Index(1).Elem().MapIndex(reflect.ValueOf("value")).IsValid())
			},
		},
		{
			// replace: duplex multivalued with filter
			Patch{Op: Remove, Path: "emails[type eq \"work\"]"},
			func(r *Resource, err error) {
				assert.Nil(t, err)
				emailsVal := reflect.ValueOf(r.GetData()["emails"])
				if emailsVal.Kind() == reflect.Interface {
					emailsVal = emailsVal.Elem()
				}
				assert.Equal(t, 1, emailsVal.Len())
				assert.NotEqual(t, "work", emailsVal.Index(0).Elem().MapIndex(reflect.ValueOf("type")).Interface())
			},
		},
	} {
		data := make(map[string]interface{}, 0)
		err := json.Unmarshal([]byte(TestUserJson), &data)
		assert.Nil(t, err)

		resource := &Resource{Complex(data)}
		err = ApplyPatch(test.patch, resource, schema)
		test.assertion(resource, err)
	}
}

const TestUserJson = `
{
  "schemas": ["urn:ietf:params:scim:schemas:core:2.0:User"],
  "id": "6B69753B-4E38-444E-8AC6-9D0E4D644D80",
  "externalId": "996624032",
  "userName": "david@example.com",
  "name": {
    "formatted": "David Qiu",
    "familyName": "Qiu",
    "givenName": "David",
    "middleName": "",
    "honorificPrefix": "Mr.",
    "honorificSuffix": ""
  },
  "displayName": "David Qiu",
  "nickName": "Q",
  "profileUrl": "https://login.example.com/davidqiu",
  "emails": [
    {
      "value": "david@example.com",
      "type": "work",
      "primary": true
    },
    {
      "value": "david@home.com",
      "type": "home"
    }
  ],
  "addresses": [
    {
      "type": "work",
      "streetAddress": "123 Main Street",
      "locality": "Toronto",
      "region": "ON",
      "postalCode": "M1M A1A",
      "country": "CA",
      "formatted": "123 Main Street, Toronto ON, CA, M1M A1A",
      "primary": true
    }
  ],
  "phoneNumbers": [
    {
      "value": "123-456-7890",
      "type": "work"
    }
  ],
  "ims": [
    {
      "value": "someaimhandle",
      "type": "aim"
    }
  ],
  "photos": [
    {
      "value": "https://photos.example.com/profilephoto/72930000000Ccne/F",
      "type": "photo"
    },
    {
      "value": "https://photos.example.com/profilephoto/72930000000Ccne/T",
      "type": "thumbnail"
    }
  ],
  "userType": "Employee",
  "title": "Tour Guide",
  "preferredLanguage": "en-US",
  "locale": "en-US",
  "timezone": "America/Los_Angeles",
  "active":true,
  "password": "t1meMa$heen",
  "meta": {
    "resourceType": "User",
    "created": "2016-01-23T04:56:22Z",
    "lastModified": "2016-05-13T04:42:34Z",
    "version": "W\/\"a330bc54f0671c9\"",
    "location": "https://example.com/v2/Users/6B69753B-4E38-444E-8AC6-9D0E4D644D80"
  }
}
`

func TestApplyPatchGroup(t *testing.T) {
	schema := &Schema{}
	err := json.Unmarshal([]byte(GroupSchemaJson), &schema)
	assert.Nil(t, err)

	for _, test := range []struct {
		patch     Patch
		assertion func(r *Resource, err error)
	}{
		// AzureAD style multivalued
		{
			func() Patch {
				src := `
          {
            "schemas": [
              "urn:ietf:params:scim:api:messages:2.0:PatchOp"
            ],
            "Operations": [{
              "op": "Add",
              "path": "members", 
              "value": [{
                "$ref": null,
                "value": "hogehoge_group_id"
              }]
            }]
          }
				`

				var mods Modification
				if err := json.Unmarshal([]byte(src), &mods); err != nil {
					t.Errorf("Failed parsing modification: %s", err.Error())
					t.Fail()
				}

				return mods.Ops[0]
			}(),
			func(r *Resource, err error) {
				assert.Nil(t, err)
				membersVal := reflect.ValueOf(r.GetData()["members"])
				if membersVal.Kind() == reflect.Interface {
					membersVal = membersVal.Elem()
				}
				assert.Equal(t, 1, membersVal.Len())
				assert.True(t, reflect.DeepEqual(membersVal.Index(0).Interface(), map[string]interface{}{
					"$ref": nil, "value": "hogehoge_group_id",
				}))
			},
		},
	} {
		data := make(map[string]interface{}, 0)
		err := json.Unmarshal([]byte(TestGroupJson), &data)
		assert.Nil(t, err)

		resource := &Resource{Complex(data)}
		err = ApplyPatch(test.patch, resource, schema)
		test.assertion(resource, err)
	}
}

const TestGroupJson = `
{
  "schemas": ["urn:ietf:params:scim:schemas:core:2.0:Group"],
  "id": "e9e30dba-f08f-4109-8486-d5c6a331660a",
  "displayName": "Tour Guides",
  "meta": {
    "resourceType": "Group",
    "created": "2010-01-23T04:56:22Z",
    "lastModified": "2011-05-13T04:42:34Z",
    "version": "W\/\"3694e05e9dff592\"",
    "location": "https://example.com/v2/Groups/e9e30dba-f08f-4109-8486-d5c6a331660a"
  }
}
`

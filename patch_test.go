package scimpatch

import (
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"reflect"
	"testing"
)

func TestApplyPatch(t *testing.T) {
	schema := &Schema{}
	err := json.Unmarshal([]byte(TestUserSchemaJson), &schema)
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
			// add: implicit path
			Patch{Op: Add, Path: "", Value: map[string]interface{}{"userName": "foo", "externalId": "bar"}},
			func(r *Resource, err error) {
				assert.Nil(t, err)
				assert.Equal(t, "foo", r.GetData()["userName"])
				assert.Equal(t, "bar", r.GetData()["externalId"])
			},
		},
		{
			// add: multiValued
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
			// add : duplex multivalued
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
			// remove: multiValued
			Patch{Op: Remove, Path: "emails"},
			func(r *Resource, err error) {
				assert.Nil(t, err)
				assert.Nil(t, r.GetData()["emails"])
			},
		},
		{
			// remove : duplex multivalued
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

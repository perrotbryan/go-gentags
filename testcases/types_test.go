package main

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMarshal(t *testing.T) {
	in := Test{
		Name: "oui",
		Bar:  55,
		Baz:  "non",
		Parent: &Test{
			Name:   "jean",
			Bar:    95,
			Baz:    "val",
			Parent: nil,
		},
	}

	expected := []byte(`{"name":"oui","bar":55,"baz":"non","parent":{"name":"jean","bar":95,"baz":"val"}}`)

	got, err := json.Marshal(in)
	assert.NoError(t, err)
	assert.Equal(t, expected, got)
}

func TestUnmarshal(t *testing.T) {
	in := []byte(`{"name":"oui","bar":55,"baz":"non","parent":{"name":"jean","bar":95,"baz":"val"}}`)
	expected := Test{
		Name: "oui",
		Bar:  55,
		Baz:  "non",
		Parent: &Test{
			Name:   "jean",
			Bar:    95,
			Baz:    "val",
			Parent: nil,
		},
	}
	var got Test
	err := json.Unmarshal(in, &got)
	assert.NoError(t, err)
	assert.Equal(t, expected, got)
}

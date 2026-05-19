package common

import (
	"encoding/json"
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConvertStringLabelsToLabels_Empty(t *testing.T) {
	v := &Value{StringLabels: ""}
	err := v.ConvertStringLabelsToLabels()
	assert.NoError(t, err)
	assert.Nil(t, v.Labels)
}

func TestConvertStringLabelsToLabels_Valid(t *testing.T) {
	input := map[string]string{"host": "localhost", "env": "dev"}
	b, _ := json.Marshal(input)

	v := &Value{StringLabels: string(b)}
	err := v.ConvertStringLabelsToLabels()
	assert.NoError(t, err)
	assert.Equal(t, input, v.Labels)
}

func TestConvertLabelsToStringLabels_Nil(t *testing.T) {
	v := &Value{Labels: nil}
	err := v.ConvertLabelsToStringLabels()
	assert.NoError(t, err)
	assert.Equal(t, "", v.StringLabels)
}

func TestConvertLabelsToStringLabels_Valid(t *testing.T) {
	labels := map[string]string{"host": "localhost", "env": "dev"}
	v := &Value{Labels: labels}

	err := v.ConvertLabelsToStringLabels()
	assert.NoError(t, err)

	// String order in JSON is not guaranteed; compare as maps
	var out map[string]string
	err = json.Unmarshal([]byte(v.StringLabels), &out)
	assert.NoError(t, err)
	assert.Equal(t, labels, out)
}

type keyCheck struct {
	Name   string   `json:"name"`
	Labels []string `json:"labels"`
}

func TestMakeKey(t *testing.T) {
	v := &Value{
		Name:   "cpu_usage",
		Labels: map[string]string{"host": "server1", "region": "us-east"},
	}

	err := v.MakeKey()
	assert.NoError(t, err)
	assert.NotEmpty(t, v.AlertId)

	// Verify the AlertId encodes the expected key content
	var k keyCheck
	err = json.Unmarshal([]byte(v.AlertId), &k)
	assert.NoError(t, err)
	assert.Equal(t, v.Name, k.Name)

	expected := []string{"host:server1", "region:us-east"}
	sort.Strings(expected)
	assert.Equal(t, expected, k.Labels)
}

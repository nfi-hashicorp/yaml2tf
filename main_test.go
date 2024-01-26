package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"
)

func assertYAMLToTF(t *testing.T, y string, tf string) {
	t.Helper()
	yn := yaml.Node{}
	yaml.Unmarshal([]byte(y), &yn)
	assert.Equal(t, tf, string(yamlToTF(&yn).Bytes()))
}

func TestYAMLToTF_fullyQuotedMap(t *testing.T) {
	assertYAMLToTF(t, `
---
# foo
"foo": "bar"
"biz": "boz"
`, `{
  # foo
  "foo" = "bar",
  "biz" = "boz",
}`)
}

func TestYAMLToTF_mixedQuotesMap(t *testing.T) {
	t.Skip("TODO: everything is quoted")
	assertYAMLToTF(t, `
---
foo: bar
"biz": boz
`, `{
  "foo" = "bar",
  biz = "boz",
}`)
}

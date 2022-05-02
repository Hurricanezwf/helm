package encoding

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestYAMLStreamToJSONChunk(t *testing.T) {
	yamlstream := `---
a: b
c: d
e:
  f: g
h:
- i
- j	

---
k: l
m: n
o:
  p: q
r:
- s
`
	jsonChunkList, err := YAMLStreamToJSONChunk(yamlstream)
	if err != nil {
		t.Fatalf(err.Error())
	}
	if len(jsonChunkList) != 2 {
		t.Fatalf("expected 2 json chunk, got %d", len(jsonChunkList))
	}
	for _, v := range jsonChunkList {
		if len(v) == 0 {
			t.Fatalf("expected non-empty json chunk, got empty")
		}
		t.Log(string(v))
	}
}

func TestYAMLStreamToJSONChunkWithEmpty(t *testing.T) {
	yamlstream := `---
a: b
c: d
e:
  f: g
h:
- i
- j	

---
`
	jsonChunkList, err := YAMLStreamToJSONChunk(yamlstream)
	if err != nil {
		t.Fatalf(err.Error())
	}
	if len(jsonChunkList) != 1 {
		t.Fatalf("expected 1 json chunk, got %d", len(jsonChunkList))
	}
	for _, v := range jsonChunkList {
		if len(v) == 0 {
			t.Fatalf("expected non-empty json chunk, got empty")
		}
		t.Log(string(v))
	}
}

func TestJSONChunkToYAMLStream(t *testing.T) {
	jsonSnippet1 := `{"hello":"world"}`
	jsonSnippet2 := `{"world":"hello"}`

	yamlDoc, err := JSONChunkToYAMLStream([]string{jsonSnippet1, jsonSnippet2})
	t.Logf("output yamlDoc: \n%s\n", yamlDoc)
	require.NoError(t, err)
	require.NotEmpty(t, yamlDoc)

	jsonChunks, err := YAMLStreamToJSONChunk(yamlDoc)
	require.NoError(t, err)
	require.Equal(t, 2, len(jsonChunks))
	require.Equal(t, jsonSnippet1, jsonChunks[0])
	require.Equal(t, jsonSnippet2, jsonChunks[1])
}

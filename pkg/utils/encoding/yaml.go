package encoding

import (
	"bufio"
	"bytes"
	"fmt"
	"strings"

	"sigs.k8s.io/yaml"
)

// JSONChunkToYAMLStream converts a JSON chunk into a YAML stream.
func JSONChunkToYAMLStream(jsonChunk []string) (string, error) {
	if len(jsonChunk) == 0 {
		return "", nil
	}
	builder := strings.Builder{}
	for idx, chunk := range jsonChunk {
		if idx == 0 {
			builder.WriteString("---\n")
		} else {
			builder.WriteString("\n---\n")
		}
		builder.WriteString(chunk)
	}
	yamlDoc := builder.String()
	if _, err := YAMLStreamToJSONChunk(yamlDoc); err != nil {
		return "", fmt.Errorf("failed to convert json chunk to yaml stream, %w", err)
	}
	return yamlDoc, nil
}

// YAMLStreamToJSONChunk converts a YAML stream into a JSON chunks.
func YAMLStreamToJSONChunk(stream string) ([]string, error) {
	var (
		docIndex      int
		jsonChunkList []string
		scanner       = bufio.NewScanner(bytes.NewBufferString(stream))
		buf           = make([]byte, 4*1024) // the size of initial allocation for buffer 4k
	)

	scanner.Buffer(buf, 5*1024*1024) // the maximum size used to buffer a token 5M
	scanner.Split(splitYAMLDocument)

	for scanner.Scan() {
		docIndex++
		b, err := yaml.YAMLToJSON(scanner.Bytes())
		if err != nil {
			return nil, fmt.Errorf("failed to convert yaml document at segment %d to json, %w", docIndex, err)
		}
		jsonChunkList = append(jsonChunkList, string(b))
	}
	return jsonChunkList, nil
}

const yamlSeparator = "\n---"
const separator = "---"

// splitYAMLDocument is a bufio.SplitFunc for splitting YAML streams into individual documents.
func splitYAMLDocument(data []byte, atEOF bool) (advance int, token []byte, err error) {
	if atEOF && len(data) == 0 {
		return 0, nil, nil
	}
	sep := len([]byte(yamlSeparator))
	if i := bytes.Index(data, []byte(yamlSeparator)); i >= 0 {
		// We have a potential document terminator
		i += sep
		after := data[i:]
		if len(after) == 0 {
			// we can't read any more characters
			if atEOF {
				return len(data), data[:len(data)-sep], nil
			}
			return 0, nil, nil
		}
		if j := bytes.IndexByte(after, '\n'); j >= 0 {
			return i + j + 1, data[0 : i-sep], nil
		}
		return 0, nil, nil
	}
	// If we're at EOF, we have a final, non-terminated line. Return it.
	if atEOF {
		return len(data), data, nil
	}
	// Request more data.
	return 0, nil, nil
}

package templateassets

import (
	"strings"
	"testing"
)

func TestOpenAPIVersionTypesAreIndependent(t *testing.T) {
	tests := []struct {
		path       string
		lowerTypes []string
	}{
		{path: "openapi/v3_1.ts", lowerTypes: []string{"V3_0", "v3_0"}},
		{path: "openapi/v3_2.ts", lowerTypes: []string{"V3_0", "v3_0", "V3_1", "v3_1"}},
	}

	for _, test := range tests {
		t.Run(test.path, func(t *testing.T) {
			contents, err := SupportFS.ReadFile(test.path)
			if err != nil {
				t.Fatal(err)
			}

			for _, lowerType := range test.lowerTypes {
				if strings.Contains(string(contents), lowerType) {
					t.Errorf("%s contains lower-version dependency %q", test.path, lowerType)
				}
			}
		})
	}
}

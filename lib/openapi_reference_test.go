package ccode

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOpenAPIReference_ValidationBoundaryMatchesGoLoader(t *testing.T) {
	references := []string{
		"./paths/countries.yaml#/countries",
		"../schemas/common.types.yaml#/Country",
		"#/components/schemas/Country",
		"./paths/countries.yaml",
		"./encoded%20path.yaml#%2FCountry%20Name",
		"./bad%2.yaml#/Country",
		"./countries.yaml?version=1#/Country",
		"https://example.com/countries.yaml#/Country",
		"//example.com/countries.yaml#/Country",
		"file:///tmp/countries.yaml#/Country",
		"./countries.yaml#Country",
	}

	type boundaryCase struct {
		Reference string `json:"reference"`
		Allowed   bool   `json:"allowed"`
	}
	cases := make([]boundaryCase, 0, len(references))
	expected := make([]bool, 0, len(references))
	sourcePath := filepath.Join(t.TempDir(), "server.yaml")
	for _, reference := range references {
		_, _, err := resolveOpenAPIReferenceLocation(reference, sourcePath)
		allowed := err == nil
		cases = append(cases, boundaryCase{Reference: reference, Allowed: allowed})
		expected = append(expected, allowed)
	}

	encodedCases, err := json.Marshal(cases)
	require.NoError(t, err)

	ctx, projectDir := setupRunnerTestProject(t, "TestOpenAPIReference_ValidationBoundaryMatchesGoLoader")
	processFile := filepath.Join(projectDir, "openapi", "reference-boundary.ts")
	process := fmt.Sprintf(`import type { Context } from "@ccode/context";
import * as OpenAPI from "@ccode/openapi";

const cases: Array<{ reference: string; allowed: boolean }> = %s;

export default function main(ctx: Context) {
  const actual = cases.map(({ reference }) => {
    try {
      OpenAPI.parseReference(reference);
      return true;
    } catch {
      return false;
    }
  });
  ctx.println(JSON.stringify(actual));
}
`, encodedCases)
	require.NoError(t, os.MkdirAll(filepath.Dir(processFile), 0755))
	require.NoError(t, os.WriteFile(processFile, []byte(process), 0644))

	var output bytes.Buffer
	ctx.stdout = &output
	require.NoError(t, ctx.Run("openapi/reference-boundary"))

	var actual []bool
	require.NoError(t, json.Unmarshal(output.Bytes(), &actual))
	assert.Equal(t, expected, actual)
}

func TestOpenAPIReference_RunnerParsesAndValidatesReferences(t *testing.T) {
	ctx, projectDir := setupRunnerTestProject(t, "TestOpenAPIReference_RunnerParsesAndValidatesReferences")
	processFile := filepath.Join(projectDir, "openapi", "references.ts")

	require.NoError(t, os.MkdirAll(filepath.Dir(processFile), 0755))
	require.NoError(t, os.WriteFile(processFile, []byte(`import type { Context } from "@ccode/context";
import * as OpenAPI from "@ccode/openapi";

function assertParsed(
  input: string,
  expected: OpenAPI.ReferenceParts,
): void {
  const actual = OpenAPI.parseReference(input);
  if (JSON.stringify(actual) !== JSON.stringify(expected)) {
    throw new Error(
      "unexpected parsed reference for " +
        actual.raw +
        ": " +
        JSON.stringify(actual),
    );
  }
}

function expectFailure(input: unknown, expectedMessage: string): void {
  try {
    OpenAPI.parseReference(input as string);
  } catch (error) {
    if (
      error instanceof Error &&
      error.message.includes(expectedMessage)
    ) {
      return;
    }
    throw error;
  }
  throw new Error("expected reference to fail: " + String(input));
}

export default function main(ctx: Context) {
  const cases: Array<{
    input: string;
    expected: OpenAPI.ReferenceParts;
  }> = [
    {
      input: "./paths/countries.yaml#/countries",
      expected: {
        raw: "./paths/countries.yaml#/countries",
        document: "./paths/countries.yaml",
        directory: "paths",
        directorySegments: ["paths"],
        filename: "countries.yaml",
        documentName: "countries",
        fragment: "/countries",
      },
    },
    {
      input: "./paths/app/countries.yaml#/countries",
      expected: {
        raw: "./paths/app/countries.yaml#/countries",
        document: "./paths/app/countries.yaml",
        directory: "paths/app",
        directorySegments: ["paths", "app"],
        filename: "countries.yaml",
        documentName: "countries",
        fragment: "/countries",
      },
    },
    {
      input: "./paths/app/common/countries.yaml#/countries",
      expected: {
        raw: "./paths/app/common/countries.yaml#/countries",
        document: "./paths/app/common/countries.yaml",
        directory: "paths/app/common",
        directorySegments: ["paths", "app", "common"],
        filename: "countries.yaml",
        documentName: "countries",
        fragment: "/countries",
      },
    },
    {
      input: "../schemas/common.types.yaml#/Country",
      expected: {
        raw: "../schemas/common.types.yaml#/Country",
        document: "../schemas/common.types.yaml",
        directory: "../schemas",
        directorySegments: ["..", "schemas"],
        filename: "common.types.yaml",
        documentName: "common.types",
        fragment: "/Country",
      },
    },
    {
      input: "#/components/schemas/Country",
      expected: {
        raw: "#/components/schemas/Country",
        document: "",
        directory: "",
        directorySegments: [],
        filename: "",
        documentName: "",
        fragment: "/components/schemas/Country",
      },
    },
    {
      input: "./paths/countries.yaml",
      expected: {
        raw: "./paths/countries.yaml",
        document: "./paths/countries.yaml",
        directory: "paths",
        directorySegments: ["paths"],
        filename: "countries.yaml",
        documentName: "countries",
        fragment: "",
      },
    },
    {
      input: "./paths/countries#/countries",
      expected: {
        raw: "./paths/countries#/countries",
        document: "./paths/countries",
        directory: "paths",
        directorySegments: ["paths"],
        filename: "countries",
        documentName: "countries",
        fragment: "/countries",
      },
    },
    {
      input:
        "./paths%20and%2Fschemas/common%2Etypes.yaml#%2FCountry%20Name",
      expected: {
        raw: "./paths%20and%2Fschemas/common%2Etypes.yaml#%2FCountry%20Name",
        document: "./paths%20and%2Fschemas/common%2Etypes.yaml",
        directory: "paths and/schemas",
        directorySegments: ["paths and", "schemas"],
        filename: "common.types.yaml",
        documentName: "common.types",
        fragment: "/Country Name",
      },
    },
  ];

  for (const testCase of cases) {
    assertParsed(testCase.input, testCase.expected);
  }

  const failures: Array<[unknown, string]> = [
    ["./bad%2.yaml#/Country", "malformed percent-encoding"],
    [" \t", "must not be empty or blank"],
    ["./countries.yaml?version=1#/Country", "must not contain a query"],
    ["https://example.com/countries.yaml#/Country", "local file path"],
    ["//example.com/countries.yaml#/Country", "local file path"],
    ["file:///tmp/countries.yaml#/Country", "local file path"],
    ["./countries.yaml#Country", "JSON Pointer"],
    [{ $ref: "./objects/country.yaml#/Country" }, "must be a string"],
    [{ $ref: 42 }, "must be a string"],
  ];
  for (const [input, message] of failures) {
    expectFailure(input, message);
  }

  const referenceObject = {
    $ref: "./objects/country.yaml#/Country",
  } as const;
  const positiveGuards = [
    OpenAPI.isReference(referenceObject),
    OpenAPI.isReference({ $ref: "#/components/schemas/Country" }),
  ];
  const negativeGuards = [
    OpenAPI.isReference(null),
    OpenAPI.isReference("./countries.yaml#/Country"),
    OpenAPI.isReference({}),
    OpenAPI.isReference({ $ref: 42 }),
  ];

  ctx.println(JSON.stringify({
    parsedCases: cases.length,
    rejectedCases: failures.length,
    positiveGuards,
    negativeGuards,
  }));
}
`), 0644))

	var output bytes.Buffer
	ctx.stdout = &output

	require.NoError(t, ctx.Run("openapi/references"))
	assert.JSONEq(t, `{
		"parsedCases": 8,
		"rejectedCases": 9,
		"positiveGuards": [true, true],
		"negativeGuards": [false, false, false, false]
	}`, output.String())
}

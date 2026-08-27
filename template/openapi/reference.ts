export interface ReferenceLike {
  readonly $ref: string;
}

export interface ReferenceParts {
  readonly raw: string;
  readonly document: string;
  readonly directory: string;
  readonly directorySegments: readonly string[];
  readonly filename: string;
  readonly documentName: string;
  readonly fragment: string;
}

const URI_SCHEME = /^[A-Za-z][A-Za-z0-9+.-]*:/;

export function isReference(value: unknown): value is ReferenceLike {
  return (
    typeof value === "object" &&
    value !== null &&
    typeof (value as { $ref?: unknown }).$ref === "string"
  );
}

export function parseReference(
  input: string | ReferenceLike,
): ReferenceParts {
  const raw = referenceValue(input);
  if (raw.trim() === "") {
    throw new Error("OpenAPI reference must not be empty or blank");
  }

  const fragmentSeparator = raw.indexOf("#");
  const document =
    fragmentSeparator === -1 ? raw : raw.slice(0, fragmentSeparator);
  const encodedFragment =
    fragmentSeparator === -1 ? "" : raw.slice(fragmentSeparator + 1);

  validateLocalDocument(document);

  const decodedDocument = decodeReferencePart(document, "file path");
  const fragment = decodeReferencePart(encodedFragment, "fragment");
  if (fragment !== "" && !fragment.startsWith("/")) {
    throw new Error(
      `Invalid OpenAPI reference fragment "#${fragment}": JSON Pointer must be empty or start with "/"`,
    );
  }

  const documentPath = decodedDocument.startsWith("./")
    ? decodedDocument.slice(2)
    : decodedDocument;
  const filenameSeparator = documentPath.lastIndexOf("/");
  const directory =
    filenameSeparator === -1 ? "" : documentPath.slice(0, filenameSeparator);
  const filename =
    filenameSeparator === -1
      ? documentPath
      : documentPath.slice(filenameSeparator + 1);

  return {
    raw,
    document,
    directory,
    directorySegments: directory === "" ? [] : directory.split("/"),
    filename,
    documentName: removeFinalExtension(filename),
    fragment,
  };
}

function referenceValue(input: string | ReferenceLike): string {
  if (typeof input === "string") {
    return input;
  }
  if (isReference(input)) {
    return input.$ref;
  }
  throw new Error(
    'OpenAPI reference must be a string or an object with a string "$ref"',
  );
}

function validateLocalDocument(document: string): void {
  if (URI_SCHEME.test(document) || document.startsWith("//")) {
    throw new Error(
      "OpenAPI reference must use a local file path without a URI scheme or host",
    );
  }
  if (document.includes("?")) {
    throw new Error("OpenAPI file reference must not contain a query string");
  }
}

function decodeReferencePart(value: string, part: string): string {
  try {
    return decodeURIComponent(value);
  } catch {
    throw new Error(
      `Invalid OpenAPI reference ${part}: malformed percent-encoding`,
    );
  }
}

function removeFinalExtension(filename: string): string {
  const extensionSeparator = filename.lastIndexOf(".");
  return extensionSeparator > 0
    ? filename.slice(0, extensionSeparator)
    : filename;
}

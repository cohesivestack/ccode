export type Initialisms = readonly string[];

export type StringTransformation = (value: string) => string;

export type InitialismTransformation = (
  value: string,
  initialisms?: Initialisms,
) => string;

export interface NativeStringUtilities {
  readonly camelCase: InitialismTransformation;
  readonly pascalCase: InitialismTransformation;
  readonly snakeCase: StringTransformation;
  readonly kebabCase: StringTransformation;
  readonly constantCase: StringTransformation;
  readonly dotCase: StringTransformation;
  readonly pathCase: StringTransformation;
  readonly titleCase: InitialismTransformation;
  readonly sentenceCase: InitialismTransformation;
  readonly upperFirst: StringTransformation;
  readonly lowerFirst: StringTransformation;
  readonly normalizeSpace: StringTransformation;
}

export interface NativeGoUtilities {
  readonly toExportedIdentifier: InitialismTransformation;
  readonly toUnexportedIdentifier: InitialismTransformation;
  readonly toPackageName: StringTransformation;
}

export type OpenAPIPathTransformation = (
  value: string,
  omitLeadingSlash?: boolean,
) => string;

export interface NativeOpenAPIPathUtilities {
  readonly toColon: OpenAPIPathTransformation;
  readonly toSquareBrackets: OpenAPIPathTransformation;
  readonly toAngleBrackets: OpenAPIPathTransformation;
  readonly toDollar: OpenAPIPathTransformation;
}

export interface NativeOpenAPIUtilities {
  readonly path: NativeOpenAPIPathUtilities;
}

export interface NativeUtilities {
  readonly string: NativeStringUtilities;
  readonly go: NativeGoUtilities;
  readonly openapi: NativeOpenAPIUtilities;
}

type NativeGlobal = typeof globalThis & {
  readonly __ccodeNative?: NativeUtilities;
};

export function getNativeUtilities(): NativeUtilities {
  const native = (globalThis as NativeGlobal).__ccodeNative;

  if (!native) {
    throw new Error(
      "Cohesive Code native utilities are not available in this runtime",
    );
  }

  return native;
}

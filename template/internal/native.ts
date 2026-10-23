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

export interface GoTypeOptions {
  readonly format?: string;
  readonly unsigned?: boolean;
  readonly source?: "openapi" | "postgresql" | "mysql" | "mariadb" | "sqlite";
  readonly overrides?: Readonly<Record<string, string>>;
  /** Defaults to null: throw when the type cannot be inferred. */
  readonly fallback?: string | null;
}

export interface NativeGoUtilities {
  readonly toExportedIdentifier: InitialismTransformation;
  readonly toUnexportedIdentifier: InitialismTransformation;
  readonly toPackageName: StringTransformation;
  readonly toType: (value: string, options?: GoTypeOptions) => string;
}

export interface NativeTypeScriptUtilities {
  readonly toTypeIdentifier: InitialismTransformation;
  readonly toValueIdentifier: InitialismTransformation;
}

export type PathTransformation = (
  value: string,
  omitLeadingSlash?: boolean,
) => string;

export interface NativePathUtilities {
  readonly toCurlyBrackets: PathTransformation;
  readonly toColon: PathTransformation;
  readonly toSquareBrackets: PathTransformation;
  readonly toAngleBrackets: PathTransformation;
  readonly toDollar: PathTransformation;
}

export interface NativeUtilities {
  readonly string: NativeStringUtilities;
  readonly go: NativeGoUtilities;
  readonly typescript: NativeTypeScriptUtilities;
  readonly path: NativePathUtilities;
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

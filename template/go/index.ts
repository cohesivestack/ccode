import { getNativeUtilities, type GoTypeOptions } from "../internal/native";

export function toExportedIdentifier(
  value: string,
  initialisms?: readonly string[],
): string {
  const transform = getNativeUtilities().go.toExportedIdentifier;

  return initialisms === undefined
    ? transform(value)
    : transform(value, initialisms);
}

export function toUnexportedIdentifier(
  value: string,
  initialisms?: readonly string[],
): string {
  const transform = getNativeUtilities().go.toUnexportedIdentifier;

  return initialisms === undefined
    ? transform(value)
    : transform(value, initialisms);
}

export function toPackageName(value: string): string {
  return getNativeUtilities().go.toPackageName(value);
}

export type { GoTypeOptions as TypeOptions } from "../internal/native";

/** Maps primitive metadata type names to Go expressions. Does not manage imports or containers. */
export function toType(value: string, options?: GoTypeOptions): string {
  return options === undefined
    ? getNativeUtilities().go.toType(value)
    : getNativeUtilities().go.toType(value, options);
}

import { getNativeUtilities } from "../internal/native";

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

import { getNativeUtilities } from "../internal/native";

export function toTypeIdentifier(
  value: string,
  initialisms?: readonly string[],
): string {
  const transform = getNativeUtilities().typescript.toTypeIdentifier;

  return initialisms === undefined
    ? transform(value)
    : transform(value, initialisms);
}

export function toValueIdentifier(
  value: string,
  initialisms?: readonly string[],
): string {
  const transform = getNativeUtilities().typescript.toValueIdentifier;

  return initialisms === undefined
    ? transform(value)
    : transform(value, initialisms);
}

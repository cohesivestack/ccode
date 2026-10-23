import { getNativeUtilities, type TypeScriptTypeOptions } from "../internal/native";

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

export type { TypeScriptTypeOptions as TypeOptions } from "../internal/native";

/** Maps primitive metadata names to TypeScript expressions. Does not build schemas or manage serialization. */
export function toType(value: string, options?: TypeScriptTypeOptions): string {
  return options === undefined
    ? getNativeUtilities().typescript.toType(value)
    : getNativeUtilities().typescript.toType(value, options);
}

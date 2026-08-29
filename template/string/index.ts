import { getNativeUtilities } from "../internal/native";

export function camelCase(
  value: string,
  initialisms?: readonly string[],
): string {
  const transform = getNativeUtilities().string.camelCase;

  return initialisms === undefined
    ? transform(value)
    : transform(value, initialisms);
}

export function pascalCase(
  value: string,
  initialisms?: readonly string[],
): string {
  const transform = getNativeUtilities().string.pascalCase;

  return initialisms === undefined
    ? transform(value)
    : transform(value, initialisms);
}

export function snakeCase(value: string): string {
  return getNativeUtilities().string.snakeCase(value);
}

export function kebabCase(value: string): string {
  return getNativeUtilities().string.kebabCase(value);
}

export function constantCase(value: string): string {
  return getNativeUtilities().string.constantCase(value);
}

export function dotCase(value: string): string {
  return getNativeUtilities().string.dotCase(value);
}

export function pathCase(value: string): string {
  return getNativeUtilities().string.pathCase(value);
}

export function titleCase(
  value: string,
  initialisms?: readonly string[],
): string {
  const transform = getNativeUtilities().string.titleCase;

  return initialisms === undefined
    ? transform(value)
    : transform(value, initialisms);
}

export function sentenceCase(
  value: string,
  initialisms?: readonly string[],
): string {
  const transform = getNativeUtilities().string.sentenceCase;

  return initialisms === undefined
    ? transform(value)
    : transform(value, initialisms);
}

export function upperFirst(value: string): string {
  return getNativeUtilities().string.upperFirst(value);
}

export function lowerFirst(value: string): string {
  return getNativeUtilities().string.lowerFirst(value);
}

export function normalizeSpace(value: string): string {
  return getNativeUtilities().string.normalizeSpace(value);
}

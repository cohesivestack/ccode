import { getNativeUtilities } from "../internal/native";

export interface TransformOptions {
  readonly omitLeadingSlash?: boolean;
}

export function toCurlyBrackets(
  path: string,
  options?: TransformOptions,
): string {
  const transform = getNativeUtilities().path.toCurlyBrackets;

  return options?.omitLeadingSlash === undefined
    ? transform(path)
    : transform(path, options.omitLeadingSlash);
}

export function toColon(
  path: string,
  options?: TransformOptions,
): string {
  const transform = getNativeUtilities().path.toColon;

  return options?.omitLeadingSlash === undefined
    ? transform(path)
    : transform(path, options.omitLeadingSlash);
}

export function toSquareBrackets(
  path: string,
  options?: TransformOptions,
): string {
  const transform = getNativeUtilities().path.toSquareBrackets;

  return options?.omitLeadingSlash === undefined
    ? transform(path)
    : transform(path, options.omitLeadingSlash);
}

export function toAngleBrackets(
  path: string,
  options?: TransformOptions,
): string {
  const transform = getNativeUtilities().path.toAngleBrackets;

  return options?.omitLeadingSlash === undefined
    ? transform(path)
    : transform(path, options.omitLeadingSlash);
}

export function toDollar(
  path: string,
  options?: TransformOptions,
): string {
  const transform = getNativeUtilities().path.toDollar;

  return options?.omitLeadingSlash === undefined
    ? transform(path)
    : transform(path, options.omitLeadingSlash);
}

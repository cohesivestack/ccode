import type * as V3_0 from "./v3_0";
import type * as V3_1 from "./v3_1";
import type * as V3_2 from "./v3_2";

export type Version = "3.0" | "3.1" | "3.2";

export interface DocumentByVersion<T extends {} = {}> {
  "3.0": V3_0.Document<T>;
  "3.1": V3_1.Document<T>;
  "3.2": V3_2.Document<T>;
}

// OpenAPI extensions can be declared using generics, for example:
// Document<"3.1", {
//   "x-amazon-apigateway-integration": AWSAPITGatewayDefinition
// }>
export type Document<
  V extends Version = Version,
  T extends {} = {},
> = DocumentByVersion<T>[V];

export type Operation<T extends {} = {}> =
  | V3_0.OperationObject<T>
  | V3_1.OperationObject<T>
  | V3_2.OperationObject<T>;

export type Parameter =
  | V3_0.ReferenceObject
  | V3_0.ParameterObject
  | V3_1.ReferenceObject
  | V3_1.ParameterObject
  | V3_2.ReferenceObject
  | V3_2.ParameterObject;

export type Parameters =
  | (V3_0.ReferenceObject | V3_0.ParameterObject)[]
  | (V3_1.ReferenceObject | V3_1.ParameterObject)[]
  | (V3_2.ReferenceObject | V3_2.ParameterObject)[];

export interface Request {
  body?: any;
  headers?: object;
  params?: object;
  query?: object;
}


type PathsWebhooksComponents<T extends {} = {}> = {
  paths: PathsObject<T>;
  webhooks: Record<string, PathItemObject | ReferenceObject>;
  components: ComponentsObject;
};

interface DocumentBase {
  openapi: `3.1.${number}`;
  info: InfoObject;
  jsonSchemaDialect?: string;
  servers?: ServerObject[];
  security?: SecurityRequirementObject[];
  tags?: TagObject[];
  externalDocs?: ExternalDocumentationObject;
  'x-express-openapi-additional-middleware'?: (
    | ((request: any, response: any, next: any) => Promise<void>)
    | ((request: any, response: any, next: any) => void)
  )[];
  'x-express-openapi-validation-strict'?: boolean;
}

export type Document<T extends {} = {}> = DocumentBase &
  (
    | (Pick<PathsWebhooksComponents<T>, 'paths'> &
        Omit<Partial<PathsWebhooksComponents<T>>, 'paths'>)
    | (Pick<PathsWebhooksComponents<T>, 'webhooks'> &
        Omit<Partial<PathsWebhooksComponents<T>>, 'webhooks'>)
    | (Pick<PathsWebhooksComponents<T>, 'components'> &
        Omit<Partial<PathsWebhooksComponents<T>>, 'components'>)
  );

export interface InfoObject {
  title: string;
  summary?: string;
  description?: string;
  termsOfService?: string;
  contact?: ContactObject;
  license?: LicenseObject;
  version: string;
}

export interface ContactObject {
  name?: string;
  url?: string;
  email?: string;
}

export interface LicenseObject {
  name: string;
  identifier?: string;
  url?: string;
}

export interface ServerObject {
  url: string;
  description?: string;
  variables?: Record<string, ServerVariableObject>;
}

export interface ServerVariableObject {
  enum?: [string, ...string[]];
  default: string | number;
  description?: string;
}

export type PathsObject<T extends {} = {}, P extends {} = {}> = Record<
  string,
  (PathItemObject<T> & P) | undefined
>;

export enum HttpMethods {
  GET = 'get',
  PUT = 'put',
  POST = 'post',
  DELETE = 'delete',
  OPTIONS = 'options',
  HEAD = 'head',
  PATCH = 'patch',
  TRACE = 'trace',
}

export type PathItemObject<T extends {} = {}> = {
  $ref?: string;
  summary?: string;
  description?: string;
  servers?: ServerObject[];
  parameters?: (ReferenceObject | ParameterObject)[];
} & {
  [method in HttpMethods]?: OperationObject<T>;
};

export type OperationObject<T extends {} = {}> = {
  tags?: string[];
  summary?: string;
  description?: string;
  externalDocs?: ExternalDocumentationObject;
  operationId?: string;
  parameters?: (ReferenceObject | ParameterObject)[];
  requestBody?: ReferenceObject | RequestBodyObject;
  responses?: ResponsesObject;
  callbacks?: Record<string, ReferenceObject | CallbackObject>;
  deprecated?: boolean;
  security?: SecurityRequirementObject[];
  servers?: ServerObject[];
} & T;

export interface ExternalDocumentationObject {
  description?: string;
  url: string;
}

export interface ParameterObject extends ParameterBaseObject {
  name: string;
  in: string;
}

export interface HeaderObject extends ParameterBaseObject {}

export interface ParameterBaseObject {
  description?: string;
  required?: boolean;
  deprecated?: boolean;
  allowEmptyValue?: boolean;
  style?: string;
  explode?: boolean;
  allowReserved?: boolean;
  schema?: ReferenceObject | SchemaObject;
  example?: any;
  examples?: Record<string, ReferenceObject | ExampleObject>;
  content?: Record<string, MediaTypeObject>;
}

export type NonArraySchemaObjectType =
  | 'boolean'
  | 'object'
  | 'number'
  | 'string'
  | 'integer'
  | 'null';

export type ArraySchemaObjectType = 'array';

/**
 * There is no way to tell TypeScript to require items when type is either
 * 'array' or an array containing 'array'. Casting a schema object to
 * ArraySchemaObject or NonArraySchemaObject will expose the precise shape.
 */
export type SchemaObject =
  | ArraySchemaObject
  | NonArraySchemaObject
  | MixedSchemaObject
  | boolean;

export interface ArraySchemaObject extends BaseSchemaObject {
  type: ArraySchemaObjectType;
  items: ReferenceObject | SchemaObject;
}

export interface NonArraySchemaObject extends BaseSchemaObject {
  type?: NonArraySchemaObjectType;
}

interface MixedSchemaObject extends BaseSchemaObject {
  type?: (ArraySchemaObjectType | NonArraySchemaObjectType)[];
  items?: ReferenceObject | SchemaObject;
}

export interface BaseSchemaObject {
  title?: string;
  description?: string;
  format?: string;
  default?: any;
  multipleOf?: number;
  maximum?: number;
  exclusiveMaximum?: boolean | number;
  minimum?: number;
  exclusiveMinimum?: boolean | number;
  maxLength?: number;
  minLength?: number;
  pattern?: string;
  additionalProperties?: boolean | ReferenceObject | SchemaObject;
  maxItems?: number;
  minItems?: number;
  uniqueItems?: boolean;
  maxProperties?: number;
  minProperties?: number;
  required?: string[];
  enum?: any[];
  properties?: Record<string, ReferenceObject | SchemaObject>;
  allOf?: (ReferenceObject | SchemaObject)[];
  oneOf?: (ReferenceObject | SchemaObject)[];
  anyOf?: (ReferenceObject | SchemaObject)[];
  not?: ReferenceObject | SchemaObject;
  discriminator?: DiscriminatorObject;
  readOnly?: boolean;
  writeOnly?: boolean;
  xml?: XMLObject;
  externalDocs?: ExternalDocumentationObject;
  example?: any;
  examples?: any[];
  deprecated?: boolean;
  contentMediaType?: string;
  $schema?: string;
  const?: any;
}

export interface DiscriminatorObject {
  propertyName: string;
  mapping?: Record<string, string>;
}

export interface XMLObject {
  name?: string;
  namespace?: string;
  prefix?: string;
  attribute?: boolean;
  wrapped?: boolean;
}

export interface ReferenceObject {
  $ref: string;
  summary?: string;
  description?: string;
}

export interface ExampleObject {
  summary?: string;
  description?: string;
  value?: any;
  externalValue?: string;
}

export interface MediaTypeObject {
  schema?: SchemaObject | ReferenceObject;
  example?: any;
  examples?: Record<string, ReferenceObject | ExampleObject>;
  encoding?: Record<string, EncodingObject>;
}

export interface EncodingObject {
  contentType?: string;
  headers?: Record<string, ReferenceObject | HeaderObject>;
  style?: string;
  explode?: boolean;
  allowReserved?: boolean;
}

export interface RequestBodyObject {
  description?: string;
  content: Record<string, MediaTypeObject>;
  required?: boolean;
}

export type ResponsesObject = Record<
  string,
  ReferenceObject | ResponseObject
>;

export interface ResponseObject {
  description: string;
  headers?: Record<string, ReferenceObject | HeaderObject>;
  content?: Record<string, MediaTypeObject>;
  links?: Record<string, ReferenceObject | LinkObject>;
}

export interface LinkObject {
  operationRef?: string;
  operationId?: string;
  parameters?: Record<string, any>;
  requestBody?: any;
  description?: string;
  server?: ServerObject;
}

export type CallbackObject = Record<string, PathItemObject | ReferenceObject>;

export type SecurityRequirementObject = Record<string, string[]>;

export interface ComponentsObject {
  schemas?: Record<string, SchemaObject>;
  responses?: Record<string, ReferenceObject | ResponseObject>;
  parameters?: Record<string, ReferenceObject | ParameterObject>;
  examples?: Record<string, ReferenceObject | ExampleObject>;
  requestBodies?: Record<string, ReferenceObject | RequestBodyObject>;
  headers?: Record<string, ReferenceObject | HeaderObject>;
  securitySchemes?: Record<string, ReferenceObject | SecuritySchemeObject>;
  links?: Record<string, ReferenceObject | LinkObject>;
  callbacks?: Record<string, ReferenceObject | CallbackObject>;
  pathItems?: Record<string, ReferenceObject | PathItemObject>;
}

export type SecuritySchemeObject =
  | HttpSecurityScheme
  | ApiKeySecurityScheme
  | OAuth2SecurityScheme
  | OpenIdSecurityScheme;

export interface HttpSecurityScheme {
  type: 'http';
  description?: string;
  scheme: string;
  bearerFormat?: string;
}

export interface ApiKeySecurityScheme {
  type: 'apiKey';
  description?: string;
  name: string;
  in: string;
}

export interface OAuth2SecurityScheme {
  type: 'oauth2';
  description?: string;
  flows: {
    implicit?: {
      authorizationUrl: string;
      refreshUrl?: string;
      scopes: Record<string, string>;
    };
    password?: {
      tokenUrl: string;
      refreshUrl?: string;
      scopes: Record<string, string>;
    };
    clientCredentials?: {
      tokenUrl: string;
      refreshUrl?: string;
      scopes: Record<string, string>;
    };
    authorizationCode?: {
      authorizationUrl: string;
      tokenUrl: string;
      refreshUrl?: string;
      scopes: Record<string, string>;
    };
  };
}

export interface OpenIdSecurityScheme {
  type: 'openIdConnect';
  description?: string;
  openIdConnectUrl: string;
}

export interface TagObject {
  name: string;
  description?: string;
  externalDocs?: ExternalDocumentationObject;
}

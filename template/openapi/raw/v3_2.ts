// OpenAPI 3.2 declarations follow the normative specification at:
// https://spec.openapis.org/oas/v3.2.0.html

type PathsWebhooksComponents<T extends {} = {}> = {
  paths: PathsObject<T>;
  webhooks: Record<string, PathItemObject<T> | ReferenceObject>;
  components: ComponentsObject;
};

interface DocumentBase {
  openapi: `3.2.${number}`;
  $self?: string;
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
  name?: string;
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
  QUERY = 'query',
}

export type PathItemObject<T extends {} = {}> = {
  $ref?: string;
  summary?: string;
  description?: string;
  servers?: ServerObject[];
  parameters?: (ReferenceObject | ParameterObject)[];
  additionalOperations?: Record<string, OperationObject<T>>;
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
  responses: ResponsesObject;
  callbacks?: Record<string, ReferenceObject | CallbackObject>;
  deprecated?: boolean;
  security?: SecurityRequirementObject[];
  servers?: ServerObject[];
} & T;

export interface ExternalDocumentationObject {
  description?: string;
  url: string;
}

export type ParameterLocation =
  | 'query'
  | 'querystring'
  | 'header'
  | 'path'
  | 'cookie';

export interface ParameterObject extends ParameterBaseObject {
  name: string;
  in: ParameterLocation;
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
  content?: Record<string, ReferenceObject | MediaTypeObject>;
}

export type NonArraySchemaObjectType =
  | 'boolean'
  | 'object'
  | 'number'
  | 'string'
  | 'integer'
  | 'null';

export type ArraySchemaObjectType = 'array';

export type SchemaObject =
  | ArraySchemaObject
  | NonArraySchemaObject
  | MixedSchemaObject
  | boolean;

export interface ArraySchemaObject extends BaseSchemaObject {
  type: ArraySchemaObjectType;
  items?: ReferenceObject | SchemaObject;
}

export interface NonArraySchemaObject extends BaseSchemaObject {
  type?: NonArraySchemaObjectType;
}

interface MixedSchemaObject extends BaseSchemaObject {
  type?: (ArraySchemaObjectType | NonArraySchemaObjectType)[];
  items?: ReferenceObject | SchemaObject;
}

export interface BaseSchemaObject {
  $id?: string;
  $anchor?: string;
  $dynamicAnchor?: string;
  $dynamicRef?: string;
  $comment?: string;
  $defs?: Record<string, ReferenceObject | SchemaObject>;
  $schema?: string;
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
  patternProperties?: Record<string, ReferenceObject | SchemaObject>;
  propertyNames?: ReferenceObject | SchemaObject;
  dependentSchemas?: Record<string, ReferenceObject | SchemaObject>;
  dependentRequired?: Record<string, string[]>;
  allOf?: (ReferenceObject | SchemaObject)[];
  oneOf?: (ReferenceObject | SchemaObject)[];
  anyOf?: (ReferenceObject | SchemaObject)[];
  not?: ReferenceObject | SchemaObject;
  if?: ReferenceObject | SchemaObject;
  then?: ReferenceObject | SchemaObject;
  else?: ReferenceObject | SchemaObject;
  prefixItems?: (ReferenceObject | SchemaObject)[];
  contains?: ReferenceObject | SchemaObject;
  minContains?: number;
  maxContains?: number;
  unevaluatedItems?: boolean | ReferenceObject | SchemaObject;
  unevaluatedProperties?: boolean | ReferenceObject | SchemaObject;
  discriminator?: DiscriminatorObject;
  readOnly?: boolean;
  writeOnly?: boolean;
  xml?: XMLObject;
  externalDocs?: ExternalDocumentationObject;
  example?: any;
  examples?: any[];
  deprecated?: boolean;
  contentMediaType?: string;
  contentEncoding?: string;
  contentSchema?: ReferenceObject | SchemaObject;
  const?: any;
}

export interface DiscriminatorObject {
  propertyName: string;
  mapping?: Record<string, string>;
  defaultMapping?: string;
}

export type XMLNodeType =
  | 'element'
  | 'attribute'
  | 'text'
  | 'cdata'
  | 'none';

export interface XMLObject {
  name?: string;
  namespace?: string;
  prefix?: string;
  attribute?: boolean;
  wrapped?: boolean;
  nodeType?: XMLNodeType;
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
  dataValue?: any;
  serializedValue?: string;
}

interface MediaTypeBaseObject {
  schema?: SchemaObject | ReferenceObject;
  itemSchema?: SchemaObject | ReferenceObject;
  example?: any;
  examples?: Record<string, ReferenceObject | ExampleObject>;
}

export type MediaTypeObject = MediaTypeBaseObject &
  (
    | {
        encoding?: Record<string, EncodingObject>;
        prefixEncoding?: never;
        itemEncoding?: never;
      }
    | {
        encoding?: never;
        prefixEncoding?: EncodingObject[];
        itemEncoding?: EncodingObject;
      }
  );

export interface EncodingObject {
  contentType?: string;
  headers?: Record<string, ReferenceObject | HeaderObject>;
  style?: string;
  explode?: boolean;
  allowReserved?: boolean;
  encoding?: Record<string, EncodingObject>;
  prefixEncoding?: EncodingObject[];
  itemEncoding?: EncodingObject;
}

export interface RequestBodyObject {
  description?: string;
  content: Record<string, ReferenceObject | MediaTypeObject>;
  required?: boolean;
}

export type ResponsesObject = Record<
  string,
  ReferenceObject | ResponseObject
>;

export interface ResponseObject {
  summary?: string;
  description?: string;
  headers?: Record<string, ReferenceObject | HeaderObject>;
  content?: Record<string, ReferenceObject | MediaTypeObject>;
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
  schemas?: Record<string, ReferenceObject | SchemaObject>;
  responses?: Record<string, ReferenceObject | ResponseObject>;
  parameters?: Record<string, ReferenceObject | ParameterObject>;
  examples?: Record<string, ReferenceObject | ExampleObject>;
  requestBodies?: Record<string, ReferenceObject | RequestBodyObject>;
  headers?: Record<string, ReferenceObject | HeaderObject>;
  securitySchemes?: Record<string, ReferenceObject | SecuritySchemeObject>;
  links?: Record<string, ReferenceObject | LinkObject>;
  callbacks?: Record<string, ReferenceObject | CallbackObject>;
  pathItems?: Record<string, ReferenceObject | PathItemObject>;
  mediaTypes?: Record<string, ReferenceObject | MediaTypeObject>;
}

interface SecuritySchemeBase {
  description?: string;
  deprecated?: boolean;
}

export type SecuritySchemeObject =
  | HttpSecurityScheme
  | ApiKeySecurityScheme
  | MutualTLSSecurityScheme
  | OAuth2SecurityScheme
  | OpenIdSecurityScheme;

export interface HttpSecurityScheme extends SecuritySchemeBase {
  type: 'http';
  scheme: string;
  bearerFormat?: string;
}

export interface ApiKeySecurityScheme extends SecuritySchemeBase {
  type: 'apiKey';
  name: string;
  in: 'query' | 'header' | 'cookie';
}

export interface MutualTLSSecurityScheme extends SecuritySchemeBase {
  type: 'mutualTLS';
}

export interface OAuth2SecurityScheme extends SecuritySchemeBase {
  type: 'oauth2';
  flows: OAuthFlowsObject;
  oauth2MetadataUrl?: string;
}

export interface OpenIdSecurityScheme extends SecuritySchemeBase {
  type: 'openIdConnect';
  openIdConnectUrl: string;
}

export interface OAuthFlowsObject {
  implicit?: OAuthImplicitFlow;
  password?: OAuthPasswordFlow;
  clientCredentials?: OAuthClientCredentialsFlow;
  authorizationCode?: OAuthAuthorizationCodeFlow;
  deviceAuthorization?: OAuthDeviceAuthorizationFlow;
}

interface OAuthFlowBase {
  refreshUrl?: string;
  scopes: Record<string, string>;
}

export interface OAuthImplicitFlow extends OAuthFlowBase {
  authorizationUrl: string;
}

export interface OAuthPasswordFlow extends OAuthFlowBase {
  tokenUrl: string;
}

export interface OAuthClientCredentialsFlow extends OAuthFlowBase {
  tokenUrl: string;
}

export interface OAuthAuthorizationCodeFlow extends OAuthFlowBase {
  authorizationUrl: string;
  tokenUrl: string;
}

export interface OAuthDeviceAuthorizationFlow extends OAuthFlowBase {
  deviceAuthorizationUrl: string;
  tokenUrl: string;
}

export interface TagObject {
  name: string;
  summary?: string;
  description?: string;
  parent?: string;
  kind?: string;
  externalDocs?: ExternalDocumentationObject;
}

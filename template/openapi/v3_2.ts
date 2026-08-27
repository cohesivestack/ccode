// OpenAPI 3.2 declarations follow the normative specification at:
// https://spec.openapis.org/oas/v3.2.0.html
import type * as V3_1 from "./v3_1";

type Modify<T, R> = Omit<T, keyof R> & R;

type PathsWebhooksComponents<T extends {} = {}> = {
  paths: PathsObject<T>;
  webhooks: Record<string, PathItemObject<T> | ReferenceObject>;
  components: ComponentsObject;
};

export type Document<T extends {} = {}> = Modify<
  Omit<
    V3_1.Document<T>,
    | 'openapi'
    | 'paths'
    | 'webhooks'
    | 'components'
    | 'servers'
    | 'tags'
  >,
  {
    openapi: `3.2.${number}`;
    $self?: string;
    info: InfoObject;
    jsonSchemaDialect?: string;
    servers?: ServerObject[];
    tags?: TagObject[];
  } & (
    | (Pick<PathsWebhooksComponents<T>, 'paths'> &
        Omit<Partial<PathsWebhooksComponents<T>>, 'paths'>)
    | (Pick<PathsWebhooksComponents<T>, 'webhooks'> &
        Omit<Partial<PathsWebhooksComponents<T>>, 'webhooks'>)
    | (Pick<PathsWebhooksComponents<T>, 'components'> &
        Omit<Partial<PathsWebhooksComponents<T>>, 'components'>)
  )
>;

export type InfoObject = V3_1.InfoObject;
export type ContactObject = V3_1.ContactObject;
export type LicenseObject = V3_1.LicenseObject;

export type ServerObject = Modify<
  V3_1.ServerObject,
  {
    name?: string;
    variables?: Record<string, ServerVariableObject>;
  }
>;

export type ServerVariableObject = V3_1.ServerVariableObject;

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

export type OperationObject<T extends {} = {}> = Modify<
  V3_1.OperationObject<T>,
  {
    parameters?: (ReferenceObject | ParameterObject)[];
    requestBody?: ReferenceObject | RequestBodyObject;
    responses: ResponsesObject;
    callbacks?: Record<string, ReferenceObject | CallbackObject>;
    servers?: ServerObject[];
  }
> &
  T;

export type ExternalDocumentationObject =
  V3_1.ExternalDocumentationObject;

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

export type ParameterBaseObject = Modify<
  V3_1.ParameterBaseObject,
  {
    schema?: ReferenceObject | SchemaObject;
    examples?: Record<string, ReferenceObject | ExampleObject>;
    content?: Record<string, ReferenceObject | MediaTypeObject>;
  }
>;

export type NonArraySchemaObjectType =
  V3_1.NonArraySchemaObjectType;
export type ArraySchemaObjectType = V3_1.ArraySchemaObjectType;

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

export type BaseSchemaObject = Modify<
  V3_1.BaseSchemaObject,
  {
    $id?: string;
    $anchor?: string;
    $dynamicAnchor?: string;
    $dynamicRef?: string;
    $comment?: string;
    $defs?: Record<string, ReferenceObject | SchemaObject>;
    additionalProperties?: boolean | ReferenceObject | SchemaObject;
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
    contentEncoding?: string;
    contentSchema?: ReferenceObject | SchemaObject;
    discriminator?: DiscriminatorObject;
    externalDocs?: ExternalDocumentationObject;
    xml?: XMLObject;
  }
>;

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

export type XMLObject = Modify<
  V3_1.XMLObject,
  {
    nodeType?: XMLNodeType;
  }
>;

export type ReferenceObject = V3_1.ReferenceObject;

export type ExampleObject = Modify<
  V3_1.ExampleObject,
  {
    dataValue?: any;
    serializedValue?: string;
  }
>;

type MediaTypeBaseObject = Modify<
  Omit<V3_1.MediaTypeObject, 'encoding'>,
  {
    schema?: SchemaObject | ReferenceObject;
    itemSchema?: SchemaObject | ReferenceObject;
    examples?: Record<string, ReferenceObject | ExampleObject>;
  }
>;

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

export type EncodingObject = Modify<
  V3_1.EncodingObject,
  {
    headers?: Record<string, ReferenceObject | HeaderObject>;
    encoding?: Record<string, EncodingObject>;
    prefixEncoding?: EncodingObject[];
    itemEncoding?: EncodingObject;
  }
>;

export type RequestBodyObject = Modify<
  V3_1.RequestBodyObject,
  {
    content: Record<string, ReferenceObject | MediaTypeObject>;
  }
>;

export type ResponsesObject = Record<
  string,
  ReferenceObject | ResponseObject
>;

export type ResponseObject = Modify<
  V3_1.ResponseObject,
  {
    summary?: string;
    description?: string;
    headers?: Record<string, ReferenceObject | HeaderObject>;
    content?: Record<string, ReferenceObject | MediaTypeObject>;
    links?: Record<string, ReferenceObject | LinkObject>;
  }
>;

export type LinkObject = Modify<
  V3_1.LinkObject,
  {
    server?: ServerObject;
  }
>;

export type CallbackObject = Record<
  string,
  PathItemObject | ReferenceObject
>;

export type SecurityRequirementObject = Record<string, string[]>;

export type ComponentsObject = Modify<
  V3_1.ComponentsObject,
  {
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
>;

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

export type TagObject = Modify<
  V3_1.TagObject,
  {
    summary?: string;
    parent?: string;
    kind?: string;
  }
>;


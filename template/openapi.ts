// Adapted from: https://github.com/kogosoftwarellc/open-api/blob/main/packages/openapi-types/index.ts
// OpenAPI v2 declarations have been removed; Cohesive Code supports OpenAPI v3+.

/* tslint:disable:no-namespace no-empty-interface */
export namespace OpenAPI {
  export type Version = '3.0' | '3.1' | '3.2';

  export interface DocumentByVersion<T extends {} = {}> {
    '3.0': OpenAPIV3.Document<T>;
    '3.1': OpenAPIV3_1.Document<T>;
    '3.2': OpenAPIV3_2.Document<T>;
  }

  // OpenAPI extensions can be declared using generics
  // e.g.:
  // OpenAPI.Document<'3.1', {
  //   'x-amazon-apigateway-integration': AWSAPITGatewayDefinition
  // }>
  export type Document<
    V extends Version = Version,
    T extends {} = {},
  > = DocumentByVersion<T>[V];

  export type Operation<T extends {} = {}> =
    | OpenAPIV3.OperationObject<T>
    | OpenAPIV3_1.OperationObject<T>
    | OpenAPIV3_2.OperationObject<T>;

  export type Parameter =
    | OpenAPIV3_2.ReferenceObject
    | OpenAPIV3_2.ParameterObject
    | OpenAPIV3_1.ReferenceObject
    | OpenAPIV3_1.ParameterObject
    | OpenAPIV3.ReferenceObject
    | OpenAPIV3.ParameterObject;

  export type Parameters =
    | (OpenAPIV3_2.ReferenceObject | OpenAPIV3_2.ParameterObject)[]
    | (OpenAPIV3_1.ReferenceObject | OpenAPIV3_1.ParameterObject)[]
    | (OpenAPIV3.ReferenceObject | OpenAPIV3.ParameterObject)[];

  export interface Request {
    body?: any;
    headers?: object;
    params?: object;
    query?: object;
  }
}

// OpenAPI 3.2 declarations follow the normative specification at:
// https://spec.openapis.org/oas/v3.2.0.html
export namespace OpenAPIV3_2 {
  type Modify<T, R> = Omit<T, keyof R> & R;

  type PathsWebhooksComponents<T extends {} = {}> = {
    paths: PathsObject<T>;
    webhooks: Record<string, PathItemObject<T> | ReferenceObject>;
    components: ComponentsObject;
  };

  export type Document<T extends {} = {}> = Modify<
    Omit<
      OpenAPIV3_1.Document<T>,
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

  export type InfoObject = OpenAPIV3_1.InfoObject;
  export type ContactObject = OpenAPIV3_1.ContactObject;
  export type LicenseObject = OpenAPIV3_1.LicenseObject;

  export type ServerObject = Modify<
    OpenAPIV3_1.ServerObject,
    {
      name?: string;
      variables?: Record<string, ServerVariableObject>;
    }
  >;

  export type ServerVariableObject = OpenAPIV3_1.ServerVariableObject;

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
    OpenAPIV3_1.OperationObject<T>,
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
    OpenAPIV3_1.ExternalDocumentationObject;

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
    OpenAPIV3_1.ParameterBaseObject,
    {
      schema?: ReferenceObject | SchemaObject;
      examples?: Record<string, ReferenceObject | ExampleObject>;
      content?: Record<string, ReferenceObject | MediaTypeObject>;
    }
  >;

  export type NonArraySchemaObjectType =
    OpenAPIV3_1.NonArraySchemaObjectType;
  export type ArraySchemaObjectType = OpenAPIV3_1.ArraySchemaObjectType;

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
    OpenAPIV3_1.BaseSchemaObject,
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
    OpenAPIV3_1.XMLObject,
    {
      nodeType?: XMLNodeType;
    }
  >;

  export type ReferenceObject = OpenAPIV3_1.ReferenceObject;

  export type ExampleObject = Modify<
    OpenAPIV3_1.ExampleObject,
    {
      dataValue?: any;
      serializedValue?: string;
    }
  >;

  type MediaTypeBaseObject = Modify<
    Omit<OpenAPIV3_1.MediaTypeObject, 'encoding'>,
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
    OpenAPIV3_1.EncodingObject,
    {
      headers?: Record<string, ReferenceObject | HeaderObject>;
      encoding?: Record<string, EncodingObject>;
      prefixEncoding?: EncodingObject[];
      itemEncoding?: EncodingObject;
    }
  >;

  export type RequestBodyObject = Modify<
    OpenAPIV3_1.RequestBodyObject,
    {
      content: Record<string, ReferenceObject | MediaTypeObject>;
    }
  >;

  export type ResponsesObject = Record<
    string,
    ReferenceObject | ResponseObject
  >;

  export type ResponseObject = Modify<
    OpenAPIV3_1.ResponseObject,
    {
      summary?: string;
      description?: string;
      headers?: Record<string, ReferenceObject | HeaderObject>;
      content?: Record<string, ReferenceObject | MediaTypeObject>;
      links?: Record<string, ReferenceObject | LinkObject>;
    }
  >;

  export type LinkObject = Modify<
    OpenAPIV3_1.LinkObject,
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
    OpenAPIV3_1.ComponentsObject,
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
    OpenAPIV3_1.TagObject,
    {
      summary?: string;
      parent?: string;
      kind?: string;
    }
  >;
}

export namespace OpenAPIV3_1 {
  type Modify<T, R> = Omit<T, keyof R> & R;

  type PathsWebhooksComponents<T extends {} = {}> = {
    paths: PathsObject<T>;
    webhooks: Record<string, PathItemObject | ReferenceObject>;
    components: ComponentsObject;
  };

  export type Document<T extends {} = {}> = Modify<
    Omit<OpenAPIV3.Document<T>, 'openapi' | 'paths' | 'components'>,
    {
      openapi: `3.1.${number}`;
      info: InfoObject;
      jsonSchemaDialect?: string;
      servers?: ServerObject[];
    } & (
      | (Pick<PathsWebhooksComponents<T>, 'paths'> &
          Omit<Partial<PathsWebhooksComponents<T>>, 'paths'>)
      | (Pick<PathsWebhooksComponents<T>, 'webhooks'> &
          Omit<Partial<PathsWebhooksComponents<T>>, 'webhooks'>)
      | (Pick<PathsWebhooksComponents<T>, 'components'> &
          Omit<Partial<PathsWebhooksComponents<T>>, 'components'>)
    )
  >;

  export type InfoObject = Modify<
    OpenAPIV3.InfoObject,
    {
      summary?: string;
      license?: LicenseObject;
    }
  >;

  export type ContactObject = OpenAPIV3.ContactObject;

  export type LicenseObject = Modify<
    OpenAPIV3.LicenseObject,
    {
      identifier?: string;
    }
  >;

  export type ServerObject = Modify<
    OpenAPIV3.ServerObject,
    {
      url: string;
      description?: string;
      variables?: Record<string, ServerVariableObject>;
    }
  >;

  export type ServerVariableObject = Modify<
    OpenAPIV3.ServerVariableObject,
    {
      enum?: [string, ...string[]];
    }
  >;

  export type PathsObject<T extends {} = {}, P extends {} = {}> = Record<
    string,
    (PathItemObject<T> & P) | undefined
  >;

  export type HttpMethods = OpenAPIV3.HttpMethods;

  export type PathItemObject<T extends {} = {}> = Modify<
    OpenAPIV3.PathItemObject<T>,
    {
      servers?: ServerObject[];
      parameters?: (ReferenceObject | ParameterObject)[];
    }
  > &
    {
      [method in HttpMethods]?: OperationObject<T>;
    };

  export type OperationObject<T extends {} = {}> = Modify<
    OpenAPIV3.OperationObject<T>,
    {
      parameters?: (ReferenceObject | ParameterObject)[];
      requestBody?: ReferenceObject | RequestBodyObject;
      responses?: ResponsesObject;
      callbacks?: Record<string, ReferenceObject | CallbackObject>;
      servers?: ServerObject[];
    }
  > &
    T;

  export type ExternalDocumentationObject = OpenAPIV3.ExternalDocumentationObject;

  export type ParameterObject = OpenAPIV3.ParameterObject;

  export type HeaderObject = OpenAPIV3.HeaderObject;

  export type ParameterBaseObject = OpenAPIV3.ParameterBaseObject;

  export type NonArraySchemaObjectType =
    | OpenAPIV3.NonArraySchemaObjectType
    | 'null';

  export type ArraySchemaObjectType = OpenAPIV3.ArraySchemaObjectType;

  /**
   * There is no way to tell typescript to require items when type is either 'array' or array containing 'array' type
   * 'items' will be always visible as optional
   * Casting schema object to ArraySchemaObject or NonArraySchemaObject will work fine
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

  export type BaseSchemaObject = Modify<
    Omit<OpenAPIV3.BaseSchemaObject, 'nullable'>,
    {
      examples?: OpenAPIV3.BaseSchemaObject['example'][];
      exclusiveMinimum?: boolean | number;
      exclusiveMaximum?: boolean | number;
      contentMediaType?: string;
      $schema?: string;
      additionalProperties?: boolean | ReferenceObject | SchemaObject;
      properties?: {
        [name: string]: ReferenceObject | SchemaObject;
      };
      allOf?: (ReferenceObject | SchemaObject)[];
      oneOf?: (ReferenceObject | SchemaObject)[];
      anyOf?: (ReferenceObject | SchemaObject)[];
      not?: ReferenceObject | SchemaObject;
      discriminator?: DiscriminatorObject;
      externalDocs?: ExternalDocumentationObject;
      xml?: XMLObject;
      const?: any;
    }
  >;

  export type DiscriminatorObject = OpenAPIV3.DiscriminatorObject;

  export type XMLObject = OpenAPIV3.XMLObject;

  export type ReferenceObject = Modify<
    OpenAPIV3.ReferenceObject,
    {
      summary?: string;
      description?: string;
    }
  >;

  export type ExampleObject = OpenAPIV3.ExampleObject;

  export type MediaTypeObject = Modify<
    OpenAPIV3.MediaTypeObject,
    {
      schema?: SchemaObject | ReferenceObject;
      examples?: Record<string, ReferenceObject | ExampleObject>;
    }
  >;

  export type EncodingObject = OpenAPIV3.EncodingObject;

  export type RequestBodyObject = Modify<
    OpenAPIV3.RequestBodyObject,
    {
      content: { [media: string]: MediaTypeObject };
    }
  >;

  export type ResponsesObject = Record<
    string,
    ReferenceObject | ResponseObject
  >;

  export type ResponseObject = Modify<
    OpenAPIV3.ResponseObject,
    {
      headers?: { [header: string]: ReferenceObject | HeaderObject };
      content?: { [media: string]: MediaTypeObject };
      links?: { [link: string]: ReferenceObject | LinkObject };
    }
  >;

  export type LinkObject = Modify<
    OpenAPIV3.LinkObject,
    {
      server?: ServerObject;
    }
  >;

  export type CallbackObject = Record<string, PathItemObject | ReferenceObject>;

  export type SecurityRequirementObject = OpenAPIV3.SecurityRequirementObject;

  export type ComponentsObject = Modify<
    OpenAPIV3.ComponentsObject,
    {
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
  >;

  export type SecuritySchemeObject = OpenAPIV3.SecuritySchemeObject;

  export type HttpSecurityScheme = OpenAPIV3.HttpSecurityScheme;

  export type ApiKeySecurityScheme = OpenAPIV3.ApiKeySecurityScheme;

  export type OAuth2SecurityScheme = OpenAPIV3.OAuth2SecurityScheme;

  export type OpenIdSecurityScheme = OpenAPIV3.OpenIdSecurityScheme;

  export type TagObject = OpenAPIV3.TagObject;
}

export namespace OpenAPIV3 {
  export interface Document<T extends {} = {}> {
    openapi: `3.0.${number}`;
    info: InfoObject;
    servers?: ServerObject[];
    paths: PathsObject<T>;
    components?: ComponentsObject;
    security?: SecurityRequirementObject[];
    tags?: TagObject[];
    externalDocs?: ExternalDocumentationObject;
    'x-express-openapi-additional-middleware'?: (
      | ((request: any, response: any, next: any) => Promise<void>)
      | ((request: any, response: any, next: any) => void)
    )[];
    'x-express-openapi-validation-strict'?: boolean;
  }

  export interface InfoObject {
    title: string;
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
    url?: string;
  }

  export interface ServerObject {
    url: string;
    description?: string;
    variables?: { [variable: string]: ServerVariableObject };
  }

  export interface ServerVariableObject {
    enum?: string[] | number[];
    default: string | number;
    description?: string;
  }

  export interface PathsObject<T extends {} = {}, P extends {} = {}> {
    [pattern: string]: (PathItemObject<T> & P) | undefined;
  }

  // All HTTP methods allowed by OpenAPI 3 spec
  // See https://swagger.io/specification/#path-item-object
  // You can use keys or values from it in TypeScript code like this:
  //     for (const method of Object.values(OpenAPIV3.HttpMethods)) { … }
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
    responses: ResponsesObject;
    callbacks?: { [callback: string]: ReferenceObject | CallbackObject };
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
    examples?: { [media: string]: ReferenceObject | ExampleObject };
    content?: { [media: string]: MediaTypeObject };
  }
  export type NonArraySchemaObjectType =
    | 'boolean'
    | 'object'
    | 'number'
    | 'string'
    | 'integer';
  export type ArraySchemaObjectType = 'array';
  export type SchemaObject = ArraySchemaObject | NonArraySchemaObject;

  export interface ArraySchemaObject extends BaseSchemaObject {
    type: ArraySchemaObjectType;
    items: ReferenceObject | SchemaObject;
  }

  export interface NonArraySchemaObject extends BaseSchemaObject {
    type?: NonArraySchemaObjectType;
  }

  export interface BaseSchemaObject {
    // JSON schema allowed properties, adjusted for OpenAPI
    title?: string;
    description?: string;
    format?: string;
    default?: any;
    multipleOf?: number;
    maximum?: number;
    exclusiveMaximum?: boolean;
    minimum?: number;
    exclusiveMinimum?: boolean;
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
    properties?: {
      [name: string]: ReferenceObject | SchemaObject;
    };
    allOf?: (ReferenceObject | SchemaObject)[];
    oneOf?: (ReferenceObject | SchemaObject)[];
    anyOf?: (ReferenceObject | SchemaObject)[];
    not?: ReferenceObject | SchemaObject;

    // OpenAPI-specific properties
    nullable?: boolean;
    discriminator?: DiscriminatorObject;
    readOnly?: boolean;
    writeOnly?: boolean;
    xml?: XMLObject;
    externalDocs?: ExternalDocumentationObject;
    example?: any;
    deprecated?: boolean;
  }

  export interface DiscriminatorObject {
    propertyName: string;
    mapping?: { [value: string]: string };
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
  }

  export interface ExampleObject {
    summary?: string;
    description?: string;
    value?: any;
    externalValue?: string;
  }

  export interface MediaTypeObject {
    schema?: ReferenceObject | SchemaObject;
    example?: any;
    examples?: { [media: string]: ReferenceObject | ExampleObject };
    encoding?: { [media: string]: EncodingObject };
  }

  export interface EncodingObject {
    contentType?: string;
    headers?: { [header: string]: ReferenceObject | HeaderObject };
    style?: string;
    explode?: boolean;
    allowReserved?: boolean;
  }

  export interface RequestBodyObject {
    description?: string;
    content: { [media: string]: MediaTypeObject };
    required?: boolean;
  }

  export interface ResponsesObject {
    [code: string]: ReferenceObject | ResponseObject;
  }

  export interface ResponseObject {
    description: string;
    headers?: { [header: string]: ReferenceObject | HeaderObject };
    content?: { [media: string]: MediaTypeObject };
    links?: { [link: string]: ReferenceObject | LinkObject };
  }

  export interface LinkObject {
    operationRef?: string;
    operationId?: string;
    parameters?: { [parameter: string]: any };
    requestBody?: any;
    description?: string;
    server?: ServerObject;
  }

  export interface CallbackObject {
    [url: string]: PathItemObject;
  }

  export interface SecurityRequirementObject {
    [name: string]: string[];
  }

  export interface ComponentsObject {
    schemas?: { [key: string]: ReferenceObject | SchemaObject };
    responses?: { [key: string]: ReferenceObject | ResponseObject };
    parameters?: { [key: string]: ReferenceObject | ParameterObject };
    examples?: { [key: string]: ReferenceObject | ExampleObject };
    requestBodies?: { [key: string]: ReferenceObject | RequestBodyObject };
    headers?: { [key: string]: ReferenceObject | HeaderObject };
    securitySchemes?: { [key: string]: ReferenceObject | SecuritySchemeObject };
    links?: { [key: string]: ReferenceObject | LinkObject };
    callbacks?: { [key: string]: ReferenceObject | CallbackObject };
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
        scopes: { [scope: string]: string };
      };
      password?: {
        tokenUrl: string;
        refreshUrl?: string;
        scopes: { [scope: string]: string };
      };
      clientCredentials?: {
        tokenUrl: string;
        refreshUrl?: string;
        scopes: { [scope: string]: string };
      };
      authorizationCode?: {
        authorizationUrl: string;
        tokenUrl: string;
        refreshUrl?: string;
        scopes: { [scope: string]: string };
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
}

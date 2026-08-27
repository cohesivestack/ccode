import type * as V3_0 from "./v3_0";

type Modify<T, R> = Omit<T, keyof R> & R;

type PathsWebhooksComponents<T extends {} = {}> = {
  paths: PathsObject<T>;
  webhooks: Record<string, PathItemObject | ReferenceObject>;
  components: ComponentsObject;
};

export type Document<T extends {} = {}> = Modify<
  Omit<V3_0.Document<T>, 'openapi' | 'paths' | 'components'>,
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
  V3_0.InfoObject,
  {
    summary?: string;
    license?: LicenseObject;
  }
>;

export type ContactObject = V3_0.ContactObject;

export type LicenseObject = Modify<
  V3_0.LicenseObject,
  {
    identifier?: string;
  }
>;

export type ServerObject = Modify<
  V3_0.ServerObject,
  {
    url: string;
    description?: string;
    variables?: Record<string, ServerVariableObject>;
  }
>;

export type ServerVariableObject = Modify<
  V3_0.ServerVariableObject,
  {
    enum?: [string, ...string[]];
  }
>;

export type PathsObject<T extends {} = {}, P extends {} = {}> = Record<
  string,
  (PathItemObject<T> & P) | undefined
>;

export type HttpMethods = V3_0.HttpMethods;

export type PathItemObject<T extends {} = {}> = Modify<
  V3_0.PathItemObject<T>,
  {
    servers?: ServerObject[];
    parameters?: (ReferenceObject | ParameterObject)[];
  }
> &
  {
    [method in HttpMethods]?: OperationObject<T>;
  };

export type OperationObject<T extends {} = {}> = Modify<
  V3_0.OperationObject<T>,
  {
    parameters?: (ReferenceObject | ParameterObject)[];
    requestBody?: ReferenceObject | RequestBodyObject;
    responses?: ResponsesObject;
    callbacks?: Record<string, ReferenceObject | CallbackObject>;
    servers?: ServerObject[];
  }
> &
  T;

export type ExternalDocumentationObject = V3_0.ExternalDocumentationObject;

export type ParameterObject = V3_0.ParameterObject;

export type HeaderObject = V3_0.HeaderObject;

export type ParameterBaseObject = V3_0.ParameterBaseObject;

export type NonArraySchemaObjectType =
  | V3_0.NonArraySchemaObjectType
  | 'null';

export type ArraySchemaObjectType = V3_0.ArraySchemaObjectType;

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
  Omit<V3_0.BaseSchemaObject, 'nullable'>,
  {
    examples?: V3_0.BaseSchemaObject['example'][];
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

export type DiscriminatorObject = V3_0.DiscriminatorObject;

export type XMLObject = V3_0.XMLObject;

export type ReferenceObject = Modify<
  V3_0.ReferenceObject,
  {
    summary?: string;
    description?: string;
  }
>;

export type ExampleObject = V3_0.ExampleObject;

export type MediaTypeObject = Modify<
  V3_0.MediaTypeObject,
  {
    schema?: SchemaObject | ReferenceObject;
    examples?: Record<string, ReferenceObject | ExampleObject>;
  }
>;

export type EncodingObject = V3_0.EncodingObject;

export type RequestBodyObject = Modify<
  V3_0.RequestBodyObject,
  {
    content: { [media: string]: MediaTypeObject };
  }
>;

export type ResponsesObject = Record<
  string,
  ReferenceObject | ResponseObject
>;

export type ResponseObject = Modify<
  V3_0.ResponseObject,
  {
    headers?: { [header: string]: ReferenceObject | HeaderObject };
    content?: { [media: string]: MediaTypeObject };
    links?: { [link: string]: ReferenceObject | LinkObject };
  }
>;

export type LinkObject = Modify<
  V3_0.LinkObject,
  {
    server?: ServerObject;
  }
>;

export type CallbackObject = Record<string, PathItemObject | ReferenceObject>;

export type SecurityRequirementObject = V3_0.SecurityRequirementObject;

export type ComponentsObject = Modify<
  V3_0.ComponentsObject,
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

export type SecuritySchemeObject = V3_0.SecuritySchemeObject;

export type HttpSecurityScheme = V3_0.HttpSecurityScheme;

export type ApiKeySecurityScheme = V3_0.ApiKeySecurityScheme;

export type OAuth2SecurityScheme = V3_0.OAuth2SecurityScheme;

export type OpenIdSecurityScheme = V3_0.OpenIdSecurityScheme;

export type TagObject = V3_0.TagObject;


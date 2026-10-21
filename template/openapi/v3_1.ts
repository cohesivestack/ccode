/** OpenAPI 3.1 ergonomic contracts. Instances are created by Context loaders in Go. */
import type * as Raw from "./raw/v3_1";

/** JSON data; null is preserved separately from an absent value. */
export type JsonValue = null | boolean | number | string | readonly JsonValue[] | JsonObject;
export type JsonObject = { readonly [key: string]: JsonValue };
/** Recursively read-only source declarations. */
export type DeepReadonly<T> = T extends object ? { readonly [K in keyof T]: DeepReadonly<T[K]> } : T;
export type ReadonlyOpenAPIDocument = DeepReadonly<Raw.Document>;
export type ParameterLocation = "path" | "query" | "header" | "cookie";
export type SchemaType = "object" | "array" | "string" | "integer" | "number" | "boolean" | "null";
export type ApiSchema = ApiBooleanSchema | ApiSchemaObject;
export type ApiSecurityScheme = ApiApiKeySecurityScheme | ApiHttpSecurityScheme | ApiOAuth2SecurityScheme | ApiOpenIdConnectSecurityScheme | ApiMutualTLSSecurityScheme;
export type ApiOAuthFlowKind = "implicit" | "password" | "clientCredentials" | "authorizationCode";

/** Location in an original parsed source document. */
export declare class ApiSource {
  private constructor();
  /** How the containing document was supplied. */
  readonly kind: "file" | "memory";
  /** Absolute file URI or stable URI for an in-memory document. */
  readonly documentUri: string;
  /** RFC 6901 JSON Pointer; empty string identifies the document root. */
  readonly pointer: string;
  /** One-based source line. */
  readonly line?: number;
  /** One-based source column. */
  readonly column?: number;
}

/** An explicit static reference and its provenance. */
export declare class ApiReference {
  protected constructor();
  /** Original reference text. */
  readonly uri: string;
  /** Absolute reference URI resolved against the applicable base URI. */
  readonly resolvedUri: string;
  /** Location of the original $ref field. */
  readonly source: ApiSource;
  /** Physical source location of the referenced declaration. */
  readonly targetSource: ApiSource;
}

/** Shared metadata within this specification version. */
export declare abstract class ApiElement<TRaw = JsonObject> {
  protected constructor();
  /** Stable identity within this loaded model. */
  readonly id: string;
  /** Original declaration location; absent for synthetic elements. */
  readonly source?: ApiSource;
  /** Original parsed declaration, before reference resolution and derived views. */
  readonly raw?: DeepReadonly<TRaw>;
  /** Declared x- extension fields. */
  readonly extensions: Readonly<Record<string, JsonValue>>;
  /** Reference at this declaration, if present. */
  readonly reference?: ApiReference;
}

/** Root graph for a single OpenAPI document. */
export declare class ApiModel extends ApiElement<Raw.Document> {
  /** Full specification version declared by the document. */
  readonly openapiVersion: `3.1.${number}`;
  /** API identity and descriptive metadata. */
  readonly info: ApiInfo;
  /** Location of the root source document. */
  readonly source: ApiSource;
  /** Original parsed root document, including unresolved reference declarations. */
  readonly raw: ReadonlyOpenAPIDocument;
  /** Scoped server declarations for ordinary routes, plus a synthetic default when needed. */
  readonly servers: readonly ApiServer[];
  /** Ordinary route templates in source order. */
  readonly paths: readonly ApiPath[];
  /** Ordinary operations in path/source order; excludes webhooks and callbacks. */
  readonly operations: readonly ApiOperation[];
  /** Named incoming webhooks; empty for OpenAPI 3.0. */
  readonly webhooks: readonly ApiWebhook[];
  /** Root components.schemas entries, including unused definitions. */
  readonly schemas: readonly NamedSchema[];
  /** Looks up an exact path template. */
  getPath(path: string): ApiPath | undefined;
  /** Looks up an operation; standard method names are case-insensitive. */
  getOperation(method: string, path: string): ApiOperation | undefined;
  /** Looks up an ordinary operation by its unique operationId. */
  getOperationById(operationId: string): ApiOperation | undefined;
  /** Looks up a webhook by its declared name. */
  getWebhook(name: string): ApiWebhook | undefined;
  /** Looks up a root component schema by name. */
  getSchema(name: string): NamedSchema | undefined;
}

/** API identity and descriptive metadata. */
export declare class ApiInfo extends ApiElement {
  /** API title. */
  readonly title: string;
  /** API version; distinct from the OpenAPI specification version. */
  readonly version: string;
  /** Short API summary. */
  readonly summary?: string;
  /** API description. */
  readonly description?: string;
  /** Terms of service URL. */
  readonly termsOfService?: string;
  /** Contact information. */
  readonly contact?: ApiContact;
  /** License information. */
  readonly license?: ApiLicense;
}

/** API contact information. */
export declare class ApiContact extends ApiElement {
  /** Contact name. */
  readonly name?: string;
  /** Contact URL. */
  readonly url?: string;
  /** Contact email. */
  readonly email?: string;
}

/** API license information. */
export declare class ApiLicense extends ApiElement {
  /** License name. */
  readonly name: string;
  /** License URL. */
  readonly url?: string;
  /** SPDX license expression. */
  readonly identifier?: string;
}

/** External documentation. */
export declare class ApiExternalDocs extends ApiElement {
  /** Documentation URL. */
  readonly url: string;
  /** Documentation description. */
  readonly description?: string;
}

/** A server declaration in its own scope. */
export declare class ApiServer extends ApiElement<Raw.ServerObject> {
  /** URL template; may be relative. */
  readonly url: string;
  /** Server description. */
  readonly description?: string;
  /** Template variables in source order. */
  readonly variables: readonly ApiServerVariable[];
  /** Scope where this server was declared. */
  readonly scope: "document" | "path" | "operation" | "implicit";
  /** Path bindings filtered to operations applicable to this server. */
  readonly paths: readonly ApiServerPath[];
  /** Applicable ordinary operations; shares instances with ApiModel.operations. */
  readonly operations: readonly ApiOperation[];
  /** Finds a URL variable. */
  getVariable(name: string): ApiServerVariable | undefined;
  /** Finds a scoped path binding. */
  getPath(path: string): ApiServerPath | undefined;
  /** Finds an applicable ordinary operation. */
  getOperation(method: string, path: string): ApiOperation | undefined;
  /** Finds an applicable operation by operationId. */
  getOperationById(operationId: string): ApiOperation | undefined;
  /** Substitutes defaults and overrides; rejects unknown variables and invalid enum values. Does not append a route path. */
  expandUrl(variables?: Readonly<Record<string, string>>): string;
}

/** One variable in a server URL template. */
export declare class ApiServerVariable extends ApiElement {
  /** Variable name. */
  readonly name: string;
  /** Default substitution value. */
  readonly defaultValue: string;
  /** Allowed values when explicitly constrained. */
  readonly enum?: readonly string[];
  /** Variable description. */
  readonly description?: string;
}

/** A route as seen from one server. */
export declare class ApiServerPath {
  private constructor();
  /** Owning server declaration. */
  readonly server: ApiServer;
  /** Shared original route model. */
  readonly pathItem: ApiPath;
  /** Route template. */
  readonly path: string;
  /** Only operations applicable to this server. */
  readonly operations: readonly ApiOperation[];
  /** Finds an operation within this binding. */
  getOperation(method: string): ApiOperation | undefined;
}

/** An ordinary route template. */
export declare class ApiPath extends ApiElement {
  /** Exact route template, including placeholders. */
  readonly path: string;
  /** Path summary. */
  readonly summary?: string;
  /** Path description. */
  readonly description?: string;
  /** Parameters declared at this path, before operation overrides. */
  readonly parameters: readonly ApiParameter[];
  /** Explicit path servers; undefined means no declaration. */
  readonly declaredServers?: readonly ApiServer[];
  /** Operations declared at this route. */
  readonly operations: readonly ApiOperation[];
  /** Looks up a method at this route. */
  getOperation(method: string): ApiOperation | undefined;
}

/** Behavior common to ordinary, webhook, and callback operations. */
export declare abstract class ApiOperationBase extends ApiElement {
  /** Standard methods are lowercase; additional method names preserve their case. */
  readonly method: string;
  /** Declared operation identifier. */
  readonly operationId?: string;
  /** Operation summary. */
  readonly summary?: string;
  /** Operation description. */
  readonly description?: string;
  /** Declared tag names. */
  readonly tags: readonly string[];
  /** Declared deprecation, defaulting to false. */
  readonly deprecated: boolean;
  /** External documentation. */
  readonly externalDocs?: ApiExternalDocs;
  /** Parameters declared directly on this operation. */
  readonly declaredParameters: readonly ApiParameter[];
  /** Effective parameters: operation declarations override path declarations by location and name. */
  readonly parameters: readonly ApiParameter[];
  /** Declared request body. */
  readonly requestBody?: ApiRequestBody;
  /** Explicit response declarations, including ranges and default. */
  readonly responses: readonly ApiResponse[];
  /** Callbacks owned by this operation. */
  readonly callbacks: readonly ApiCallback[];
  /** Explicit operation servers. */
  readonly declaredServers?: readonly ApiServer[];
  /** Explicit security alternatives; an empty array disables inherited security. */
  readonly declaredSecurity?: readonly ApiSecurityRequirement[];
  /** Effective security alternatives, inherited from the root if absent. */
  readonly security: readonly ApiSecurityRequirement[];
  /** Finds an effective parameter by exact location and name. */
  getParameter(location: ParameterLocation, name: string): ApiParameter | undefined;
  /** Finds an exact declared response key; does not infer range/default fallback. */
  getResponse(status: string): ApiResponse | undefined;
  /** Finds a declared callback. */
  getCallback(name: string): ApiCallback | undefined;
}

/** An ordinary operation addressable through an API server. */
export declare class ApiOperation extends ApiOperationBase {
  /** Route template. */
  readonly path: string;
  /** Owning route. */
  readonly pathItem: ApiPath;
  /** Effective operation servers, using operation/path/document precedence. */
  readonly servers: readonly ApiServer[];
}

/** A named webhook; its destination is supplied by the consumer. */
export declare class ApiWebhook extends ApiElement {
  /** Webhook name. */
  readonly name: string;
  /** Path summary. */
  readonly summary?: string;
  /** Path description. */
  readonly description?: string;
  /** Parameters declared at this path, before operation overrides. */
  readonly parameters: readonly ApiParameter[];
  /** Explicit path servers; undefined means no declaration. */
  readonly declaredServers?: readonly ApiServer[];
  /** Webhook operations. */
  readonly operations: readonly ApiWebhookOperation[];
  /** Finds a webhook method. */
  getOperation(method: string): ApiWebhookOperation | undefined;
}

/** An operation on a named webhook. */
export declare class ApiWebhookOperation extends ApiOperationBase {
  /** Owning webhook. */
  readonly webhook: ApiWebhook;
}

/** A callback declared by an operation. */
export declare class ApiCallback extends ApiElement {
  /** Callback name. */
  readonly name: string;
  /** Operation declaring this callback. */
  readonly owner: ApiOperationBase;
  /** Earlier view reached by a recursive callback reference; paths is empty at this boundary. */
  readonly recursiveTarget?: ApiCallback;
  /** Callback destinations expressed as runtime expressions; follow recursiveTarget at a recursive boundary. */
  readonly paths: readonly ApiCallbackPath[];
  /** Finds an exact callback expression. */
  getPath(expression: string): ApiCallbackPath | undefined;
}

/** A callback destination expressed at runtime. */
export declare class ApiCallbackPath extends ApiElement {
  /** Owning callback. */
  readonly callback: ApiCallback;
  /** Runtime expression identifying the callback URL. */
  readonly expression: string;
  /** Path summary. */
  readonly summary?: string;
  /** Path description. */
  readonly description?: string;
  /** Parameters declared at this path, before operation overrides. */
  readonly parameters: readonly ApiParameter[];
  /** Explicit path servers; undefined means no declaration. */
  readonly declaredServers?: readonly ApiServer[];
  /** Callback operations. */
  readonly operations: readonly ApiCallbackOperation[];
  /** Finds a callback method. */
  getOperation(method: string): ApiCallbackOperation | undefined;
}

/** An operation on a callback destination. */
export declare class ApiCallbackOperation extends ApiOperationBase {
  /** Owning callback destination. */
  readonly callbackPath: ApiCallbackPath;
}

/** A named request parameter. */
export declare class ApiParameter extends ApiElement {
  /** Parameter name. */
  readonly name: string;
  /** Wire location. */
  readonly location: ParameterLocation;
  /** Description. */
  readonly description?: string;
  /** Declared requirement, defaulting to false. */
  readonly required: boolean;
  /** Declared deprecation, defaulting to false. */
  readonly deprecated: boolean;
  /** Explicit serialization style. */
  readonly style?: string;
  /** Explicit explode setting. */
  readonly explode?: boolean;
  /** Declared schema when schema-based. */
  readonly schema?: ApiSchema;
  /** Media types when content-based. */
  readonly content: readonly ApiMediaType[];
  /** Explicit example; null is preserved. */
  readonly example?: JsonValue;
  /** Named example declarations. */
  readonly examples: readonly ApiExample[];
  /** Explicit reserved-character handling. */
  readonly allowReserved?: boolean;
  /** Explicit empty-value handling. */
  readonly allowEmptyValue?: boolean;
  /** Finds an exact declared media type. */
  getMediaType(name: string): ApiMediaType | undefined;
}

/** A named response or encoding header. */
export declare class ApiHeader extends ApiElement {
  /** Header name. */
  readonly name: string;
  /** Description. */
  readonly description?: string;
  /** Declared requirement, defaulting to false. */
  readonly required: boolean;
  /** Declared deprecation, defaulting to false. */
  readonly deprecated: boolean;
  /** Explicit serialization style. */
  readonly style?: string;
  /** Explicit explode setting. */
  readonly explode?: boolean;
  /** Declared schema when schema-based. */
  readonly schema?: ApiSchema;
  /** Media types when content-based. */
  readonly content: readonly ApiMediaType[];
  /** Explicit example; null is preserved. */
  readonly example?: JsonValue;
  /** Named example declarations. */
  readonly examples: readonly ApiExample[];
  /** Finds an exact declared media type. */
  getMediaType(name: string): ApiMediaType | undefined;
}

/** An operation request body. */
export declare class ApiRequestBody extends ApiElement {
  /** Body description. */
  readonly description?: string;
  /** Whether the body is required; defaults to false. */
  readonly required: boolean;
  /** Declared request media types. */
  readonly content: readonly ApiMediaType[];
  /** Finds an exact declared media type. */
  getMediaType(name: string): ApiMediaType | undefined;
}

/** One explicitly declared response. */
export declare class ApiResponse extends ApiElement {
  /** Exact response key, such as 200, 2XX, or default. */
  readonly status: string;
  /** Response description. */
  readonly description?: string;
  /** Declared response headers. */
  readonly headers: readonly ApiHeader[];
  /** Declared response media types. */
  readonly content: readonly ApiMediaType[];
  /** Declared response links. */
  readonly links: readonly ApiLink[];
  /** Finds a header case-insensitively. */
  getHeader(name: string): ApiHeader | undefined;
  /** Finds an exact declared media type. */
  getMediaType(name: string): ApiMediaType | undefined;
  /** Finds a named link. */
  getLink(name: string): ApiLink | undefined;
}

/** One media type declaration. */
export declare class ApiMediaType extends ApiElement {
  /** Declared media type key. */
  readonly name: string;
  /** Whole-body schema. */
  readonly schema?: ApiSchema;
  /** Explicit example. */
  readonly example?: JsonValue;
  /** Named examples. */
  readonly examples: readonly ApiExample[];
  /** Property-based encodings. */
  readonly encoding: readonly ApiNamedEncoding[];
  /** Finds a named example. */
  getExample(name: string): ApiExample | undefined;
  /** Finds the encoding for a property. */
  getEncoding(propertyName: string): ApiEncoding | undefined;
}

/** One named example declaration. */
export declare class ApiExample extends ApiElement {
  /** Example name. */
  readonly name: string;
  /** Example summary. */
  readonly summary?: string;
  /** Example description. */
  readonly description?: string;
  /** Embedded example value. */
  readonly value?: JsonValue;
  /** URL containing the example. */
  readonly externalValue?: string;
}

/** A property name paired with its encoding. */
export declare class ApiNamedEncoding {
  private constructor();
  /** Encoded property name. */
  readonly propertyName: string;
  /** Encoding declaration. */
  readonly encoding: ApiEncoding;
}

/** Serialization details for a body property or item. */
export declare class ApiEncoding extends ApiElement {
  /** Explicit content type. */
  readonly contentType?: string;
  /** Encoding headers. */
  readonly headers: readonly ApiHeader[];
  /** Explicit serialization style. */
  readonly style?: string;
  /** Explicit explode setting. */
  readonly explode?: boolean;
  /** Explicit reserved-character handling. */
  readonly allowReserved?: boolean;
}

/** A response link to another operation. */
export declare class ApiLink extends ApiElement {
  /** Link name. */
  readonly name: string;
  /** Link description. */
  readonly description?: string;
  /** Target operation identifier. */
  readonly operationId?: string;
  /** Target operation reference URI. */
  readonly operationRef?: string;
  /** Parameter values or runtime expressions. */
  readonly parameters: Readonly<Record<string, JsonValue>>;
  /** Body value or runtime expression. */
  readonly requestBody?: JsonValue;
  /** Explicit target server; excluded from ordinary server navigation. */
  readonly server?: ApiServer;
}

/** A named schema declaration. */
export declare class NamedSchema extends ApiElement<JsonValue> {
  /** Component name, definition name, or pattern expression. */
  readonly name: string;
  /** Schema at this declaration. */
  readonly schema: ApiSchema;
}

/** A static schema reference; cycles preserve object identity. */
export declare class ApiSchemaReference extends ApiReference {
  /** Referenced schema; sibling constraints remain on the referring schema. */
  readonly target: ApiSchema;
}

/** A schema accepting all or no values. */
export declare class ApiBooleanSchema extends ApiElement<boolean> {
  /** Discriminant for boolean schemas. */
  readonly kind: "boolean";
  /** True accepts everything; false accepts nothing. */
  readonly value: boolean;
}

/** Schema object preserving declared constraints and composition. */
export declare class ApiSchemaObject extends ApiElement {
  /** Discriminant for schema objects. */
  readonly kind: "schema";
  /** Static schema reference; target is separate from local constraints. */
  readonly reference?: ApiSchemaReference;
  /** Declared schema title. */
  readonly title?: string;
  /** Schema description. */
  readonly description?: string;
  /** Declared format annotation. */
  readonly format?: string;
  /** Explicit type declaration normalized to an array; no type inference or nullability expansion. */
  readonly types: readonly SchemaType[];
  /** Directly declared properties; references and compositions are not flattened. */
  readonly properties: readonly ApiProperty[];
  /** Directly declared required property names. */
  readonly required: readonly string[];
  /** Schema for array items. */
  readonly items?: ApiSchema;
  /** Explicit additional-properties constraint; undefined preserves absence. */
  readonly additionalProperties?: ApiSchema;
  /** Conjunctive subschemas. */
  readonly allOf: readonly ApiSchema[];
  /** Exactly-one subschemas. */
  readonly oneOf: readonly ApiSchema[];
  /** Alternative subschemas. */
  readonly anyOf: readonly ApiSchema[];
  /** Negated subschema. */
  readonly not?: ApiSchema;
  /** Explicit allowed values; absence differs from an empty enum. */
  readonly enum?: readonly JsonValue[];
  /** Default annotation. */
  readonly default?: JsonValue;
  /** Legacy singular example annotation. */
  readonly example?: JsonValue;
  /** Read-only annotation. */
  readonly readOnly?: boolean;
  /** Write-only annotation. */
  readonly writeOnly?: boolean;
  /** Deprecation annotation. */
  readonly deprecated?: boolean;
  /** Declared discriminator. */
  readonly discriminator?: ApiDiscriminator;
  /** XML serialization metadata. */
  readonly xml?: ApiXml;
  /** Schema documentation. */
  readonly externalDocs?: ApiExternalDocs;
  /** Explicit multipleOf constraint. */
  readonly multipleOf?: number;
  /** Explicit minimum constraint. */
  readonly minimum?: number;
  /** Explicit maximum constraint. */
  readonly maximum?: number;
  /** Explicit minLength constraint. */
  readonly minLength?: number;
  /** Explicit maxLength constraint. */
  readonly maxLength?: number;
  /** Explicit minItems constraint. */
  readonly minItems?: number;
  /** Explicit maxItems constraint. */
  readonly maxItems?: number;
  /** Explicit minProperties constraint. */
  readonly minProperties?: number;
  /** Explicit maxProperties constraint. */
  readonly maxProperties?: number;
  /** Version-specific exclusive lower-bound constraint. */
  readonly exclusiveMinimum?: number;
  /** Version-specific exclusive upper-bound constraint. */
  readonly exclusiveMaximum?: number;
  /** Declared string pattern. */
  readonly pattern?: string;
  /** Whether array values must be unique. */
  readonly uniqueItems?: boolean;
  /** Effective schema dialect, respecting jsonSchemaDialect and $schema. */
  readonly dialect?: string;
  /** Declared $id. */
  readonly schemaId?: string;
  /** Declared $anchor. */
  readonly anchor?: string;
  /** Declared $dynamicAnchor. */
  readonly dynamicAnchor?: string;
  /** Original $dynamicRef; context-dependent targets are not reduced to a static target. */
  readonly dynamicRef?: string;
  /** $defs entries. */
  readonly definitions: readonly NamedSchema[];
  /** Pattern-indexed property schemas. */
  readonly patternProperties: readonly NamedSchema[];
  /** Property-dependent schemas. */
  readonly dependentSchemas: readonly NamedSchema[];
  /** Property-dependent required names. */
  readonly dependentRequired: Readonly<Record<string, readonly string[]>>;
  /** Position-specific item schemas. */
  readonly prefixItems: readonly ApiSchema[];
  /** Declared examples. */
  readonly examples: readonly JsonValue[];
  /** Explicit constant constraint. */
  readonly const?: JsonValue;
  /** Declared propertyNames subschema. */
  readonly propertyNames?: ApiSchema;
  /** Declared unevaluatedProperties subschema. */
  readonly unevaluatedProperties?: ApiSchema;
  /** Declared unevaluatedItems subschema. */
  readonly unevaluatedItems?: ApiSchema;
  /** Declared contains subschema. */
  readonly contains?: ApiSchema;
  /** Declared if subschema. */
  readonly if?: ApiSchema;
  /** Declared then subschema. */
  readonly then?: ApiSchema;
  /** Declared else subschema. */
  readonly else?: ApiSchema;
  /** Declared contentSchema subschema. */
  readonly contentSchema?: ApiSchema;
  /** Declared minContains constraint. */
  readonly minContains?: number;
  /** Declared maxContains constraint. */
  readonly maxContains?: number;
  /** Declared contentEncoding annotation. */
  readonly contentEncoding?: string;
  /** Declared contentMediaType annotation. */
  readonly contentMediaType?: string;
  /** Finds a directly declared property. */
  getProperty(name: string): ApiProperty | undefined;
  /** Finds a directly declared $defs entry. */
  getDefinition(name: string): NamedSchema | undefined;
}

/** A directly declared schema property. */
export declare class ApiProperty extends ApiElement<JsonValue> {
  /** Property name. */
  readonly name: string;
  /** Schema declaring this property. */
  readonly parent: ApiSchemaObject;
  /** Property schema. */
  readonly schema: ApiSchema;
  /** Membership in the declaring schema required array; excludes conditional or composed requirements. */
  readonly required: boolean;
}

/** Declared discriminator metadata. */
export declare class ApiDiscriminator extends ApiElement {
  /** Discriminator property. */
  readonly propertyName: string;
  /** Explicit value-to-schema mappings. */
  readonly mapping: Readonly<Record<string, string>>;
}

/** Declared XML serialization metadata. */
export declare class ApiXml extends ApiElement {
  /** Declared name. */
  readonly name?: string;
  /** Declared namespace. */
  readonly namespace?: string;
  /** Declared prefix. */
  readonly prefix?: string;
  /** Declared attribute. */
  readonly attribute?: boolean;
  /** Declared wrapped. */
  readonly wrapped?: boolean;
}

/** One security alternative; its schemes must all be satisfied. */
export declare class ApiSecurityRequirement extends ApiElement {
  /** AND requirements; an empty array represents anonymous access. */
  readonly schemes: readonly ApiSecurityUse[];
}

/** A security scheme used by a requirement. */
export declare class ApiSecurityUse {
  private constructor();
  /** Referenced security component name. */
  readonly name: string;
  /** Resolved scheme definition. */
  readonly scheme: ApiSecurityScheme;
  /** Required scopes or roles. */
  readonly scopes: readonly string[];
}

/** Metadata common to security schemes. */
export declare abstract class ApiSecuritySchemeBase extends ApiElement {
  /** Security component name. */
  readonly name: string;
  /** Scheme description. */
  readonly description?: string;
}

/** API key authentication. */
export declare class ApiApiKeySecurityScheme extends ApiSecuritySchemeBase {
  /** Scheme discriminant. */
  readonly type: "apiKey";
  /** API key location. */
  readonly location: "query" | "header" | "cookie";
  /** Actual wire name, distinct from the component name. */
  readonly parameterName: string;
}

/** HTTP authentication. */
export declare class ApiHttpSecurityScheme extends ApiSecuritySchemeBase {
  /** Scheme discriminant. */
  readonly type: "http";
  /** HTTP authentication scheme. */
  readonly scheme: string;
  /** Bearer token format hint. */
  readonly bearerFormat?: string;
}

/** Mutual TLS authentication. */
export declare class ApiMutualTLSSecurityScheme extends ApiSecuritySchemeBase {
  /** Scheme discriminant. */
  readonly type: "mutualTLS";
}

/** OAuth 2 authentication. */
export declare class ApiOAuth2SecurityScheme extends ApiSecuritySchemeBase {
  /** Scheme discriminant. */
  readonly type: "oauth2";
  /** Declared OAuth flows. */
  readonly flows: readonly ApiOAuthFlow[];
  /** Finds a declared OAuth flow. */
  getFlow(kind: ApiOAuthFlowKind): ApiOAuthFlow | undefined;
}

/** OpenID Connect authentication. */
export declare class ApiOpenIdConnectSecurityScheme extends ApiSecuritySchemeBase {
  /** Scheme discriminant. */
  readonly type: "openIdConnect";
  /** Discovery URL. */
  readonly openIdConnectUrl: string;
}

/** One OAuth flow declaration. */
export declare class ApiOAuthFlow extends ApiElement {
  /** Flow kind. */
  readonly kind: ApiOAuthFlowKind;
  /** Authorization endpoint. */
  readonly authorizationUrl?: string;
  /** Token endpoint. */
  readonly tokenUrl?: string;
  /** Refresh endpoint. */
  readonly refreshUrl?: string;
  /** Available scopes and descriptions. */
  readonly scopes: Readonly<Record<string, string>>;
}

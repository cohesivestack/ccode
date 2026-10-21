// Type contract checks. Run tsc --noEmit --strict --target es2020 --moduleResolution node on this file.
import type { Context } from "../context";
import * as OpenAPI from "../openapi";
import type * as OpenAPIRaw from "../openapi/raw";

declare const ctx: Context;
const api30: OpenAPI.V3_0.ApiModel = ctx.loadOpenAPIFromFile("api.yaml", { expectedVersion: "3.0" });
const api31: OpenAPI.V3_1.ApiModel = ctx.loadOpenAPIFromString("...", { expectedVersion: "3.1" });
const api32: OpenAPI.V3_2.ApiModel = ctx.loadOpenAPIFromBytes([], { expectedVersion: "3.2" });
const anyVersion: OpenAPI.ApiModel = ctx.loadOpenAPIFromFile("api.yaml");
const raw31: OpenAPIRaw.V3_1.Document = ctx.parseOpenAPIFromFile("api.yaml", { expectedVersion: "3.1" });
const original31: OpenAPI.V3_1.DeepReadonly<OpenAPIRaw.V3_1.Document> = api31.raw;
void [anyVersion, raw31, original31];

for (const server of api31.servers) {
  server.expandUrl({ region: "eu" });
  const operation = server.getPath("/users/{id}")?.getOperation("GET");
  operation?.getParameter("path", "id");
  operation?.getResponse("200")?.getMediaType("application/json");
  // @ts-expect-error Collections are immutable.
  server.operations.push(operation);
}

const legacy = api30.getSchema("User")?.schema;
if (legacy?.kind === "schema") {
  const nullable: boolean | undefined = legacy.nullable;
  const bound: boolean | undefined = legacy.exclusiveMinimum;
  void [nullable, bound];
  // @ts-expect-error Modern schema dialect metadata is absent in the 3.0 model.
  legacy.dialect;
}
const modern = api31.getSchema("User")?.schema;
if (modern?.kind === "schema") {
  const bound: number | undefined = modern.exclusiveMinimum;
  modern.getDefinition("Nested");
  void bound;
  // @ts-expect-error Nullable is an OpenAPI 3.0 keyword.
  modern.nullable;
}
api32.operations[0]?.getParameter("querystring", "filter");
api32.operations[0]?.getResponse("200")?.getMediaType("application/jsonl")?.itemSchema;
api32.servers[0]?.name;
// @ts-expect-error Querystring parameters were added in 3.2.
api31.operations[0]?.getParameter("querystring", "filter");
// @ts-expect-error Item schemas were added in 3.2.
api31.operations[0]?.getResponse("200")?.getMediaType("application/jsonl")?.itemSchema;
// @ts-expect-error The server name field was added in 3.2.
api31.servers[0]?.name;
// @ts-expect-error Expected versions are restricted to supported minor versions.
ctx.loadOpenAPIFromFile("api.yaml", { expectedVersion: "3.3" });
// @ts-expect-error Version-specific documents are not interchangeable.
const wrong: OpenAPI.V3_2.ApiModel = api30;
// @ts-expect-error Raw source declarations are deeply immutable.
api31.raw.info.title = "changed";
// @ts-expect-error Native model classes cannot be instantiated in TypeScript.
new OpenAPI.V3_1.ApiModel();
void wrong;

import type * as Database from "./database";
import type * as OpenAPI from "./openapi";
import type * as OpenAPIRaw from "./openapi/raw";

export interface Context {
  println: (message: string) => void;
  runProcess: (process: string) => void;
  renderTemplate: (templatePath: string, data: any) => string;
  generate: (templatePath: string, filePath: string, data: any) => void;
  accelerate: (
    id: string,
    templatePath: string,
    data: any,
    guidePath?: string,
  ) => void;
  /** Removes matching files under output_path, preserving artifacts emitted so far in this run. */
  prune: (pathPatterns: string[]) => void;
  parseJSONFromBytes: (jsonBytes: number[]) => Record<string, any>;
  parseJSONFromString: (jsonString: string) => Record<string, any>;
  parseJSONFromFile: (filePath: string) => Record<string, any>;
  parseOpenAPIFromBytes: (specBytes: number[]) => OpenAPIRaw.Document;
  parseOpenAPIFromString: (spec: string) => OpenAPIRaw.Document;
  parseOpenAPIFromFile(filePath: string): OpenAPIRaw.Document;
  parseOpenAPIFromFile<V extends OpenAPIRaw.Version>(
    filePath: string,
    options: { expectedVersion: V },
  ): OpenAPIRaw.Document<V>;
  /** Loads a read-only ergonomic graph using the native Go implementation. */
  loadOpenAPIFromBytes(specBytes: number[]): OpenAPI.ApiModel;
  loadOpenAPIFromBytes<V extends OpenAPI.Version>(
    specBytes: number[],
    options: { expectedVersion: V },
  ): OpenAPI.ApiModel<V>;
  loadOpenAPIFromString(spec: string): OpenAPI.ApiModel;
  loadOpenAPIFromString<V extends OpenAPI.Version>(
    spec: string,
    options: { expectedVersion: V },
  ): OpenAPI.ApiModel<V>;
  loadOpenAPIFromFile(filePath: string): OpenAPI.ApiModel;
  loadOpenAPIFromFile<V extends OpenAPI.Version>(
    filePath: string,
    options: { expectedVersion: V },
  ): OpenAPI.ApiModel<V>;
  inspectDatabase(
    connectionURL: Database.PostgreSQL.ConnectionURL,
  ): Database.PostgreSQL.Inspection;
  inspectDatabase(
    connectionURL: Database.MySQL.ConnectionURL,
  ): Database.MySQL.Inspection;
  inspectDatabase(
    connectionURL: Database.MariaDB.ConnectionURL,
  ): Database.MariaDB.Inspection;
  inspectDatabase(
    connectionURL: Database.SQLite.ConnectionURL,
  ): Database.SQLite.Inspection;
  inspectDatabase<E extends Database.Engine>(
    connectionURL: string,
    options: Database.InspectOptions<E>,
  ): Database.InspectionByEngine[E];
  inspectDatabase(connectionURL: string): Database.Inspection;
}

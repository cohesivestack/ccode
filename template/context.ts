import type * as Database from "./database";
import type * as OpenAPI from "./openapi";

export interface Context {
  println: (message: string) => void;
  setScope: (scopeName: string) => void;
  scope: () => string;
  renderTemplate: (templatePath: string, data: any) => string;
  generate: (templatePath: string, filePath: string, data: any) => void;
  accelerate: (
    id: string,
    templatePath: string,
    data: any,
    instructionsPath?: string,
  ) => void;
  parseJSONFromBytes: (jsonBytes: number[]) => Record<string, any>;
  parseJSONFromString: (jsonString: string) => Record<string, any>;
  parseJSONFromFile: (filePath: string) => Record<string, any>;
  parseOpenAPIFromBytes: (specBytes: number[]) => OpenAPI.Document;
  parseOpenAPIFromString: (spec: string) => OpenAPI.Document;
  parseOpenAPIFromFile(filePath: string): OpenAPI.Document;
  parseOpenAPIFromFile<V extends OpenAPI.Version>(
    filePath: string,
    options: { expectedVersion: V },
  ): OpenAPI.Document<V>;
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

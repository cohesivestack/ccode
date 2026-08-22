import type {
  DatabaseEngine,
  DatabaseInspection,
  DatabaseInspectionByEngine,
  InspectDatabaseOptions,
  MariaDBConnectionURL,
  MariaDBInspection,
  MySQLConnectionURL,
  MySQLInspection,
  PostgreSQLConnectionURL,
  PostgreSQLInspection,
  SQLiteConnectionURL,
  SQLiteInspection,
} from "./database";
import type { OpenAPIDocument, OpenAPIVersion } from "./openapi";

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
  parseOpenAPIFromBytes: (specBytes: number[]) => OpenAPIDocument;
  parseOpenAPIFromString: (spec: string) => OpenAPIDocument;
  parseOpenAPIFromFile(filePath: string): OpenAPIDocument;
  parseOpenAPIFromFile<V extends OpenAPIVersion>(
    filePath: string,
    options: { expectedVersion: V },
  ): OpenAPIDocument<V>;
  inspectDatabase(
    connectionURL: PostgreSQLConnectionURL,
  ): PostgreSQLInspection;
  inspectDatabase(connectionURL: MySQLConnectionURL): MySQLInspection;
  inspectDatabase(connectionURL: MariaDBConnectionURL): MariaDBInspection;
  inspectDatabase(connectionURL: SQLiteConnectionURL): SQLiteInspection;
  inspectDatabase<Engine extends DatabaseEngine>(
    connectionURL: string,
    options: InspectDatabaseOptions<Engine>,
  ): DatabaseInspectionByEngine[Engine];
  inspectDatabase(connectionURL: string): DatabaseInspection;
}

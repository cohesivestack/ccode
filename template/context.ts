import type { OpenAPI } from "./openapi";

export type OpenAPIDocument = OpenAPI.Document;

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
  parseOpenAPIFromFile<V extends OpenAPI.Version>(
    filePath: string,
    options: { expectedVersion: V },
  ): OpenAPI.Document<V>;
}

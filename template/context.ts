import type { OpenAPIV3, OpenAPIV3_1 } from "./openapi";

export type OpenAPIDocument = OpenAPIV3.Document | OpenAPIV3_1.Document;

export interface Context {
  println: (message: string) => void;
  templateToString: (templatePath: string, data: any) => string;
  templateToFile: (templatePath: string, filePath: string, data: any) => void;
  parseJSONFromBytes: (jsonBytes: number[]) => Record<string, any>;
  parseJSONFromString: (jsonString: string) => Record<string, any>;
  parseJSONFromFile: (filePath: string) => Record<string, any>;
  parseOpenAPIFromBytes: (specBytes: number[]) => OpenAPIDocument;
  parseOpenAPIFromString: (spec: string) => OpenAPIDocument;
  parseOpenAPIFromFile: (filePath: string) => OpenAPIDocument;
}

export interface Context {
  println: (message: string) => void;
  templateToString: (templatePath: string, data: any) => string;
  templateToFile: (templatePath: string, filePath: string, data: any) => void;
  parseJSONFromBytes: (jsonBytes: number[]) => Record<string, any>;
  parseJSONFromString: (jsonString: string) => Record<string, any>;
  parseJSONFromFile: (filePath: string) => Record<string, any>;
  parseOpenAPIFromBytes: (specBytes: number[]) => Record<string, any>;
  parseOpenAPIFromString: (spec: string) => Record<string, any>;
  parseOpenAPIFromFile: (filePath: string) => Record<string, any>;
}

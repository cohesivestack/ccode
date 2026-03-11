export interface Context {
  println: (message: string) => void;
  templateToString: (templatePath: string, data: any) => string;
  templateToFile: (templatePath: string, filePath: string, data: any) => void;
  parseJSONFromBytes: (jsonBytes: number[]) => Record<string, any>;
  parseJSONFromString: (jsonString: string) => Record<string, any>;
  parseJSONFromFile: (filePath: string) => Record<string, any>;
}

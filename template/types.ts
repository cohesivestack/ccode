export interface Context {
  println: (message: string) => void;
  templateToString: (templatePath: string, data: any) => string;
  templateToFile: (templatePath: string, filePath: string, data: any) => void;
}

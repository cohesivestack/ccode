package ccode

// func RunOpenAPIProcess() {
// 	petstore, _ := os.ReadFile("petstorev3.yaml")
// 	document, err := libopenapi.NewDocument(petstore)
// 	if err != nil {
// 		panic(fmt.Sprintf("cannot create new document: %e", err))
// 	}
// 	docModel, err := document.BuildV3Model()
// 	if err != nil {
// 		panic(fmt.Sprintf("cannot create v3 model from document: %e", err))
// 	}

// 	x := docModel.Model.Paths.PathItems.

// 	// The following fails after the first iteration
// 	for schemaName, schema := range docModel.Model.Components.Schemas.FromOldest() {
// 		if schema.Schema().Properties != nil {
// 			fmt.Printf("Schema '%s' has %d properties\n", schemaName, schema.Schema().Properties.Len())
// 		}
// 	}
// }

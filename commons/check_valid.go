package commons

import (
	"encoding/json"
	"log"
		
	"github.com/ianlancetaylor/jsonschema"
	"github.com/ianlancetaylor/jsonschema/draft7"
)

func Validate(schema string, dataAny any) error {
	data1 := StructToMapRecursive(dataAny)
	data := BsonAnyToMap(data1)

	content := []byte(schema)

	var v any
	if err := json.Unmarshal(content, &v); err != nil {
		log.Printf("Failed to decode json %#v", err)

		return err
	}

	vldtor, err := jsonschema.SchemaFromJSON(draft7.SchemaID, nil, v)
	if err != nil {
		log.Printf("Failed to load schema %v", err)
		return err
	}

	// validate
	valid := vldtor.Validate(data)

	if valid != nil {
		log.Printf("Invalid json: %v", valid)
				
		return valid
	}

	return nil
}
package structful

import (
	"log"
	"fmt"
	"os"
	"encoding/json"
	"crypto/md5"
	
	"github.com/spf13/cast"
	
	"github.com/gsuntres/pkg/commons"
	
)

func SeedStructful(files []string) error {
	for _, file := range files {

		// 1. Read the file bytes
		content, err := os.ReadFile(file)
		if err != nil {
			log.Fatalf("Error reading file: %v", err)
		}

		// 2. Define the target map
		var data map[string]any

		// 3. Unmarshal (parse) the JSON into the map
		err = json.Unmarshal(content, &data)
		if err != nil {
			commons.JsonPrintError(err, content)
			
			return err
		}

		// 4. Check if composite
		isComposite := false
		isCompositeVal, ok := data["_composite"]
		if ok {
			isComposite = isCompositeVal.(bool)
		}

		var name string
		nameVal, nameOk := data["_name"]
		if nameOk {
			name = cast.ToString(nameVal)
		}

		var version string
		versionVal, versionOk := data["_version"]
		if versionOk {
			version = cast.ToString(versionVal)
		}

		if isComposite {
			// if composite extract entries
			entries := make([]any, 0)
			entriesVal, entriesOk := data["_entries"]
			if entriesOk {
				entries = entriesVal.([]any)

				for _, entry := range entries {
					reg := Current()

					d, err := cast.ToStringMapE(entry)
					if err != nil {
						log.Printf("failed to cast %v", err)

						continue
					}

					toHash, err := json.Marshal(d)
					if err != nil {
						log.Printf("failed to hash %v", err)

						continue
					}

					hashed := fmt.Sprintf("%x", md5.Sum(toHash))

					d["_hash"] = hashed
					d["_name"] = name
					d["_version"] = version

					if !reg.CheckHash(hashed) {
						err = reg.Save(d)
						if err != nil {
							log.Printf("failed to save %v", err)

							continue
						}
					}
				}
			}
		} else {
			toHash, err := json.Marshal(data)
			if err != nil {
				log.Printf("failed to hash %v", err)

				continue
			}

			hashed := fmt.Sprintf("%x", md5.Sum(toHash))

			// 5. Save it
			reg := Current()
			data["_hash"] = hashed
			if !reg.CheckHash(hashed) {
				err = reg.Save(data)
				if err != nil {
					return err
				}
			}
		}
	}

	return nil
}

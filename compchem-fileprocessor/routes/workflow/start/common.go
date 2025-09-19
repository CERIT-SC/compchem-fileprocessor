package start_workflow_route

import (
	"fmt"

	"fi.muni.cz/invenio-file-processor/v2/services"
)

func validateFiles(files []services.File, errors []string) {
	mimetype := files[0].Mimetype

	for index, file := range files {
		if file.FileName == "" {
			errors = append(errors, fmt.Sprintf("fileName-%d", index))
		}

		if file.Mimetype == "" || file.Mimetype != mimetype {
			errors = append(errors, fmt.Sprintf("mimetype-%d", index))
		}
	}
}

package argointegration

import (
	"fi.muni.cz/upload-processor/v2/config"
	"go.uber.org/zap"
)

func SubmitWorkflow(logger *zap.Logger, config config.ArgoApi, ) error {
	SUBMIT_TEMPLATE = "/api/v1/worfklows/%s"
  
  logger.Info("Submitting workflow", fields ...zap.Field)

}

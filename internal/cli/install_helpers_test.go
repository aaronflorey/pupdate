package cli

import (
	"github.com/aaronflorey/pupdate/internal/detection"
	"github.com/aaronflorey/pupdate/internal/install"
)

type managerPlan = install.Plan

func selectManagerPlan(result detection.DetectionResult, allowScripts bool) (install.Plan, bool, string) {
	return install.SelectPlan(result, allowScripts)
}

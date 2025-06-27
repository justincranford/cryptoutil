package businessmodel

import (
	"errors"
	"fmt"

	cryptoutilBusinessLogicModel "cryptoutil/internal/openapi/model"
)

var validTransitions = func() map[cryptoutilBusinessLogicModel.ElasticKeyStatus]map[cryptoutilBusinessLogicModel.ElasticKeyStatus]bool {
	transitions := map[cryptoutilBusinessLogicModel.ElasticKeyStatus][]cryptoutilBusinessLogicModel.ElasticKeyStatus{
		cryptoutilBusinessLogicModel.Creating:                       {cryptoutilBusinessLogicModel.PendingGenerate, cryptoutilBusinessLogicModel.PendingImport},
		cryptoutilBusinessLogicModel.ImportFailed:                   {cryptoutilBusinessLogicModel.PendingDeleteWasImportFailed, cryptoutilBusinessLogicModel.PendingImport},
		cryptoutilBusinessLogicModel.PendingImport:                  {cryptoutilBusinessLogicModel.PendingDeleteWasPendingImport, cryptoutilBusinessLogicModel.ImportFailed, cryptoutilBusinessLogicModel.Active},
		cryptoutilBusinessLogicModel.PendingGenerate:                {cryptoutilBusinessLogicModel.GenerateFailed, cryptoutilBusinessLogicModel.Active},
		cryptoutilBusinessLogicModel.GenerateFailed:                 {cryptoutilBusinessLogicModel.PendingDeleteWasGenerateFailed, cryptoutilBusinessLogicModel.PendingGenerate},
		cryptoutilBusinessLogicModel.Active:                         {cryptoutilBusinessLogicModel.PendingDeleteWasActive, cryptoutilBusinessLogicModel.Disabled},
		cryptoutilBusinessLogicModel.Disabled:                       {cryptoutilBusinessLogicModel.PendingDeleteWasDisabled, cryptoutilBusinessLogicModel.Active},
		cryptoutilBusinessLogicModel.PendingDeleteWasImportFailed:   {cryptoutilBusinessLogicModel.FinishedDelete, cryptoutilBusinessLogicModel.ImportFailed},
		cryptoutilBusinessLogicModel.PendingDeleteWasPendingImport:  {cryptoutilBusinessLogicModel.FinishedDelete, cryptoutilBusinessLogicModel.PendingImport},
		cryptoutilBusinessLogicModel.PendingDeleteWasActive:         {cryptoutilBusinessLogicModel.FinishedDelete, cryptoutilBusinessLogicModel.Active},
		cryptoutilBusinessLogicModel.PendingDeleteWasDisabled:       {cryptoutilBusinessLogicModel.FinishedDelete, cryptoutilBusinessLogicModel.Disabled},
		cryptoutilBusinessLogicModel.PendingDeleteWasGenerateFailed: {cryptoutilBusinessLogicModel.FinishedDelete, cryptoutilBusinessLogicModel.GenerateFailed},
		cryptoutilBusinessLogicModel.StartedDelete:                  {cryptoutilBusinessLogicModel.FinishedDelete},
		cryptoutilBusinessLogicModel.FinishedDelete:                 {},
	}
	convertedTransitions := make(map[cryptoutilBusinessLogicModel.ElasticKeyStatus]map[cryptoutilBusinessLogicModel.ElasticKeyStatus]bool)
	for current, nextStates := range transitions {
		convertedTransitions[current] = make(map[cryptoutilBusinessLogicModel.ElasticKeyStatus]bool)
		for _, next := range nextStates {
			convertedTransitions[current][next] = true
		}
	}
	return convertedTransitions
}()

func EkasticKeyStatusTransition(current cryptoutilBusinessLogicModel.ElasticKeyStatus, next cryptoutilBusinessLogicModel.ElasticKeyStatus) error {
	allowedTransitions, exists := validTransitions[current]
	if !exists {
		return errors.New("invalid current state")
	}

	if allowedTransitions[next] {
		return nil
	}

	return fmt.Errorf("invalid transition from current %s to next %s, allowed next %v", current, next, allowedTransitions)
}

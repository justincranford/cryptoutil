// Copyright (c) 2025 Justin Cranford
//
//

package businesslogic

import (
	"errors"
	"fmt"

	cryptoutilOpenapiModel "cryptoutil/api/model"
)

var validTransitions = func() map[cryptoutilOpenapiModel.ElasticKeyStatus]map[cryptoutilOpenapiModel.ElasticKeyStatus]bool {
	transitions := map[cryptoutilOpenapiModel.ElasticKeyStatus][]cryptoutilOpenapiModel.ElasticKeyStatus{
		cryptoutilOpenapiModel.Creating:                       {cryptoutilOpenapiModel.PendingGenerate, cryptoutilOpenapiModel.PendingImport},
		cryptoutilOpenapiModel.ImportFailed:                   {cryptoutilOpenapiModel.PendingDeleteWasImportFailed, cryptoutilOpenapiModel.PendingImport},
		cryptoutilOpenapiModel.PendingImport:                  {cryptoutilOpenapiModel.PendingDeleteWasPendingImport, cryptoutilOpenapiModel.ImportFailed, cryptoutilOpenapiModel.Active},
		cryptoutilOpenapiModel.PendingGenerate:                {cryptoutilOpenapiModel.GenerateFailed, cryptoutilOpenapiModel.Active},
		cryptoutilOpenapiModel.GenerateFailed:                 {cryptoutilOpenapiModel.PendingDeleteWasGenerateFailed, cryptoutilOpenapiModel.PendingGenerate},
		cryptoutilOpenapiModel.Active:                         {cryptoutilOpenapiModel.PendingDeleteWasActive, cryptoutilOpenapiModel.Disabled},
		cryptoutilOpenapiModel.Disabled:                       {cryptoutilOpenapiModel.PendingDeleteWasDisabled, cryptoutilOpenapiModel.Active},
		cryptoutilOpenapiModel.PendingDeleteWasImportFailed:   {cryptoutilOpenapiModel.FinishedDelete, cryptoutilOpenapiModel.ImportFailed},
		cryptoutilOpenapiModel.PendingDeleteWasPendingImport:  {cryptoutilOpenapiModel.FinishedDelete, cryptoutilOpenapiModel.PendingImport},
		cryptoutilOpenapiModel.PendingDeleteWasActive:         {cryptoutilOpenapiModel.FinishedDelete, cryptoutilOpenapiModel.Active},
		cryptoutilOpenapiModel.PendingDeleteWasDisabled:       {cryptoutilOpenapiModel.FinishedDelete, cryptoutilOpenapiModel.Disabled},
		cryptoutilOpenapiModel.PendingDeleteWasGenerateFailed: {cryptoutilOpenapiModel.FinishedDelete, cryptoutilOpenapiModel.GenerateFailed},
		cryptoutilOpenapiModel.StartedDelete:                  {cryptoutilOpenapiModel.FinishedDelete},
		cryptoutilOpenapiModel.FinishedDelete:                 {},
	}
	convertedTransitions := make(map[cryptoutilOpenapiModel.ElasticKeyStatus]map[cryptoutilOpenapiModel.ElasticKeyStatus]bool)
	for current, nextStates := range transitions {
		convertedTransitions[current] = make(map[cryptoutilOpenapiModel.ElasticKeyStatus]bool)
		for _, next := range nextStates {
			convertedTransitions[current][next] = true
		}
	}

	return convertedTransitions
}()

// TransitionElasticKeyStatus validates an ElasticKey status state transition.
func TransitionElasticKeyStatus(current, next cryptoutilOpenapiModel.ElasticKeyStatus) error {
	allowedTransitions, exists := validTransitions[current]
	if !exists {
		return errors.New("invalid current state")
	}

	if allowedTransitions[next] {
		return nil
	}

	return fmt.Errorf("invalid transition from current %s to next %s, allowed next %v", current, next, allowedTransitions)
}

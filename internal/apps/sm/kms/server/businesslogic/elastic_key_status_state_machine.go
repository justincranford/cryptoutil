// Copyright (c) 2025 Justin Cranford
//
//

package businesslogic

import (
	"errors"
	"fmt"

	cryptoutilKmsServer "cryptoutil/api/kms/server"
	cryptoutilOpenapiModel "cryptoutil/api/model"
)

var validTransitions = func() map[cryptoutilKmsServer.ElasticKeyStatus]map[cryptoutilKmsServer.ElasticKeyStatus]bool {
	transitions := map[cryptoutilKmsServer.ElasticKeyStatus][]cryptoutilKmsServer.ElasticKeyStatus{
		cryptoutilKmsServer.ElasticKeyStatus(cryptoutilOpenapiModel.Creating):        {cryptoutilKmsServer.ElasticKeyStatus(cryptoutilOpenapiModel.PendingGenerate), cryptoutilKmsServer.ElasticKeyStatus(cryptoutilOpenapiModel.PendingImport)},
		cryptoutilKmsServer.ElasticKeyStatus(cryptoutilOpenapiModel.ImportFailed):    {cryptoutilKmsServer.ElasticKeyStatus(cryptoutilOpenapiModel.PendingDeleteWasImportFailed), cryptoutilKmsServer.ElasticKeyStatus(cryptoutilOpenapiModel.PendingImport)},
		cryptoutilKmsServer.ElasticKeyStatus(cryptoutilOpenapiModel.PendingImport):   {cryptoutilKmsServer.ElasticKeyStatus(cryptoutilOpenapiModel.PendingDeleteWasPendingImport), cryptoutilKmsServer.ElasticKeyStatus(cryptoutilOpenapiModel.ImportFailed), cryptoutilKmsServer.Active},
		cryptoutilKmsServer.ElasticKeyStatus(cryptoutilOpenapiModel.PendingGenerate): {cryptoutilKmsServer.ElasticKeyStatus(cryptoutilOpenapiModel.GenerateFailed), cryptoutilKmsServer.Active},
		cryptoutilKmsServer.ElasticKeyStatus(cryptoutilOpenapiModel.GenerateFailed):  {cryptoutilKmsServer.ElasticKeyStatus(cryptoutilOpenapiModel.PendingDeleteWasGenerateFailed), cryptoutilKmsServer.ElasticKeyStatus(cryptoutilOpenapiModel.PendingGenerate)},
		cryptoutilKmsServer.Active: {cryptoutilKmsServer.ElasticKeyStatus(cryptoutilOpenapiModel.PendingDeleteWasActive), cryptoutilKmsServer.ElasticKeyStatus(cryptoutilOpenapiModel.Disabled)},
		cryptoutilKmsServer.ElasticKeyStatus(cryptoutilOpenapiModel.Disabled):                       {cryptoutilKmsServer.ElasticKeyStatus(cryptoutilOpenapiModel.PendingDeleteWasDisabled), cryptoutilKmsServer.Active},
		cryptoutilKmsServer.ElasticKeyStatus(cryptoutilOpenapiModel.PendingDeleteWasImportFailed):   {cryptoutilKmsServer.ElasticKeyStatus(cryptoutilOpenapiModel.FinishedDelete), cryptoutilKmsServer.ElasticKeyStatus(cryptoutilOpenapiModel.ImportFailed)},
		cryptoutilKmsServer.ElasticKeyStatus(cryptoutilOpenapiModel.PendingDeleteWasPendingImport):  {cryptoutilKmsServer.ElasticKeyStatus(cryptoutilOpenapiModel.FinishedDelete), cryptoutilKmsServer.ElasticKeyStatus(cryptoutilOpenapiModel.PendingImport)},
		cryptoutilKmsServer.ElasticKeyStatus(cryptoutilOpenapiModel.PendingDeleteWasActive):         {cryptoutilKmsServer.ElasticKeyStatus(cryptoutilOpenapiModel.FinishedDelete), cryptoutilKmsServer.Active},
		cryptoutilKmsServer.ElasticKeyStatus(cryptoutilOpenapiModel.PendingDeleteWasDisabled):       {cryptoutilKmsServer.ElasticKeyStatus(cryptoutilOpenapiModel.FinishedDelete), cryptoutilKmsServer.ElasticKeyStatus(cryptoutilOpenapiModel.Disabled)},
		cryptoutilKmsServer.ElasticKeyStatus(cryptoutilOpenapiModel.PendingDeleteWasGenerateFailed): {cryptoutilKmsServer.ElasticKeyStatus(cryptoutilOpenapiModel.FinishedDelete), cryptoutilKmsServer.ElasticKeyStatus(cryptoutilOpenapiModel.GenerateFailed)},
		cryptoutilKmsServer.ElasticKeyStatus(cryptoutilOpenapiModel.StartedDelete):                  {cryptoutilKmsServer.ElasticKeyStatus(cryptoutilOpenapiModel.FinishedDelete)},
		cryptoutilKmsServer.ElasticKeyStatus(cryptoutilOpenapiModel.FinishedDelete):                 {},
	}
	convertedTransitions := make(map[cryptoutilKmsServer.ElasticKeyStatus]map[cryptoutilKmsServer.ElasticKeyStatus]bool)
	for current, nextStates := range transitions {
		convertedTransitions[current] = make(map[cryptoutilKmsServer.ElasticKeyStatus]bool)
		for _, next := range nextStates {
			convertedTransitions[current][next] = true
		}
	}

	return convertedTransitions
}()

// TransitionElasticKeyStatus validates an ElasticKey status state transition.
func TransitionElasticKeyStatus(current, next cryptoutilKmsServer.ElasticKeyStatus) error {
	allowedTransitions, exists := validTransitions[current]
	if !exists {
		return errors.New("invalid current state")
	}

	if allowedTransitions[next] {
		return nil
	}

	return fmt.Errorf("invalid transition from current %s to next %s, allowed next %v", current, next, allowedTransitions)
}

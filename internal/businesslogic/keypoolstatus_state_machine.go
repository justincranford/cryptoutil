package service

import (
	cryptoutilBusinessModel "cryptoutil/internal/openapi/model"
	"errors"
	"fmt"
)

var validTransitions = func() map[cryptoutilBusinessModel.KeyPoolStatus]map[cryptoutilBusinessModel.KeyPoolStatus]bool {
	transitions := map[cryptoutilBusinessModel.KeyPoolStatus][]cryptoutilBusinessModel.KeyPoolStatus{
		cryptoutilBusinessModel.Creating:                       {cryptoutilBusinessModel.PendingGenerate, cryptoutilBusinessModel.PendingImport},
		cryptoutilBusinessModel.ImportFailed:                   {cryptoutilBusinessModel.PendingDeleteWasImportFailed, cryptoutilBusinessModel.PendingImport},
		cryptoutilBusinessModel.PendingImport:                  {cryptoutilBusinessModel.PendingDeleteWasPendingImport, cryptoutilBusinessModel.ImportFailed, cryptoutilBusinessModel.Active},
		cryptoutilBusinessModel.PendingGenerate:                {cryptoutilBusinessModel.GenerateFailed, cryptoutilBusinessModel.Active},
		cryptoutilBusinessModel.GenerateFailed:                 {cryptoutilBusinessModel.PendingDeleteWasGenerateFailed, cryptoutilBusinessModel.PendingGenerate},
		cryptoutilBusinessModel.Active:                         {cryptoutilBusinessModel.PendingDeleteWasActive, cryptoutilBusinessModel.Disabled},
		cryptoutilBusinessModel.Disabled:                       {cryptoutilBusinessModel.PendingDeleteWasDisabled, cryptoutilBusinessModel.Active},
		cryptoutilBusinessModel.PendingDeleteWasImportFailed:   {cryptoutilBusinessModel.FinishedDelete, cryptoutilBusinessModel.ImportFailed},
		cryptoutilBusinessModel.PendingDeleteWasPendingImport:  {cryptoutilBusinessModel.FinishedDelete, cryptoutilBusinessModel.PendingImport},
		cryptoutilBusinessModel.PendingDeleteWasActive:         {cryptoutilBusinessModel.FinishedDelete, cryptoutilBusinessModel.Active},
		cryptoutilBusinessModel.PendingDeleteWasDisabled:       {cryptoutilBusinessModel.FinishedDelete, cryptoutilBusinessModel.Disabled},
		cryptoutilBusinessModel.PendingDeleteWasGenerateFailed: {cryptoutilBusinessModel.FinishedDelete, cryptoutilBusinessModel.GenerateFailed},
		cryptoutilBusinessModel.StartedDelete:                  {cryptoutilBusinessModel.FinishedDelete},
		cryptoutilBusinessModel.FinishedDelete:                 {},
	}
	convertedTransitions := make(map[cryptoutilBusinessModel.KeyPoolStatus]map[cryptoutilBusinessModel.KeyPoolStatus]bool)
	for current, nextStates := range transitions {
		convertedTransitions[current] = make(map[cryptoutilBusinessModel.KeyPoolStatus]bool)
		for _, next := range nextStates {
			convertedTransitions[current][next] = true
		}
	}
	return convertedTransitions
}()

func TransitionState(current cryptoutilBusinessModel.KeyPoolStatus, next cryptoutilBusinessModel.KeyPoolStatus) error {
	allowedTransitions, exists := validTransitions[current]
	if !exists {
		return errors.New("invalid current state")
	}

	if allowedTransitions[next] {
		return nil
	}

	return fmt.Errorf("invalid transition from current %s to next %s, allowed next %v", current, next, allowedTransitions)
}

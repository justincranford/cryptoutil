package businesslogic

import (
	cryptoutilServiceModel "cryptoutil/internal/openapi/model"
	"errors"
	"fmt"
)

var validTransitions = func() map[cryptoutilServiceModel.KeyPoolStatus]map[cryptoutilServiceModel.KeyPoolStatus]bool {
	transitions := map[cryptoutilServiceModel.KeyPoolStatus][]cryptoutilServiceModel.KeyPoolStatus{
		cryptoutilServiceModel.Creating:                       {cryptoutilServiceModel.PendingGenerate, cryptoutilServiceModel.PendingImport},
		cryptoutilServiceModel.ImportFailed:                   {cryptoutilServiceModel.PendingDeleteWasImportFailed, cryptoutilServiceModel.PendingImport},
		cryptoutilServiceModel.PendingImport:                  {cryptoutilServiceModel.PendingDeleteWasPendingImport, cryptoutilServiceModel.ImportFailed, cryptoutilServiceModel.Active},
		cryptoutilServiceModel.PendingGenerate:                {cryptoutilServiceModel.GenerateFailed, cryptoutilServiceModel.Active},
		cryptoutilServiceModel.GenerateFailed:                 {cryptoutilServiceModel.PendingDeleteWasGenerateFailed, cryptoutilServiceModel.PendingGenerate},
		cryptoutilServiceModel.Active:                         {cryptoutilServiceModel.PendingDeleteWasActive, cryptoutilServiceModel.Disabled},
		cryptoutilServiceModel.Disabled:                       {cryptoutilServiceModel.PendingDeleteWasDisabled, cryptoutilServiceModel.Active},
		cryptoutilServiceModel.PendingDeleteWasImportFailed:   {cryptoutilServiceModel.FinishedDelete, cryptoutilServiceModel.ImportFailed},
		cryptoutilServiceModel.PendingDeleteWasPendingImport:  {cryptoutilServiceModel.FinishedDelete, cryptoutilServiceModel.PendingImport},
		cryptoutilServiceModel.PendingDeleteWasActive:         {cryptoutilServiceModel.FinishedDelete, cryptoutilServiceModel.Active},
		cryptoutilServiceModel.PendingDeleteWasDisabled:       {cryptoutilServiceModel.FinishedDelete, cryptoutilServiceModel.Disabled},
		cryptoutilServiceModel.PendingDeleteWasGenerateFailed: {cryptoutilServiceModel.FinishedDelete, cryptoutilServiceModel.GenerateFailed},
		cryptoutilServiceModel.StartedDelete:                  {cryptoutilServiceModel.FinishedDelete},
		cryptoutilServiceModel.FinishedDelete:                 {},
	}
	convertedTransitions := make(map[cryptoutilServiceModel.KeyPoolStatus]map[cryptoutilServiceModel.KeyPoolStatus]bool)
	for current, nextStates := range transitions {
		convertedTransitions[current] = make(map[cryptoutilServiceModel.KeyPoolStatus]bool)
		for _, next := range nextStates {
			convertedTransitions[current][next] = true
		}
	}
	return convertedTransitions
}()

func TransitionState(current cryptoutilServiceModel.KeyPoolStatus, next cryptoutilServiceModel.KeyPoolStatus) error {
	allowedTransitions, exists := validTransitions[current]
	if !exists {
		return errors.New("invalid current state")
	}

	if allowedTransitions[next] {
		return nil
	}

	return fmt.Errorf("invalid transition from current %s to next %s, allowed next %v", current, next, allowedTransitions)
}

package service

import (
	"errors"
	"fmt"
)

type KeyStatus string

const (
	Creating                       KeyStatus = "Creating"
	ImportFailed                   KeyStatus = "ImportFailed"
	PendingImport                  KeyStatus = "PendingImport"
	PendingGenerate                KeyStatus = "PendingGenerate"
	GenerateFailed                 KeyStatus = "GenerateFailed"
	Active                         KeyStatus = "Active"
	Disabled                       KeyStatus = "Disabled"
	PendingDeleteWasImportFailed   KeyStatus = "PendingDeleteWasImportFailed"
	PendingDeleteWasPendingImport  KeyStatus = "PendingDeleteWasPendingImport"
	PendingDeleteWasActive         KeyStatus = "PendingDeleteWasActive"
	PendingDeleteWasDisabled       KeyStatus = "PendingDeleteWasDisabled"
	PendingDeleteWasGenerateFailed KeyStatus = "PendingDeleteWasGenerateFailed"
	StartedDelete                  KeyStatus = "StartedDelete"
	FinishedDelete                 KeyStatus = "FinishedDelete"
)

var validTransitions = func() map[KeyStatus]map[KeyStatus]bool {
	transitions := map[KeyStatus][]KeyStatus{
		Creating:                       {PendingGenerate, PendingImport},
		ImportFailed:                   {PendingDeleteWasImportFailed, PendingImport},
		PendingImport:                  {PendingDeleteWasPendingImport, ImportFailed, Active},
		PendingGenerate:                {GenerateFailed, Active},
		GenerateFailed:                 {PendingDeleteWasGenerateFailed, PendingGenerate},
		Active:                         {PendingDeleteWasActive, Disabled},
		Disabled:                       {PendingDeleteWasDisabled, Active},
		PendingDeleteWasImportFailed:   {FinishedDelete, ImportFailed},
		PendingDeleteWasPendingImport:  {FinishedDelete, PendingImport},
		PendingDeleteWasActive:         {FinishedDelete, Active},
		PendingDeleteWasDisabled:       {FinishedDelete, Disabled},
		PendingDeleteWasGenerateFailed: {FinishedDelete, GenerateFailed},
		StartedDelete:                  {FinishedDelete},
		FinishedDelete:                 {},
	}
	convertedTransitions := make(map[KeyStatus]map[KeyStatus]bool)
	for current, nextStates := range transitions {
		convertedTransitions[current] = make(map[KeyStatus]bool)
		for _, next := range nextStates {
			convertedTransitions[current][next] = true
		}
	}
	return convertedTransitions
}()

func TransitionState(current KeyStatus, next KeyStatus) error {
	allowedTransitions, exists := validTransitions[current]
	if !exists {
		return errors.New("invalid current state")
	}

	if allowedTransitions[next] {
		return nil
	}

	return fmt.Errorf("invalid transition from current %s to next %s, allowed next %v", current, next, allowedTransitions)
}

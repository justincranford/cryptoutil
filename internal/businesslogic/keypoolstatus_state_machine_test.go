package service

import (
	"testing"
)

func TestTransitionInvalidState(t *testing.T) {
	err := TransitionState("DoesNotExist", Creating)
	if err == nil {
		t.Errorf("Expected transition from DoesNotExist to %s to fail, but it succeeded", Creating)
	}
}

func TestTransitionValidStateNextValid(t *testing.T) {
	for current, allowedNextStatuses := range validTransitions {
		for next := range allowedNextStatuses {
			t.Run("valid_"+string(current)+"_to_"+string(next), func(t *testing.T) {
				err := TransitionState(current, next)
				if err != nil {
					t.Errorf("Expected transition from %s to %s to succeed, but it failed with error: %v", current, next, err)
				}
			})
		}
	}
}

func TestTransitionValidStateNextInvalid(t *testing.T) {
	for current := range validTransitions {
		for potentialNext := range validTransitions {
			if potentialNext == current {
				continue // skip self, it is covered in TestTransitionStateSelfInvalid
			}
			if !validTransitions[current][potentialNext] {
				t.Run("invalid_"+string(current)+"_to_"+string(potentialNext), func(t *testing.T) {
					err := TransitionState(current, potentialNext)
					if err == nil {
						t.Errorf("Expected transition from %s to %s to fail, but it succeeded", current, potentialNext)
					}
				})
			}
		}
	}
}

func TestTransitionValidStateNextSelfInvalid(t *testing.T) {
	for current := range validTransitions {
		t.Run("self_"+string(current), func(t *testing.T) {
			err := TransitionState(current, current)
			if err == nil {
				t.Errorf("Expected self-transition for %s to fail, but it succeeded", current)
			}
		})
	}
}

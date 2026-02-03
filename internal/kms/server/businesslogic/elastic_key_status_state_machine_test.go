// Copyright (c) 2025 Justin Cranford
//
//

package businesslogic

import (
	"testing"

	cryptoutilKmsServer "cryptoutil/api/kms/server"
	cryptoutilOpenapiModel "cryptoutil/api/model"

	"github.com/stretchr/testify/require"
)

func TestTransitionInvalidState(t *testing.T) {
	err := TransitionElasticKeyStatus("DoesNotExist", cryptoutilKmsServer.ElasticKeyStatus(cryptoutilOpenapiModel.Creating))
	require.Error(t, err)
}

func TestTransitionValidStateNextValid(t *testing.T) {
	for current, allowedNextStatuses := range validTransitions {
		for next := range allowedNextStatuses {
			t.Run("valid_"+string(current)+"_to_"+string(next), func(t *testing.T) {
				err := TransitionElasticKeyStatus(current, next)
				require.NoError(t, err)
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
					err := TransitionElasticKeyStatus(current, potentialNext)
					require.Error(t, err)
				})
			}
		}
	}
}

func TestTransitionValidStateNextSelfInvalid(t *testing.T) {
	for current := range validTransitions {
		t.Run("self_"+string(current), func(t *testing.T) {
			err := TransitionElasticKeyStatus(current, current)
			require.Error(t, err)
		})
	}
}

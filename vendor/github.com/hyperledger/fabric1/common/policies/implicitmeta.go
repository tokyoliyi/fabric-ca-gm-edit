/*
Copyright IBM Corp. 2016 All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package policies

import (
	"fmt"

	cb "github.com/hyperledger/fabric/protos/common"

	"github.com/golang/protobuf/proto"
)

type implicitMetaPolicy struct {
	conf        *cb.ImplicitMetaPolicy
	threshold   int
	subPolicies []Policy
}

// NewPolicy creates a new policy based on the policy bytes
func newImplicitMetaPolicy(data []byte) (*implicitMetaPolicy, error) {
	imp := &cb.ImplicitMetaPolicy{}
	if err := proto.Unmarshal(data, imp); err != nil {
		return nil, fmt.Errorf("Error unmarshaling to ImplicitMetaPolicy: %s", err)
	}

	return &implicitMetaPolicy{
		conf: imp,
	}, nil
}

func (imp *implicitMetaPolicy) initialize(config *policyConfig) {
	imp.subPolicies = make([]Policy, len(config.managers))
	i := 0
	for _, manager := range config.managers {
		imp.subPolicies[i], _ = manager.GetPolicy(imp.conf.SubPolicy)
		i++
	}

	switch imp.conf.Rule {
	case cb.ImplicitMetaPolicy_ANY:
		imp.threshold = 1
	case cb.ImplicitMetaPolicy_ALL:
		imp.threshold = len(imp.subPolicies)
	case cb.ImplicitMetaPolicy_MAJORITY:
		imp.threshold = len(imp.subPolicies)/2 + 1
	}

	// In the special case that there are no policies, consider 0 to be a majority or any
	if len(imp.subPolicies) == 0 {
		imp.threshold = 0
	}
}

// Evaluate takes a set of SignedData and evaluates whether this set of signatures satisfies the policy
func (imp *implicitMetaPolicy) Evaluate(signatureSet []*cb.SignedData) error {
	remaining := imp.threshold
	for _, policy := range imp.subPolicies {
		if policy.Evaluate(signatureSet) == nil {
			remaining--
			if remaining == 0 {
				return nil
			}
		}
	}
	if remaining == 0 {
		return nil
	}
	return fmt.Errorf("Failed to reach implicit threshold of %d sub-policies, required %d remaining", imp.threshold, remaining)
}

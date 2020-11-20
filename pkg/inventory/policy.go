// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package inventory

import (
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

// InventoryPolicy defines if an inventory object can take over
// objects that belong to another inventory object or don't
// belong to any inventory object.
// This is done by determining if the apply/prune operation
// can go through for a resource based on the comparison
// the inventory-d annotation value in the package and that
// in the live object.
type InventoryPolicy int

const (
	// InvnetoryPolicyMustMatch: This policy enforces that the resources being applied can not
	// have any overlap with objects in other inventories or objects that already exist
	// in the cluster but don't belong to an inventory.
	//
	// The apply operation can go through when
	// - A new resources in the package doesn't exist in the cluster
	// - An existing resource in the package doesn't exist in the cluster
	// - An existing resource exist in the cluster. The inventory-id annotation in the live object
	//   matches with that in the package.
	//
	// The prune operation can go through when
	// - The inventory-id annotation in the live object match with that
	//   in the package.
	InventoryPolicyMustMatch InventoryPolicy = iota

	// AdoptIfNoInventory: This policy enforces that resources being applied
	// can not have any overlap with objects in other inventories, but are
	// permitted to take ownership of objects that don't belong to any inventories.
	//
	// The apply operation can go through when
	// - New resource in the package doesn't exist in the cluster
	// - If a new resource exist in the cluster, its inventory-id annotation is empty
	// - Existing resource in the package doesn't exist in the cluster
	// - If existing resource exist in the cluster, its inventory-id annotation in the live object
	//   is empty
	// - An existing resource exist in the cluster. The inventory-id annotation in the live object
	//   matches with that in the package.
	//
	// The prune operation can go through when
	// - The inventory-id annotation in the live object match with that
	//   in the package.
	AdoptIfNoInventory

	// AdoptAll: This policy will let the current inventory take ownership of any objects.
	//
	// The apply operation can go through for any resource in the package even if the
	// live object has an unmatched inventory-id annotation.
	//
	// The prune operation can go through when
	// - The inventory-id annotation in the live object match with that
	//   in the package.
	AdoptAll
)

const owningInventoryKey = "config.kubernetes.io/owning-inventory"

// inventoryIDMatchStatus represents the result of comparing the
// id from current inventory info and the inventory-id from a live object.
type inventoryIDMatchStatus int

const (
	Empty inventoryIDMatchStatus = iota
	Match
	Unmatch
)

func inventoryIDMatch(inv InventoryInfo, obj *unstructured.Unstructured) inventoryIDMatchStatus {
	annotations := obj.GetAnnotations()
	value, found := annotations[owningInventoryKey]
	if !found {
		return Empty
	}
	if value == inv.ID() {
		return Match
	}
	return Unmatch
}

func CanApply(inv InventoryInfo, obj *unstructured.Unstructured, policy InventoryPolicy) bool {
	if obj == nil {
		return true
	}
	matchStatus := inventoryIDMatch(inv, obj)
	switch matchStatus {
	case Empty:
		return policy != InventoryPolicyMustMatch
	case Match:
		return true
	case Unmatch:
		return policy == AdoptAll
	}
	return false
}

func CanPrune(inv InventoryInfo, obj *unstructured.Unstructured, policy InventoryPolicy) bool {
	if obj == nil {
		return false
	}
	matchStatus := inventoryIDMatch(inv, obj)
	switch matchStatus {
	case Empty:
		return false
	case Match:
		return true
	case Unmatch:
		return false
	}
	return false
}

package main

import (
	"fmt"
	"time"
)

type Scope struct {
	Key    string
	Type   string
	Path   string
	Demand func(scopedProps ScopedDemandProps) chan any
}

type Supplier func(data any, scope Scope) chan any

type DemandProps struct {
	Key       string
	Type      string
	Path      string
	Data      any
	Suppliers map[string]Supplier
}

type SuppliersMerge struct {
	Clear  bool
	Add    map[string]Supplier
	Remove map[string]bool
}

type ScopedDemandProps struct {
	Key            string
	Type           string
	SuppliersMerge SuppliersMerge
	Data           any
}

func mergeSuppliers(original map[string]Supplier, suppliersOp SuppliersMerge) map[string]Supplier {
	merged := make(map[string]Supplier)
	if !suppliersOp.Clear {
		for k, v := range original {
			merged[k] = v
		}
	}
	for k, v := range suppliersOp.Add {
		merged[k] = v
	}
	for k, remove := range suppliersOp.Remove {
		if remove {
			delete(merged, k)
		}
	}
	return merged
}

func globalDemand(props DemandProps) chan any {
	resultCh := make(chan any)
	go func() {
		defer close(resultCh)
		fmt.Println("Global demand function called with:")
		fmt.Printf("Key: %s\nType: %s\nPath: %s\n", props.Key, props.Type, props.Path)

		supplier, found := props.Suppliers[props.Type]
		if !found {
			fmt.Println("Supplier not found for type:", props.Type)
			return
		}
		scope := Scope{
			Key:    props.Key,
			Type:   props.Type,
			Path:   props.Path,
			Demand: createScopedDemand(props),
		}
		result := <-supplier(props.Data, scope)
		resultCh <- result
	}()
	return resultCh
}

func createScopedDemand(superProps DemandProps) func(ScopedDemandProps) chan any {
	return func(scopedProps ScopedDemandProps) chan any {
		demandKey := scopedProps.Key
		if demandKey == "" {
			demandKey = superProps.Key
		}
		path := superProps.Path + "/" + demandKey + "(" + scopedProps.Type + ")"

		newSuppliers := mergeSuppliers(superProps.Suppliers, scopedProps.SuppliersMerge)

		demandProps := DemandProps{
			Key:       demandKey,
			Type:      scopedProps.Type,
			Path:      path,
			Data:      scopedProps.Data,
			Suppliers: newSuppliers,
		}
		return globalDemand(demandProps)
	}
}

func supplyDemand(rootSupplier Supplier, suppliers map[string]Supplier) chan any {
	suppliers["$$root"] = rootSupplier
	demand := DemandProps{
		Key:  "root",
		Type: "$$root",
		Path: "root",
		Data: nil,
		Suppliers: suppliers,
	}
	return globalDemand(demand)
}

func thirdSupplier(data any, scope Scope) chan any {
	fmt.Println("Third supplier function called.")
	resultCh := make(chan any)
	go func() {
		defer close(resultCh)
		result := <-scope.Demand(ScopedDemandProps{Type: "first"})
		resultCh <- result
	}()
	return resultCh
}

func main() {
	suppliers := map[string]Supplier{
		"first": func(data any, scope Scope) chan any {
			resultCh := make(chan any)
			go func() {
				defer close(resultCh)
				fmt.Println("First supplier function called.")
				time.Sleep(1 * time.Second) // Simulate async work
				resultCh <- "First result"
			}()
			return resultCh
		},
		"second": func(data any, scope Scope) chan any {
			resultCh := make(chan any)
			go func() {
				defer close(resultCh)
				fmt.Println("Second supplier function called.")
				time.Sleep(1 * time.Second) // Simulate async work
				resultCh <- 2
			}()
			return resultCh
		},
	}

	rootSupplier := func(data any, scope Scope) chan any {
		resultCh := make(chan any)
		go func() {
			defer close(resultCh)
			fmt.Println("Root supplier function called.")
			mergeOps := SuppliersMerge{
				Add: map[string]Supplier{
					"third": thirdSupplier,
				},
			}
			result := <-scope.Demand(ScopedDemandProps{Type: "third", SuppliersMerge: mergeOps})
			if str, ok := result.(string); ok {
				fmt.Println("Root supplier received result:", str)
			}
			resultCh <- result
		}()
		return resultCh
	}

	result := <-supplyDemand(rootSupplier, suppliers)
	if str, ok := result.(string); ok {
		fmt.Println("Final result:", str)
	}
}
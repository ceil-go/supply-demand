package supply_demand

import (
	"fmt"
	"testing"
	"time"
)

func TestBasicSupplyDemandFlow(t *testing.T) {
	suppliers := map[string]Supplier{
		"first": func(data any, scope Scope) chan any {
			resultCh := make(chan any)
			go func() {
				defer close(resultCh)
				resultCh <- "OK"
			}()
			return resultCh
		},
	}
	rootSupplier := func(data any, scope Scope) chan any {
		resultCh := make(chan any)
		go func() {
			defer close(resultCh)
			res := <-scope.Demand(ScopedDemandProps{Type: "first"})
			resultCh <- res
		}()
		return resultCh
	}

	out := <-SupplyDemand(rootSupplier, suppliers)
	if out != "OK" {
		t.Errorf("expected OK, got %#v", out)
	}
}

func TestSupplierMergeAddAndRemove(t *testing.T) {
	supplierA := func(data any, scope Scope) chan any {
		resultCh := make(chan any)
		go func() {
			defer close(resultCh)
			resultCh <- "A"
		}()
		return resultCh
	}
	supplierB := func(data any, scope Scope) chan any {
		resultCh := make(chan any)
		go func() {
			defer close(resultCh)
			resultCh <- "B"
		}()
		return resultCh
	}

	suppliers := map[string]Supplier{"A": supplierA}
	rootSupplier := func(data any, scope Scope) chan any {
		resultCh := make(chan any)
		go func() {
			defer close(resultCh)
			mergeOps := SuppliersMerge{Add: map[string]Supplier{"B": supplierB}, Remove: map[string]bool{"A": true}}
			res := <-scope.Demand(ScopedDemandProps{Type: "B", SuppliersMerge: mergeOps})
			resultCh <- res
		}()
		return resultCh
	}

	out := <-SupplyDemand(rootSupplier, suppliers)
	if out != "B" {
		t.Errorf("expected B, got %#v", out)
	}
}

func TestChainedSuppliers(t *testing.T) {
	supplier2 := func(data any, scope Scope) chan any {
		resultCh := make(chan any)
		go func() {
			defer close(resultCh)
			time.Sleep(50 * time.Millisecond)
			resultCh <- "2"
		}()
		return resultCh
	}
	supplier1 := func(data any, scope Scope) chan any {
		resultCh := make(chan any)
		go func() {
			defer close(resultCh)
			res := <-scope.Demand(ScopedDemandProps{Type: "second"})
			resultCh <- fmt.Sprintf("1&%v", res)
		}()
		return resultCh
	}
	rootSupplier := func(data any, scope Scope) chan any {
		resultCh := make(chan any)
		go func() {
			defer close(resultCh)
			res := <-scope.Demand(ScopedDemandProps{Type: "first"})
			resultCh <- res
		}()
		return resultCh
	}
	suppliers := map[string]Supplier{
		"first":  supplier1,
		"second": supplier2,
	}

	val := <-SupplyDemand(rootSupplier, suppliers)
	if val != "1&2" {
		t.Errorf("expected 1&2, got %#v", val)
	}
}

func TestMissingSupplier(t *testing.T) {
	suppliers := map[string]Supplier{
		"first": func(data any, scope Scope) chan any {
			resultCh := make(chan any)
			go func() {
				defer close(resultCh)
				resultCh <- "first"
			}()
			return resultCh
		},
	}
	rootSupplier := func(data any, scope Scope) chan any {
		resultCh := make(chan any)
		go func() {
			defer close(resultCh)
			// This demand type does not exist
			res := <-scope.Demand(ScopedDemandProps{Type: "does-not-exist"})
			if res != nil {
				t.Errorf("expected nil, got %#v", res)
			}
			resultCh <- nil
		}()
		return resultCh
	}
	<-SupplyDemand(rootSupplier, suppliers) // If there's panic, test will fail
}

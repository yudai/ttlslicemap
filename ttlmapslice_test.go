package ttlslicemap

import (
	"sync"
	"testing"
	"time"
)

func TestBasicFunctions(t *testing.T) {
	tsm := New(time.Hour * 1)

	var new bool
	var items []interface{}
	var exists bool

	// start
	matchCount(t, tsm, 0)

	// create a slice
	new = tsm.Add("one", "foo")
	if !new {
		t.Errorf("It must be a new entry")
	}
	matchCount(t, tsm, 1)
	items, exists = tsm.Get("one")
	if !exists {
		t.Errorf("The added item is missing")
	}
	if len(items) != 1 || items[0] != "foo" {
		t.Errorf("Items must include the given item")
	}

	// add another item to the slice
	new = tsm.Add("one", "bar")
	if new {
		t.Errorf("It must not be a new entry")
	}
	matchCount(t, tsm, 1)
	items, exists = tsm.Get("one")
	if !exists {
		t.Errorf("The added items are missing")
	}
	if len(items) != 2 || items[0] != "foo" || items[1] != "bar" {
		t.Errorf("Items must include the given 2 items")
	}

	// create another slice
	new = tsm.Add("two", "Supersix")
	if !new {
		t.Errorf("It must be a new entry")
	}
	matchCount(t, tsm, 2)
	items, exists = tsm.Get("two")
	if !exists {
		t.Errorf("The added item is missing")
	}
	if len(items) != 1 || items[0] != "Supersix" {
		t.Errorf("Items must include the given item")
	}

	// add some items to the newer slice
	tsm.Add("two", "CAAD")
	tsm.Add("two", "Synapse")
	matchCount(t, tsm, 2)
	items, exists = tsm.Get("two")
	if !exists {
		t.Errorf("The added item is missing")
	}
	if len(items) != 3 || items[0] != "Supersix" || items[1] != "CAAD" || items[2] != "Synapse" {
		t.Errorf("Items must include the given item")
	}

	// Add slices and remove some of them
	tsm.Add("three", "san")
	tsm.Add("four", "yon")
	tsm.Add("five", "go")
	matchCount(t, tsm, 5)
	exists = tsm.Remove("three")
	if !exists {
		t.Errorf("Items should not have been removed")
	}
	matchCount(t, tsm, 4)
	exists = tsm.Remove("one")
	if !exists {
		t.Errorf("Items should not have been removed")
	}
	matchCount(t, tsm, 3)
}

func TestExpire(t *testing.T) {
	tsm := New(time.Second * 5)

	var exists bool

	// Check self removing
	tsm.Add("one", "Dura-Ace")
	matchCount(t, tsm, 1)
	time.Sleep(time.Second * 6)
	matchCount(t, tsm, 0)

	// Extend by Add()
	tsm.Add("two", "Ultegra")
	matchCount(t, tsm, 1)
	time.Sleep(time.Second * 3)
	tsm.Add("two", "105")
	time.Sleep(time.Second * 3)
	matchCount(t, tsm, 1)
	time.Sleep(time.Second * 3)
	matchCount(t, tsm, 0)

	// Extend by Get()
	tsm.Add("three", "Tiagra")
	matchCount(t, tsm, 1)
	time.Sleep(time.Second * 3)
	tsm.Get("three")
	time.Sleep(time.Second * 3)
	matchCount(t, tsm, 1)
	time.Sleep(time.Second * 3)
	matchCount(t, tsm, 0)

	// More complicated pattern
	tsm.Add("four", "san francisco")
	tsm.Add("five", "tokyo")
	time.Sleep(time.Second * 3)
	matchCount(t, tsm, 2)
	tsm.Add("four", "palo alto")
	tsm.Get("five")
	time.Sleep(time.Second * 3)
	matchCount(t, tsm, 2)
	tsm.Get("four")
	tsm.Add("five", "saitama")
	matchCount(t, tsm, 2)
	time.Sleep(time.Second * 3)
	tsm.Get("four")
	time.Sleep(time.Second * 3)
	matchCount(t, tsm, 1)
	if _, exists := tsm.Get("five"); exists {
		t.Errorf("five must have been deleted")
	}
	time.Sleep(time.Second * 6)
	matchCount(t, tsm, 0)

	// Remove before self removal
	tsm.Add("six", "gopher")
	matchCount(t, tsm, 1)
	exists = tsm.Remove("six")
	if !exists {
		t.Errorf("Items should not have been removed")
	}
	matchCount(t, tsm, 0)
	time.Sleep(time.Second * 6) // nothing happens
	matchCount(t, tsm, 0)

	// Remove and Add again
	tsm.Add("seven", "Mark")
	exists = tsm.Remove("seven")
	if !exists {
		t.Errorf("Items should not have been removed")
	}
	matchCount(t, tsm, 0)
	time.Sleep(time.Second * 3)
	tsm.Add("seven", "Cavendish") // readding, but this entry this should not be deleted
	matchCount(t, tsm, 1)
	time.Sleep(time.Second * 3)
	matchCount(t, tsm, 1)
	time.Sleep(time.Second * 3)
	matchCount(t, tsm, 0)
}

func TestConcurrent(t *testing.T) {
	tsm := New(time.Hour * 10)
	numRoutines := 100
	numAdd := 100

	wg := new(sync.WaitGroup)
	wg.Add(numRoutines * 2)
	for i := 0; i < numRoutines; i++ {
		go func() {
			defer wg.Done()
			for j := 0; j < numAdd; j++ {
				tsm.Add("foo", "something")
				tsm.Get("foo")
			}
		}()
		go func() {
			defer wg.Done()
			for j := 0; j < numAdd; j++ {
				tsm.Add("bar", "something")
				tsm.Get("bar")
			}
		}()
	}
	wg.Wait()

	matchCount(t, tsm, 2)
	expected := numAdd * numRoutines
	if slice, _ := tsm.Get("foo"); len(slice) != expected {
		t.Errorf("Add count error: expected %d, actual %d", expected, len(slice))
	}
	if slice, _ := tsm.Get("bar"); len(slice) != expected {
		t.Errorf("Add count error: expected %d, actual %d", expected, len(slice))
	}
}

func matchCount(t *testing.T, tsm *TTLSliceMap, expected int) {
	count := tsm.Count()
	if count != expected {
		t.Errorf("Item count must be %d, but got %d", expected, count)
	}
}

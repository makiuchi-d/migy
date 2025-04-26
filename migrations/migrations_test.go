package migrations

import (
	"reflect"
	"testing"
)

func TestFromSnapshot(t *testing.T) {
	migs, err := Load("testdata/migrations")
	if err != nil {
		t.Fatal(err)
	}

	exp := []int{3, 4}
	nums := make([]int, 0, len(migs))
	for _, m := range migs.FromSnapshot() {
		nums = append(nums, m.Number)
	}
	if !reflect.DeepEqual(nums, exp) {
		t.Fatalf("from snapshot: %v wants %v", nums, exp)
	}

	exp = []int{0, 1, 3}
	nums = nums[:0]
	mi, err := migs.FromSnapshotTo(3)
	if err != nil {
		t.Fatal(err)
	}
	for _, m := range mi {
		nums = append(nums, m.Number)
	}
	if !reflect.DeepEqual(nums, exp) {
		t.Fatalf("from snapshot to '3': %v wants %v", nums, exp)
	}
}

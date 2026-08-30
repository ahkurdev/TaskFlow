package service

import (
	"reflect"
	"testing"

	"taskflow/internal/model"
)

func idOfList(l model.List) int64 { return l.ID }

func TestReorderByID_MoveToFront(t *testing.T) {
	items := []model.List{{ID: 1}, {ID: 2}, {ID: 3}}

	got := reorderByID(items, 3, 0, idOfList)

	want := []int64{3, 1, 2}
	assertOrder(t, got, want)
}

func TestReorderByID_MoveToEnd(t *testing.T) {
	items := []model.List{{ID: 1}, {ID: 2}, {ID: 3}}

	got := reorderByID(items, 1, 10, idOfList) // out-of-range clamps to end

	want := []int64{2, 3, 1}
	assertOrder(t, got, want)
}

func TestReorderByID_MoveToMiddle(t *testing.T) {
	items := []model.List{{ID: 1}, {ID: 2}, {ID: 3}, {ID: 4}}

	got := reorderByID(items, 4, 1, idOfList)

	want := []int64{1, 4, 2, 3}
	assertOrder(t, got, want)
}

func TestReorderByID_NegativePositionClampsToStart(t *testing.T) {
	items := []model.List{{ID: 1}, {ID: 2}, {ID: 3}}

	got := reorderByID(items, 2, -5, idOfList)

	want := []int64{2, 1, 3}
	assertOrder(t, got, want)
}

func TestReorderByID_SingleItem(t *testing.T) {
	items := []model.List{{ID: 1}}

	got := reorderByID(items, 1, 0, idOfList)

	want := []int64{1}
	assertOrder(t, got, want)
}

func assertOrder(t *testing.T, got []model.List, want []int64) {
	t.Helper()

	gotIDs := make([]int64, len(got))
	for i, l := range got {
		gotIDs[i] = l.ID
	}

	if !reflect.DeepEqual(gotIDs, want) {
		t.Errorf("order = %v, want %v", gotIDs, want)
	}
}

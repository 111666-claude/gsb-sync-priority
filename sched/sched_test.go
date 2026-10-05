package sched

import "testing"

func TestSingleChangeIsSelected(t *testing.T) {
	s := &Scheduler{Radius: 1000, BudgetBytes: 1000}
	rows := s.Select([]Change{{Tick: 1, Entity: "a", Priority: 1, Dist: 10, Bytes: 32}})
	if len(rows) != 1 || rows[0].Entity != "a" {
		t.Fatalf("单条变更应被选中：%+v", rows)
	}
}

func TestHigherPriorityComesFirst(t *testing.T) {
	s := &Scheduler{Radius: 1000, BudgetBytes: 1000}
	rows := s.Select([]Change{
		{Tick: 1, Entity: "low", Priority: 1, Dist: 10, Bytes: 32},
		{Tick: 1, Entity: "high", Priority: 9, Dist: 10, Bytes: 32},
	})
	if len(rows) != 2 || rows[0].Entity != "high" {
		t.Fatalf("高优先级应排前面：%+v", rows)
	}
}

func TestBudgetHasRoomForBoth(t *testing.T) {
	s := &Scheduler{Radius: 1000, BudgetBytes: 1000}
	rows := s.Select([]Change{
		{Tick: 1, Entity: "a", Priority: 5, Dist: 10, Bytes: 16},
		{Tick: 1, Entity: "b", Priority: 1, Dist: 10, Bytes: 16},
	})
	if len(rows) != 2 {
		t.Fatalf("预算充足时两条都该选中：%+v", rows)
	}
}

func TestRowsCarryByteCost(t *testing.T) {
	s := &Scheduler{Radius: 1000, BudgetBytes: 1000}
	rows := s.Select([]Change{{Tick: 1, Entity: "a", Priority: 1, Dist: 10, Bytes: 48}})
	if rows[0].Bytes != 48 {
		t.Fatalf("字节数应为 48：%+v", rows)
	}
}

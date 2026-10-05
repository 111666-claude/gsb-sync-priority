// Package sched 按带宽配额挑选每个 tick 要同步给客户端的实体。
package sched

import "sort"

// Change 是某个 tick 上某个实体的一次变更。
type Change struct {
	Tick     int
	Entity   string
	Priority int
	Dist     int
	Bytes    int
}

// Row 是一条被选中的同步记录。
type Row struct {
	Tick   int
	Entity string
	Bytes  int
}

// Scheduler 挑选同步目标。
type Scheduler struct {
	Radius      int
	BudgetBytes int

	Scanned int
}

// Select 依次处理每个 tick，返回被选中的同步记录。
func (s *Scheduler) Select(changes []Change) []Row {
	byTick := map[int][]Change{}
	var ticks []int
	for _, c := range changes {
		s.Scanned++
		if _, ok := byTick[c.Tick]; !ok {
			ticks = append(ticks, c.Tick)
		}
		byTick[c.Tick] = append(byTick[c.Tick], c)
	}
	sort.Ints(ticks)

	var rows []Row
	for _, tick := range ticks {
		seen := map[string]bool{}
		cand := make([]Change, 0, len(byTick[tick]))
		for _, c := range byTick[tick] {
			if c.Dist > s.Radius {
				continue
			}
			if seen[c.Entity] {
				continue
			}
			seen[c.Entity] = true
			cand = append(cand, c)
		}

		sort.SliceStable(cand, func(i, j int) bool {
			if cand[i].Priority != cand[j].Priority {
				return cand[i].Priority > cand[j].Priority
			}
			if cand[i].Dist != cand[j].Dist {
				return cand[i].Dist < cand[j].Dist
			}
			return cand[i].Entity < cand[j].Entity
		})

		used := 0
		for _, c := range cand {
			if used+c.Bytes > s.BudgetBytes {
				continue
			}
			rows = append(rows, Row{Tick: tick, Entity: c.Entity, Bytes: c.Bytes})
			used += c.Bytes
		}
	}
	return rows
}

// Command plan 从 CSV 读变更，按 tick 打印要同步的实体表。
package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"
	"text/tabwriter"

	"example.com/sync-priority/sched"
)

func parseCSV(text string) []sched.Change {
	var out []sched.Change
	sc := bufio.NewScanner(strings.NewReader(text))
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" || strings.HasPrefix(line, "tick,") {
			continue
		}
		f := strings.Split(line, ",")
		if len(f) != 5 {
			continue
		}
		tick, _ := strconv.Atoi(strings.TrimSpace(f[0]))
		pri, _ := strconv.Atoi(strings.TrimSpace(f[2]))
		dist, _ := strconv.Atoi(strings.TrimSpace(f[3]))
		size, _ := strconv.Atoi(strings.TrimSpace(f[4]))
		out = append(out, sched.Change{
			Tick: tick, Entity: strings.TrimSpace(f[1]),
			Priority: pri, Dist: dist, Bytes: size,
		})
	}
	return out
}

func scaleCSV(ticks, perTick, entities int) string {
	var b strings.Builder
	b.WriteString("tick,entity,priority,dist,bytes\n")
	n := 0
	for t := 1; t <= ticks; t++ {
		for k := 0; k < perTick; k++ {
			fmt.Fprintf(&b, "%d,e%d,%d,%d,%d\n", t, n%entities, (n%5)+1, (n%40)+1, 16)
			n++
		}
	}
	return b.String()
}

func printRows(rows []sched.Row) {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 1, ' ', 0)
	fmt.Fprintln(w, "tick\tentity\tbytes")
	cur := -1
	total := 0
	for _, r := range rows {
		if r.Tick != cur {
			if cur >= 0 {
				fmt.Fprintf(w, "%d\ttotal\t%d\n", cur, total)
			}
			cur = r.Tick
			total = 0
		}
		fmt.Fprintf(w, "%d\t%s\t%d\n", r.Tick, r.Entity, r.Bytes)
		total += r.Bytes
	}
	if cur >= 0 {
		fmt.Fprintf(w, "%d\ttotal\t%d\n", cur, total)
	}
	w.Flush()
}

func main() {
	sample := flag.String("sample", "", "场景名；留空则从标准输入读 CSV")
	radius := flag.Int("radius", 100, "相关半径")
	budget := flag.Int("budget", 64, "每 tick 带宽配额（字节）")
	flag.Parse()

	text := ""
	radiusSet, budgetSet := *radius, *budget

	switch *sample {
	case "":
		sc := bufio.NewScanner(os.Stdin)
		var b strings.Builder
		for sc.Scan() {
			b.WriteString(sc.Text())
			b.WriteString("\n")
		}
		text = b.String()
	case "radius":
		radiusSet, budgetSet = 100, 1000
		text = "tick,entity,priority,dist,bytes\n1,a,5,50,10\n1,b,9,500,10\n"
	case "budget":
		radiusSet, budgetSet = 1000, 20
		text = "tick,entity,priority,dist,bytes\n1,a,3,10,10\n1,b,2,10,10\n1,c,1,10,10\n"
	case "order":
		radiusSet, budgetSet = 1000, 1000
		text = "tick,entity,priority,dist,bytes\n1,a,2,10,10\n1,b,2,5,10\n"
	case "dup":
		radiusSet, budgetSet = 1000, 20
		text = "tick,entity,priority,dist,bytes\n1,a,1,10,10\n1,a,1,10,10\n1,b,1,10,10\n"
	case "work":
		radiusSet, budgetSet = 1000, 100000
		text = scaleCSV(200, 15, 300)
	default:
		fmt.Fprintf(os.Stderr, "未知场景 %q\n", *sample)
		os.Exit(2)
	}

	changes := parseCSV(text)
	s := &sched.Scheduler{Radius: radiusSet, BudgetBytes: budgetSet}
	rows := s.Select(changes)

	if *sample == "work" {
		fmt.Printf("{\"scanned\":%d,\"rows\":%d}\n", s.Scanned, len(rows))
		return
	}
	sort.SliceStable(rows, func(i, j int) bool { return rows[i].Tick < rows[j].Tick })
	printRows(rows)
}

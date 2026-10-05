# sync-priority

按带宽配额挑选每个 tick 要同步给客户端的实体。只用 Go 标准库。

```
go test ./...
go run ./cmd/plan --sample=radius
go run ./cmd/plan --sample=budget
go run ./cmd/plan --sample=order
go run ./cmd/plan --sample=dup
go run ./cmd/plan --sample=work
```

不带 `--sample` 时从标准输入读 CSV，列序为 `tick,entity,priority,dist,bytes`。

## 口径

- **候选范围**：只有本次 `tick` 内发生变更的实体参与挑选，`dist` 大于 `Radius` 的不发。
- **排序**：按 `priority` 降序，同优先级按 `dist` 升序，再同则按 `entity` 升序。顺序必须确定。
- **带宽配额**：按累计 `bytes` 计，装不下 `BudgetBytes` 的实体整条跳过，不做截断。
- **幂等**：同一 tick 内重复出现的同一实体只算一次，只占一份配额。

## 不变量

- 每个 tick 的输出只由本 tick 的变更决定，与之前 tick 选过什么无关。
- 每个 tick 选中的记录字节总和不超过 `BudgetBytes`。
- 挑选只按变更索引进行，`Scanned` 不得随累计变更数放大。

## 输出契约

每个 tick 一段对齐表格，末尾一行 `total`：

```
tick  entity  bytes
1     b       10
1     a       10
1     total   20
```

`--sample=work` 只打印一行汇总：`{"rows": N, "scanned": N}`。

# Sniphunt 高性能搜索库设计方案

## 1. 背景与目标
Sniphunt 目前是一个简单的代码片段相似度搜索工具。为了提升其工程价值和性能，本项目将升级为一个高性能的 Go 语言搜索库，不仅支持原有的代码片段相似度搜索，还将引入类似 ripgrep 的高性能正则搜索功能。

### 核心目标
*   **高性能**：通过并发 I/O 和高效正则引擎，达到接近 ripgrep 的搜索速度。
*   **库化 (Library-first)**：提供简洁、易用的 Go API，方便集成到其他工具中。
*   **低资源消耗**：优化内存分配和 GC 压力。

---

## 2. 核心架构设计：三阶段流水线
采用生产者-消费者模型，解耦 I/O 遍历与 CPU 计算。

### 阶段 A：任务分发器 (Producer)
*   **职责**：快速遍历文件树，生成搜索任务。
*   **技术选型**：`github.com/karrick/godirwalk`。相比 `path/filepath.Walk`，它能显著减少 `Lstat` 调用，性能提升巨大。
*   **功能**：支持忽略文件/目录（类似 .gitignore）、过滤特定扩展名。

### 阶段 B：工作池 (Worker Pool)
*   **职责**：读取文件内容并执行匹配逻辑（正则或相似度计算）。
*   **技术选型**：`runtime.NumCPU()` 数量的 Worker。
*   **优化**：每个 Worker 持有一个预分配的 `sync.Pool` 缓冲区，减少堆内存分配。

### 阶段 C：结果汇总器 (Collector)
*   **职责**：异步收集匹配结果，通过 Channel 返回给调用者。
*   **功能**：支持流式处理结果，避免大内存占用。

---

## 3. 关键技术栈选型

| 模块 | 推荐方案 | 理由 |
| --- | --- | --- |
| **目录遍历** | `godirwalk` | 避免标准库性能坑，支持高效过滤。 |
| **正则引擎** | `regexp` / `regexp2` | 标准库 `regexp` 保证 $O(n)$ 时间；`regexp2` 支持 PCRE 断言。 |
| **并发控制** | `golang.org/x/sync/errgroup` | 优雅处理并发错误和 Context 取消。 |
| **对象复用** | `sync.Pool` | 极大减少 GC 压力。 |

---

## 4. 性能优化策略 (Performance "Three Axes")

### 第一招：预过滤 (Literal Pre-filter)
在执行昂贵的正则匹配前，先提取正则中的固定字符串。利用 `strings.Index`（汇编优化，支持 SIMD）进行快速扫描，如果不包含固定字符串则直接跳过该行。

### 第二招：零拷贝与对象池 (Zero-copy)
*   复用 `[]byte` 切片。
*   尽量避免 `string([]byte)` 转换。
*   在 `[]byte` 上直接进行正则匹配。

### 第三招：自适应读取策略
*   **普通文件**：`os.ReadFile`。
*   **大型文件**：使用 `bufio.Scanner` 并配合 64KB 以上的大 Buffer。
*   **超大文件**：考虑使用 `mmap`。

---

## 5. 项目结构与库接口设计

### 项目布局
遵循 Go 官方建议的工程布局：
*   `cmd/sniphunt/`：命令行工具入口。
*   `pkg/search/`：搜索核心库逻辑。
*   `pkg/similarity/`：相似度计算库。
*   `docs/`：项目文档。

### 库接口 (Library API)

```go
package sniphunt

import (
    "context"
)

// Searcher 定义搜索配置
type Searcher struct {
    Workers   int      // 并发数，默认 NumCPU
    Ignore    []string // 忽略路径
    Extensions []string // 允许的后缀
    // ... 其他配置
}

// Match 搜索结果
type Match struct {
    Path    string
    LineNum int
    Text    []byte
    Score   float64 // 用于相似度搜索
}

// Search 执行正则搜索
func (s *Searcher) Search(ctx context.Context, root string, pattern string) (<-chan Match, <-chan error) {
    // 内部实现：
    // 1. 编译正则
    // 2. 启动 Worker Pool
    // 3. 遍历文件并分发任务
}

// MatchSnippet 执行相似度搜索
func (s *Searcher) MatchSnippet(ctx context.Context, root string, snippet string) (<-chan Match, <-chan error) {
    // 内部实现：
    // 1. 计算相似度
}
```

---

## 6. 待办事项与路线图
1.  [x] 初始化 Go Module。
2.  [x] 集成 `godirwalk` 实现基础遍历框架。
3.  [x] 实现基于 `errgroup` 的 Worker Pool 并发模型。
4.  [x] 引入 `sync.Pool` 优化缓冲区管理。
5.  [x] 实现正则搜索核心逻辑（支持预过滤）。
6.  [x] 将原有的相似度搜索重构为库接口。
7.  [x] 编写单元测试和 Benchmark 压测。

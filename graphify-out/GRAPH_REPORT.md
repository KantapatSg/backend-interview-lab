# Graph Report - interview-crash-kit  (2026-08-21)

## Corpus Check
- cluster-only mode — file stats not available

## Summary
- 210 nodes · 315 edges · 26 communities (18 shown, 8 thin omitted)
- Extraction: 97% EXTRACTED · 3% INFERRED · 0% AMBIGUOUS · INFERRED: 8 edges (avg confidence: 0.85)
- Token cost: 0 input · 0 output

## Graph Freshness
- Built from commit: `f58d66ba`
- Run `git rev-parse HEAD` and compare to check if the graph is stale.
- Run `graphify update .` after code changes (no API cost).

## Community Hubs (Navigation)
- Community 0
- Community 1
- Community 2
- Community 3
- Community 4
- Community 5
- Community 6
- Community 7
- Community 8
- Community 9
- Community 10
- Community 11
- Community 12
- Community 13
- Community 14
- Community 15
- Community 16
- Community 17
- Community 18
- Community 21
- Community 23

## God Nodes (most connected - your core abstractions)
1. `Job` - 19 edges
2. `compilerOptions` - 15 edges
3. `JobService` - 9 edges
4. `NewJobService()` - 9 edges
5. `Repository` - 9 edges
6. `Event` - 8 edges
7. `store` - 8 edges
8. `Server` - 7 edges
9. `writeJSON()` - 7 edges
10. `Cache` - 7 edges

## Surprising Connections (you probably didn't know these)
- `TestJobTransitionIncrementsVersion()` --calls--> `NewJob()`  [INFERRED]
  platform/internal/domain/job_test.go → platform/internal/domain/job.go
- `Server` --references--> `JobService`  [EXTRACTED]
  platform/internal/httpapi/server.go → platform/internal/app/jobs.go
- `NewServer()` --references--> `JobService`  [EXTRACTED]
  platform/internal/httpapi/server.go → platform/internal/app/jobs.go
- `main()` --calls--> `NewJobService()`  [EXTRACTED]
  platform/cmd/api/main.go → platform/internal/app/jobs.go
- `TestCreateAndGetUsesCacheButKeepsDBSourceOfTruth()` --calls--> `NewJobService()`  [INFERRED]
  platform/internal/app/jobs_test.go → platform/internal/app/jobs.go

## Import Cycles
- None detected.

## Communities (26 total, 8 thin omitted)

### Community 0 - "Community 0"
Cohesion: 0.15
Nodes (15): EventPublisher, fakeCache, fakeRepo, JobCache, JobRepository, JobStatus, time.Duration, time.Time (+7 more)

### Community 1 - "Community 1"
Cohesion: 0.08
Nodes (25): compilerOptions, allowJs, esModuleInterop, incremental, isolatedModules, jsx, lib, module (+17 more)

### Community 2 - "Community 2"
Cohesion: 0.08
Nodes (23): next, dependencies, next, react, react-dom, devDependencies, @types/node, @types/react (+15 more)

### Community 3 - "Community 3"
Cohesion: 0.19
Nodes (12): net/http.Handler, testing.T, fetch(), main(), TestFetchCompletesBeforeDeadline(), TestFetchHonorsDeadline(), TestJobTransitionIncrementsVersion(), NewServer() (+4 more)

### Community 4 - "Community 4"
Cohesion: 0.22
Nodes (6): context.Context, github.com/jackc/pgx/v5/pgxpool.Pool, fakeRepository, Repository, isUniqueViolation(), New()

### Community 5 - "Community 5"
Cohesion: 0.21
Nodes (9): Consumer, Producer, kafka.Reader, kafka.Writer, env(), main(), publishOutbox(), NewConsumer() (+1 more)

### Community 6 - "Community 6"
Cohesion: 0.37
Nodes (8): encoding/json.RawMessage, net/http.Request, net/http.ResponseWriter, createRequest, Server, valueOrDefault(), writeError(), writeJSON()

### Community 7 - "Community 7"
Cohesion: 0.30
Nodes (7): order, orderCreated, store, sync.RWMutex, main(), newStore(), TestProjectionIsEventuallyVisible()

### Community 8 - "Community 8"
Cohesion: 0.24
Nodes (5): github.com/redis/go-redis/v9.Client, env(), main(), New(), Cache

### Community 9 - "Community 9"
Cohesion: 0.47
Nodes (4): result, main(), runPool(), TestRunPoolProcessesEveryInput()

### Community 10 - "Community 10"
Cohesion: 0.40
Nodes (3): Order, OrderStatus(), Order

### Community 13 - "Community 13"
Cohesion: 0.50
Nodes (3): orders, outbox_events, processed_events

### Community 14 - "Community 14"
Cohesion: 0.50
Nodes (3): jobs, outbox_events, processed_events

## Knowledge Gaps
- **48 isolated node(s):** `Order`, `Order`, `Job`, `ImportJob`, `EventPublisher` (+43 more)
  These have ≤1 connection - possible missing edges or undocumented components.
- **8 thin communities (<3 nodes) omitted from report** — run `graphify query` to explore isolated nodes.

## Suggested Questions
_Questions this graph is uniquely positioned to answer:_

- **Why does `Server` connect `Community 6` to `Community 0`, `Community 4`?**
  _High betweenness centrality (0.053) - this node is a cross-community bridge._
- **Why does `Job` connect `Community 0` to `Community 8`, `Community 4`?**
  _High betweenness centrality (0.028) - this node is a cross-community bridge._
- **Why does `publishOutbox()` connect `Community 5` to `Community 4`?**
  _High betweenness centrality (0.025) - this node is a cross-community bridge._
- **Are the 2 inferred relationships involving `NewJobService()` (e.g. with `TestCreateAndGetUsesCacheButKeepsDBSourceOfTruth()` and `TestGetFallsBackWhenCacheMisses()`) actually correct?**
  _`NewJobService()` has 2 INFERRED edges - model-reasoned connections that need verification._
- **What connects `Order`, `Order`, `Job` to the rest of the system?**
  _48 weakly-connected nodes found - possible documentation gaps or missing edges._
- **Should `Community 0` be split into smaller, more focused modules?**
  _Cohesion score 0.14814814814814814 - nodes in this community are weakly interconnected._
- **Should `Community 1` be split into smaller, more focused modules?**
  _Cohesion score 0.07692307692307693 - nodes in this community are weakly interconnected._
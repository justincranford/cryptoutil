# Quizme v3 — framework/suite/ and framework/product/ Scope

**Plan**: docs/framework-v5/plan.md
**Context**: All other quizme-v2 questions resolved. This one requires concrete
code evidence before choosing between options B, C, and D.

---

## What Currently Exists

### Suite CLI — NOT yet using framework (`internal/apps/cryptoutil/cryptoutil.go`)

The suite entry point has hand-coded routing with an inline `printUsage()`:

```go
// Routes by product name via switch statement.
// No framework function used — only inline code.
func Suite(args []string, ...) int {
    switch product {
    case "identity": return cryptoutilAppsIdentity.Identity(...)
    case "jose":     return cryptoutilAppsJose.Jose(...)
    case "pki":      return cryptoutilAppsPki.Pki(...)
    case "skeleton": return cryptoutilAppsSkeleton.Skeleton(...)
    case "sm":       return cryptoutilAppsSm.Sm(...)
    case "pki-init": return cryptoutilAppsPkiinit.Run(...)
    }
}
```

There is **no** `RouteSuite()` function in `framework/service/cli/`.

### Product CLI — ALREADY using framework (`internal/apps/sm/sm.go`, etc.)

All 5 products already call `framework/service/cli.RouteProduct()`:

```go
// sm.go — identical pattern across all products
func Sm(args []string, ...) int {
    return cryptoutilTemplateCli.RouteProduct(
        cryptoutilTemplateCli.ProductConfig{
            ProductName: "sm",
            UsageText:   usageText,
            VersionText: versionText,
        },
        args, stdin, stdout, stderr,
        []cryptoutilTemplateCli.ServiceEntry{
            {Name: "im",  Handler: cryptoutilAppsSmIm.Im},
            {Name: "kms", Handler: cryptoutilAppsSmKms.Kms},
        },
    )
}
```

### Framework CLI — current contents (`internal/apps/framework/service/cli/`)

| File | Provides |
|------|----------|
| `product_router.go` | `RouteProduct()`, `ProductConfig`, `ServiceEntry` |
| `service_router.go` | `RouteService()`, `ServiceConfig`, `SubcommandFunc` |
| `health_commands.go` | `HealthCommand()`, `LivezCommand()`, `ReadyzCommand()`, `ShutdownCommand()` |
| `http_client.go` | HTTP client helpers for health checks |
| `constants.go` | Shared CLI constants |

### The Gap

| Tier | Framework function | Status |
|------|--------------------|--------|
| Suite → Products | `RouteSuite()` | **MISSING** — `cryptoutil.go` uses inline switch |
| Product → Services | `RouteProduct()` | ✅ Exists in `framework/service/cli/` |
| Service → Subcommands | `RouteService()` | ✅ Exists in `framework/service/cli/` |

---

## What `framework/suite/` and `framework/product/` Would Contain

The only concrete missing piece at the suite level is a `RouteSuite()` function
(the suite-level equivalent of `RouteProduct()`). The product level is already
fully served by `framework/service/cli/`.

Three structural options for where to add `RouteSuite()` (and optionally move
`RouteProduct()`):

---

## Question 1: Where should `RouteSuite()` live?

**Question**: The suite entry point (`cryptoutil.go`) uses an inline switch and
`printUsage()`. A `RouteSuite()` framework function would extract this pattern,
matching how `RouteProduct()` already extracts the product routing. Where should
`RouteSuite()` (and a `SuiteConfig`/`ProductEntry` type pair) be defined?

**A)** Add to existing `framework/service/cli/` as `suite_router.go` — minimal
   change, same package, no package restructure needed

**B)** Create new `framework/suite/cli/` package — clean separation by tier,
   mirrors how `service/` has its own `cli/` subdirectory

**C)** Create `framework/suite/` with just a `suite_router.go` (flat, no `cli/`
   subdirectory) — single file, no over-nesting

**D)** Keep inline in `cryptoutil.go` — suite is a 1-of-1 entry point, not worth
   abstracting; product/service were worth it because there are 5+10 instances

**E)**

**Answer**:

**Rationale**: There is exactly ONE suite (`cryptoutil`). The `RouteProduct()` and
`RouteService()` abstractions pay off because they are used 5 times and 10 times
respectively. `RouteSuite()` would only ever be called once. Option D argues the
abstraction is not worth it. Options A-C argue consistency and future-proofing.

---

## Question 2: Should `RouteProduct()` stay in `framework/service/cli/`?

**Question**: Currently `RouteProduct()` lives in `framework/service/cli/` even
though it handles product-level (not service-level) routing. This is a naming
inconsistency. Should it move?

**A)** Keep `RouteProduct()` in `framework/service/cli/` — it works, moving it
   breaks all 5 product imports, not worth the churn

**B)** Move `RouteProduct()` to `framework/product/cli/` — correct tier placement,
   mirrors the `framework/service/cli/` pattern; update all 5 product imports

**C)** Move `RouteProduct()` to `framework/product/` (flat, no `cli/` subdir) —
   single file, avoids over-nesting

**D)** Create a new `framework/cli/` package that holds suite, product, AND service
   routing together — all CLI infrastructure unified in one place

**E)**

**Answer**:

**Rationale**: There are exactly 5 products that import `RouteProduct()`. Moving
it is a well-defined refactor. Options A and D minimize churn. Options B and C
improve structural accuracy at the cost of import path changes.

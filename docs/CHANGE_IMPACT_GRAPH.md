# Change Impact Graph

Potential blast radius for a changed service, derived from persisted directed dependencies.

Edge semantics: **`from → to` means `from` depends on `to`.**

If `checkout-api → payments-service`, a regression in payments can affect checkout. Checkout is a **dependent** of payments. Payments is a **dependency** of checkout.

## Algorithm

Blast radius of changed service `S`:

1. Build adjacency lists:
   - `dependsOn[from] = [to, ...]`
   - `dependedBy[to] = [from, ...]`
2. **Downstream impact** (who could feel this change): BFS on `dependedBy` starting at `S`, excluding `S`.
3. **Upstream dependencies** (what this service relies on): BFS on `dependsOn` starting at `S`, excluding `S`.
4. Record `depth` as BFS level.
5. **Cycles:** color-state DFS on `dependsOn`. Nodes in a cycle are still visited once in BFS via a `seen` set. Cycle membership is returned as `cycles[]` (lists of service IDs). Back-edges do not infinite-loop.

BFS is the default because depth is the natural blast-radius explanation ("two hops from payments to storefront").

## Output

```json
{
  "changed_service_id": "payments-service",
  "direct_dependents": ["checkout-api"],
  "transitive_dependents": ["web-storefront"],
  "upstream_dependencies": ["notification-worker"],
  "max_dependent_depth": 2,
  "critical_in_radius": ["checkout-api"],
  "cycles": [],
  "algorithm": "bfs",
  "model_version": "graph-v1"
}
```

`critical_in_radius` = services in direct or transitive dependents whose criticality is HIGH or CRITICAL.

Unknown `S`: empty radius plus error code `UNKNOWN_SERVICE`. Do not invent neighbors.

Missing graph (no edges at all): radius empty, `graph_status: EMPTY`, not a fake singleton.

## UI contract

The graph is rendered from this payload. The frontend must not contain a hardcoded map of "payments affects checkout and storefront."

Recruiter-readable caption:

> This service was changed. These systems could potentially be affected.

Caption must include the limitation:

> Reachability is not observed breakage. A dependent may be unaffected in production.

## Test obligations

- linear chain depth
- diamond (two paths, node visited once)
- cycle A→B→A
- disconnected node
- unknown service
- empty catalog
- critical_in_radius filter

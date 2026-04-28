# Public Release Readiness Decision

Date: 2026-04-28
Repository: `silentspike/netbird-machine-tunnel`
Decision: **GO for public repository visibility after this decision record is merged and final maintainer approval is recorded**

This record separates repository visibility from publishing a tagged binary
release. The repository is ready to become public after the protected-branch
documentation update lands. A final tagged GitHub Release remains a separate
promotion step because it needs a version/tag decision and release notes.

## Current Decision

| Area | State | Evidence |
|------|-------|----------|
| Repository visibility | PASS / still private | GitHub reports `visibility=PRIVATE`, `isPrivate=true`; no public flip has been performed yet. |
| Current launch candidate | PASS | Protected `main` is at `600de6963790e12bf3bea15f78f01714e054663b`. |
| Main CI | PASS | Final `main` workflows for `600de696` completed successfully: Secret Scan `25047595566`, Test installation `25047595572`, License `25047595557`, Wasm `25047595585`, Infrastructure `25047595550`, CodeQL `25047595588`, Darwin `25047595571`, Mobile `25047595626`, FreeBSD `25047595590`, Windows `25047595584`, Release `25047595583`, Linux `25047595587`, plus Dependabot update runs. |
| Branch protection | PASS | `main` uses strict required checks with 48 contexts, including `CodeQL (go)` and `CodeQL (javascript-typescript)`. |
| CodeQL / code scanning | PASS | #170 is dispositioned. PR #186 re-enabled CodeQL pull-request scanning. Current open `main` code-scanning baseline is `142 medium`, `2 warning`, `0 critical`, `0 high`. The inherited upstream OIDC SSRF alert was explicitly accepted/dismissed for this launch scope. |
| Dependabot alerts | PASS | #167 is closed. Open Dependabot alerts are `0` after PR #185 merged `go-jose` remediation and the no-patch/transitive alerts were dismissed with documented `tolerable_risk` dispositions. |
| Signal trust model | PASS | #114 and #109 are closed. Public docs now describe the real Signal trust boundary instead of claiming certificate pinning. |
| RC artifact generation | PASS | Release workflow run `25047595583` produced the final Actions snapshot artifact set for `600de696`. |
| Checksums and SBOM | PASS | The downloaded `release` artifact passed `sha256sum -c netbird_0.1.0-SNAPSHOT-600de696_checksums.txt`; the Windows Machine archive and SBOM are checksum-covered. The `netbird-machine` SBOM exists and contains 86 packages. |
| Downloaded-artifact lab smoke | PASS | VM102 ran the downloaded `netbird-machine.exe` from the final Actions artifact, not a local build. Installed service binary SHA256: `1ddc5069ec364157be07a874621d3d64f9e4cbe367c656d2858fd867aa155440`. |
| Machine Tunnel lab function | PASS | `NetBirdMachine` is Running/Automatic, `wg-nb-machine` is Up, route `192.168.100.0/24` is installed via `wg-nb-machine`, and DC ports `53`, `88`, `389`, `445`, and `636` are reachable through the tunnel. Logs show WireGuard handshake evidence after restart. |
| Stale GitHub Release exposure | PASS | Old `v0.1.0` release from 2026-02-06 was moved to Draft; `gh release list --exclude-drafts` returns no public releases. |
| Public documentation | PASS after this PR | README/license/security/ADR/fork-diff documentation has been cleaned up; this file is the final stale NO-GO document that must be merged before public visibility. |

## Go Conditions

Public repository visibility is approved only when all of the following remain
true at the moment of the visibility change:

1. This decision record is merged to protected `main`.
2. The post-merge `main` CI for the documentation update is green.
3. `gh repo view silentspike/netbird-machine-tunnel` still reports
   `visibility=PRIVATE` immediately before the flip.
4. `gh release list --exclude-drafts` still returns no stale public releases.
5. The maintainer explicitly confirms the visibility flip after reviewing this
   record.

## Release Hold

Do not publish a final tagged GitHub Release in the same step as the visibility
flip. The binary release should be promoted separately after choosing:

1. public version/tag name,
2. release notes,
3. whether to publish a prerelease first,
4. whether to reuse the validated Actions snapshot artifacts or generate a
   tag-triggered release run.

Until then, the final validated launch artifact is the private Actions snapshot
from run `25047595583` at commit `600de6963790e12bf3bea15f78f01714e054663b`.

## Issue Disposition

| Issue | State |
|-------|-------|
| #167 | Closed completed after dependency remediation and alert disposition. |
| #170 | Closeable after posting PR #186, branch-protection, CodeQL baseline, and final main CI evidence. |
| #168 | Closeable after posting final `600de696` CI, artifact checksum/SBOM, lab smoke, and stale-release-draft evidence. |
| #149 | Closeable after #168/#170 are closed; the v0.69.0 sync is complete on `main`. |

## Final Position

The project is **GO for public repository visibility** after this decision
record merges through the protected branch and final maintainer approval is
recorded.

The project is **HOLD for a public tagged binary release** until the explicit
release-promotion decision above is made.

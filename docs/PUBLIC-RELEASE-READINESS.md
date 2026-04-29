# Public Release Readiness Decision

Date: 2026-04-28
Repository: `silentspike/netbird-machine-tunnel`
Decision: **HOLD for public repository visibility until the v0.70.0 sync branch has green post-merge CI, refreshed artifact evidence, and final maintainer approval**

This record separates repository visibility from publishing a tagged binary
release. The previous public-readiness pass was valid for the v0.69.0 baseline,
but this branch intentionally moves the fork to upstream NetBird v0.70.0 before
public visibility. That resets the final visibility decision until the v0.70.0
sync lands on protected `main`, post-merge CI is green, release-candidate
artifact checks are refreshed, and the maintainer gives final approval. A final
tagged GitHub Release remains a separate promotion step because it needs a
version/tag decision, release notes, and release-artifact promotion approval.

## Current Decision

| Area | State | Evidence |
|------|-------|----------|
| Repository visibility | PASS / still private | GitHub reports `visibility=PRIVATE`, `isPrivate=true`; no public flip has been performed yet. |
| Current launch candidate | PASS | Protected `main` is at `48f9f3371e7dd832019cc1cd1b2d60632ce7c268`. |
| Main CI | PASS | Final `main` workflows for `48f9f337` completed successfully: Secret Scan `25065555333`, Test installation `25065555263`, License `25065555281`, Wasm `25065555287`, Infrastructure `25065555404`, CodeQL `25065555318`, Darwin `25065555293`, Mobile `25065555264`, FreeBSD `25065555238`, Windows `25065555332`, Release `25065555244`, Linux `25065555278`. |
| Branch protection | PASS | `main` uses strict required checks with 48 contexts, including `CodeQL (go)` and `CodeQL (javascript-typescript)`, with admin enforcement enabled. |
| CodeQL / code scanning | PASS | #170 is closed. PR #186 re-enabled CodeQL pull-request scanning. Current open `main` code-scanning baseline by `rule.security_severity_level` is `142 medium`, `2 warning`, `0 critical`, `0 high`. The inherited upstream OIDC SSRF alert was explicitly accepted/dismissed for this launch scope. |
| Dependabot alerts | PASS | #167 is closed. Open Dependabot alerts are `0` after PR #185 merged `go-jose` remediation and the no-patch/transitive alerts were dismissed with documented `tolerable_risk` dispositions. |
| Signal trust model | PASS | #114 and #109 are closed. Public docs now describe the real Signal trust boundary instead of claiming certificate pinning. |
| Issue board public hygiene | PASS | Open `priority:critical` and `priority:high` issue counts are both `0`. Historical umbrella/docs/flake issues were commented and downgraded to backlog/medium where needed. |
| RC artifact generation | PASS / historical snapshot evidence | Release workflow run `25047595583` produced the validated private Actions snapshot artifact set for commit `600de6963790e12bf3bea15f78f01714e054663b`. Later changes through `48f9f337` are documentation/workflow/test-only public-readiness changes and have green `main` CI. |
| Checksums and SBOM | PASS / historical snapshot evidence | The downloaded `release` artifact from run `25047595583` passed `sha256sum -c netbird_0.1.0-SNAPSHOT-600de696_checksums.txt`; the Windows Machine archive and SBOM are checksum-covered. The `netbird-machine` SBOM exists and contains 86 packages. |
| Downloaded-artifact lab smoke | PASS / historical snapshot evidence | VM102 ran the downloaded `netbird-machine.exe` from the private Actions artifact for `600de696`, not a local build. Installed service binary SHA256: `1ddc5069ec364157be07a874621d3d64f9e4cbe367c656d2858fd867aa155440`. |
| Machine Tunnel lab function | PASS / historical snapshot evidence | For the validated private Actions snapshot, `NetBirdMachine` was Running/Automatic, `wg-nb-machine` was Up, route `192.168.100.0/24` was installed via `wg-nb-machine`, and DC ports `53`, `88`, `389`, `445`, and `636` were reachable through the tunnel. Logs showed WireGuard handshake evidence after restart. |
| Stale GitHub Release exposure | PASS | Old `v0.1.0` release from 2026-02-06 was moved to Draft; current release listing shows only draft `v0.1.0` and no published release. |
| Public documentation | PASS / refreshed in current sprint branch | README/license/security/ADR/fork-diff documentation has been cleaned up. This decision record now distinguishes current `48f9f337` main/CI state from historical artifact/lab evidence at `600de696`. |
| Upstream drift | IN PROGRESS | This branch syncs upstream NetBird `v0.70.0` through #189. Final public visibility remains on HOLD until the v0.70.0 post-merge CI, security, RC artifact, and lab evidence are refreshed. |

## Go Conditions

Public repository visibility is approved only when all of the following remain
true at the moment of the visibility change:

1. This refreshed decision record is merged to protected `main`.
2. The post-merge `main` CI for this public-readiness update is green.
3. `gh repo view silentspike/netbird-machine-tunnel` still reports
   `visibility=PRIVATE` immediately before the flip.
4. `gh release list --repo silentspike/netbird-machine-tunnel --json tagName,isDraft`
   still shows no published non-draft release that would become public by
   accident.
5. Open Dependabot alerts remain `0`.
6. Open CodeQL critical/high alerts remain `0`.
7. Open `priority:critical` and `priority:high` issue counts remain `0` or every
   remaining item has an explicit non-blocking disposition.
8. The maintainer explicitly confirms the visibility flip after reviewing this
   record.

## Release Hold

Do not publish a final tagged GitHub Release in the same step as the repository
visibility flip. The binary release should be promoted separately after
choosing:

1. public version/tag name,
2. release notes,
3. whether to publish a prerelease first,
4. whether to reuse the validated private Actions snapshot artifacts or generate
   a tag-triggered release run,
5. whether to run a fresh downloaded-artifact lab smoke for the final release
   candidate.

Until then, the final validated launch artifact is the private Actions snapshot
from run `25047595583` at commit `600de6963790e12bf3bea15f78f01714e054663b`.
The current public-visibility launch candidate is protected `main` at
`48f9f3371e7dd832019cc1cd1b2d60632ce7c268`, with green CI and no public tagged
binary release.

## Issue Disposition

| Issue | State |
|-------|-------|
| #149 | Closed completed after upstream NetBird v0.69.0 sync landed on protected `main`. |
| #166 | Closed completed; follow-up comment records current 48 strict branch-protection contexts including both CodeQL contexts. |
| #167 | Closed completed after dependency remediation, no-patch/transitive alert disposition, and default-branch open Dependabot alert count `0`. |
| #168 | Closed completed after final `48f9f337` CI, draft-release exposure check, dependency/CodeQL status, and historical artifact/lab evidence were recorded. |
| #170 | Closed completed after CodeQL PR scanning was re-enabled, branch protection required both CodeQL contexts, and the final baseline reached `0` critical/high. |
| #189 | In progress for upstream NetBird v0.70.0 sync before public visibility. |

## Final Position

The project is **HOLD for public repository visibility** until the v0.70.0 sync
branch merges through the protected branch, post-merge checks are green, final
artifact/lab/security evidence is refreshed, and final maintainer approval is
recorded immediately before the visibility flip.

The project is **HOLD for a public tagged binary release** until the explicit
release-promotion decision above is made.

# Public Release Readiness Decision

Date: 2026-04-29
Repository: `silentspike/netbird-machine-tunnel`
Decision: **GO for public repository visibility after this refreshed decision record is merged to protected `main` and its post-merge CI is green**

This record separates repository visibility from publishing a tagged binary
release. Repository visibility can move to public after the gates below remain
true immediately before the flip. A public tagged GitHub Release remains on
hold and must be promoted separately.

## Current Decision

| Area | State | Evidence |
|------|-------|----------|
| Repository visibility | PASS / still private | `gh repo view silentspike/netbird-machine-tunnel` reports `visibility=PRIVATE`, `isPrivate=true`. |
| Current launch candidate | PASS | The validated feature launch commit is `0acb92e1ac82bd65194a0b03f9864bda5b657011`, the merge commit for PR #191. Protected `main` may include later documentation-only readiness commits; the current protected `main` commit must have green post-merge CI immediately before the visibility flip. |
| Upstream baseline | PASS | Fork is synced through upstream NetBird `v0.70.0`; `upstream/main` commits after the tag remain intentionally out of scope. |
| PR #191 | PASS | PR #191 merged at `2026-04-29T06:44:54Z` with merge commit `0acb92e1ac82bd65194a0b03f9864bda5b657011`. |
| PR #191 required checks | PASS | Final PR head `b7775adb` completed 48/48 required checks successfully. |
| Post-merge main CI | PASS | Main push CI for the feature launch commit `0acb92e1` completed 12/12 push workflows successfully. The Linux workflow was recovered by rerunning only the cancelled failed jobs after two unit jobs hung. The readiness-record update merge commit `0f9abc98a33a935770ce9cdc44d96d61e24fc0c9` also completed 12/12 post-merge `main` push workflows successfully. |
| Branch protection | PASS | `main` uses strict required checks with 48 contexts, including `CodeQL (go)` and `CodeQL (javascript-typescript)`, with admin enforcement enabled. |
| PR conversations | PASS | PR #191 has 0 unresolved review conversations after the CodeQL log-injection threads were fixed and resolved. |
| CodeQL / code scanning | PASS | Open main code-scanning alerts by `rule.security_severity_level`: `142 medium`, `2 null`, `0 critical`, `0 high`. The PR-blocking `go/log-injection` alert for `management/server/policy.go` is not open on `main`. |
| Dependabot alerts | PASS | Open Dependabot alerts are `0`. |
| Issue board public hygiene | PASS | Open `priority:critical` issue count is `0`; open `priority:high` issue count is `0`. Issue #189 remains open as a medium upstream-sync tracking issue, but the v0.70.0 sync implementation has merged via PR #191. |
| Published releases | PASS | `gh release list` shows only draft `v0.1.0`; there is no published non-draft release that would become public accidentally. |
| RC artifact generation | PASS | Main Release workflow run `25094796270` for commit `0acb92e1` produced 7 non-expired artifacts, including `release`, `windows-packages`, `linux-packages`, and UI artifacts. |
| Checksums | PASS | Downloaded `release` artifact from run `25094796270`; `sha256sum -c netbird_0.1.0-SNAPSHOT-0acb92e1_checksums.txt` passed. |
| SBOM | PASS | `netbird-machine_0.1.0-SNAPSHOT-0acb92e1_windows_amd64.tar.gz.sbom.spdx.json` exists, is checksum-covered, and contains 86 packages. |
| Downloaded-artifact lab smoke | PASS | VM102 ran the downloaded `netbird-machine.exe` from the main Release artifact, not a local build. Active service binary SHA256: `b2121736b4cdaf8668d41ab42c8566452e31762efa221499174479cd49e9ff9a`. |
| Machine Tunnel lab function | PASS | On VM102, `NetBirdMachine` was `Running`, process path was `C:\temp\netbird-machine.exe`, `wg-nb-machine` was `Up`, IPv4 `100.95.231.226/16` was present, and routes including `192.168.100.0/24` were installed via the interface. Logs showed repeated `Interface wg-nb-machine verified` lines after restart. |
| Lab cleanup | PASS | Temporary Proxmox firewall rules, HTTP servers, upload server, and transfer binaries used for VM102 artifact transfer were removed after evidence capture. |

## Go Conditions

Public repository visibility is approved only when all of the following remain
true at the moment of the visibility change:

1. This refreshed decision record is merged to protected `main`.
2. The post-merge `main` CI for this decision-record update is green.
3. `gh repo view silentspike/netbird-machine-tunnel` still reports
   `visibility=PRIVATE` immediately before the flip.
4. `gh release list --repo silentspike/netbird-machine-tunnel --json tagName,isDraft`
   still shows no published non-draft release that would become public by
   accident.
5. Open Dependabot alerts remain `0`.
6. Open CodeQL critical/high alerts remain `0`.
7. Open `priority:critical` and `priority:high` issue counts remain `0` or every
   remaining item has an explicit non-blocking disposition.
8. The maintainer explicitly confirms the final visibility flip after reviewing
   this record.

## Release Hold

Do not publish a final tagged GitHub Release in the same step as the repository
visibility flip. The binary release should be promoted separately after
choosing:

1. public version/tag name,
2. release notes,
3. whether to publish a prerelease first,
4. whether to reuse the validated private Actions snapshot artifacts or generate
   a tag-triggered release run,
5. whether to run another downloaded-artifact lab smoke for the final tagged
   release candidate.

Until then, the final validated launch artifact is the private Actions snapshot
from main Release run `25094796270` at commit
`0acb92e1ac82bd65194a0b03f9864bda5b657011`.

## Issue Disposition

| Issue | State |
|-------|-------|
| #149 | Closed completed after upstream NetBird v0.69.0 sync landed on protected `main`. |
| #166 | Closed completed; follow-up comment records current 48 strict branch-protection contexts including both CodeQL contexts. |
| #167 | Closed completed after dependency remediation, no-patch/transitive alert disposition, and default-branch open Dependabot alert count `0`. |
| #168 | Closed completed after final CI, draft-release exposure check, dependency/CodeQL status, and artifact/lab evidence were recorded. |
| #170 | Closed completed after CodeQL PR scanning was re-enabled, branch protection required both CodeQL contexts, and the final baseline reached `0` critical/high. |
| #189 | Open tracking issue for upstream NetBird v0.70.0 sync; implementation merged via PR #191 and can be closed after the maintainer accepts the final disposition. |

## Final Position

The project is **GO for public repository visibility** after this refreshed
record lands on protected `main`, the resulting post-merge `main` CI is green,
and the maintainer confirms the final visibility flip.

The project is **HOLD for a public tagged binary release** until the explicit
release-promotion decision above is made.

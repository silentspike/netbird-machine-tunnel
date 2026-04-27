# Public Readiness Execution Progress

Status: BLOCKED_PUBLIC_GO_LIVE
Started: 2026-04-27
Repo: `silentspike/netbird-machine-tunnel`
Branch at start: `security/codeql-high-baseline`
Plan SSOT: `docs/internal/public-readiness-living-checklist.md`

This file is the `$start` execution tracker. It mirrors the living checklist
as executable tasks. Every completed task must have fresh evidence and one
atomic commit before the next task starts.

## Current Status

- Hooks: registered in project-local `.claude/settings.json`.
- Context: global rules, `/work` rules, `/work/jobs/AGENTS.md`, relevant memory,
  and the living checklist were re-read on 2026-04-27.
- Pull policy: no `main` pull was performed.
- Public release: NO-GO until the listed public readiness gates pass.

## Task Table

| # | Task | Scope | Status | Commit |
|---|------|-------|--------|--------|
| 1 | Update #168 with main CI evidence | Post the verified PR #183 `main` push green result to issue #168 and record the evidence. | DONE | task commit |
| 2 | Continue #170 CodeQL baseline | Push `security/codeql-high-baseline`, run CodeQL on that ref, inspect result, and prepare the OIDC SSRF policy decision. | DONE | task commit |
| 3 | Fix README license wording | Align public README License section with `LICENSE` and `NOTICE.md` dual-license structure. | DONE | task commit |
| 4 | Update ADR statuses | Mark ADR-001 superseded/amended and ADR-002 implemented with implementation evidence. | DONE | task commit |
| 5 | Create `FORK_DIFF.md` | Document fork-specific contribution from verified paths and upstream diff, then link it from README. | DONE | task commit |
| 6 | Promote CRL limitation | Move "No CRL checking" into the public security surface with scope and mitigations. | DONE | task commit |
| 7 | Reassess #108 and #109 | Verify current code/lab behavior and decide blocker vs known limitation vs stale closeable issue. | DONE | task commit |
| 8 | Continue RC gates | Validate RC artifacts, checksums, SBOMs, `netbird-machine.exe`, and downloaded-artifact lab smoke. | DONE | task commit |
| 9 | Prepare public Go/No-Go | Produce final evidence-backed public release decision and remaining blockers. | DONE | task commit |
| 10 | Plan-Verifikation | Reread the plan line by line, compare implementation, run final required checks, and update this file. | DONE | task commit |
| 11 | Fix fork-owned CodeQL follow-up | Classify remaining critical/high CodeQL findings by fork ownership, fix fork-added/fork-modified findings locally, and rerun CodeQL. | DONE | task commit |
| 12 | Patch dependency alerts | Use the reliable GraphQL Dependabot export, patch actionable Go/NPM dependency alerts, and document no-patch dispositions. | DONE | task commit |

## Task Details

### Task 1: Update #168 with main CI evidence

Scope:
- Recheck live `main` Actions for commit `bb2682231cb2ff3f191f51c691f95459e6f9921f`.
- Post a concise evidence comment to issue #168.
- Update `docs/internal/public-readiness-living-checklist.md` inline.
- Update this `PROGRESS.md`.

Acceptance criteria:
- AC1: Live `main` push workflows for the PR #183 merge commit are all completed successfully.
  - Evidence: `gh run list` command output summarized in this file.
- AC2: Issue #168 receives a new comment with the exact evidence summary.
  - Evidence: GitHub issue comment URL or latest-comment check.
- AC3: Internal living checklist records the posted update.
  - Evidence: `rg`/line inspection of `docs/internal/public-readiness-living-checklist.md`.
- AC4: No unrelated files are changed.
  - Evidence: `git status --short` review.

### Task 2: Continue #170 CodeQL baseline

Scope:
- Push `security/codeql-high-baseline`.
- Trigger CodeQL by `workflow_dispatch` for that ref.
- Inspect CodeQL result and record whether high/critical findings decrease.
- Prepare but do not rush the OIDC issuer SSRF policy decision.

Acceptance criteria:
- AC1: Branch is present on origin at the local head.
  - Evidence: `git ls-remote` or `git rev-parse` comparison.
- AC2: CodeQL workflow run for the branch is started and reaches a terminal state.
  - Evidence: `gh run list/view` output.
- AC3: Findings delta is recorded for #170/public go-live.
  - Evidence: Code scanning API summary.
- AC4: OIDC SSRF policy is documented as decision-needed or implemented with tests.
  - Evidence: issue comment or code/test evidence depending on outcome.

Pre-task self-check:
- Must push the current `security/codeql-high-baseline` branch to origin.
- Must trigger CodeQL explicitly on that ref.
- Must wait for CodeQL to reach a terminal state before claiming AC2.
- Must inspect code scanning state after the run.
- Must not blindly change OIDC issuer SSRF behavior; private/internal issuers can
  be legitimate in self-hosted deployments.
- Expected changed files before commit: `PROGRESS.md` only, unless a CodeQL
  follow-up fix is required by fresh evidence.

### Task 3: Fix README license wording

Scope:
- Read current `README.md`, `LICENSE`, `NOTICE.md`, and `LICENSES/REUSE.toml`.
- Change README License section to match the dual-license structure.
- Keep wording clear for public readers and consistent with NetBird attribution.

Acceptance criteria:
- AC1: README no longer claims the entire repo is AGPL-only.
  - Evidence: `rg` command against README License section.
- AC2: README points to `LICENSE`, `NOTICE.md`, and component license locations.
  - Evidence: `rg` command output.
- AC3: License check still passes.
  - Evidence: repository license/dependency workflow or local available check.

### Task 4: Update ADR statuses

Scope:
- Update `docs/ADR-001-mTLS-Port-Strategy.md`.
- Update `docs/ADR-002-CNG-Signer-Interface.md`.
- Add implementation evidence links that are true in current code.

Acceptance criteria:
- AC1: ADR-001 no longer presents stale single-port routing as the active final state without caveat.
  - Evidence: `rg` command over ADR status and supersession/amendment text.
- AC2: ADR-002 status reflects implementation state.
  - Evidence: `rg` command over ADR status and evidence section.
- AC3: Referenced implementation paths exist.
  - Evidence: `test -e` or `rg --files` output.

### Task 5: Create `FORK_DIFF.md`

Scope:
- Create public root `FORK_DIFF.md`.
- Build content from verified paths and upstream comparison, not only marker comments.
- Link it from README near the top.

Acceptance criteria:
- AC1: `FORK_DIFF.md` exists and summarizes client, auth, management, tests, docs, and workflow deltas.
  - Evidence: file existence and targeted `rg`.
- AC2: No false claim that all fork additions are marker-commented.
  - Evidence: `rg` against `FORK_DIFF.md`.
- AC3: README links to `FORK_DIFF.md`.
  - Evidence: `rg` against README.

### Task 6: Promote CRL limitation

Scope:
- Add prominent public security limitation for missing CRL checking.
- Include mitigations and scope.
- Link from README or release/go-live docs if appropriate.

Acceptance criteria:
- AC1: Public security surface contains a dedicated CRL limitation section.
  - Evidence: `rg` against `SECURITY.md` or selected public doc.
- AC2: Mitigations include certificate lifetime/rotation and issuer/account constraints.
  - Evidence: `rg` output.
- AC3: README known limitation remains consistent.
  - Evidence: `rg` output.

### Task 7: Reassess #108 and #109

Scope:
- Re-read both issues.
- Inspect current code paths for WireGuard peer config, Signal, and Relay behavior.
- Use lab/runtime evidence where required before changing issue state.

Acceptance criteria:
- AC1: Current issue state and latest comments are captured.
  - Evidence: `gh issue view` output.
- AC2: Current implementation state is checked with commands/tests.
  - Evidence: targeted tests or lab smoke output.
- AC3: Each issue has a documented disposition: blocker, known limitation, or stale/closeable.
  - Evidence: issue comment or internal decision record.

### Task 8: Continue RC gates

Scope:
- Validate public release candidate artifacts as a consumer would.
- Verify checksums, SBOMs, release notes, and `netbird-machine.exe`.
- Install downloaded RC artifact in lab and run Machine Tunnel smoke.

Acceptance criteria:
- AC1: Artifact list includes fork-specific `netbird-machine.exe`.
  - Evidence: downloaded artifact listing.
- AC2: Checksums and SBOMs are verified or missing-state is explicitly blocked.
  - Evidence: checksum/SBOM commands.
- AC3: Lab smoke uses downloaded artifact, not local build.
  - Evidence: transfer/install command output and smoke result.

### Task 9: Prepare public Go/No-Go

Scope:
- Gather final state for security, CI, branch protection, dependencies, docs, release artifacts, lab, and visibility.
- Produce evidence-backed decision.
- Do not flip public visibility unless all gates pass and user explicitly approves.

Acceptance criteria:
- AC1: Every public readiness blocker has final PASS/BLOCKED/ACCEPTED-RISK state.
  - Evidence: decision record.
- AC2: User approval is recorded before visibility flip.
  - Evidence: explicit user approval in conversation or issue record.
- AC3: Visibility remains private unless AC1 and AC2 are satisfied.
  - Evidence: `gh repo view` output.

### Task 10: Plan-Verifikation

Scope:
- Reread `docs/internal/public-readiness-living-checklist.md`.
- Compare each checkbox/task against implementation and evidence.
- Run final required live checks.
- Mark this tracker `COMPLETE` or `BLOCKED`.

Acceptance criteria:
- AC1: Plan was reread and each task checked.
  - Evidence: updated findings table.
- AC2: Final verification commands were run.
  - Evidence: command outputs summarized.
- AC3: Remaining blockers are explicit.
  - Evidence: `Blocked Items` section updated.

## Blocked Items

- Public visibility flip is blocked until #168 RC gates and #170 CodeQL/security
  disposition are complete or explicitly accepted in the go-live decision.
- #108 original peer-configuration claim is stale/closeable after maintainer
  approval; #109 functional Signal/Relay connectivity is implemented and
  lab-validated, but remains a public go-live blocker until #114 Signal Server
  trust-model review is resolved or explicitly split out.
- #170 remains open after Task 11 remote CodeQL: fork-added/fork-modified
  high findings are cleared on `security/codeql-high-baseline`. The branch now
  has 145 open alerts: 1 critical, 0 high, 142 medium, 2 warning. The remaining
  `go/request-forgery` finding in `management/server/identity_provider.go` is
  unchanged from upstream `v0.69.0` and must be dispositioned as inherited
  upstream risk instead of blindly patched in the fork.
- Public visibility remains blocked after Task 9/10 by #170 CodeQL
  critical/high disposition, #167 unstable Dependabot alert export, #114 Signal
  trust-model review, missing final approval, no final tagged public-launch
  release, and local readiness commits not yet merged to `main`.
- #167 remains open after Task 12 local remediation because Dependabot alerts
  are default-branch based. The branch patches actionable alerts locally, but
  the final close decision needs a protected-main merge, default-branch rescan,
  and explicit disposition for no-patch/transitive alerts.

## Findings

- 2026-04-27: `main` CI for PR #183 merge commit is green and should be posted
  to #168 as the first execution task.
- 2026-04-27: README License section conflicts with `LICENSE`/`NOTICE.md` by
  presenting the whole repo as AGPL-only.
- 2026-04-27: fork marker coverage is partial; `FORK_DIFF.md` must not rely only
  on `MACHINE-TUNNEL-FORK` comments.
- 2026-04-27: CodeQL branch `security/codeql-high-baseline` reduced open branch
  alerts from `main` 164 total / 19 high to branch 148 total / 3 high; the
  critical `go/request-forgery` remains and needs an OIDC issuer SSRF policy.
- 2026-04-27: README License section now matches the repository dual-license
  structure from `LICENSE` and `NOTICE.md`; local internal and external license
  dependency checks passed.
- 2026-04-27: ADR-001 is now explicitly superseded by the dedicated mTLS port
  implementation; ADR-002 is now implemented with current auth implementation
  evidence.
- 2026-04-27: `FORK_DIFF.md` now gives reviewers a verified, narrow summary of
  fork-specific Machine Tunnel additions and is linked from README.
- 2026-04-27: CRL/OCSP revocation checking is now a prominent public security
  limitation in `SECURITY.md`, with README pointing readers to the detailed
  scope and mitigations.
- 2026-04-27: #108/#109 were reassessed with current code, tests, and live lab
  evidence. #108 is stale/closeable after maintainer approval; #109 remains
  open because #114 trust-model review is still open.
- 2026-04-27: RC artifact gate passed for the `main` Actions snapshot artifact
  from run `24984503525`: fork-specific `netbird-machine` archive and SBOM are
  present, full checksums pass, and VM102 smoke passed after installing the
  downloaded artifact binary.
- 2026-04-27: `docs/PUBLIC-RELEASE-READINESS.md` records an explicit NO-GO for
  public visibility/release while allowing continued private preparation.
- 2026-04-27: remaining #170 critical/high findings were classified by fork
  ownership. `management/internals/shared/mtls/dnslabel.go` is fork-added,
  `management/server/geolocation/utils.go` is fork-modified, and
  `management/server/identity_provider.go` has no fork delta against upstream
  `v0.69.0`.
- 2026-04-27: local #170 follow-up fixes are complete: DNSLabel suffix
  generation now uses deterministic HMAC-SHA256 with a 64-bit suffix,
  geolocation archive extraction writes only expected filenames and skips
  traversal entries, and the branch-local gRPC sync-limit parser now uses
  `strconv.ParseInt(..., 32)` to satisfy targeted lint/security checks.
- 2026-04-27: CodeQL run `25009836010` on
  `security/codeql-high-baseline` passed for Go and JavaScript/TypeScript.
  Code-scanning API for the branch now reports `145` open alerts:
  `1 critical`, `0 high`, `142 medium`, `2 warning`. #170 was updated with the
  rerun evidence:
  `https://github.com/silentspike/netbird-machine-tunnel/issues/170#issuecomment-4329195301`.
- 2026-04-27: Dependabot REST returned `0` open alerts, but GraphQL
  `vulnerabilityAlerts(states: OPEN)` returned the authoritative default-branch
  set of 12 alerts: 1 high, 10 moderate, 1 low. Local branch remediation
  updates patched Go modules and `proxy/web/package-lock.json`; remaining
  no-patch/transitive alerts need final disposition after main rescan.

## Task Evidence

### Task 1: Update #168 with main CI evidence

Status: DONE, committed in the Task 1 commit.

Pre-task self-check:
- Must prove the live `main` push workflows for PR #183 merge commit are green.
- Must post that evidence to issue #168.
- Must update the internal living checklist and this tracker.
- Expected tracked file change: `PROGRESS.md`.
- Expected ignored file change: `docs/internal/public-readiness-living-checklist.md`.

AC results:
- AC1 PASS: `gh run list --repo silentspike/netbird-machine-tunnel --branch main`
  for commit `bb2682231cb2ff3f191f51c691f95459e6f9921f` returned `completed`
  and `success` for `FreeBSD`, `Test installation`, `Secret Scan`,
  `Check License Dependencies`, `Windows`, `Release`, `Linux`, `Darwin`,
  `Mobile`, `Wasm`, `Test Infrastructure files`, and `CodeQL`.
- AC2 PASS: issue #168 received comment
  `https://github.com/silentspike/netbird-machine-tunnel/issues/168#issuecomment-4328257712`.
  A latest-comment recheck returned author `obtFusi`, timestamp
  `2026-04-27T15:27:19Z`, and the expected evidence body.
- AC3 PASS: `docs/internal/public-readiness-living-checklist.md` now marks
  "Update #168 with the new main CI green evidence" as checked and includes the
  issue comment URL.
- AC4 PASS: worktree review shows the only tracked normal file change for this
  task is `PROGRESS.md`; `.claude/settings.json` and `docs/internal/` are
  ignored local execution/handoff files.

### Task 2: Continue #170 CodeQL baseline

Status: DONE, committed in the Task 2 commit. #170 remains a public go-live blocker.

Pre-task self-check:
- Must push the current `security/codeql-high-baseline` branch to origin.
- Must trigger CodeQL explicitly on that ref.
- Must wait for CodeQL to reach a terminal state before claiming AC2.
- Must inspect code scanning state after the run.
- Must not blindly change OIDC issuer SSRF behavior; private/internal issuers can
  be legitimate in self-hosted deployments.
- Expected changed files before commit: `PROGRESS.md` only, unless a CodeQL
  follow-up fix is required by fresh evidence.

AC results:
- AC1 PASS: `git ls-remote --heads origin security/codeql-high-baseline` and
  `git rev-parse HEAD` both returned
  `a579940b5024b2e805bd17a07e712ef1381b1057`.
- AC2 PASS: manual CodeQL run
  `https://github.com/silentspike/netbird-machine-tunnel/actions/runs/25004182847`
  completed with conclusion `success`; `CodeQL (go)` and
  `CodeQL (javascript-typescript)` both completed successfully.
- AC3 PASS: code-scanning API was queried for `refs/heads/main` and
  `refs/heads/security/codeql-high-baseline`. Open alerts changed from
  `main`: 164 total, 1 critical, 19 high, 142 medium, 2 warning to branch:
  148 total, 1 critical, 3 high, 142 medium, 2 warning. Remaining
  high/critical branch alerts are:
  `go/request-forgery` at `management/server/identity_provider.go:45`,
  `go/zipslip` at `management/server/geolocation/utils.go:37`,
  `go/zipslip` at `management/server/geolocation/utils.go:75`, and
  `go/weak-sensitive-data-hashing` at
  `management/internals/shared/mtls/dnslabel.go:47`.
- AC4 PASS: issue #170 received comment
  `https://github.com/silentspike/netbird-machine-tunnel/issues/170#issuecomment-4328369339`.
  The latest-comment recheck confirmed the branch name, `148 total` delta, and
  "OIDC SSRF policy decision needed" wording.

### Task 3: Fix README license wording

Status: DONE, committed in the Task 3 commit.

Pre-task self-check:
- Must change only the public-facing README License section and progress
  tracking unless verification finds a related license-file issue.
- Must align README wording with `LICENSE`, `NOTICE.md`, and
  `LICENSES/REUSE.toml`.
- Must prove the old AGPL-only wording is gone.
- Must prove README links to license and attribution files.
- Must run available local license checks.

AC results:
- AC1 PASS: `rg -n "AGPL-only|Licensed under \\*\\*GNU Affero|dual-license|BSD-3-Clause|management/LICENSE|signal/LICENSE|relay/LICENSE|combined/LICENSE|NOTICE.md" README.md`
  showed the new dual-license wording and did not show the previous
  "Licensed under **GNU Affero..." AGPL-only sentence.
- AC2 PASS: README now links to `[LICENSE](LICENSE)`, `[NOTICE.md](NOTICE.md)`,
  `[management/LICENSE](management/LICENSE)`, `[signal/LICENSE](signal/LICENSE)`,
  `[relay/LICENSE](relay/LICENSE)`, and `[combined/LICENSE](combined/LICENSE)`.
- AC3 PASS: local internal workflow check returned
  `PASS internal AGPL dependency check`. Local external workflow-equivalent
  `go-licenses` check reported GPL dependencies
  `github.com/netbirdio/management-integrations/integrations` and
  `goauthentik.io/api/v3`, confirmed both are only used by internal AGPL
  packages, and returned `PASS external GPL/AGPL license dependency check`.

### Task 4: Update ADR statuses

Status: DONE, committed in the Task 4 commit.

Pre-task self-check:
- Must update only ADR-001, ADR-002, and progress tracking.
- Must preserve historical context while making the current implementation state
  clear.
- Must prove ADR-001 no longer presents single-port routing as the active final
  state without caveat.
- Must prove ADR-002 no longer presents CNG signer work as blocked/pending.
- Must prove referenced implementation paths exist and compile/test at package
  level where feasible.

AC results:
- AC1 PASS: `rg -n "Status:|Superseded|Historical Implementation Sketch|Current Implementation Evidence|management_grpc.pb.go|RegisterMachinePeer" docs/ADR-001-mTLS-Port-Strategy.md`
  shows ADR-001 status `Superseded by dedicated mTLS port implementation`, the
  historical single-port sketch is labeled, and current evidence references the
  dedicated mTLS server and generated Machine Tunnel RPC code.
- AC2 PASS: `rg -n "Status.*Implemented|Original Interface Sketch|shipped implementation|Implementation Evidence|WinCertSigner|CryptAcquireCertificatePrivateKey|Remaining Validation" docs/ADR-002-CNG-Signer-Interface.md`
  shows ADR-002 status `Implemented`, labels the old snippet as original sketch,
  and references the shipped Windows signer/certificate discovery implementation.
- AC3 PASS: `test -e` confirmed every referenced implementation path exists.
  `go test ./client/internal/auth ./management/internals/server` returned:
  `ok github.com/netbirdio/netbird/client/internal/auth` and
  `ok github.com/netbirdio/netbird/management/internals/server`.

### Task 5: Create `FORK_DIFF.md`

Status: DONE, committed in the Task 5 commit.

Pre-task self-check:
- Must create a public root `FORK_DIFF.md`.
- Must use verified path inventory and `git diff --name-status v0.69.0...HEAD`
  evidence, not only marker comments.
- Must not claim that all fork additions are marked with `MACHINE-TUNNEL-FORK`.
- Must link the file from README.

AC results:
- AC1 PASS: `test -f FORK_DIFF.md` and targeted `rg` found the document title,
  upstream baseline `v0.69.0`, client/service/auth/server/proto/test sections,
  and the known security limitation section.
- AC2 PASS: `rg` check for false marker completeness claims returned
  `PASS no false marker completeness claim`; the file states that
  `MACHINE-TUNNEL-FORK` markers are not a complete index.
- AC3 PASS: `rg -n "FORK_DIFF.md|Fork Contribution Summary" README.md`
  returned the intro link, table-of-contents link, and documentation-table link.
  `git diff --name-status v0.69.0...HEAD -- ...` confirmed the representative
  Machine Tunnel paths used by the document.

### Task 6: Promote CRL limitation

Status: DONE, committed in the Task 6 commit.

Pre-task self-check:
- Must add a dedicated public CRL/revocation limitation section to the security
  surface, preferably `SECURITY.md`.
- Must keep README known limitations consistent and link readers to the
  detailed security posture.
- Must explicitly cover scope and mitigations: certificate lifetime/rotation,
  issuer fingerprint constraints, and per-account AllowedDomains/account-domain
  constraints.
- Expected tracked file changes: `SECURITY.md`, `README.md`, `PROGRESS.md`.
- Expected ignored file change: `docs/internal/public-readiness-living-checklist.md`.

AC results:
- AC1 PASS: `rg -n "Known Security Limitations|Certificate Revocation Checking|CRL|OCSP|dedicated mTLS endpoint|upstream NetBird authentication" SECURITY.md`
  found the dedicated public limitation section and scope text in `SECURITY.md`.
- AC2 PASS: `rg -n "short machine-certificate lifetimes|routine certificate rotation|per-account AllowedDomains|issuer fingerprint constraints|account/domain mapping|accepted issuer fingerprint" SECURITY.md README.md FORK_DIFF.md`
  found certificate lifetime/rotation, issuer fingerprint constraints, and
  account/domain or per-account AllowedDomains mitigations in the public docs.
- AC3 PASS: `rg -n "No CRL checking|Certificate Revocation Checking|SECURITY.md#certificate-revocation-checking|CRL|OCSP|short certificate lifetimes|per-account AllowedDomains" README.md`
  found the README known-limitation row with a link to the detailed
  `SECURITY.md` section.
- Hygiene PASS: `git diff --check` returned clean; `git status --short` showed
  only `PROGRESS.md`, `README.md`, and `SECURITY.md` as tracked changes for the
  task. The internal checklist update remains ignored under `docs/internal/`.

### Task 7: Reassess #108 and #109

Status: DONE, committed in the Task 7 commit.

Pre-task self-check:
- Must read live GitHub issue state and latest discussion for #108 and #109.
- Must inspect current code paths for the behavior each issue claims, rather
  than relying on old plan text.
- Must use targeted tests or lab/runtime evidence where the issue cannot be
  dispositioned from code and existing evidence alone.
- Must record each issue as blocker, known limitation, or stale/closeable with
  evidence.
- Expected tracked file change before commit: `PROGRESS.md` only, unless code or
  public documentation must change based on fresh evidence.
- Expected ignored file change: `docs/internal/public-readiness-living-checklist.md`.

AC results:
- AC1 PASS: live issue state was captured with `gh issue view` for #108, #109,
  #113, and #114. #108 remains `OPEN` and says it is blocked by #109. #109
  remains `OPEN`; its latest prior status says E2E tests passed except #114.
  #110, #111, #112, and #113 are `CLOSED`; #114 remains `OPEN`.
- AC2 PASS: current implementation state was checked from code, tests, and lab:
  `machine.go` builds Signal/Relay config, starts `PeerEngine`, and calls
  `connectToRemotePeers`; `peerengine.go` reuses `peer.Conn` with Signal,
  Relay, `SRWatcher`, and status dependencies. `GOOS=windows GOARCH=amd64 go
  test -c -o /tmp/netbird-tunnel-task7.test.exe ./client/internal/tunnel`
  produced a 38M Windows test binary, and `go test ./client/internal/tunnel -run
  'TestHealth|TestReconnect|TestBootstrap|TestMTLS|TestTrust' -count=1`
  returned `ok`.
- AC2 PASS: live lab was reachable after starting VM102. Proxmox reported
  VM100/101/103 running and VM102 started successfully with QEMU Guest Agent
  ready. Management, Signal, Relay, and Zitadel DB containers on VM103 were up.
  Windows VM102 reported `NetBirdMachine` Running, `wg-nb-machine` Up,
  `100.95.231.226/16`, route `192.168.100.0/24` via `wg-nb-machine`, DC ports
  `53/88/389/445/636` with `TcpTestSucceeded: True`, and recent log evidence
  including `Received handshake response`.
- AC3 PASS: issue reassessment comments were posted:
  #108 `https://github.com/silentspike/netbird-machine-tunnel/issues/108#issuecomment-4328594963`
  and #109
  `https://github.com/silentspike/netbird-machine-tunnel/issues/109#issuecomment-4328595119`.
  Latest-comment checks confirmed both comments by `obtFusi` with expected body
  matches.
- Disposition PASS: #108 is `stale/closeable after maintainer approval` for the
  original peer-configuration claim. #109 is `functional Signal/Relay
  connectivity implemented and lab-validated`, but remains a public go-live
  blocker until #114 is resolved or explicitly split out.

### Task 8: Continue RC gates

Status: DONE, committed in the Task 8 commit.

Pre-task self-check:
- Must inspect current GitHub release/RC state before assuming an artifact
  exists.
- Must verify whether a downloadable release artifact includes
  `netbird-machine.exe`.
- Must verify checksums and SBOMs if release assets exist; if they do not exist,
  record a hard BLOCKED state instead of substituting local builds.
- Must use downloaded release artifacts for any lab smoke. Local workstation
  builds are not acceptable release evidence.
- Expected tracked file change before commit: `PROGRESS.md` only, unless a
  release workflow/docs bug must be fixed based on fresh evidence.
- Expected ignored file change: `docs/internal/public-readiness-living-checklist.md`.

AC results:
- AC1 PASS: `gh release list` showed the only visible GitHub Release is old
  `v0.1.0` from 2026-02-06, so the current RC validation used the successful
  `main` Actions snapshot run `24984503525` at commit
  `bb2682231cb2ff3f191f51c691f95459e6f9921f`. GitHub Actions artifact API
  showed non-expired artifacts including `release`, `linux-packages`,
  `windows-packages`, `macos-packages`, `release-ui`, and `release-ui-darwin`.
- AC1 PASS: downloaded the full `release` artifact from run `24984503525`.
  Artifact inventory contained
  `netbird-machine_0.1.0-SNAPSHOT-bb268223_windows_amd64.tar.gz`,
  `netbird-machine_0.1.0-SNAPSHOT-bb268223_windows_amd64.tar.gz.sbom.spdx.json`,
  and extracted `netbird-machine_windows_amd64_v1/netbird-machine.exe`.
- AC2 PASS: `sha256sum -c netbird_0.1.0-SNAPSHOT-bb268223_checksums.txt`
  passed for the complete downloaded artifact set, including `netbird-machine`
  and SBOM entries. Direct SHA256 for the downloaded binary was
  `be656553c08aaf620f7dde652223ed0909d77320805ed66423c973ef7ff645c9`.
  `file` identified it as `PE32+ executable for MS Windows`, and the
  `netbird-machine` archive contained `netbird-machine.exe` plus expected
  project files.
- AC3 PASS: VM102 lab smoke used the downloaded artifact binary, not a local
  build. Direct SCP to `admin@10.0.0.160` was used after the Proxmox HTTP hop
  timed out from the Windows guest. The service binary
  `C:\temp\netbird-machine.exe` was replaced with the downloaded artifact,
  service restarted, and its SHA256 matched
  `be656553c08aaf620f7dde652223ed0909d77320805ed66423c973ef7ff645c9`.
- AC3 PASS: post-restart lab smoke returned `NetBirdMachine` Running,
  `wg-nb-machine` Up, `100.95.231.226/16`, route `192.168.100.0/24` via
  `wg-nb-machine`, DC ports `53`, `88`, `389`, `445`, and `636` all
  `TcpTestSucceeded: True` via `wg-nb-machine`, and recent log evidence
  `Received handshake response`.
- AC3 PASS: issue #168 received the RC artifact update comment
  `https://github.com/silentspike/netbird-machine-tunnel/issues/168#issuecomment-4328746572`;
  latest-comment recheck confirmed author `obtFusi`, timestamp
  `2026-04-27T16:37:49Z`, and expected body match.
- Scope note: this validates the current `main` Actions snapshot artifact as an
  RC candidate. It is not yet a public GitHub Release/Pre-release asset.

### Task 9: Prepare public Go/No-Go

Status: DONE, committed in the Task 9 commit.

Pre-task self-check:
- Must gather current security, dependency, CI, artifact, lab, issue, and repo
  visibility state from live sources.
- Must produce an evidence-backed public release decision record.
- Must not change repository visibility without explicit user approval recorded
  after all gates are green or accepted risk.
- Must distinguish `GO for continued preparation` from `GO for public
  visibility/release`.
- Expected tracked file changes: `PROGRESS.md` and a public-safe decision record
  if needed.
- Expected ignored file change: `docs/internal/public-readiness-living-checklist.md`.

AC results:
- AC1 PASS: live repo, issue, branch-protection, CI, dependency-alert,
  code-scanning, artifact, and lab states were gathered. GitHub reports
  `silentspike/netbird-machine-tunnel` as `PRIVATE`; `main` branch protection
  has strict required checks with 46 contexts, admin enforcement, conversation
  resolution, force-push protection, deletion protection, and CODEOWNERS
  `errors=[]`.
- AC1 PASS: main CI for commit
  `bb2682231cb2ff3f191f51c691f95459e6f9921f` remains green for the core
  post-merge workflow set. Release artifact run `24984503525` and VM102
  downloaded-artifact smoke are recorded as PASS from Task 8.
- AC1 CORRECTED: an initial Dependabot API call returned `0` open alerts, but
  final verification repeated the call and observed inconsistent results (`0`
  and `12` open alerts, including at least one `postcss` alert). Issue #167 was
  updated with the correction and remains a public go-live blocker:
  `https://github.com/silentspike/netbird-machine-tunnel/issues/167#issuecomment-4328769698`.
- AC1 PASS: live CodeQL/code-scanning remains a hard blocker. `main` open
  alerts are still 1 critical, 19 high, 142 medium, and 2 warning. The
  `security/codeql-high-baseline` branch reduces this to 1 critical, 3 high,
  142 medium, and 2 warning, but #170 remains open until those findings are
  fixed, dismissed with evidence, or explicitly accepted.
- AC1 PASS: `docs/PUBLIC-RELEASE-READINESS.md` was created as the decision
  record. It records **NO-GO for public visibility and public release** and
  lists hard blockers: #170, #167, #114/#109 trust-model disposition, unmerged
  local readiness commits, missing final public approval, and missing final
  tagged public-launch release.
- AC2 PASS: no final public visibility/release approval was recorded in this
  task after the current blockers were known.
- AC3 PASS: `gh repo view` reports `visibility=PRIVATE`; no repository
  visibility change was performed.

### Task 10: Plan-Verifikation

Status: DONE, committed in the Task 10 commit. Overall public go-live remains
BLOCKED.

Pre-task self-check:
- Must reread `docs/internal/public-readiness-living-checklist.md` and compare
  remaining unchecked items against completed tasks.
- Must verify current tracked/ignored worktree state and repo visibility.
- Must verify whether any public-release blocker remains.
- Must update `PROGRESS.md` to the honest final execution state: complete for
  executed tasks, blocked for public go-live if blockers remain.
- Expected tracked file change before commit: `PROGRESS.md`.
- Expected ignored file change: `docs/internal/public-readiness-living-checklist.md`
  only if final checklist notes need updating.

AC results:
- AC1 PASS: `docs/internal/public-readiness-living-checklist.md` was reread and
  scanned for unchecked items. Completed execution items from Tasks 1-9 are
  recorded. Remaining unchecked items are deliberate NO-GO/deferred work, not
  skipped execution: license metadata recheck, #166 close decision, #167 stable
  dependency-alert disposition, #170 OIDC/CodeQL disposition, CRL release-note
  link, public first-screen polish, visibility flip, post-public checks, and
  final tagged public release artifacts.
- AC2 PASS: final live checks were run: repo visibility remains `PRIVATE`; open
  issues still include #170, #168, #167, #166, #114, #109, and #108; CODEOWNERS
  API returns `errors=[]`; branch protection has 46 strict required checks and
  admin enforcement; `security/codeql-high-baseline` code-scanning state remains
  1 critical, 3 high, 142 medium, 2 warning.
- AC2 PASS: final verification also caught and corrected the Dependabot alert
  disposition. A previous Task 9 comment claiming `0` open alerts was updated
  because repeated API checks produced inconsistent `0` and `12` results.
  #167 remains a public go-live blocker.
- AC3 PASS: remaining blockers are explicit in this file and
  `docs/PUBLIC-RELEASE-READINESS.md`: #170 CodeQL critical/high disposition,
  #167 stable dependency-alert disposition, #114/#109 Signal trust-model
  disposition, local readiness commits not merged to `main`, missing final
  public approval, and missing final tagged public-launch release.
- Hygiene PASS: worktree review before commit showed only tracked
  `PROGRESS.md` and `docs/PUBLIC-RELEASE-READINESS.md` plus ignored
  `docs/internal/` updates; the large downloaded artifact directory is outside
  the repository under `/work/vpn/.artifacts/task8-release-24984503525`.

### Task 11: Fix fork-owned CodeQL follow-up

Status: DONE. Public go-live still remains blocked by #170 final inherited-risk
disposition, #167, #114/#109, final mainline CI/release, and explicit approval.

Scope correction:
- The remaining critical/high CodeQL findings are not all fork-owned.
- `git diff --name-status v0.69.0...HEAD -- management/server/identity_provider.go management/server/geolocation/utils.go management/internals/shared/mtls/dnslabel.go`
  returned `A management/internals/shared/mtls/dnslabel.go` and
  `M management/server/geolocation/utils.go`; it returned no fork delta for
  `management/server/identity_provider.go`.
- Therefore `dnslabel.go` is fork-added, `geolocation/utils.go` is
  fork-modified, and `identity_provider.go` is inherited upstream NetBird
  surface. The OIDC SSRF finding should be dispositioned as inherited risk or
  tracked upstream, not blindly patched in the fork.

Local changes:
- `management/internals/shared/mtls/dnslabel.go`: replaced the raw SHA-256
  suffix slice with deterministic HMAC-SHA256 fingerprinting and increased the
  suffix to 64 bits / 16 hex chars.
- `management/server/geolocation/{database.go,utils.go}`: changed archive
  extraction to write only expected filenames
  (`GeoLite2-City.mmdb`, `GeoLite2-City-Locations-en.csv`) and skip traversal
  archive entries.
- `management/internals/shared/grpc/server.go`: fixed a branch-local
  `strconv.Atoi` narrowing/lint issue with `strconv.ParseInt(..., 32)`.

Local verification:
- PASS: `go test ./management/internals/shared/...`
- PASS: `go test ./management/server/geolocation`
- PASS: `golangci-lint run ./management/server/geolocation ./management/internals/shared/mtls ./management/internals/shared/grpc`
- PASS: `git diff --check`

Remote verification:
- PASS: pushed `security/codeql-high-baseline` at
  `926cb5b0a698f7de1b0086c45c04979ae05ea571`.
- PASS: CodeQL run
  `https://github.com/silentspike/netbird-machine-tunnel/actions/runs/25009836010`
  completed successfully for `CodeQL (javascript-typescript)` and
  `CodeQL (go)`.
- PASS: code-scanning API for
  `refs/heads/security/codeql-high-baseline` now reports 145 open alerts:
  1 critical, 0 high, 142 medium, 2 warning.
- PASS: issue #170 received the update:
  `https://github.com/silentspike/netbird-machine-tunnel/issues/170#issuecomment-4329195301`.

Remaining for #170:
- Formally disposition the remaining inherited upstream OIDC SSRF finding in
  `management/server/identity_provider.go`.
- Decide whether to accept it for public launch and/or track/report upstream.

### Task 12: Patch dependency alerts

Status: DONE. #167 remains open until protected `main` receives the dependency
remediation, GitHub refreshes default-branch alerts, and no-patch/transitive
alerts receive final dispositions.

Reliable export:
- REST endpoint
  `repos/silentspike/netbird-machine-tunnel/dependabot/alerts?state=open`
  returned `0` repeatedly despite Git push warnings.
- GraphQL `repository.vulnerabilityAlerts(states: OPEN)` returned 12
  default-branch alerts: 1 high, 10 moderate, 1 low.

Local remediation:
- Updated `github.com/okta/okta-sdk-golang/v2` to `v2.20.0`; `go mod why -m
  gopkg.in/square/go-jose.v2` now reports that the main module does not need
  `gopkg.in/square/go-jose.v2`.
- Updated patched Go dependencies including `github.com/quic-go/quic-go`,
  `github.com/pion/dtls/v3`, `golang.org/x/image`,
  `github.com/aws/aws-sdk-go-v2/aws/protocol/eventstream`,
  `github.com/aws/aws-sdk-go-v2/service/s3`, `github.com/jackc/pgx/v5`, and
  `github.com/Azure/go-ntlmssp`.
- Updated `proxy/web/package-lock.json`; `postcss` is now `8.5.12`.
- Updated `github.com/pion/dtls/v2` to latest v2 `v2.2.12`; advisory still has
  no patched v2 version.

Local verification:
- PASS: `go test ./management/server/idp ./relay/server/listener/quic ./client/internal/relay ./client/iface/bind ./upload-server/server ./management/server/store ./idp/dex`
- PASS: `npm --prefix proxy/web audit --audit-level=low`
- PASS: `git diff --check`

Remote verification:
- PASS: pushed `security/codeql-high-baseline` at
  `3e6e791eea352472705ba72c6e614d87e4fc58f7`.
- PASS: CodeQL run
  `https://github.com/silentspike/netbird-machine-tunnel/actions/runs/25010818374`
  completed successfully for `CodeQL (go)` and
  `CodeQL (javascript-typescript)`.
- PASS: code-scanning API for
  `refs/heads/security/codeql-high-baseline` remains 145 open alerts:
  1 critical, 0 high, 142 medium, 2 warning.
- PASS: issue #167 received the dependency remediation update:
  `https://github.com/silentspike/netbird-machine-tunnel/issues/167#issuecomment-4329357241`.

Remaining for #167:
- Let protected-main/default-branch dependency scanning refresh after merge.
- Dismiss or explicitly accept no-patch/transitive alerts with evidence,
  especially `github.com/docker/docker` via `management/server/testutil` and
  `github.com/pion/dtls/v2` via `client/internal/relay`.

## Commits

- Task 1: `Task 1: Update #168 with main CI evidence`
- Task 2: `Task 2: Continue #170 CodeQL baseline`
- Task 3: `Task 3: Fix README license wording`
- Task 4: `Task 4: Update ADR statuses`
- Task 5: `Task 5: Create fork contribution summary`
- Task 6: `Task 6: Promote CRL limitation`
- Task 7: `Task 7: Reassess #108 and #109`
- Task 8: `Task 8: Continue RC gates`
- Task 9: `Task 9: Prepare public Go/No-Go`
- Task 10: `Task 10: Plan-Verifikation`
- Task 11: `Task 11: Fix fork-owned CodeQL findings`
- Task 12: `Task 12: Patch dependency alerts`

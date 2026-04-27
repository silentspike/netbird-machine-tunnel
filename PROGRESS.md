# Public Readiness Execution Progress

Status: IN_PROGRESS
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
| 7 | Reassess #108 and #109 | Verify current code/lab behavior and decide blocker vs known limitation vs stale closeable issue. | PENDING | |
| 8 | Continue RC gates | Validate RC artifacts, checksums, SBOMs, `netbird-machine.exe`, and downloaded-artifact lab smoke. | PENDING | |
| 9 | Prepare public Go/No-Go | Produce final evidence-backed public release decision and remaining blockers. | PENDING | |
| 10 | Plan-Verifikation | Reread the plan line by line, compare implementation, run final required checks, and update this file. | PENDING | |

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
- #108/#109 cannot be treated as non-showstoppers until current code/lab
  evidence supports that decision.
- #170 remains open after Task 2: the branch reduced High alerts, but
  `go/request-forgery`, two `go/zipslip` alerts, and one
  `go/weak-sensitive-data-hashing` alert remain.

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

## Commits

- Task 1: `Task 1: Update #168 with main CI evidence`
- Task 2: `Task 2: Continue #170 CodeQL baseline`
- Task 3: `Task 3: Fix README license wording`
- Task 4: `Task 4: Update ADR statuses`
- Task 5: `Task 5: Create fork contribution summary`
- Task 6: `Task 6: Promote CRL limitation`

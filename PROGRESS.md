# Upgrade Execution Progress: NetBird Fork v0.66.4 -> v0.69.0

Status: IN_PROGRESS
Started: 2026-04-24
Branch: `sync/upstream-v0.69.0`
Plan: `docs/internal/upgrade-plan-v0.69.0.md`
Lessons learned: `docs/internal/upgrade-lessons-learned.md`
GitHub issue: `#149` - `[deps] sync upstream NetBird v0.69.0`
Backup tag: `backup-pre-sync-v0.69.0-20260424-1017`

## Execution Rules Loaded

- Project rules: `/work/vpn/.claude/CLAUDE.md`
- Global rules: `/home/jan/.claude/CLAUDE.md`
- Start skill: `/home/jan/.codex/skills/start/SKILL.md`
- Memory files: none found under `/work/vpn` at start

## Hook Setup

Project-local hooks registered in `.claude/settings.json`:

- `PreToolUse` / `TaskUpdate`: `~/bin/pretooluse-task-checklist-gate.sh`
- `PreToolUse` / `TaskUpdate`: `~/bin/pretooluse-start-progress-gate.sh`
- `PostToolUse` / catch-all: `~/bin/posttooluse-start-enforcer.sh`

## Current Findings

- Local `HEAD` and `origin/main` matched at start: `693c367caef77ddc4396ddf5a4641584f891a191`.
- No `git pull` was run during start; branch was created from the current local `main`.
- Existing `.gitignore` change adds `docs/upgrade.md`; treat it as pre-existing user/workspace state until Phase 0 resolves it.
- `.claude/settings.json` was created by the `$start` hook registration step and is intentionally ignored because it contains machine-local hook paths.
- Internal plan and lessons files are ignored by `.gitignore` through `docs/internal/`.
- Current fork and upstream release configs do not show `sbom`, `syft`, or `cyclonedx` entries in the checked release workflow/config files. AC-SEC-08 remains a release blocker until SBOM generation is added, verified through GoReleaser/upstream release changes, or explicitly documented with an approved alternative.
- Current `upstream-sync.yml` detects the latest tag but merges `upstream/main`; this must be fixed or disabled before the automation is trusted.

## Task Table

| Task | Scope | Status | Commit |
|---|---|---|---|
| 0. Phase 0: Preparation | Resolve preflight, issue, backup tag, tool and runner checks, SBOM/upstream-sync prechecks | DONE | this task commit |
| 1. Phase 1: Branch and Merge | Ensure upgrade branch, merge `v0.69.0` only | PENDING | pending |
| 2. Phase 2: Resolve Conflicts | Resolve predicted conflicts and inspect fork-sensitive auto-merges | PENDING | pending |
| 3. Phase 3: Implement Fork Adaptations | Bootstrap API, proto, release workflow, PeerEngine decision, Go version, docs | PENDING | pending |
| 4. Phase 4: Local Validation | Go tests, lint, Windows builds, management build, targeted tests | PENDING | pending |
| 5. Phase 5: Lab Deployment and E2E | Snapshots, management deploy, Windows fresh/upgrade, security and connectivity evidence | PENDING | pending |
| 6. Phase 6: PR and CI | Push branch, create PR, monitor required CI checks | PENDING | pending |
| 7. Phase 7: Merge Gate | Evidence summary and explicit user merge approval | PENDING | pending |
| 8. Phase 8: Post-Merge Main Smoke | Verify merged `main` before tagging | PENDING | pending |
| 9. Phase 9: Public Release Candidate | RC artifacts, checksums, SBOM, lab artifact smoke | PENDING | pending |
| 10. Phase 10: Go-Live Decision | Explicit GO/NO-GO with evidence and rollback state | PENDING | pending |
| 11. Phase 11: Public Release and Monitoring | User-approved release and 24h monitoring | PENDING | pending |
| 12. Plan-Verifikation | Line-by-line plan verification and final state | PENDING | pending |

## Task Details

### Task 0: Phase 0 - Preparation

Checklist:

- Resolve or intentionally isolate the existing `.gitignore` change.
- Confirm `upstream` push URL is disabled.
- Run required MainRag searches and record evidence.
- Verify local tools: `golangci-lint`, MinGW, `protoc`, proto plugins, `gitleaks`.
- Identify or create a GitHub issue in `obtFusi/netbird-fork`.
- Create and push a backup tag.
- Verify runner, lab, Actions cache, Actions permissions, upstream-sync risk, and SBOM generation state.

Acceptance criteria:

- AC0-1: Work is not on `main`; evidence: `git branch --show-current`.
- AC0-2: Dirty state is understood and no user changes are reverted; evidence: `git status --short --branch`, `git diff -- .gitignore`.
- AC0-3: `upstream` is fetch-only; evidence: `git remote -v`.
- AC0-4: Required tools are available; evidence: version/path commands.
- AC0-5: MainRag searches work; evidence: command output or artifact paths.
- AC0-6: GitHub issue is identified or created; evidence: issue URL/number.
- AC0-7: Backup tag exists locally and on origin; evidence: `git show-ref`, `git ls-remote`.
- AC0-8: Runner/lab/CI prechecks are current; evidence: `gh api`, SSH, workflow greps.

Evidence:

- AC0-1 PASS: `git branch --show-current` -> `sync/upstream-v0.69.0`.
- AC0-2 PASS: `git status --short --branch` showed `.gitignore` modified plus new `PROGRESS.md`; `git diff -- .gitignore` showed `docs/upgrade.md` and local `.claude/settings*.json` ignore rules only. No user change was reverted.
- AC0-3 PASS: `git remote -v` showed `upstream` fetch URL and `upstream DISABLED (push)`.
- AC0-4 PASS: tool evidence:
  - `golangci-lint` 2.11.4
  - `x86_64-w64-mingw32-gcc` at `/usr/bin/x86_64-w64-mingw32-gcc`, GCC 13.2.0
  - `protoc` `libprotoc 34.1`
  - `protoc-gen-go` v1.36.11
  - `protoc-gen-go-grpc` 1.6.1
  - `gitleaks` at `/usr/bin/gitleaks`, RPM package `gitleaks-8.30.1-1.1.x86_64`
- AC0-5 PASS: MainRag searches completed and were saved to:
  - `/work/vpn/.artifacts/mainrag-start-v069-01.log`
  - `/work/vpn/.artifacts/mainrag-start-v069-02.log`
  - `/work/vpn/.artifacts/mainrag-start-v069-03.log`
- AC0-6 PASS: GitHub issue created and normalized to labels `status:in-progress`, `priority:high`, `type:ci`, `upstream-sync`: `https://github.com/obtFusi/netbird-fork/issues/149`.
- AC0-7 PASS: backup tag `backup-pre-sync-v0.69.0-20260424-1017` exists locally and on `origin`, pointing to `693c367caef77ddc4396ddf5a4641584f891a191`.
- AC0-8 PASS with findings:
  - Proxmox host reachable, VMs 100-103 running, QGA OK for all after starting VM 102.
  - Management compose stack running; ports 80, 443, 33073, and 33074 listening.
  - Management logs show SQLite store engine, completed migrations, and mTLS Machine Tunnel server on 33074.
  - CT 150 runner services `github-runner.service` and `github-runner-netbird.service` active.
  - GitHub repo runner online with labels `self-hosted`, `Linux`, `X64`, `netbird-fork`.
  - Actions cache count is zero; first CI run is expected cold-cache.
  - Actions permissions currently allow all actions and do not require SHA pinning; public hardening item.
  - `upstream-sync.yml` still merges `upstream/main`; implementation must fix or explicitly disable it.
  - SBOM release generation was not found by grep in current fork release workflow/config or upstream v0.69.0 checked files; implementation must resolve AC-SEC-08.

### Task 1: Phase 1 - Branch and Merge

Checklist:

- Verify branch `sync/upstream-v0.69.0` is based on `origin/main`.
- Merge release tag `v0.69.0`, not `upstream/main`.
- Record conflict list.

Acceptance criteria:

- AC-SRC-01 and AC-SRC-02 have fresh command evidence.
- Merge target is tag `v0.69.0`.
- No unplanned branch strategy change occurs without user-visible decision.

### Task 2: Phase 2 - Resolve Conflicts

Checklist:

- Resolve predicted conflicts in workflow files, README, and generated proto.
- Inspect high-risk auto-merges.
- Run conflict-marker scan.

Acceptance criteria:

- AC-SRC-03, AC-SRC-04, AC-SRC-05, AC-SRC-06, and AC-SRC-07 have structural and command evidence.
- Proto generated files match merged `.proto`.

### Task 3: Phase 3 - Implement Fork Adaptations

Checklist:

- Update Machine Tunnel bootstrap for new management client API.
- Preserve and regenerate management proto.
- Merge release workflow guard logic and avoid unavailable runner labels.
- Decide/document PeerEngine `PortForwardManager` and `MetricsRecorder` behavior.
- Apply Go version decision.
- Update public docs, CHANGELOG, version, `llms.txt`, upstream-sync task/fix state, SBOM state.

Acceptance criteria:

- AC-SRC-04 through AC-SRC-08 are proven.
- Public repository ACs touched by docs/version changes are mapped for later verification.

### Task 4: Phase 4 - Local Validation

Checklist:

- Run `go test ./...`.
- Run focused fork tests.
- Run `golangci-lint run --timeout=12m`.
- Build Windows Machine Tunnel binary with standard and CGO paths as applicable.
- Build management binary and verify executable type.
- Run secret scan.

Acceptance criteria:

- AC-BLD-01 through AC-BLD-06 and AC-SEC-06 have command evidence.
- Any failing validation blocks downstream tasks unless explicitly isolated.

### Task 5: Phase 5 - Lab Deployment and E2E

Checklist:

- Snapshot lab VMs.
- Deploy management and Windows artifacts.
- Execute lab matrix evidence pack.
- Run fresh install, upgrade-in-place, mTLS, DPAPI, DC, router, DNS, firewall, NRPT, DB/store, metrics, and redaction checks.

Acceptance criteria:

- AC-MT-01 through AC-MT-10 and AC-SEC-01 through AC-SEC-07 have evidence.
- Evidence pack contains command output and redaction scan.

### Task 6: Phase 6 - PR and CI

Checklist:

- Create persistent PR body under `/work/vpn/.artifacts`.
- Push branch to `origin`.
- Create PR against `obtFusi/netbird-fork`.
- Monitor PR checks with branch filter.

Acceptance criteria:

- AC-BLD-07 and AC-PUB-07 have GitHub evidence.
- CI failures are triaged with logs, not guessed.

### Task 7: Phase 7 - Merge Gate

Checklist:

- Compile evidence summary.
- Present required checklist.
- Ask user exactly whether to merge.

Acceptance criteria:

- No merge without explicit user approval.
- All required evidence is present or the PR remains open.

### Task 8: Phase 8 - Post-Merge Main Smoke

Checklist:

- After merge, update local `main`.
- Run focused tests and builds before tag creation.

Acceptance criteria:

- Main smoke passes before any tag/release task starts.
- Failure results in no tag and a fix path.

### Task 9: Phase 9 - Public Release Candidate

Checklist:

- Use release skill if available.
- Produce RC artifacts, checksums, SBOM/signing status.
- Install downloaded RC artifact in lab and repeat public demo test.

Acceptance criteria:

- AC-SEC-08 and release artifact checklist have evidence.
- Lab validates released artifact, not local build.

### Task 10: Phase 10 - Go-Live Decision

Checklist:

- Fill GO/NO-GO decision record.
- Document residual risks and rollback state.
- Ask for public release approval.

Acceptance criteria:

- Every non-negotiable AC is PASS or the release is NO-GO.
- User approval exists before public release.

### Task 11: Phase 11 - Public Release and Monitoring

Checklist:

- Publish only after explicit approval.
- Verify final release workflow, assets, notes, checksums, SBOM/signing.
- Monitor lab for at least 24 hours.

Acceptance criteria:

- Published release artifacts match reviewed commit.
- Monitoring evidence exists or release remains in observation.

### Task 12: Plan-Verifikation

Checklist:

- Reread the plan completely.
- Compare implementation against every plan section.
- Run final full verification required by the plan.
- Update this file to `COMPLETE` or `BLOCKED`.

Acceptance criteria:

- Every plan item is either completed with evidence or documented as blocked with a required user decision.

## Blocked Items

- None yet.

## Commit References

- Pending.

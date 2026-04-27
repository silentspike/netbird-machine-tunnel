# Public Release Readiness Decision

Date: 2026-04-27
Repository: `silentspike/netbird-machine-tunnel`
Decision: **NO-GO for public visibility and public release**

This decision record separates release preparation progress from permission to
make the repository public or publish a release. The project is making strong
progress, but public exposure should wait until the remaining security and
governance gates are resolved or explicitly accepted.

## Current Decision

| Area | State | Evidence |
|------|-------|----------|
| Repository visibility | PASS | GitHub reports `visibility=PRIVATE`; no public flip was performed. |
| Main CI | PASS | `main` commit `bb2682231cb2ff3f191f51c691f95459e6f9921f` has the post-merge workflow set green, including FreeBSD, Windows, Linux, Darwin, Mobile, Wasm, Release, Secret Scan, dependency license checks, and CodeQL. |
| Branch protection | PASS with solo-maintainer policy | `main` has strict required status checks with 46 contexts, admin enforcement, conversation resolution, force-push protection, and deletion protection. Required approvals are intentionally `0` for the solo-maintainer repository. |
| CODEOWNERS syntax | PASS | GitHub CODEOWNERS API returns `errors=[]`. |
| RC artifact generation | PASS for Actions snapshot | Release workflow run `24984503525` produced non-expired Actions artifacts from `main` commit `bb2682231cb2ff3f191f51c691f95459e6f9921f`. |
| Fork-specific binary | PASS | Downloaded artifact contains `netbird-machine_0.1.0-SNAPSHOT-bb268223_windows_amd64.tar.gz` and extracted `netbird-machine.exe`. |
| Checksums and SBOM | PASS | `sha256sum -c netbird_0.1.0-SNAPSHOT-bb268223_checksums.txt` passed; `netbird-machine` archive and SBOM are checksum-covered. |
| Downloaded-artifact lab smoke | PASS | VM102 ran the downloaded `netbird-machine.exe` with SHA256 `be656553c08aaf620f7dde652223ed0909d77320805ed66423c973ef7ff645c9`; service, tunnel interface, route, DC ports, and handshake evidence passed. |
| Dependabot alerts | PASS pending issue close approval | Live Dependabot API returned `0` open alerts; issue #167 is now stale/closeable after maintainer approval. |
| Public documentation | PASS for current preparation | README license wording, ADR status, `FORK_DIFF.md`, and CRL limitation documentation have been updated in the readiness branch. |
| CodeQL/security baseline | **NO-GO** | #170 remains open. `main` still has 1 critical and 19 high CodeQL alerts. Baseline branch reduces this to 1 critical and 3 high, but those findings are not resolved or accepted yet. |
| Signal trust model | **NO-GO** | #114 remains open; #109 remains a public go-live blocker until #114 is resolved or explicitly split out. |
| Public approval | **NO-GO** | No final public visibility/release approval has been recorded after the current blocker list. |
| Mainline inclusion | **NO-GO** | The current public-readiness commits are local on `security/codeql-high-baseline` and are not yet pushed, reviewed, and merged to `main`. |
| Public GitHub Release | NO-GO for final release | Current artifact validation used an Actions snapshot artifact. The visible GitHub Release `v0.1.0` is old and is not the current public-launch RC. |

## Hard Blockers

1. Resolve or formally accept #170 CodeQL critical/high findings.
2. Resolve #114 Signal Server trust-model review, then close or split #109.
3. Push the public-readiness branch, open/review/merge it through protected
   `main`, and rerun the required checks on the resulting mainline state.
4. Record explicit maintainer approval before changing repository visibility.
5. Create a final tagged release or pre-release only after the above gates pass.

## Closeable After Approval

The following issues appear stale or closeable based on current evidence, but
were not closed automatically:

| Issue | Proposed disposition |
|-------|----------------------|
| #167 | Close after maintainer approval; live Dependabot API returned `0` open alerts. |
| #108 | Close after maintainer approval; the original peer-configuration issue is stale after current code and lab verification. |
| #168 | Keep open until public release approval is recorded, even though post-merge CI and RC artifact/lab gates now passed. |
| #166 | Keep open until the final decision record is accepted; current branch protection is materially hardened for solo development. |

## Final Position

The project is **GO for continued private preparation**.

The project is **NO-GO for public visibility or public release** until the hard
blockers above are resolved or explicitly accepted in a final maintainer
approval record.

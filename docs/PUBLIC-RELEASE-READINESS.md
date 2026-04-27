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
| Dependabot alerts | **NO-GO pending default-branch rescan/disposition** | GraphQL is the reliable export path: it reports 12 open default-branch alerts, while the REST endpoint currently returns `0`. Branch `security/codeql-high-baseline` head `3e6e791e` patches the actionable Go/NPM alerts, verifies `npm audit` clean, and keeps CodeQL green in run `25010818374`; remaining no-patch/transitive alerts (`docker/docker`, `pion/dtls/v2`) still need final disposition after merge/default-branch rescan. Evidence posted to #167: https://github.com/silentspike/netbird-machine-tunnel/issues/167#issuecomment-4329357241 |
| Public documentation | PASS for current preparation | README license wording, ADR status, `FORK_DIFF.md`, and CRL limitation documentation have been updated in the readiness branch. |
| CodeQL/security baseline | **NO-GO pending inherited-risk disposition** | #170 remains open. CodeQL run `25009836010` on `security/codeql-high-baseline` completed successfully and reduced the branch to `145` open alerts: 1 critical, 0 high, 142 medium, 2 warning. The fork-added `dnslabel.go` and fork-modified `geolocation/utils.go` high findings are cleared. The only remaining critical/high finding is `go/request-forgery` in `management/server/identity_provider.go`, which is unchanged from upstream v0.69.0 and must be dispositioned as inherited upstream risk instead of blindly patched. Evidence posted to #170: https://github.com/silentspike/netbird-machine-tunnel/issues/170#issuecomment-4329195301 |
| Signal trust model | **NO-GO pending maintainer acceptance** | #114 code review found no Signal certificate pinning. Machine Tunnel reuses upstream Signal: HTTPS uses standard system-root/embedded-root TLS validation, Signal message bodies use NaCl box encryption with WireGuard keys, and Signal metadata/availability risk is accepted. Public docs now document this explicitly; #109 remains a public go-live blocker until #114 is accepted and the issue state is updated. |
| Public approval | **NO-GO** | No final public visibility/release approval has been recorded after the current blocker list. |
| Mainline inclusion | **NO-GO** | The current public-readiness commits are on `security/codeql-high-baseline` and are not yet reviewed, merged to protected `main`, and rerun through the required mainline checks. |
| Public GitHub Release | NO-GO for final release | Current artifact validation used an Actions snapshot artifact. The visible GitHub Release `v0.1.0` is old and is not the current public-launch RC. |

## Hard Blockers

1. Resolve #170 by formally dispositioning the remaining inherited upstream
   `go/request-forgery` OIDC SSRF finding. Fork-added/fork-modified
   critical/high findings are cleared on the CodeQL branch; the remaining
   critical finding must be accepted explicitly for public launch or tracked
   upstream.
2. Resolve #167 Dependabot alert disposition after merging the dependency
   remediation branch to `main`: verify the default-branch rescan, then dismiss
   or explicitly accept no-patch/transitive alerts with evidence.
3. Accept the documented #114 Signal Server trust-model disposition, then close
   #114 and close or split #109 after maintainer approval.
4. Push the public-readiness branch, open/review/merge it through protected
   `main`, and rerun the required checks on the resulting mainline state.
5. Record explicit maintainer approval before changing repository visibility.
6. Create a final tagged release or pre-release only after the above gates pass.

## Closeable After Approval

The following issues appear stale or closeable based on current evidence, but
were not closed automatically:

| Issue | Proposed disposition |
|-------|----------------------|
| #167 | Keep open until default-branch rescan and no-patch/transitive dispositions are recorded. Branch remediation is pushed at `3e6e791e`; GraphQL remains the reliable export path. |
| #108 | Close after maintainer approval; the original peer-configuration issue is stale after current code and lab verification. |
| #168 | Keep open until public release approval is recorded, even though post-merge CI and RC artifact/lab gates now passed. |
| #166 | Keep open until the final decision record is accepted; current branch protection is materially hardened for solo development. |

## Final Position

The project is **GO for continued private preparation**.

The project is **NO-GO for public visibility or public release** until the hard
blockers above are resolved or explicitly accepted in a final maintainer
approval record.

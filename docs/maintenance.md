# Maintenance Policy

This document describes how this fork is maintained as a public, solo-maintained
NetBird fork. It is intended for operators and reviewers who need to understand
the update cadence, security gates, release policy, and issue-routing model.

This fork is not the official NetBird project. General NetBird improvements
should still go upstream to `netbirdio/netbird`; this repository focuses on the
Windows Machine Tunnel additions and the operational changes needed to support
them.

## Current Baseline

| Area | Position |
|------|----------|
| Upstream baseline | NetBird `v0.69.0` |
| Next upstream sync | NetBird `v0.70.0`, tracked in #189 after this public-readiness pass |
| Maintainer model | Solo-maintained fork |
| Repository visibility | Public-readiness approved only after final maintainer approval and green protected-branch checks |
| Public binary release | Separate HOLD decision until a tag, notes, artifacts, checksums, SBOM, and downloaded-artifact lab smoke are approved |

## Upstream Sync Cadence

This fork tracks upstream NetBird by release tags, not by blindly merging
`upstream/main`.

- Review upstream tagged releases at least monthly.
- Prioritize upstream security releases or dependency fixes ahead of feature
  syncs.
- Open or update an `upstream-sync` issue for each planned upstream tag.
- Keep the sync branch private or draft until local CI, generated code checks,
  and Machine Tunnel compatibility checks pass.
- Do not fold the deferred `v0.70.0` sync into the current public-readiness
  sprint; it is a follow-up tracked in #189.

Each upstream sync must record:

1. upstream tag,
2. fork base commit,
3. generated diff or conflict summary,
4. required code adaptations,
5. test/lab evidence,
6. release decision.

## Security Maintenance

Critical and high security findings are release blockers until fixed or
explicitly dispositioned. The target triage window is 2 business days for:

- CodeQL critical/high findings,
- Dependabot critical/high alerts,
- secret-scanning findings,
- credible reports affecting Machine Tunnel authentication, key handling,
  certificate validation, or management-server isolation.

Medium and lower findings are triaged by exploitability, upstream status, and
Machine Tunnel impact. They can remain open only with an explicit disposition.

Security automation:

- CodeQL scans Go and JavaScript/TypeScript.
- Dependabot monitors supported package ecosystems.
- Secret Scan runs in CI.
- Branch protection requires the security checks that gate protected `main`.

Public reports should be filed through GitHub Security Advisories when
appropriate. Do not include private keys, setup keys, customer hostnames, or
internal network details in public issues.

## CI and Branch Protection

Protected `main` is the source of truth for public readiness.

Expected gates:

- strict required status checks,
- CodeQL for Go and JavaScript/TypeScript,
- Linux, Windows, Darwin, FreeBSD, Mobile, Wasm, release, install, license, and
  infrastructure workflows,
- Secret Scan,
- generated-protobuf checks,
- admin enforcement.

Public pull requests must not run untrusted code on privileged local runners.
GitHub-hosted runners are the default for untrusted PR validation. Any
self-hosted runner use must be limited to trusted release or lab workflows.

## Release Policy

Repository visibility and binary releases are separate decisions.

Making the repository public does not publish a supported binary release. A
public tagged binary release requires:

1. explicit tag/version decision,
2. release notes,
3. green protected-branch CI,
4. generated artifacts from the release workflow,
5. checksum verification,
6. SBOM verification,
7. downloaded-artifact lab smoke test, not a local-build smoke test,
8. final Go/No-Go record.

Until those release gates pass, published GitHub releases must remain absent or
draft-only.

## Lab Validation

Machine Tunnel changes need Windows-focused validation before a public binary
release. Minimum release-candidate lab evidence:

- fresh install with short-lived setup key,
- upgrade-in-place from the previous validated build,
- Windows service starts as SYSTEM and remains Automatic,
- DPAPI-protected config remains loadable,
- machine certificate mTLS transition succeeds,
- `wg-nb-machine` interface is up,
- Domain Controller ports needed for AD login are reachable through the tunnel,
- downloaded release artifact hash matches the recorded checksum.

Lab evidence should be treated as release evidence only when the tested binary
came from the release artifact being promoted.

## Issue Routing

Use labels to keep the public board readable:

- `upstream-sync`: upstream tag tracking and merge planning,
- `type:security`: security findings and dispositions,
- `type:bug`: confirmed fork behavior defects,
- `type:docs`: documentation fixes,
- `priority:critical` or `priority:high`: public readiness or release blockers,
- `status:backlog`: known non-blocking follow-up work.

Before a public visibility or binary-release decision, open critical/high issues
must be zero or explicitly documented as non-blocking.

## Recent Maintenance Decisions

| Date | Decision |
|------|----------|
| 2026-04-28 | Public-readiness sprint keeps upstream baseline at `v0.69.0`. |
| 2026-04-28 | NetBird `v0.70.0` sync deferred to #189 after public-readiness work. |
| 2026-04-28 | Repository visibility and tagged binary release treated as separate gates. |

## Operator Checklist

Before public visibility:

```bash
gh repo view silentspike/netbird-machine-tunnel --json visibility,isPrivate
gh release list --repo silentspike/netbird-machine-tunnel --json tagName,isDraft,isPrerelease
gh issue list --repo silentspike/netbird-machine-tunnel --state open --label priority:critical --json number --jq length
gh issue list --repo silentspike/netbird-machine-tunnel --state open --label priority:high --json number --jq length
```

Before public binary release:

```bash
gh run list --repo silentspike/netbird-machine-tunnel --branch main --limit 20
gh api repos/silentspike/netbird-machine-tunnel/dependabot/alerts --jq '[.[] | select(.state == "open")] | length'
gh api 'repos/silentspike/netbird-machine-tunnel/code-scanning/alerts?state=open&per_page=100' --paginate --slurp \
  | jq 'flatten | group_by(.rule.security_severity_level // "null") | map({level: (.[0].rule.security_severity_level // "null"), count: length})'
```

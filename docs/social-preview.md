# Social Preview Spec

This repository keeps a committed social-preview artifact so the public GitHub
repo can be configured consistently after the visibility flip.

## Artifact

- File: `docs/media/social-preview.svg`
- Canvas: `1200x630`
- Intended use: GitHub repository Social Preview / OpenGraph image
- Upload path: GitHub repository settings UI

`gh repo edit` does not currently expose a Social Preview/OpenGraph image upload
flag in this environment, so the committed artifact is the source of truth until
the image is uploaded manually through GitHub settings.

## Message

Primary message:

> NetBird Machine Tunnel

Supporting message:

> Unofficial fork: pre-login Windows VPN for Active Directory

Diagram:

```text
Windows laptop -> Machine Tunnel mTLS -> NetBird Management -> Domain Controller
```

## Constraints

- Do not imply this is an official NetBird product.
- Do not use hiring, AI-assistant, or unrelated personal-project framing.
- Do not include internal hostnames, private IPs, setup keys, or customer data.
- Keep the visual focused on the operational value: machine-certificate mTLS
  before Windows user login.

## Alt Text

Neutral enterprise diagram showing a remote Windows laptop using Machine Tunnel
mTLS to reach NetBird Management and a Domain Controller before user login.

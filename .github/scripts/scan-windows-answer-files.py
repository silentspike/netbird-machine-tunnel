"""Scan complete Windows answer files introduced by the incoming commits."""

from pathlib import Path
import re
import subprocess
import sys
from tempfile import TemporaryDirectory


def read_git(*args, input_data=None):
    result = subprocess.run(
        ["git", *args],
        input=input_data, check=True, capture_output=True,
    )
    return result.stdout


def answer_blobs(revisions):
    commits = read_git("rev-list", revisions)
    changes = read_git(
        "diff-tree", "--stdin", "--root", "-m", "-r", "--no-commit-id",
        "--no-abbrev", "--no-renames", "--diff-filter=AMT", "--raw", "-z", "--",
        ":(glob,icase)**/unattend.xml", ":(glob,icase)**/autounattend.xml",
        input_data=commits,
    )
    if not changes:
        return set()
    fields = changes.split(b"\0")
    if fields[-1] != b"" or (len(fields) - 1) % 2:
        raise ValueError("Unexpected Git diff output")
    blobs = set()
    for header in fields[:-1:2]:
        _, mode, _, oid, _ = header.split()
        if not re.fullmatch(rb"[0-9a-f]{40}|[0-9a-f]{64}", oid):
            raise ValueError("Unexpected Git object ID")
        if mode in (b"100644", b"100755"):
            blobs.add(oid.decode("ascii"))
    return blobs


def main():
    if len(sys.argv) != 2 or not re.fullmatch(r"[0-9a-f]{40}(?:\.\.[0-9a-f]{40})?", sys.argv[1]):
        print("Expected validated Git revisions.", file=sys.stderr)
        return 2
    try:
        config = Path(".gitleaks.toml").resolve(strict=True)
        blobs = answer_blobs(sys.argv[1])
        if not blobs:
            return 0
        with TemporaryDirectory(prefix="check-answer-files-") as tmp:
            scratch = Path(tmp)
            for oid in blobs:
                directory = scratch / oid
                directory.mkdir()
                (directory / "autounattend.xml").write_bytes(read_git("cat-file", "blob", oid))
            return subprocess.run(
                [
                    "gitleaks", "dir", "--no-banner", "--redact=100", "--config", str(config),
                    "--enable-rule", "windows-answer-file-plaintext-password", str(scratch),
                ],
                check=False,
            ).returncode
    except (ValueError, OSError, subprocess.CalledProcessError):
        print("Windows answer-file scan could not finish.", file=sys.stderr)
        return 2


if __name__ == "__main__":
    raise SystemExit(main())

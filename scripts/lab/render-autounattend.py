#!/usr/bin/env python3
"""Render the Windows lab setup template using a separate installation password."""

import argparse
import os
from pathlib import Path
import tempfile
from xml.etree import ElementTree
from xml.sax.saxutils import escape

PASSWORD_VARIABLE = "NETBIRD_LAB_ADMIN_PASSWORD"
PLACEHOLDER = "CHANGE-ME-Lab-Password!"
NAMESPACE = "{urn:schemas-microsoft-com:unattend}"


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("output", type=Path, help="New answer file outside all Git checkouts")
    args = parser.parse_args()

    password = os.environ.get(PASSWORD_VARIABLE)
    if not password or password == PLACEHOLDER:
        parser.error(f"Set {PASSWORD_VARIABLE} to the password for this installation")

    output = args.output.expanduser().absolute()
    parent = output.parent.resolve()
    if any((path / ".git").exists() for path in (parent, *parent.parents)):
        parser.error("Choose an output directory outside Git checkouts")
    output = parent / output.name

    template = Path(__file__).with_name("autounattend.xml").read_text(encoding="utf-8")
    if template.count(PLACEHOLDER) != 2:
        parser.error("The answer file template must contain the two lab password placeholders")

    value = escape(password, {"\r": "&#13;"})
    rendered = template.replace(PLACEHOLDER, value)
    try:
        document = ElementTree.fromstring(rendered)
    except (ElementTree.ParseError, UnicodeError):
        parser.error(f"The template or {PASSWORD_VARIABLE} cannot be represented as valid XML")

    tags = {NAMESPACE + "AdministratorPassword", NAMESPACE + "Password"}
    entries = [element for element in document.iter() if element.tag in tags]
    if len(entries) != 2 or any(element.findtext(NAMESPACE + "Value") != password for element in entries):
        parser.error("The two Windows setup password values did not round-trip")

    try:
        with tempfile.NamedTemporaryFile(
            mode="w", encoding="utf-8", newline="\n", dir=parent, prefix=".unattend-"
        ) as temporary:
            temporary.write(rendered)
            temporary.flush()
            os.link(temporary.name, output)
    except FileExistsError:
        parser.error("The output file already exists")
    except OSError:
        parser.error("Unable to create a private answer file in the selected directory")


if __name__ == "__main__":
    main()

# Windows lab answer file

Set `NETBIRD_LAB_ADMIN_PASSWORD` from the private environment for the new lab
installation. Use a fresh password for each installation.

On the Linux provisioning host, render the supplied template to a private
directory on a local filesystem outside your Git checkouts.

```bash
python3 scripts/lab/render-autounattend.py "$LAB_OUTPUT_DIR/autounattend.xml"
```

The rendered file is readable only by its owner and contains the password
Windows Setup reads during installation. Keep it with the private installation
media and remove it when setup is complete. The repository contains the
template for reuse.

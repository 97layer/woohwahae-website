# Linux VM Bootstrap

## Intent

Bootstrap the drained Ubuntu VM as the single Layer OS continuity host without
relying on a remote Go or Node toolchain.

## Related Docs

- `docs/operator.md`
- `docs/validation.md`

## Related Artifacts

- `scripts/deploy_vm.sh`
- `scripts/systemd/layer-osd.service`
- `scripts/systemd/layer-osd.env.example`

## Continuity Host Shape

- host: `97layer-vm`
- repo root: `/srv/layer-os/current`
- releases: `/srv/layer-os/releases/<stamp>`
- runtime data: `/var/lib/layer-os`
- runtime logs: `/var/log/layer-os`
- daemon env: `/etc/layer-os/layer-osd.env`
- provider env: `/etc/layer-os/providers.env`
- service: `layer-osd.service`

## Canonical Bootstrap

From the local repo:

```bash
./scripts/deploy_vm.sh --host 97layer-vm
```

The script:

1. builds `layer-osd` and `layer-osctl` for `linux/amd64`
2. syncs the repo snapshot to a remote release directory
3. installs the Linux `systemd` unit if needed
4. creates example env files on first boot
5. restarts the remote `layer-osd`
6. verifies `http://127.0.0.1:17808/healthz` on the VM

Use `--check` when you only want to validate the local build and remote layout:

```bash
./scripts/deploy_vm.sh --host 97layer-vm --check
```

## Register As Deploy Target

Once the remote service is healthy, register the VM as the real deploy target
from the local control daemon:

```bash
env GOCACHE=/tmp/gocache GOMODCACHE=/tmp/gomodcache \
go run ./cmd/layer-osctl target put \
  --id vm \
  --command '/bin/bash,/Users/97layer/layer OS/scripts/deploy_vm.sh,--host,97layer-vm'
```

That keeps the founder/admin deploy corridor pointed at one canonical write path:
the local runtime launches a single deploy script, and that script owns the
remote sync + systemd restart sequence.

## First Smoke

After target registration, close the first evidence loop:

1. create a work item for the VM rollout
2. create and approve the release approval
3. create a release pointing at `vm`
4. execute the deploy against target `vm`
5. verify the release/deploy timeline appears in founder/admin

## Notes

- The VM currently has `git`, `rsync`, and `jq`, but no remote Go or Node
  toolchain. The deploy script deliberately ships prebuilt binaries.
- The deploy script only seeds example env files on first boot; it does not
  overwrite existing secrets.

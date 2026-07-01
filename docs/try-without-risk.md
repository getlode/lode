# Try lode without risk

`lode` is designed to be tested on a real DVC repository without migration. For the
commands it supports, it writes standard DVC metadata and cache objects. The important
safety property is no format lock-in: you can stop using `lode` and keep using `dvc`.

## Fast path: existing DVC repo

```bash
brew install getlode/tap/lode
cd your-dvc-repo
lode doctor
lode status
lode add path/to/data
lode verify
dvc status
```

If `dvc status` reports the repo normally after `lode add`, DVC is reading the files
that `lode` produced.

## Copy-based trial

Use this when you want a completely disposable test run.

```bash
cp -a your-dvc-repo /tmp/lode-trial
cd /tmp/lode-trial
lode doctor
lode status
lode add path/to/data
lode verify
dvc status
```

Delete `/tmp/lode-trial` when finished. Your original repo was not touched.

## What lode may write

Depending on the command, `lode` can update the same files DVC would update:

- `.dvc/*.dvc` or `<path>.dvc` pointer files
- `.dvc/cache/files/md5/...` cache objects
- `.dvc/tmp/...` state and lock files
- `.gitignore` entries for tracked data paths
- objects in the configured S3-compatible remote when running `push`

## Rollback

There is no export step because there is no new format. Stop running `lode` and
continue with `dvc`. If you tested in a copied repo, delete the copy. If you tested
in-place and do not want the new metadata, use normal Git review/revert on the `.dvc`
and `.gitignore` changes before committing. If you pushed during a trial, remote cache
objects may remain until normal DVC/lode garbage collection removes unreferenced data.

## When not to use lode yet

- You need `dvc repro` or pipeline orchestration. Keep using DVC for that.
- Your primary remote is native GCS, Azure, or SSH. Use DVC until native support lands.
- Your workflow depends on a DVC command not listed in the compatibility matrix.
- You are on NFS/restored backups/safety-critical checks and cannot run `--rehash`.

#!/usr/bin/env bash
# Deliberately branch-bounded, non-executable upstream source snapshot import.
set -euo pipefail

[[ "${GITHUB_REPOSITORY:-}" == 'GoreeCloud/code' ]] || { echo 'wrong repository' >&2; exit 2; }
[[ "${GITHUB_REF:-}" == 'refs/heads/feat/forgejo-bootstrap' ]] || { echo 'wrong branch' >&2; exit 2; }
[[ "$(git branch --show-current)" == 'feat/forgejo-bootstrap' ]] || { echo 'wrong local branch' >&2; exit 2; }
if [[ -e vendor/forgejo/LICENSE || -e vendor/forgejo/UPSTREAM-SNAPSHOT.md ]]; then
  echo 'Source snapshot exists: refusing implicit overwrite.'
  exit 0
fi
if [[ -n "$(git status --porcelain)" ]]; then
  echo 'Dirty repository: refusing import.' >&2
  exit 2
fi

tag='v15.0.9'
upstream='https://codeberg.org/forgejo/forgejo.git'
tmp="$(mktemp -d)"
trap 'rm -rf "$tmp"' EXIT
git clone --filter=blob:none --depth=1 --branch "$tag" "$upstream" "$tmp/source"
test -s "$tmp/source/LICENSE"
test -s "$tmp/source/go.mod"
sha="$(git -C "$tmp/source" rev-parse HEAD)"
[[ "$sha" =~ ^[0-9a-f]{40}$ ]] || exit 3

mkdir -p vendor/forgejo
git -C "$tmp/source" archive --format=tar HEAD | tar -xf - -C vendor/forgejo
cat > vendor/forgejo/UPSTREAM-SNAPSHOT.md <<EOF
# Official Forgejo source snapshot

Origin: $upstream
Release tag: $tag
Resolved commit: $sha
Import method: complete tracked-source archive; not upstream commit history
Licensing: GPL-3.0-or-later (preserve LICENSE, notices and corresponding-source obligations)
Status: unreviewed imported source, not GoreeCloud Stable or production acceptance

Security: independently verify upstream tag, provenance, dependencies, build and release suitability before packaging or deployment.
EOF

git add -- vendor/forgejo
if git diff --cached --quiet; then
  echo 'No files imported'; exit 3
fi
git -c user.name='GoreeCloud Code Import' \
    -c user.email='295260680+GoreeCloud@users.noreply.github.com' \
    commit -m "vendor: import Forgejo $tag tracked snapshot ($sha)"
git push origin HEAD:refs/heads/feat/forgejo-bootstrap

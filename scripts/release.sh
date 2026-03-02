#!/usr/bin/env bash
set -euo pipefail

usage() {
  cat <<'EOF'
Usage:
  scripts/release.sh <version> [--push]

Example:
  scripts/release.sh v1.2.3
  scripts/release.sh v2.0.0-alpha.1
  scripts/release.sh v2.0.0-rc.1 --push

What it does:
  1) Validate semantic version format (vMAJOR.MINOR.PATCH[-{alpha|beta|rc}.N])
  2) Generate docs/releases/<version>.md from git commits
  3) Create annotated git tag
  4) Optionally push the tag to origin
EOF
}

if [[ $# -lt 1 ]]; then
  usage
  exit 1
fi

if [[ "${1:-}" == "-h" || "${1:-}" == "--help" ]]; then
  usage
  exit 0
fi

VERSION="$1"
PUSH_TAG="false"

for arg in "${@:2}"; do
  case "$arg" in
    --push)
      PUSH_TAG="true"
      ;;
    -h|--help)
      usage
      exit 0
      ;;
    *)
      echo "Unknown option: $arg" >&2
      usage
      exit 1
      ;;
  esac
done

if ! [[ "$VERSION" =~ ^v[0-9]+\.[0-9]+\.[0-9]+(-(alpha|beta|rc)\.[0-9]+)?$ ]]; then
  echo "Version must match vMAJOR.MINOR.PATCH or vMAJOR.MINOR.PATCH-(alpha|beta|rc).N, got: $VERSION" >&2
  exit 1
fi

if ! git rev-parse --is-inside-work-tree >/dev/null 2>&1; then
  echo "This script must be run inside a git repository." >&2
  exit 1
fi

if ! git diff --quiet || ! git diff --cached --quiet; then
  echo "Tracked files have unstaged or staged changes. Commit or stash them first." >&2
  exit 1
fi

if git rev-parse -q --verify "refs/tags/$VERSION" >/dev/null 2>&1; then
  echo "Tag already exists: $VERSION" >&2
  exit 1
fi

LAST_TAG="$(git tag -l 'v[0-9]*.[0-9]*.[0-9]*' --sort=-v:refname | head -n 1)"
if [[ -n "$LAST_TAG" ]]; then
  newest="$(printf "%s\n%s\n" "${LAST_TAG#v}" "${VERSION#v}" | sort -V | tail -n 1)"
  if [[ "v${newest}" != "$VERSION" ]]; then
    echo "Version $VERSION must be greater than latest tag $LAST_TAG" >&2
    exit 1
  fi
fi

RANGE="HEAD"
if [[ -n "$LAST_TAG" ]]; then
  RANGE="${LAST_TAG}..HEAD"
fi

COMMITS="$(git log --no-merges --pretty=format:'- %s (%h)' "$RANGE")"
if [[ -z "$COMMITS" ]]; then
  COMMITS="- No non-merge commits found in range ${RANGE}."
fi

mkdir -p docs/releases
RELEASE_FILE="docs/releases/${VERSION}.md"
if [[ -e "$RELEASE_FILE" ]]; then
  echo "Release note file already exists: $RELEASE_FILE" >&2
  exit 1
fi

DATE_UTC="$(date -u +%F)"
PREVIOUS_LABEL="${LAST_TAG:-initial release}"

cat > "$RELEASE_FILE" <<EOF
# Release ${VERSION}

- Date (UTC): ${DATE_UTC}
- Previous tag: ${PREVIOUS_LABEL}

## Highlights

- TODO: summarize the key changes for this release.

## Commits

${COMMITS}
EOF

git tag -a "$VERSION" -m "release: ${VERSION}"

if [[ "$PUSH_TAG" == "true" ]]; then
  git push origin "$VERSION"
fi

echo "Created release note draft: ${RELEASE_FILE}"
echo "Created git tag: ${VERSION}"
if [[ "$PUSH_TAG" == "true" ]]; then
  echo "Pushed git tag to origin: ${VERSION}"
else
  echo "Tag not pushed. Run: git push origin ${VERSION}"
fi

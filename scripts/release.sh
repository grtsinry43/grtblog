#!/usr/bin/env bash
set -euo pipefail

section() {
  printf '\n==> %s\n' "$1"
}

info() {
  printf '  - %s\n' "$1"
}

warn() {
  printf '  ! %s\n' "$1" >&2
}

next_step() {
  printf '    %s\n' "$1"
}

replace_section() {
  local file="$1"
  local start_marker="$2"
  local end_marker="$3"
  local content="$4"
  local tmp_file
  local body_file
  tmp_file="$(mktemp)"
  body_file="$(mktemp)"

  printf '%s\n' "$content" > "$body_file"

  awk -v start="$start_marker" -v end="$end_marker" -v body_file="$body_file" '
    $0 == start {
      print
      while ((getline line < body_file) > 0) {
        print line
      }
      close(body_file)
      skip = 1
      next
    }
    $0 == end {
      skip = 0
    }
    !skip {
      print
    }
  ' "$file" > "$tmp_file"

  rm -f "$body_file"
  mv "$tmp_file" "$file"
}

usage() {
  cat <<'EOF'
Usage:
  scripts/release.sh <version> [--notes-only] [--continue] [--push]

Example:
  scripts/release.sh v1.2.3
  scripts/release.sh v2.1.0-beta.1
  scripts/release.sh v2.1.0-rc.1 --notes-only
  scripts/release.sh v2.1.0-rc.1 --continue
  scripts/release.sh v1.2.3 --push

What it does:
  1) Validate semantic version format and detect channel (stable or preview)
  2) Generate or refresh docs/releases/<version>.md from git commits
  3) Require the release notes files to be committed in HEAD before tagging
  4) Optionally auto-commit only the release docs with --continue
  5) Create annotated git tag
  6) Optionally push the tag to origin

Recommended flow:
  1. Run this script once to generate release notes
  2. Review and commit docs/releases/<version>.md and docs/releases/index.md
  3. Run the script again to create the tag

Fast path:
  scripts/release.sh <version> --continue [--push]
  This will auto-commit only the release docs, then create the tag.

Behavior by channel:
  - stable: for GitHub Release + GHCR + Docker Hub
  - preview: for Git tag + GHCR only
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
NOTES_ONLY="false"
CONTINUE_RELEASE="false"
RERUN_CMD="scripts/release.sh ${VERSION}"
CONTINUE_CMD="scripts/release.sh ${VERSION} --continue"

for arg in "${@:2}"; do
  case "$arg" in
    --notes-only)
      NOTES_ONLY="true"
      ;;
    --continue)
      CONTINUE_RELEASE="true"
      ;;
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

if [[ "$PUSH_TAG" == "true" ]]; then
  RERUN_CMD+=" --push"
  CONTINUE_CMD+=" --push"
fi

if [[ "$NOTES_ONLY" == "true" && "$CONTINUE_RELEASE" == "true" ]]; then
  warn "--notes-only and --continue cannot be used together."
  exit 1
fi

if ! [[ "$VERSION" =~ ^v[0-9]+\.[0-9]+\.[0-9]+(-(alpha|beta|rc)\.[0-9]+)?$ ]]; then
  warn "Version must match vMAJOR.MINOR.PATCH or vMAJOR.MINOR.PATCH-(alpha|beta|rc).N"
  next_step "received: ${VERSION}"
  exit 1
fi

CHANNEL="stable"
STAGE="stable"
if [[ "$VERSION" =~ -(alpha|beta|rc)\.[0-9]+$ ]]; then
  CHANNEL="preview"
  STAGE="${BASH_REMATCH[1]}"
fi

section "Release Plan"
info "version: ${VERSION}"
info "channel: ${CHANNEL}"
info "stage: ${STAGE}"
if [[ "$NOTES_ONLY" == "true" ]]; then
  info "mode: notes-only"
elif [[ "$CONTINUE_RELEASE" == "true" && "$PUSH_TAG" == "true" ]]; then
  info "mode: auto-commit, tag and push"
elif [[ "$CONTINUE_RELEASE" == "true" ]]; then
  info "mode: auto-commit and tag"
elif [[ "$PUSH_TAG" == "true" ]]; then
  info "mode: tag and push"
else
  info "mode: tag only"
fi

if ! git rev-parse --is-inside-work-tree >/dev/null 2>&1; then
  warn "This script must be run inside a git repository."
  exit 1
fi

LAST_TAG="$(git tag -l 'v[0-9]*.[0-9]*.[0-9]*' --sort=-v:refname | head -n 1)"
if [[ -n "$LAST_TAG" ]]; then
  # SemVer comparison: a release (no pre-release suffix) is always greater than
  # a pre-release of the same numeric version (e.g. 2.0.0 > 2.0.0-rc.2).
  # sort -V gets this wrong, so we handle it explicitly.
  semver_greater_or_equal() {
    local a="$1" b="$2"
    local a_base="${a%%-*}" b_base="${b%%-*}"
    local a_pre="" b_pre=""
    [[ "$a" == *-* ]] && a_pre="${a#*-}"
    [[ "$b" == *-* ]] && b_pre="${b#*-}"

    local newest
    newest="$(printf "%s\n%s\n" "$a_base" "$b_base" | sort -V | tail -n 1)"

    if [[ "$a_base" != "$b_base" ]]; then
      [[ "$newest" == "$a_base" ]] && return 0 || return 1
    fi

    # Same base version: release > pre-release
    if [[ -z "$a_pre" && -n "$b_pre" ]]; then return 0; fi
    if [[ -n "$a_pre" && -z "$b_pre" ]]; then return 1; fi
    if [[ -z "$a_pre" && -z "$b_pre" ]]; then return 0; fi

    # Both pre-release: fall back to sort -V
    newest="$(printf "%s\n%s\n" "$a" "$b" | sort -V | tail -n 1)"
    [[ "$newest" == "$a" ]] && return 0 || return 1
  }

  if ! semver_greater_or_equal "${VERSION#v}" "${LAST_TAG#v}"; then
    warn "Version ${VERSION} must be greater than latest tag ${LAST_TAG}."
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
INDEX_FILE="docs/releases/index.md"
AUTO_COMMIT_MSG="docs(release): prepare ${VERSION}"

DATE_UTC="$(date -u +%F)"
PREVIOUS_LABEL="${LAST_TAG:-initial release}"

list_worktree_changes() {
  {
    git diff --name-only
    git diff --cached --name-only
    git ls-files --others --exclude-standard
  } | awk 'NF > 0' | sort -u
}

list_unrelated_changes() {
  local allowed_one="$1"
  local allowed_two="$2"
  list_worktree_changes | awk -v a="$allowed_one" -v b="$allowed_two" '
    $0 != a && $0 != b { print }
  '
}

EXISTING_UNRELATED_CHANGES="$(list_unrelated_changes "$RELEASE_FILE" "$INDEX_FILE")"
if [[ -n "$EXISTING_UNRELATED_CHANGES" ]]; then
  warn "Working tree contains unrelated changes. Release automation only allows release doc files."
  section "Blocked Files"
  while IFS= read -r path; do
    next_step "$path"
  done <<< "$EXISTING_UNRELATED_CHANGES"
  section "Next Step"
  next_step "commit or stash those files first, then rerun:"
  if [[ "$CONTINUE_RELEASE" == "true" ]]; then
    next_step "  ${CONTINUE_CMD}"
  else
    next_step "  ${RERUN_CMD}"
    next_step "or use auto-commit for release docs only:"
    next_step "  ${CONTINUE_CMD}"
  fi
  exit 1
fi

if [[ "$CONTINUE_RELEASE" != "true" ]]; then
  EXISTING_CHANGES="$(list_worktree_changes)"
  if [[ -n "$EXISTING_CHANGES" ]]; then
    warn "Working tree is not clean."
    section "Changed Files"
    while IFS= read -r path; do
      next_step "$path"
    done <<< "$EXISTING_CHANGES"
    section "Next Step"
    next_step "commit or stash changes, then rerun:"
    next_step "  ${RERUN_CMD}"
    next_step "or use auto-commit for release docs only:"
    next_step "  ${CONTINUE_CMD}"
    exit 1
  fi
fi

TMP_RELEASE_FILE="$(mktemp)"
trap 'rm -f "$TMP_RELEASE_FILE"' EXIT

section "Generate Release Notes"
info "source range: ${RANGE}"
info "release notes file: ${RELEASE_FILE}"

NOTES_CHANGED="false"
if [[ "$CONTINUE_RELEASE" == "true" && -f "$RELEASE_FILE" ]]; then
  # --continue: preserve user-edited release notes, do not overwrite with template
  info "release notes already exist, preserving user edits (--continue)"
else
  cat > "$TMP_RELEASE_FILE" <<EOF
# Release ${VERSION}

- Date (UTC): ${DATE_UTC}
- Previous tag: ${PREVIOUS_LABEL}
- Channel: ${CHANNEL}
- Stage: ${STAGE}

## Highlights

- TODO: summarize the key changes for this release.

## Commits

${COMMITS}
EOF

  if [[ ! -f "$RELEASE_FILE" ]] || ! cmp -s "$TMP_RELEASE_FILE" "$RELEASE_FILE"; then
    mkdir -p "$(dirname "$RELEASE_FILE")"
    cp "$TMP_RELEASE_FILE" "$RELEASE_FILE"
    NOTES_CHANGED="true"
  fi

  if [[ "$NOTES_CHANGED" == "true" ]]; then
    info "release notes generated or refreshed"
  else
    info "release notes already up to date"
  fi
fi

section "Update Release Archive"
info "archive index: ${INDEX_FILE}"
if [[ ! -f "$INDEX_FILE" ]]; then
  cat > "$INDEX_FILE" <<'EOF'
# 版本发布

这里汇总 GrtBlog 的版本发布记录，包含两个通道：

- `stable`：正式版本，会同步发布 GitHub Release、GHCR，以及已配置时的 Docker Hub
- `preview`：预发布版本，只保留 Git tag 和 GHCR 镜像，用于 beta / rc 验证

## Stable

<!-- stable:start -->
- _暂无条目_
<!-- stable:end -->

## Preview

<!-- preview:start -->
- _暂无条目_
<!-- preview:end -->
EOF
fi

ENTRY_DESC="${DATE_UTC} · ${CHANNEL}"
if [[ "$CHANNEL" == "preview" ]]; then
  ENTRY_DESC="${ENTRY_DESC} · ${STAGE}"
fi
ENTRY_LINE="- [${VERSION}](/releases/${VERSION}) - ${ENTRY_DESC}"

SECTION_START="<!-- stable:start -->"
SECTION_END="<!-- stable:end -->"
if [[ "$CHANNEL" == "preview" ]]; then
  SECTION_START="<!-- preview:start -->"
  SECTION_END="<!-- preview:end -->"
fi

CURRENT_SECTION="$(awk -v start="$SECTION_START" -v end="$SECTION_END" '
  $0 == start { flag = 1; next }
  $0 == end { flag = 0 }
  flag { print }
' "$INDEX_FILE")"

FILTERED_SECTION="$(printf '%s\n' "$CURRENT_SECTION" | awk -v version="$VERSION" '
  $0 ~ "\\(" "/releases/" version "\\)" { next }
  $0 ~ "_暂无条目_" { next }
  NF > 0 { print }
')"

NEW_SECTION="$ENTRY_LINE"
if [[ -n "$FILTERED_SECTION" ]]; then
  NEW_SECTION+=$'\n'"$FILTERED_SECTION"
fi

TMP_INDEX_FILE="$(mktemp)"
cp "$INDEX_FILE" "$TMP_INDEX_FILE"
replace_section "$TMP_INDEX_FILE" "$SECTION_START" "$SECTION_END" "$NEW_SECTION"

INDEX_CHANGED="false"
if ! cmp -s "$TMP_INDEX_FILE" "$INDEX_FILE"; then
  mv "$TMP_INDEX_FILE" "$INDEX_FILE"
  INDEX_CHANGED="true"
else
  rm -f "$TMP_INDEX_FILE"
fi

if [[ "$INDEX_CHANGED" == "true" ]]; then
  info "release archive updated"
else
  info "release archive already up to date"
fi

if [[ "$NOTES_ONLY" == "true" ]]; then
  section "Next Step"
  next_step "review:"
  next_step "  ${RELEASE_FILE}"
  next_step "  ${INDEX_FILE}"
  next_step "commit it if needed:"
  next_step "  git add ${RELEASE_FILE} ${INDEX_FILE}"
  next_step "  git commit -m \"${AUTO_COMMIT_MSG}\""
  next_step "then create the tag:"
  next_step "  ${RERUN_CMD}"
  next_step "or auto-commit and continue:"
  next_step "  ${CONTINUE_CMD}"
  exit 0
fi

if git rev-parse -q --verify "refs/tags/$VERSION" >/dev/null 2>&1; then
  warn "Tag already exists: ${VERSION}"
  exit 1
fi

if [[ "$CONTINUE_RELEASE" == "true" ]]; then
  section "Auto Commit Release Docs"
  CURRENT_UNRELATED_CHANGES="$(list_unrelated_changes "$RELEASE_FILE" "$INDEX_FILE")"
  if [[ -n "$CURRENT_UNRELATED_CHANGES" ]]; then
    warn "Detected unrelated changes after generating release docs. Auto-commit is blocked."
    section "Blocked Files"
    while IFS= read -r path; do
      next_step "$path"
    done <<< "$CURRENT_UNRELATED_CHANGES"
    exit 1
  fi

  RELEASE_DOCS_CHANGED="false"
  if ! git cat-file -e "HEAD:${RELEASE_FILE}" >/dev/null 2>&1 || ! git diff --quiet HEAD -- "$RELEASE_FILE"; then
    RELEASE_DOCS_CHANGED="true"
  fi
  if ! git cat-file -e "HEAD:${INDEX_FILE}" >/dev/null 2>&1 || ! git diff --quiet HEAD -- "$INDEX_FILE"; then
    RELEASE_DOCS_CHANGED="true"
  fi

  if [[ "$RELEASE_DOCS_CHANGED" == "true" ]]; then
    git add "$RELEASE_FILE" "$INDEX_FILE"
    git commit -m "$AUTO_COMMIT_MSG"
    info "created commit: ${AUTO_COMMIT_MSG}"
  else
    info "release docs already committed in HEAD"
  fi
elif [[ "$NOTES_CHANGED" == "true" || "$INDEX_CHANGED" == "true" ]]; then
  warn "Release notes changed, so tagging is intentionally blocked."
  section "Next Step"
  next_step "review:"
  next_step "  ${RELEASE_FILE}"
  next_step "  ${INDEX_FILE}"
  next_step "commit it:"
  next_step "  git add ${RELEASE_FILE} ${INDEX_FILE}"
  next_step "  git commit -m \"${AUTO_COMMIT_MSG}\""
  next_step "rerun:"
  next_step "  ${RERUN_CMD}"
  next_step "or auto-commit and continue:"
  next_step "  ${CONTINUE_CMD}"
  exit 1
fi

if ! git cat-file -e "HEAD:${RELEASE_FILE}" >/dev/null 2>&1; then
  warn "Release note file exists but is not committed in HEAD."
  section "Next Step"
  next_step "git add ${RELEASE_FILE} ${INDEX_FILE}"
  next_step "git commit -m \"${AUTO_COMMIT_MSG}\""
  next_step "${RERUN_CMD}"
  next_step "or auto-commit and continue:"
  next_step "${CONTINUE_CMD}"
  exit 1
fi

if ! git diff --quiet HEAD -- "$RELEASE_FILE"; then
  warn "Release note file differs from HEAD."
  section "Next Step"
  next_step "commit the latest release notes:"
  next_step "  git add ${RELEASE_FILE} ${INDEX_FILE}"
  next_step "  git commit -m \"${AUTO_COMMIT_MSG}\""
  next_step "rerun:"
  next_step "  ${RERUN_CMD}"
  next_step "or auto-commit and continue:"
  next_step "  ${CONTINUE_CMD}"
  exit 1
fi

if ! git cat-file -e "HEAD:${INDEX_FILE}" >/dev/null 2>&1; then
  warn "Release archive index exists but is not committed in HEAD."
  section "Next Step"
  next_step "git add ${RELEASE_FILE} ${INDEX_FILE}"
  next_step "git commit -m \"${AUTO_COMMIT_MSG}\""
  next_step "${RERUN_CMD}"
  next_step "or auto-commit and continue:"
  next_step "${CONTINUE_CMD}"
  exit 1
fi

if ! git diff --quiet HEAD -- "$INDEX_FILE"; then
  warn "Release archive index differs from HEAD."
  section "Next Step"
  next_step "commit the updated archive index:"
  next_step "  git add ${RELEASE_FILE} ${INDEX_FILE}"
  next_step "  git commit -m \"${AUTO_COMMIT_MSG}\""
  next_step "rerun:"
  next_step "  ${RERUN_CMD}"
  next_step "or auto-commit and continue:"
  next_step "  ${CONTINUE_CMD}"
  exit 1
fi

section "Create Tag"
git tag -a "$VERSION" -m "release(${CHANNEL}): ${VERSION}"
info "created annotated tag: ${VERSION}"

if [[ "$PUSH_TAG" == "true" ]]; then
  section "Push Tag"
  git push origin "$VERSION"
  info "pushed tag to origin: ${VERSION}"
fi

section "CI Follow-up"
if [[ "$CHANNEL" == "stable" ]]; then
  info "GitHub Actions will build and publish stable images"
  info "targets: GHCR and, if configured, Docker Hub"
  info "GitHub Release: will be created automatically for ${VERSION}"
  info "image tags: ${VERSION}, ${VERSION#v}"
  info "rolling tags: stable, latest, major.minor"
else
  info "GitHub Actions will build and publish preview images"
  info "targets: GHCR only"
  info "GitHub Release: will be created as prerelease for ${VERSION}"
  info "distribution: Git tag + GHCR preview images + GitHub prerelease"
  info "image tags: ${VERSION#v}, preview, ${STAGE}"
fi

section "Done"
info "release channel: ${CHANNEL}"
info "release notes in HEAD: ${RELEASE_FILE}"
info "release archive in HEAD: ${INDEX_FILE}"
info "created git tag: ${VERSION}"
if [[ "$PUSH_TAG" == "true" ]]; then
  info "remote push: completed"
else
  next_step "push when ready:"
  next_step "  git push origin ${VERSION}"
fi

#!/usr/bin/env sh
set -e

MSG_FILE="$1"
SUBJECT="$(grep -v '^#' "$MSG_FILE" | head -1 | sed 's/^[[:space:]]*//;s/[[:space:]]*$//')"

if [ -z "$SUBJECT" ]; then
    echo "[ptk] commit message is empty"
    exit 1
fi

TYPES="feat|fix|docs|style|refactor|perf|test|build|ci|chore|revert"
if ! echo "$SUBJECT" | grep -qE "^($TYPES)(\([a-z0-9/._-]+\))?(!)?: .+"; then
    echo "[ptk] invalid commit message: $SUBJECT"
    echo "      format: <type>(<scope>): <subject>"
    echo "      types:  feat fix docs style refactor perf test build ci chore revert"
    echo "      example: feat(ps): add get-eventlog filter"
    exit 1
fi

AFTER_COLON="$(echo "$SUBJECT" | sed 's/^[^:]*: //')"
FIRST_CHAR="$(echo "$AFTER_COLON" | cut -c1)"
if echo "$FIRST_CHAR" | grep -qE "[A-Z]"; then
    echo "[ptk] subject must be lower-case: $SUBJECT"
    echo "      change '${FIRST_CHAR}' to '$(echo "$FIRST_CHAR" | tr '[:upper:]' '[:lower:]')'"
    exit 1
fi

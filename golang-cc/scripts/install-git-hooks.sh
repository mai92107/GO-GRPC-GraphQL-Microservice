#!/bin/sh
set -eu

HOOK_NAME="${1:-pre-push}"
REPO_ROOT="$(git rev-parse --show-toplevel)"
PROJECT_PREFIX="$(git rev-parse --show-prefix)"
PROJECT_DIR="${PROJECT_PREFIX%/}"
HOOKS_DIR="$REPO_ROOT/.git/hooks"
HOOK_PATH="$HOOKS_DIR/$HOOK_NAME"

case "$HOOK_NAME" in
	pre-push|post-commit)
		;;
	*)
		echo "Unsupported hook: $HOOK_NAME" >&2
		echo "Usage: $0 [pre-push|post-commit]" >&2
		exit 1
		;;
esac

mkdir -p "$HOOKS_DIR"

if [ -z "$PROJECT_DIR" ]; then
	PROJECT_DIR="."
fi

if [ ! -f "$REPO_ROOT/$PROJECT_DIR/Makefile" ]; then
	echo "Makefile not found at $REPO_ROOT/$PROJECT_DIR/Makefile" >&2
	echo "Run this installer from the project directory that contains Makefile." >&2
	exit 1
fi

if [ -f "$HOOK_PATH" ]; then
	BACKUP_PATH="$HOOK_PATH.bak"
	if [ -f "$BACKUP_PATH" ]; then
		BACKUP_PATH="$HOOK_PATH.bak.$(date +%Y%m%d%H%M%S)"
	fi
	cp "$HOOK_PATH" "$BACKUP_PATH"
	echo "Backed up existing hook to $BACKUP_PATH"
fi

cat > "$HOOK_PATH" <<HOOK
#!/bin/sh
set -eu

REPO_ROOT="$(git rev-parse --show-toplevel)"
PROJECT_DIR="$PROJECT_DIR"
cd "\$REPO_ROOT/\$PROJECT_DIR"

make deploy
HOOK

chmod +x "$HOOK_PATH"
echo "Installed $HOOK_NAME hook at $HOOK_PATH"

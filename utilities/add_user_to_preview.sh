#!/usr/bin/env bash
#
# add_user_to_preview.sh
#
# Copy a production user record into a preview environment's users table.
# The user is looked up in the production `cocktails-users` table via the
# username-index, then written into the `cocktails-pr-<PR>-users` table.

set -euo pipefail

PROG="$(basename "$0")"

usage() {
    cat <<EOF
Usage: $PROG -p PR -u USERNAME [-r AWS_REGION]

Copy an existing production user into a PR preview's users table.

Options:
  -p PR          Preview PR number (required).
  -u USERNAME    Existing production username to copy (required).
  -r AWS_REGION  AWS region (optional, defaults to us-east-1).
  -h, --help     Show this help message and exit.

Examples:
  $PROG -p 42 -u jorge
  $PROG -p 42 -u jorge -r eu-west-1
EOF
}

# Print an error message followed by the usage instructions, then exit.
die() {
    echo "Error: $*" >&2
    echo >&2
    usage >&2
    exit 1
}

# --- Parse arguments -------------------------------------------------------

PR=""
USERNAME=""
AWS_REGION="${AWS_REGION:-us-east-1}"

# Support long-form help before getopts (getopts only handles short options).
for arg in "$@"; do
    case "$arg" in
        -h|--help)
            usage
            exit 0
            ;;
    esac
done

while getopts ":p:u:r:h" opt; do
    case "$opt" in
        p) PR="$OPTARG" ;;
        u) USERNAME="$OPTARG" ;;
        r) AWS_REGION="$OPTARG" ;;
        h)
            usage
            exit 0
            ;;
        :)  die "Option -$OPTARG requires an argument." ;;
        \?) die "Unknown option: -$OPTARG" ;;
    esac
done

# --- Validate inputs -------------------------------------------------------

[ -n "$PR" ]       || die "PR number is required (-p)."
[ -n "$USERNAME" ] || die "Username is required (-u)."

case "$PR" in
    ''|*[!0-9]*) die "PR must be a positive integer, got: $PR" ;;
esac

command -v aws >/dev/null 2>&1 || die "'aws' CLI not found on PATH."

export AWS_REGION

# --- Look the user up in production ---------------------------------------

# Query the production username-index and pull out the first matching item.
ITEM=$(aws dynamodb query \
    --table-name cocktails-users \
    --index-name username-index \
    --key-condition-expression 'username = :u' \
    --expression-attribute-values "{\":u\":{\"S\":\"$USERNAME\"}}" \
    --region "$AWS_REGION" \
    --query 'Items[0]' \
    --output json) || die "Failed to query production user '$USERNAME'."

if [ -z "$ITEM" ] || [ "$ITEM" = "null" ]; then
    die "No production user found with username '$USERNAME'."
fi

# --- Write the record into the preview table ------------------------------

aws dynamodb put-item \
    --table-name "cocktails-pr-${PR}-users" \
    --item "$ITEM" \
    --region "$AWS_REGION" || die "Failed to write user into cocktails-pr-${PR}-users."

echo "Copied user '$USERNAME' into cocktails-pr-${PR}-users (region: $AWS_REGION)."

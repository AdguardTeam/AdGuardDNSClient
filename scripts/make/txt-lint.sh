#!/bin/sh

# This comment is used to simplify checking local copies of the script.  Bump
# this number every time a remarkable change is made to this script.
#
# AdGuard-Project-Version: 5

# TODO(e.burkov):  Add xmllint.  See msi/product.wxs.

verbose="${VERBOSE:-0}"
readonly verbose

if [ "$verbose" -gt '0' ]
then
	set -x
fi

# Set $EXIT_ON_ERROR to zero to see all errors.
if [ "${EXIT_ON_ERROR:-1}" -eq '0' ]
then
	set +e
else
	set -e
fi

# We don't need glob expansions and we want to see errors about unset variables.
set -f -u

# Source the common helpers, including not_found.
. ./scripts/make/helper.sh

# Simple analyzers

# trailing_newlines is a simple check that makes sure that all plain-text files
# have a trailing newlines to make sure that all tools work correctly with them.
trailing_newlines() (
	nl="$( printf "\n" )"
	readonly nl

	git ls-files\
		':!*.bmp'\
		| while read -r f
		do
			final_byte="$( tail -c -1 "$f" )"
			if [ "$final_byte" != "$nl" ]
			then
				printf '%s: must have a trailing newline\n' "$f"
			fi
		done
)

# trailing_whitespace is a simple check that makes sure that there are no
# trailing whitespace in plain-text files.
trailing_whitespace() {
	git ls-files\
		':!*.bmp'\
		| while read -r f
		do
			grep -e '[[:space:]]$' -n -- "$f"\
				| sed -e "s:^:${f}\::" -e 's/ \+$/>>>&<<</'
		done
}

run_linter -e trailing_newlines

run_linter -e trailing_whitespace

# TODO(e.burkov): Consider moving this script into the "Misc Lint" stage.
git ls-files -- '*.conf' '*.md' '*.txt' '*.yaml' '*.yml'\
	| xargs misspell --error\
	| sed -e 's/^/misspell: /'

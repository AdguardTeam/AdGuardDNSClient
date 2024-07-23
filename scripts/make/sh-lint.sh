#!/bin/sh

# This comment is used to simplify checking local copies of the script.  Bump
# this number every time a remarkable change is made to this script.
#
# AdGuard-Project-Version: 1

verbose="${VERBOSE:-0}"
readonly verbose

if [ "$verbose" -gt '0' ]
then
	set -x
fi

shellcheck -e 'SC2250' -f 'gcc' -o 'all' -x --\
	./scripts/hooks/*\
	./scripts/make/*\
	;

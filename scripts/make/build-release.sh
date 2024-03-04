#!/bin/sh

# AdGuardDNSClient Release Script
#
# The commentary in this file is written with the assumption that the reader
# only has superficial knowledge of the POSIX shell language and alike.
# Experienced readers may find it overly verbose.

# The default verbosity level is 0.  Show log messages if the caller requested
# verbosity level greater than 0.  Show the environment and every command that
# is run if the verbosity level is greater than 1.  Otherwise, print nothing.
#
# The level of verbosity for the build script is the same minus one level.  See
# below in build().
verbose="${VERBOSE:-0}"
readonly verbose

if [ "$verbose" -gt '1' ]
then
	env
	set -x
fi

# By default, sign the packages, but allow users to skip that step.
sign="${SIGN:-1}"
readonly sign

# Exit the script if a pipeline fails (-e), prevent accidental filename
# expansion (-f), and consider undefined variables as errors (-u).
set -e -f -u

# Function log is an echo wrapper that writes to stderr if the caller requested
# verbosity level greater than 0.  Otherwise, it does nothing.
log() {
	if [ "$verbose" -gt '0' ]
	then
		# Don't use quotes to get word splitting.
		echo "$1" 1>&2
	fi
}

log 'starting to build AdGuardDNSClient release'

# Require the channel to be set.  Additional validation is performed later by
# go-build.sh.
channel="${CHANNEL:?please set CHANNEL}"
readonly channel

# Check VERSION against the default value from the Makefile.  If it is that, use
# the version calculation script.
version="${VERSION:-}"
if [ "$version" = 'v0.0.0' ] || [ "$version" = '' ]
then
	version="$( sh ./scripts/make/version.sh )"
fi
readonly version

log "channel '$channel'"
log "version '$version'"

# Check architecture and OS limiters.  Add spaces to the local versions for
# better pattern matching.
if [ "${ARCH:-}" != '' ]
then
	log "arches: '$ARCH'"
	arches=" $ARCH "
else
	arches=''
fi
readonly arches

if [ "${OS:-}" != '' ]
then
	log "oses: '$OS'"
	oses=" $OS "
else
	oses=''
fi
readonly oses

# Require the gpg key and passphrase to be set if the signing is required.
if [ "$sign" -eq '1' ]
then
	gpg_key_passphrase="${GPG_KEY_PASSPHRASE:?please set GPG_KEY_PASSPHRASE or unset SIGN}"
	gpg_key="${GPG_KEY:?please set GPG_KEY or unset SIGN}"
	signer_api_key="${SIGNER_API_KEY:?please set SIGNER_API_KEY or unset SIGN}"
	deploy_script_path="${DEPLOY_SCRIPT_PATH:?please set DEPLOY_SCRIPT_PATH or unset SIGN}"
else
	gpg_key_passphrase=''
	gpg_key=''
	signer_api_key=''
	deploy_script_path=''
fi
readonly gpg_key_passphrase gpg_key signer_api_key deploy_script_path

# The default distribution files directory is dist.
dist="${DIST_DIR:-dist}"
readonly dist

log "checking tools"

# Make sure we fail gracefully if one of the tools we need is missing.  Use
# alternatives when available.
use_shasum='0'
for tool in gpg gzip sed sha256sum tar zip
do
	if ! command -v "$tool" > /dev/null
	then
		if [ "$tool" = 'sha256sum' ] && command -v 'shasum' > /dev/null
		then
			# macOS doesn't have sha256sum installed by default, but it does
			# have shasum.
			log 'replacing sha256sum with shasum -a 256'
			use_shasum='1'
		else
			log "pieces don't fit, '$tool' not found"

			exit 1
		fi
	fi
done
readonly use_shasum

# Data section.  Arrange data into space-separated tables for read -r to read.
# Use a hyphen for missing values.

# os     arch
platforms="\
darwin   arm64
darwin   amd64
linux    386
linux    amd64
linux    arm64
windows  amd64
windows  arm64"
readonly platforms

# Function sign signs the specified build as intended by the target operating
# system.
sign() {
	# Get the arguments.  Here and below, use the "sign_" prefix for all
	# variables local to function sign.
	sign_os="$1"
	sign_bin_path="$2"

	if [ "$sign_os" != 'windows' ]
	then
		gpg\
			--default-key "$gpg_key"\
			--detach-sig\
			--passphrase "$gpg_key_passphrase"\
			--pinentry-mode loopback\
			-q\
			"$sign_bin_path"\
			;

		return
	fi

	signed_bin_path="${sign_bin_path}.signed"

	env\
		INPUT_FILE="$sign_bin_path"\
		OUTPUT_FILE="$signed_bin_path"\
		SIGNER_API_KEY="$signer_api_key"\
		"$deploy_script_path" sign-executable\
		;

	mv "$signed_bin_path" "$sign_bin_path"
}

# Function build builds the release for one platform.  It builds a binary and an
# archive.
build() {
	# Get the arguments.  Here and below, use the "build_" prefix for all
	# variables local to function build.
	build_dir="${dist}/${1}/AdGuardDNSClient"\
		build_ar="$2"\
		build_os="$3"\
		build_arch="$4"\
		;

	# Use the ".exe" filename extension if we build a Windows release.
	if [ "$build_os" = 'windows' ]
	then
		build_output="./${build_dir}/AdGuardDNSClient.exe"
	else
		build_output="./${build_dir}/AdGuardDNSClient"
	fi

	mkdir -p "./${build_dir}"

	# Build the binary.
	env\
		GOARCH="$build_arch"\
		GOOS="$os"\
		VERBOSE="$(( verbose - 1 ))"\
		VERSION="$version"\
		OUT="$build_output"\
		sh ./scripts/make/go-build.sh\
		;

	log "$build_output"

	# Sign the binary if needed.
	if [ "$sign" -eq '1' ]
	then
		sign "$build_os" "$build_output"
	fi

	# Prepare the build directory for archiving.
	#
	# TODO(e.burkov):  Add CHANGELOG.md and LICENSE.txt.
	cp ./README.md "$build_dir"

	# Make archives.  Windows and macOS prefer ZIP archives; the rest, gzipped
	# tarballs.
	case "$build_os"
	in
	('darwin'|'windows')
		build_archive="./${dist}/${build_ar}.zip"
		# TODO(a.garipov): Find an option similar to the -C option of tar for
		# zip.
		( cd "${dist}/${1}" && zip -9 -q -r "../../${build_archive}" "./AdGuardDNSClient" )
		;;
	(*)
		build_archive="./${dist}/${build_ar}.tar.gz"
		tar -C "./${dist}/${1}" -c -f - "./AdGuardDNSClient" | gzip -9 - > "$build_archive"
		;;
	esac

	log "$build_archive"
}

log "starting builds"

# Go over all platforms defined in the space-separated table above, tweak the
# values where necessary, and feed to build.
echo "$platforms" | while read -r os arch
do
	# See if the architecture or the OS is in the allowlist.  To do so, try
	# removing everything that matches the pattern (well, a prefix, but that
	# doesn't matter here) containing the arch or the OS.
	#
	# For example, when $arches is " amd64 arm64 " and $arch is "amd64",
	# then the pattern to remove is "* amd64 *", so the whole string becomes
	# empty.  On the other hand, if $arch is "windows", then the pattern is
	# "* windows *", which doesn't match, so nothing is removed.
	#
	# See https://stackoverflow.com/a/43912605/1892060.
	if [ "${arches##* $arch *}" != '' ]
	then
		log "$arch excluded, continuing"

		continue
	elif [ "${oses##* $os *}" != '' ]
	then
		log "$os excluded, continuing"

		continue
	fi

	dir="AdGuardDNSClient_${os}_${arch}"
	# Name archive the same as the corresponding distribution directory.
	ar="$dir"

	build "$dir" "$ar" "$os" "$arch"
done

log "calculating checksums"

# calculate_checksums uses the previously detected SHA-256 tool to calculate
# checksums.  Do not use find with -exec, since shasum requires arguments.
calculate_checksums() {
	if [ "$use_shasum" -eq '0' ]
	then
		sha256sum "$@"
	else
		shasum -a 256 "$@"
	fi
}

# Calculate the checksums of the files in a subshell with a different working
# directory.  Don't use ls, because files matching one of the patterns may be
# absent, which will make ls return with a non-zero status code.
#
# TODO(a.garipov): Consider calculating these as the build goes.
(
	set +f

	cd "./${dist}"

	: > ./checksums.txt

	for archive in ./*.zip ./*.tar.gz
	do
		# Make sure that we don't try to calculate a checksum for a glob pattern
		# that matched no files.
		if [ ! -f "$archive" ]
		then
			continue
		fi

		calculate_checksums "$archive" >> ./checksums.txt
	done
)

log "writing versions"

echo "version=$version" > "./${dist}/version.txt"

log "finished"

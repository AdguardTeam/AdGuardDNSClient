# Keep the Makefile POSIX-compliant.  We currently allow hyphens in
# target names, but that may change in the future.
#
# See https://pubs.opengroup.org/onlinepubs/9699919799/utilities/make.html.
.POSIX:

# This comment is used to simplify checking local copies of the
# Makefile.  Bump this number every time a significant change is made to
# this Makefile.
#
# AdGuard-Project-Version: 5

# Don't name these macros "GO" etc., because GNU Make apparently makes
# them exported environment variables with the literal value of
# "${GO:-go}" and so on, which is not what we need.  Use a dot in the
# name to make sure that users don't have an environment variable with
# the same name.
#
# See https://unix.stackexchange.com/q/646255/105635.
GO.MACRO = $${GO:-go}
VERBOSE.MACRO = $${VERBOSE:-0}

BRANCH = $$( git rev-parse --abbrev-ref HEAD )
CHANNEL = development
DEPLOY_SCRIPT_PATH = not/a/real/path
DIST_DIR = dist
GOAMD64 = v1
GOPROXY = https://proxy.golang.org|direct
GOTOOLCHAIN = go1.22.4
GPG_KEY = devteam@adguard.com
GPG_KEY_PASSPHRASE = not-a-real-password
MSI = 1
RACE = 0
REVISION = $$( git rev-parse --short HEAD )
SIGN = 1
SIGNER_API_KEY = not-a-real-key
VERSION = v0.0.0

ENV = env\
	BRANCH="$(BRANCH)"\
	CHANNEL="$(CHANNEL)"\
	DEPLOY_SCRIPT_PATH='$(DEPLOY_SCRIPT_PATH)' \
	DIST_DIR='$(DIST_DIR)'\
	GO="$(GO.MACRO)"\
	GOAMD64='$(GOAMD64)'\
	GOPROXY='$(GOPROXY)'\
	GOTOOLCHAIN='$(GOTOOLCHAIN)'\
	GPG_KEY='$(GPG_KEY)'\
	GPG_KEY_PASSPHRASE='$(GPG_KEY_PASSPHRASE)'\
	MSI='$(MSI)'\
	PATH="$${PWD}/bin:$$( "$(GO.MACRO)" env GOPATH )/bin:$${PATH}"\
	RACE='$(RACE)'\
	REVISION="$(REVISION)"\
	SIGN='$(SIGN)'\
	SIGNER_API_KEY='$(SIGNER_API_KEY)' \
	VERBOSE="$(VERBOSE.MACRO)"\
	VERSION="$(VERSION)"\

# Keep the line above blank.

# Keep this target first, so that a naked make invocation triggers a
# full build.
build: go-deps go-build

init: ; git config core.hooksPath ./scripts/hooks

test: go-test

go-build: ; $(ENV)          "$(SHELL)" ./scripts/make/go-build.sh
go-deps:  ; $(ENV)          "$(SHELL)" ./scripts/make/go-deps.sh
go-lint:  ; $(ENV)          "$(SHELL)" ./scripts/make/go-lint.sh
go-test:  ; $(ENV) RACE='1' "$(SHELL)" ./scripts/make/go-test.sh
go-tools: ; $(ENV)          "$(SHELL)" ./scripts/make/go-tools.sh

go-upd-tools: ; $(ENV) "$(SHELL)" ./scripts/make/go-upd-tools.sh

go-check: go-tools go-lint go-test

# A quick check to make sure that all operating systems relevant to the
# development of the project can be typechecked and built successfully.
go-os-check:
	env GOOS='darwin'  "$(GO.MACRO)" vet ./internal/...
	env GOOS='linux'   "$(GO.MACRO)" vet ./internal/...
	env GOOS='windows' "$(GO.MACRO)" vet ./internal/...

txt-lint: ; $(ENV) "$(SHELL)" ./scripts/make/txt-lint.sh

build-release: ; $(ENV) "$(SHELL)" ./scripts/make/build-release.sh

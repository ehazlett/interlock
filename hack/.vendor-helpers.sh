#!/usr/bin/env bash

PROJECT=github.com/docker/containerd

# Downloads dependencies into vendor/ directory
mkdir -p vendor

export GOPATH="$GOPATH:${PWD}/vendor"

find='find'
if [ "$(go env GOHOSTOS)" = 'windows' ]; then
	find='/usr/bin/find'
fi

clone() {
	local vcs="$1"
	local pkg="$2"
	local rev="$3"
	local url="$4"

	: ${url:=https://$pkg}
	local target="vendor/src/$pkg"

	echo -n "$pkg @ $rev: "

	if [ -d "$target" ]; then
		echo -n 'rm old, '
		rm -rf "$target"
	fi

	echo -n 'clone, '
	case "$vcs" in
		git)
			git clone --quiet --no-checkout "$url" "$target"
			( cd "$target" && git checkout --quiet "$rev" && git reset --quiet --hard "$rev" )
			;;
		hg)
			hg clone --quiet --updaterev "$rev" "$url" "$target"
			;;
	esac

	echo -n 'rm VCS, '
	( cd "$target" && rm -rf .{git,hg} )

	echo -n 'rm vendor, '
	( cd "$target" && rm -rf vendor Godeps/_workspace )

	echo done
}

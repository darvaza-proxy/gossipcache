#!/bin/sh

set -eu

: ${GO:=go}

MODULES=$(find * -name go.mod -exec dirname '{}' \;)
GROUPS="pkg" # Add all groups as space-separated list

mod() {
	local d="${1:-.}"
	grep ^module "$d/go.mod" | cut -d' ' -f2
}

namedir() {
	local d="$1" g= n=

	if [ "." = "$d" ]; then
		echo "root"
		return
	fi

	for g in $GROUPS; do
		n="${d#$g/}"
		if [ "x$n" != "x$d" ]; then
			echo "$n" | tr '/' '-'
			return
		fi
	done

	echo "$d" | tr '/' '-'
}

mod_replace() {
	local d="$1"
	if [ ! -f "$d/go.mod" ] || [ ! -r "$d/go.mod" ]; then
		echo "Error: Cannot read $d/go.mod" >&2
		return 1
	fi
	grep "=>" "$d/go.mod" | sed -n -e "s;^.*\($ROOT_MODULE.*\)[ \t]\+=>.*;\1;p"
}

gen_index() {
	local d= n=

	for d; do
		n=$(namedir "$d")
		m=$(mod "$d")
		echo "$n:$d:$m"
	done
}

ROOT_MODULE=$(mod) || exit 1
if [ -z "$ROOT_MODULE" ]; then
    echo "Error: Could not determine root module" >&2
    exit 1
fi
INDEX=$(gen_index $MODULES)

echo "$INDEX" | while IFS=: read name dir mod; do
	deps=
	for dep in $(mod_replace "$dir"); do
		depname=$(echo "$INDEX" | grep ":$dep$" | cut -d: -f1 | tr '\n' ',' | sed -e 's|,\+$||g')
		if [ -n "$depname" ]; then
			deps="${deps:+$deps,}$depname"
		fi
	done

	echo "$name:$dir:$mod:$deps"
done | sort -V

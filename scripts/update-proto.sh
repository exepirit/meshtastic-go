#!/usr/bin/env bash

tmpdir="/tmp/meshtastic-protobufs"

if [ ! -d "$tmpdir" ]; then
  git clone https://github.com/meshtastic/protobufs "$tmpdir"
fi
protobufs_version=$(cd "$tmpdir" && git describe --tags --abbrev=0)
(cd "$tmpdir" && git checkout "$protobufs_version")

rm -r pkg/meshtastic/proto || true
protoc -I "$tmpdir" \
  --go_out=pkg/meshtastic \
  --go_opt=module=github.com/meshtastic/go \
  "$tmpdir/nanopb.proto" $tmpdir/meshtastic/*.proto

find pkg/meshtastic/generated -type f | while read file; do
  sed -i 's/package generated/package proto/g' "$file"
done

mv pkg/meshtastic/generated pkg/meshtastic/proto

echo "Protocols upgraded to version $protobufs_version"

#! /bin/bash

for plugin in `ls -d plugins/*/`; do
  echo "Building $plugin"

  sha1=`shasum -a 1 $plugin/plugin.go | awk '{print $1}' | cut -c -7`
  plugin_dir="$plugin/plugin.$sha1"
  if [ -d "$plugin_dir" ]; then
    rm -rf "$plugin_dir"
  fi
  cp -r $plugin $plugin_dir

  sed -i '' "s/package [a-zA-Z0-9]*/package main/g" $plugin_dir/plugin.go
  go build -buildmode=plugin -o $plugin/plugin.$sha1.so $plugin_dir/plugin.go
  rm -rf "$plugin_dir"
done
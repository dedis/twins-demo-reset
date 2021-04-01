#! /usr/bin/env bash
set -xe

# cleanup previous installations
rm -rf deb

mkdir -p deb/opt/twins/bin     # install dir

# binary built by CI
mv twins-demo-reset deb/opt/twins/bin

VERSION=$(./twins-demo-reset -v | tr - +)

# add systemd units
cp -a pkg/usr deb

# adjust permissions
find deb ! -perm -a+r -exec chmod a+r {} \;

fpm \
    --force -t deb -a all -s dir -C deb -n twins-demo-reset -v ${VERSION:1} \
    --before-install pkg/before-install.sh \
    --after-install pkg/after-install.sh \
    --before-remove pkg/before-remove.sh \
    --after-remove pkg/after-remove.sh \
    --url https://github.com/dedis/twins-demo-reset \
    --description 'An HTTP server to reset twins demo' \
    --package dist \
    .

# cleanup
rm -rf ./deb


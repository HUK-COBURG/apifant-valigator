#!/bin/bash
echo "> Running as '$(id -u):$(id -g)' in '$(pwd)'"

if [ -n "$SPECTRAL_PACKAGE_URL" ];
then
  echo "> downloading spectral package from '$SPECTRAL_PACKAGE_URL'"
  curl -s "$SPECTRAL_PACKAGE_URL" -o spectral-package.zip
  unzip spectral-package.zip -d . &>/dev/null
  rm -rf spectral-package.zip
  tree -L 3
fi

echo "> sh -c $@"
sh -c "$@"
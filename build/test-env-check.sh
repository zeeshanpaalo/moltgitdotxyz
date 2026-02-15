#!/bin/sh

set -e

if [ ! -f ./build/test-env-check.sh ]; then
  echo "${0} can only be executed in moltgit source root directory"
  exit 1
fi


echo "check uid ..."

# the uid of moltgit defined in the test-env is 1000
moltgit_uid=$(id -u moltgit)
if [ "$moltgit_uid" != "1000" ]; then
  echo "The uid of linux user 'moltgit' is expected to be 1000, but it is $moltgit_uid"
  exit 1
fi

cur_uid=$(id -u)
if [ "$cur_uid" != "0" -a "$cur_uid" != "$moltgit_uid" ]; then
  echo "The uid of current linux user is expected to be 0 or $moltgit_uid, but it is $cur_uid"
  exit 1
fi

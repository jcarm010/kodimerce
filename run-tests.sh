#!/usr/bin/env bash
set -e
root=${PWD}
#cd entities && goapp test && cd ${root}
cd initial_config && goapp test && cd ${root}
echo "All Tests Completed Successfully"
#!/bin/bash

set -eux

if [ $# -ne 1 ]; then
    echo "usage: ./compile.sh <target>" 1>&2
    exit 1
fi

CURDIR=$(
    cd $(dirname $0)
    pwd
)

TARGET=$1

cp ${CURDIR}/${TARGET}.jack ${CURDIR}/${TARGET}Test
${CURDIR}/../../tools/JackCompiler.sh ${CURDIR}/${TARGET}Test

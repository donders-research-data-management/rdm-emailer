#!/bin/bash

for e in $( iquest --no-page "%s,%s" "SELECT USER_NAME,META_USER_ATTR_VALUE WHERE USER_TYPE = 'rodsuser' AND META_USER_ATTR_NAME = 'email'" ); do
    uid=$( echo $e | awk -F ',' '{print $1}')
    email=$( echo $e | awk -F ',' '{print $2}')
    uname=$( iquest "%s" "SELECT META_USER_ATTR_VALUE WHERE USER_NAME = '$uid' AND META_USER_ATTR_NAME = 'displayName'" )

    echo ${email},"\"${uname}\""
done

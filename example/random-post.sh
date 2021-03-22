#!/bin/bash

URL=$LB_ADDR

MAXCOUNT=100000
count=1

while [ "$count" -le $MAXCOUNT ]
do
  let "cellname = $RANDOM % 256"

  let "operation=$RANDOM % 5"
  #echo $operation
  if [ "$operation" -lt 4 ]
  then
    ./simple-post -print 2
  else
    curl -s -N -H "Content-type: application/json" $URL/rsrp/$cellname | head -c 40 > /dev/null
  fi

  let "count += 1"
done

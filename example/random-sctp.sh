#!/bin/bash

IP=10.0.0.234
PORT=30001

MAXCOUNT=1500
count=1
while [ "$count" -le $MAXCOUNT ]
do
  ./sctp -ip $IP -port $PORT -print 2
  #sleep 1
  let "count += 1"
done

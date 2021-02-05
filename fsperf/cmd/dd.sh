#!/bin/bash

dd_perfdir=$1
size_outfile_mb=$2
timeout=$3

if [[ $# -lt 3 ]]
then
  echo "I need 3 arguments - dir to perf, timeout, size of file to write/read in mb"
  echo "eg. test a 1GiB file at /some/directory, timeout after 4s)"
  echo "dd.sh /some/directory 1024 4"
  exit 1
fi

if [ ! -d "$dd_perfdir" ];then echo "Perf dir $dd_perfdir doesn't exist, exiting!";exit 1;fi

#dd_bs=$2
#dd_count=$3
#dd_results_file=$4
dd_bs=64k
dd_count=$(($size_outfile_mb*1024/64))
dd_results_file=$(mktemp)
#dd_outfile=${dd_perfdir}/__dd_test_outfile
dd_outfile=`mktemp -p ${dd_perfdir}`
kill_timeout=$(($timeout+1))

write_timeout=0
read_timeout=0
if [[ "$timeout" != "" ]]
then
  /usr/bin/timeout -k $kill_timeout -s SIGINT $timeout dd if=/dev/zero of=$dd_outfile bs=$dd_bs count=$dd_count 2> $dd_results_file
  #echo 3 > /proc/sys/vm/drop_caches
  if [[ $? -eq 137 ]]
  then
    bytes=0
    time_write=0
    rate=0
    write_timeout=1
  else
    bytes=`cat $dd_results_file | grep copied | sed -e 's/\(^[0-9]*\) .*/\1/g'`
    time_write=`cat $dd_results_file | grep copied | sed -e 's/.*copied, \([^ ]*\) .*/\1/g'`
    rate=`echo "$bytes / $time_write" | bc`
  fi
  /usr/bin/timeout -k $kill_timeout -s SIGINT $timeout dd if=$dd_outfile of=/dev/null bs=$dd_bs iflag=nocache 2> ${dd_results_file}_read
  if [[ $? -eq 137 ]]
  then 
    bytes_read=0
    time_read=0
    rate_read=0
    read_timeout=1
  else 
    bytes_read=`cat ${dd_results_file}_read | grep copied | sed -e 's/\(^[0-9]*\) .*/\1/g'`
    time_read=`cat ${dd_results_file}_read | grep copied | sed -e 's/.*copied, \([^ ]*\) .*/\1/g'`
    rate_read=`echo "$bytes_read / $time_read" | bc`
  fi
else
  dd if=/dev/zero of=$dd_outfile bs=$dd_bs count=$dd_count 2> $dd_results_file
  #echo 3 > /proc/sys/vm/drop_caches
  dd if=$dd_outfile of=/dev/null bs=$dd_bs 2> ${dd_results_file}_read
fi

jq -n --arg bs "$dd_bs" --arg count "$dd_count" --arg wtm "$time_write" --arg wbts "$bytes" --arg wrt "$rate" --arg wto "$write_timeout" --arg rtm "$time_read" --arg rbts "$bytes_read" --arg rrt "$rate_read" --arg rto "$read_timeout" '{dd: {bs: $bs, count: $count, write_time: $wtm, write_bytes: $wbts, write_rate: $wrt, write_timeout: $wto, read_time: $rtm, read_bytes: $rbts, read_rate: $rrt, read_timeout: $rto} }' 

rm -f $dd_results_file
rm -f ${dd_results_file}_read
rm -f ${dd_outfile}

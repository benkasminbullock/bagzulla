#!/bin/sh
./stop.pl
make
nohup ./bagzulla --url "http://mikan/bagpub" --display "http://mikan/files/fileserver.cgi?file=./projects/" -port 1919 > log 2>&1 &

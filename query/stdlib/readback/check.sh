#!/bin/bash

passed()
{
	echo -n "tests that passed where ID contains $2: "
	grep "^test passed but checkid $1 failed:" log-*.txt | grep $2 | wc -l
}

echo FAILURES: 
grep '^readback was empty:' log-*-*.txt | sed 's/^.*empty: //'

passed 20 20
passed 2c 2c
passed 5c 5c

passed 5c 5c20
passed 5c 5c22
passed 5c 5c2c
passed 5c 5c3d

passed 20 2020
passed 2c 2c2c
passed 5c 5c5c

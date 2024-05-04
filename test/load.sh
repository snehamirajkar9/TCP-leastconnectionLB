#!/bin/bash
for j in $(seq 1 30)
    do bash load-test.sh $j&
done

wait
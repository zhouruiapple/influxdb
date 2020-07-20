# How many test case processes to run in parallel.
PARALLEL=16

# How many iterations of of parallel runs to make.
ITERATIONS=32

logfn() {
	printf "log-%02d-%02d.txt" $1 $2
}

for run in `seq 1 $ITERATIONS`; do
	for i in `seq 1 $PARALLEL`; do
		LOG=`logfn $run $i`
		./readback.test &>$LOG &
		sleep 0.5
	done

	for i in `seq 1 $PARALLEL`; do
			wait
	done
done

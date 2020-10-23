
Build binaries

    make build

Start services running in background or start these in another terminal.

    PORT=8080 ./server &
    PORT=5000 ./serverA &
    PORT=6000 ./serverB &

Run load tests over the services

Case 1

    ab -n 500000 -c 50 0.0.0.0:8080/withoutgoroutine
    ...
    Percentage of the requests served within a certain time (ms)
      50%      3
      66%      4
      75%      4
      80%      4
      90%      5
      95%      7
      98%     10
      99%     11
     100%     38 (longest request)

Case 2

    ab -n 500000 -c 50 0.0.0.0:8080/withgoroutine
    ...
    Percentage of the requests served within a certain time (ms)
      50%      5
      66%      5
      75%      6
      80%      6
      90%      8
      95%     10
      98%     14
      99%     16
     100%     42 (longest request)


Case 3

    ab -n 500000 -c 50 0.0.0.0:8080/withsleepygoroutine
    Percentage of the requests served within a certain time (ms)
    ...
      50%      3
      66%      4
      75%      4
      80%      4
      90%      5
      95%      7
      98%     10
      99%     12
     100%     31 (longest request)


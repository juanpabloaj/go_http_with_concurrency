
Build binaries

    make build

Start services running in background or start these in another terminal.

    PORT=8080 ./server &
    PORT=5000 ./serverA &
    PORT=6000 ./serverB &

Run load tests over the services

    ab -n 100000 -c 50 0.0.0.0:8080/withoutgoroutine
    ab -n 100000 -c 50 0.0.0.0:8080/withgoroutine
    ab -n 100000 -c 50 0.0.0.0:8080/withsleepygoroutine

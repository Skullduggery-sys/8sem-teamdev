## Результаты теста c 2мя доп. инстансами
```h
k6 run ./nginx/test/script.js

         /\      Grafana   /‾‾/
    /\  /  \     |\  __   /  /
   /  \/    \    | |/ /  /   ‾‾\
  /          \   |   (  |  (‾)  |
 / __________ \  |_|\_\  \_____/

     execution: local
        script: ./nginx/test/script.js
        output: -

     scenarios: (100.00%) 1 scenario, 20 max VUs, 1m30s max duration (incl. graceful stop):
              * default: Up to 20 looping VUs for 1m0s over 3 stages (gracefulRampDown: 30s, gracefulStop: 30s)


     ✓ status is 200
     ✓ response body is not empty

     checks.........................: 100.00% 668 out of 668
     data_received..................: 206 kB  3.4 kB/s
     data_sent......................: 63 kB   1.0 kB/s
     http_req_blocked...............: avg=14.95µs min=2.74µs   med=4.97µs  max=342.51µs p(90)=7.06µs  p(95)=141.22µs
     http_req_connecting............: avg=7.06µs  min=0s       med=0s      max=232.1µs  p(90)=0s      p(95)=98.56µs
     http_req_duration..............: avg=3.06s   min=890.44ms med=2.88s   max=5.44s    p(90)=4.89s   p(95)=5.06s
       { expected_response:true }...: avg=3.06s   min=890.44ms med=2.88s   max=5.44s    p(90)=4.89s   p(95)=5.06s
     http_req_failed................: 0.00%   0 out of 334
     http_req_receiving.............: avg=74.16µs min=32.57µs  med=69.53µs max=194.33µs p(90)=92.31µs p(95)=107.71µs
     http_req_sending...............: avg=41.5µs  min=10.04µs  med=19.88µs max=6.45ms   p(90)=30.68µs p(95)=40.16µs
     http_req_tls_handshaking.......: avg=0s      min=0s       med=0s      max=0s       p(90)=0s      p(95)=0s
     http_req_waiting...............: avg=3.06s   min=890.23ms med=2.88s   max=5.44s    p(90)=4.89s   p(95)=5.06s
     http_reqs......................: 334     5.542208/s
     iteration_duration.............: avg=3.06s   min=890.92ms med=2.88s   max=5.44s    p(90)=4.89s   p(95)=5.06s
     iterations.....................: 334     5.542208/s
     vus............................: 2       min=2          max=20
     vus_max........................: 20      min=20         max=20


running (1m00.3s), 00/20 VUs, 334 complete and 0 interrupted iterations
default ✓ [======================================] 00/20 VUs  1m0s
```

## Результат грепа по логам nginx
```h
$ awk '{print $NF}' /var/log/nginx/access.log | sort | uniq -c
      1 "-"
    506 "127.0.0.1:9000"
     84 "127.0.0.1:9001"
     83 "127.0.0.1:9002"
```


### Снепшот логов nginx
```h
127.0.0.1 - - [02/Mar/2025:19:04:28 +0300] "POST /api/v2/sign-in HTTP/1.1" 200 200 "-" "k6/0.55.0 (https://k6.io/)" "-" "127.0.0.1:9000"
127.0.0.1 - - [02/Mar/2025:19:04:28 +0300] "POST /api/v2/sign-in HTTP/1.1" 200 200 "-" "k6/0.55.0 (https://k6.io/)" "-" "127.0.0.1:9002"
127.0.0.1 - - [02/Mar/2025:19:04:29 +0300] "POST /api/v2/sign-in HTTP/1.1" 200 200 "-" "k6/0.55.0 (https://k6.io/)" "-" "127.0.0.1:9000"
127.0.0.1 - - [02/Mar/2025:19:04:29 +0300] "POST /api/v2/sign-in HTTP/1.1" 200 200 "-" "k6/0.55.0 (https://k6.io/)" "-" "127.0.0.1:9001"
127.0.0.1 - - [02/Mar/2025:19:04:29 +0300] "POST /api/v2/sign-in HTTP/1.1" 200 200 "-" "k6/0.55.0 (https://k6.io/)" "-" "127.0.0.1:9002"
127.0.0.1 - - [02/Mar/2025:19:04:29 +0300] "POST /api/v2/sign-in HTTP/1.1" 200 200 "-" "k6/0.55.0 (https://k6.io/)" "-" "127.0.0.1:9000"
127.0.0.1 - - [02/Mar/2025:19:04:30 +0300] "POST /api/v2/sign-in HTTP/1.1" 200 200 "-" "k6/0.55.0 (https://k6.io/)" "-" "127.0.0.1:9000"
127.0.0.1 - - [02/Mar/2025:19:04:30 +0300] "POST /api/v2/sign-in HTTP/1.1" 200 200 "-" "k6/0.55.0 (https://k6.io/)" "-" "127.0.0.1:9000"
127.0.0.1 - - [02/Mar/2025:19:04:30 +0300] "POST /api/v2/sign-in HTTP/1.1" 200 200 "-" "k6/0.55.0 (https://k6.io/)" "-" "127.0.0.1:9002"
127.0.0.1 - - [02/Mar/2025:19:04:30 +0300] "POST /api/v2/sign-in HTTP/1.1" 200 200 "-" "k6/0.55.0 (https://k6.io/)" "-" "127.0.0.1:9001"
127.0.0.1 - - [02/Mar/2025:19:04:30 +0300] "POST /api/v2/sign-in HTTP/1.1" 200 200 "-" "k6/0.55.0 (https://k6.io/)" "-" "127.0.0.1:9000"
127.0.0.1 - - [02/Mar/2025:19:04:31 +0300] "POST /api/v2/sign-in HTTP/1.1" 200 200 "-" "k6/0.55.0 (https://k6.io/)" "-" "127.0.0.1:9000"
127.0.0.1 - - [02/Mar/2025:19:04:31 +0300] "POST /api/v2/sign-in HTTP/1.1" 200 200 "-" "k6/0.55.0 (https://k6.io/)" "-" "127.0.0.1:9001"
```

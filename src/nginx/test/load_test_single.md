## Результаты одиночного теста
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

     checks.........................: 100.00% 648 out of 648
     data_received..................: 200 kB  3.3 kB/s
     data_sent......................: 61 kB   1.0 kB/s
     http_req_blocked...............: avg=15.1µs  min=2.83µs   med=5.01µs  max=212.78µs p(90)=7.68µs  p(95)=142.93µs
     http_req_connecting............: avg=6.99µs  min=0s       med=0s      max=149.93µs p(90)=0s      p(95)=95.76µs
     http_req_duration..............: avg=3.17s   min=898.72ms med=3.68s   max=4.51s    p(90)=4.01s   p(95)=4.16s
       { expected_response:true }...: avg=3.17s   min=898.72ms med=3.68s   max=4.51s    p(90)=4.01s   p(95)=4.16s
     http_req_failed................: 0.00%   0 out of 324
     http_req_receiving.............: avg=76.19µs min=29.7µs   med=71.2µs  max=228.43µs p(90)=91.35µs p(95)=112.1µs
     http_req_sending...............: avg=33.84µs min=10.54µs  med=19.77µs max=3.72ms   p(90)=30.74µs p(95)=41.81µs
     http_req_tls_handshaking.......: avg=0s      min=0s       med=0s      max=0s       p(90)=0s      p(95)=0s
     http_req_waiting...............: avg=3.17s   min=898.63ms med=3.68s   max=4.51s    p(90)=4.01s   p(95)=4.16s
     http_reqs......................: 324     5.340421/s
     iteration_duration.............: avg=3.17s   min=898.86ms med=3.68s   max=4.51s    p(90)=4.01s   p(95)=4.16s
     iterations.....................: 324     5.340421/s
     vus............................: 1       min=1          max=20
     vus_max........................: 20      min=20         max=20


running (1m00.7s), 00/20 VUs, 324 complete and 0 interrupted iterations
default ✓ [======================================] 00/20 VUs  1m0s
```

## Результат грепа по логам nginx
```h
$ awk '{print $NF}' /var/log/nginx/access.log | sort | uniq -c
    1 "-"
    339 "127.0.0.1:9000"
```


### Снепшот логов nginx
```h
127.0.0.1 - - [02/Mar/2025:18:57:50 +0300] "POST /api/v2/sign-in HTTP/1.1" 200 200 "-" "k6/0.55.0 (https://k6.io/)" "-" "127.0.0.1:9000"
127.0.0.1 - - [02/Mar/2025:18:57:51 +0300] "POST /api/v2/sign-in HTTP/1.1" 200 200 "-" "k6/0.55.0 (https://k6.io/)" "-" "127.0.0.1:9000"
127.0.0.1 - - [02/Mar/2025:18:57:51 +0300] "POST /api/v2/sign-in HTTP/1.1" 200 200 "-" "k6/0.55.0 (https://k6.io/)" "-" "127.0.0.1:9000"
127.0.0.1 - - [02/Mar/2025:18:57:51 +0300] "POST /api/v2/sign-in HTTP/1.1" 200 200 "-" "k6/0.55.0 (https://k6.io/)" "-" "127.0.0.1:9000"
127.0.0.1 - - [02/Mar/2025:18:57:51 +0300] "POST /api/v2/sign-in HTTP/1.1" 200 200 "-" "k6/0.55.0 (https://k6.io/)" "-" "127.0.0.1:9000"
127.0.0.1 - - [02/Mar/2025:18:57:51 +0300] "POST /api/v2/sign-in HTTP/1.1" 200 200 "-" "k6/0.55.0 (https://k6.io/)" "-" "127.0.0.1:9000"
127.0.0.1 - - [02/Mar/2025:18:57:52 +0300] "POST /api/v2/sign-in HTTP/1.1" 200 200 "-" "k6/0.55.0 (https://k6.io/)" "-" "127.0.0.1:9000"
127.0.0.1 - - [02/Mar/2025:18:57:52 +0300] "POST /api/v2/sign-in HTTP/1.1" 200 200 "-" "k6/0.55.0 (https://k6.io/)" "-" "127.0.0.1:9000"
127.0.0.1 - - [02/Mar/2025:18:57:52 +0300] "POST /api/v2/sign-in HTTP/1.1" 200 200 "-" "k6/0.55.0 (https://k6.io/)" "-" "127.0.0.1:9000"
127.0.0.1 - - [02/Mar/2025:18:57:52 +0300] "POST /api/v2/sign-in HTTP/1.1" 200 200 "-" "k6/0.55.0 (https://k6.io/)" "-" "127.0.0.1:9000"
127.0.0.1 - - [02/Mar/2025:18:57:52 +0300] "POST /api/v2/sign-in HTTP/1.1" 200 200 "-" "k6/0.55.0 (https://k6.io/)" "-" "127.0.0.1:9000"
127.0.0.1 - - [02/Mar/2025:18:57:53 +0300] "POST /api/v2/sign-in HTTP/1.1" 200 200 "-" "k6/0.55.0 (https://k6.io/)" "-" "127.0.0.1:9000"
```

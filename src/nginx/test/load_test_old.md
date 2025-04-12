## Результаты теста c 2мя доп. инстансами
```h
     checks.........................: 100.00% 1148560 out of 1148560
     data_received..................: 350 MB  2.9 MB/s
     data_sent......................: 94 MB   785 kB/s
     http_req_blocked...............: avg=12.29µs min=1.49µs   med=4.42µs  max=23.61ms p(90)=6.87µs   p(95)=9.53µs
     http_req_connecting............: avg=3.28µs  min=0s       med=0s      max=20.68ms p(90)=0s       p(95)=0s
     http_req_duration..............: avg=2.93ms  min=304.78µs med=2.19ms  max=56ms    p(90)=5.81ms   p(95)=7.59ms
       { expected_response:true }...: avg=2.93ms  min=304.78µs med=2.19ms  max=56ms    p(90)=5.81ms   p(95)=7.59ms
     http_req_failed................: 0.00%   0 out of 574280
     http_req_receiving.............: avg=91.89µs min=14.85µs  med=43.95µs max=27.89ms p(90)=105.51µs p(95)=215.22µs
     http_req_sending...............: avg=25.21µs min=4.88µs   med=12.9µs  max=18.89ms p(90)=23.03µs  p(95)=34.17µs
     http_req_tls_handshaking.......: avg=0s      min=0s       med=0s      max=0s      p(90)=0s       p(95)=0s
     http_req_waiting...............: avg=2.81ms  min=268.24µs med=2.08ms  max=55.91ms p(90)=5.64ms   p(95)=7.39ms
     http_reqs......................: 574280  4785.646798/s
     iteration_duration.............: avg=3.13ms  min=356.35µs med=2.38ms  max=56.12ms p(90)=6.11ms   p(95)=7.94ms
     iterations.....................: 574280  4785.646798/s
     vus............................: 1       min=1                  max=20
     vus_max........................: 20      min=20                 max=20


running (2m00.0s), 00/20 VUs, 574280 complete and 0 interrupted iterations
```

## Результат грепа по логам nginx
```h
$ awk '{print $NF}' /var/log/nginx/access.log | sort | uniq -c

 287149 "127.0.0.1:9000"
 143571 "127.0.0.1:9001"
 143571 "127.0.0.1:9002"
```


### Снепшот логов nginx
```h
127.0.0.1 - - [24/Feb/2025:22:06:30 +0300] "POST /api/v1/get-user-token?admin_secret=local_admin&login=alivasilyev HTTP/1.1" 200 193 "-" "k6/0.55.0 (https://k6.io/)" "-" "127.0.0.1:9000"
127.0.0.1 - - [24/Feb/2025:22:06:30 +0300] "POST /api/v1/get-user-token?admin_secret=local_admin&login=alivasilyev HTTP/1.1" 200 193 "-" "k6/0.55.0 (https://k6.io/)" "-" "127.0.0.1:9001"
127.0.0.1 - - [24/Feb/2025:22:06:30 +0300] "POST /api/v1/get-user-token?admin_secret=local_admin&login=alivasilyev HTTP/1.1" 200 193 "-" "k6/0.55.0 (https://k6.io/)" "-" "127.0.0.1:9002"
127.0.0.1 - - [24/Feb/2025:22:06:30 +0300] "POST /api/v1/get-user-token?admin_secret=local_admin&login=alivasilyev HTTP/1.1" 200 193 "-" "k6/0.55.0 (https://k6.io/)" "-" "127.0.0.1:9000"
127.0.0.1 - - [24/Feb/2025:22:06:30 +0300] "POST /api/v1/get-user-token?admin_secret=local_admin&login=alivasilyev HTTP/1.1" 200 193 "-" "k6/0.55.0 (https://k6.io/)" "-" "127.0.0.1:9000"
127.0.0.1 - - [24/Feb/2025:22:06:30 +0300] "POST /api/v1/get-user-token?admin_secret=local_admin&login=alivasilyev HTTP/1.1" 200 193 "-" "k6/0.55.0 (https://k6.io/)" "-" "127.0.0.1:9000"
127.0.0.1 - - [24/Feb/2025:22:06:30 +0300] "POST /api/v1/get-user-token?admin_secret=local_admin&login=alivasilyev HTTP/1.1" 200 193 "-" "k6/0.55.0 (https://k6.io/)" "-" "127.0.0.1:9002"
127.0.0.1 - - [24/Feb/2025:22:06:30 +0300] "POST /api/v1/get-user-token?admin_secret=local_admin&login=alivasilyev HTTP/1.1" 200 193 "-" "k6/0.55.0 (https://k6.io/)" "-" "127.0.0.1:9001"
127.0.0.1 - - [24/Feb/2025:22:06:30 +0300] "POST /api/v1/get-user-token?admin_secret=local_admin&login=alivasilyev HTTP/1.1" 200 193 "-" "k6/0.55.0 (https://k6.io/)" "-" "127.0.0.1:9001"
127.0.0.1 - - [24/Feb/2025:22:06:30 +0300] "POST /api/v1/get-user-token?admin_secret=local_admin&login=alivasilyev HTTP/1.1" 200 193 "-" "k6/0.55.0 (https://k6.io/)" "-" "127.0.0.1:9000"
127.0.0.1 - - [24/Feb/2025:22:06:30 +0300] "POST /api/v1/get-user-token?admin_secret=local_admin&login=alivasilyev HTTP/1.1" 200 193 "-" "k6/0.55.0 (https://k6.io/)" "-" "127.0.0.1:9002"
127.0.0.1 - - [24/Feb/2025:22:06:30 +0300] "POST /api/v1/get-user-token?admin_secret=local_admin&login=alivasilyev HTTP/1.1" 200 193 "-" "k6/0.55.0 (https://k6.io/)" "-" "127.0.0.1:9000"
127.0.0.1 - - [24/Feb/2025:22:06:30 +0300] "POST /api/v1/get-user-token?admin_secret=local_admin&login=alivasilyev HTTP/1.1" 200 193 "-" "k6/0.55.0 (https://k6.io/)" "-" "127.0.0.1:9002"
127.0.0.1 - - [24/Feb/2025:22:06:30 +0300] "POST /api/v1/get-user-token?admin_secret=local_admin&login=alivasilyev HTTP/1.1" 200 193 "-" "k6/0.55.0 (https://k6.io/)" "-" "127.0.0.1:9000"
127.0.0.1 - - [24/Feb/2025:22:06:30 +0300] "POST /api/v1/get-user-token?admin_secret=local_admin&login=alivasilyev HTTP/1.1" 200 193 "-" "k6/0.55.0 (https://k6.io/)" "-" "127.0.0.1:9001"
```

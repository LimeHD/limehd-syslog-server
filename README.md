## syslog сервер который собирает статистику с nginx серверов и складывае ее в influxdb

### Запуск

`$ go run . --help`

`$ go run . --debug --bind-address 0.0.0.0:514 --maxmind ./GeoLite2-City.mmdb --influx-url http://0.0.0.0:8086 --influx-db polina`

#### Подробности для разработки

- Собрать influx: `$ docker run -p 8086:8086 -d --name influx_docker --rm -v $PWD:/var/lib/influxdb influxdb`
- Посмотреть результаты: 
    - `$ docker run --rm --link=influx_docker -it influxdb influx -host influx_docker`
    - `> create database polina`
    - `> show databases`
    - `> use polina`
    - `> select * from syslog`
    
    <details>
      <summary>Output</summary>
      
      ```shell script
        > select * from syslog
        name: syslog
        time                bytes_sent channel   connections country_id quality streaming_server
        ----                ---------- -------   ----------- ---------- ------- ----------------
        1596799651430918846 404        domashniy 154                   vh1w    syslog-server.local
        1596799652461060025 404        domashniy 154                   vh1w    syslog-server.local
        1596799653820469346 404        domashniy 154                   vh1w    syslog-server.local
        1596799654459295760 404        domashniy 154                   vh1w    syslog-server.local
        1596799655117672443 404        domashniy 154                   vh1w    syslog-server.local
        ...
      ```
    </details>

### Боевой сервер

* influx-host: `influx.iptv2022.com`
* database: `polina`

### Метрики

* [ ] Трафик (`value` в measurement `bytes_sent`)
* [ ] online-пользователи (`value` в measurement `connections`)

### Тэги

* `country_id` - ID страны из [maxmind](https://dev.maxmind.com/geoip/legacy/codes/iso3166/)
* `channel`
* `streaming_server`
* `quality`

### Схема

```
+------------------------+                   +-------------------------+
|                        |                   |                         |
|  MEDIA STREAMAING      |   SYSLOG (UDP)    |                         |
|  NGINX SERVERS         |                   |   LIMEHD SYSLOG SERVER  |
|                        |   +----------->   |                         |
+------------------------+                   +-------------------------+
                                                          +
                                                          |
                                                          |
                                                          v
                                              +-----------------------+
                                              |                       |
                                              |        INFLUXDB       |
                                              |                       |
                                              +-----------------------+
```


### Формат

```
29/Apr/2020:21:21:14 +0300|1588184474.870|83.219.236.137|HTTP/1.1|GET|mhd.limehd.tv|/streaming/domashniy/324/vh1w/playlist.m3u8|-|206|4004|0.000|-|-|-|-|-|-|Mozilla/5.0 (Web0S; Linux/SmartTV) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/53.0.2785.34 Safari/537.36 WebAppManager|-|1|2313149490
```

Где:

* `vh1w` - `quality`
* `domashniy` - `channel`
* `83.219.236.137` - IP для определения `country_id`
* `mhd.limehd.tv` - `streaming_server`

```
oleg_maksimov  8:25 AM
08:24:43.763016 IP 172.19.95.111.22221 > 194.35.48.67.514: SYSLOG local7.info, length: 527
08:24:43.766297 IP 172.19.95.111.61437 > 194.35.48.67.514: SYSLOG local7.info, length: 356
08:24:43.766326 IP 172.19.95.111.61437 > 194.35.48.67.514: SYSLOG local7.info, length: 295
08:24:43.767118 IP 172.19.95.111.26344 > 194.35.48.67.514: SYSLOG local7.info, length: 470
08:24:43.767415 IP 172.19.95.111.8950 > 194.35.48.67.514: SYSLOG local7.info, length: 452
```


```
log_format csv
                $time_local|
                $msec|
                $remote_addr|
                $server_protocol|
                $request_method|
                $host|
                $uri|
                $args|
                $status|
                $body_bytes_sent|
                $request_time|
                $upstream_response_time|
                $upstream_addr|
                $upstream_status|
                $http_referer|
                $http_via|
                $http_x_forwarded_for|
                $http_user_agent|
                $sent_http_x_profile|
                $connection_requests|
                $connection|
                $bytes_sent;
access_log syslog:server=127.0.0.1:PORT csv;
```

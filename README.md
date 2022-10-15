# ProberMesh
#### 分布式 C/S 网络网格探测框架

> ICMP 视野落在 Region to Region ；
>
> HTTP 视野落在 Region to URL ；
>
> 同 Region 下做 agg；

- ##### 支持 ICMP/HTTP 分阶段耗时

- ##### ICMP 支持 resolve/rtt 分阶段耗时

- ##### ICMP 支持丢包率， 抖动标准差

- ##### HTTP 支持 resolve/connect/tls/processing/transfer 分阶段耗时 (httpstat)

#### 

```golang
// ping 指标
prober_icmp_failed
prober_icmp_duration_seconds
prober_icmp_packet_loss_rate
prober_icmp_jitter_stddev_seconds
prober_icmp_duration_seconds_total (histogram)

// http 指标
prober_http_failed
prober_http_duration_seconds

// 健康检查
prober_agent_is_alive
```


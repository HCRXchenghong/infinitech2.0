# Observability

生产目标：

- Prometheus 采集 API、BFF、Realtime、Worker、PostgreSQL、Redis、Kafka 指标。
- Grafana 展示订单、支付、钱包、实时、RTC、基础设施看板。
- Loki 收集结构化日志。
- Tempo 收集 OpenTelemetry trace。
- 告警必须覆盖支付回调失败、钱包扣减冲突、抢单冲突异常、Kafka 堆积、Realtime 断连风暴。


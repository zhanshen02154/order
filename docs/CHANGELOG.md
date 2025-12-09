
<a name="v3.0.0"></a>
## [v3.0.0](https://github.com/zhanshen02154/order/compare/v2.0.0...v3.0.0) (2025-12-09)

### Bug Fixes

* 更改应用层事件侦听器为新名称
* 修复事件包装器元数据的时间戳转换
* 降低pprof采样频率
* **ETCD分布式锁:** 释放锁采用单独的超时上下文
* **ETCD分布式锁:** ETCD分布式锁由共享session改为每个锁独立维护session。
* **事件侦听器:** 消息的Key强制字符串类型

### Code Refactoring

* 接口层事件处理器结构体改为私有
* 事件侦听器结构体改为私有，仅开放接口
* Etcd分布式锁结构体改为私有
* 支付回调接口收到信息改为订单处理中
* 调整目录结构
* **获取DB实例:** 没有事务实例则用WithContext

### Features

* 新增确认支付方法
* 新增响应代码判断是否移入死信队列
* 新增商品事件处理器
* 新增更新支付状态
* 事件侦听器支持传入消息的Key
* 应用死信队列包装器
* 新增死信队列
* 新增事件总线
* **broker:** 新增基于kafka的broker
* **接口层:** 新增订单事件处理器
* **配置结构体:** 新增Broker配置结构体

### Performance Improvements

* 优化依赖包
* **pprof:** 调整pprof采样频率

### BREAKING CHANGE


基础设施层移除的目录：
- broker
- config
- server
- registory

- 新增事件总线

- 新增订单事件处理器

- 新增基于kafka的broker
- 新增连接broker的事件侦听器
- 新增service客户端发布元数据处理包装器


<a name="v2.0.0"></a>
## [v2.0.0](https://github.com/zhanshen02154/order/compare/v1.0.1...v2.0.0) (2025-11-28)

### Bug Fixes

* 调整打印日志
* 恢复TTL到默认值
* **ETCD分布式锁:** 恢复 DialKeepAliveTime参数
* **ETCD分布式锁:** 加锁和解锁使用单独的context
* **ETCD分布式锁:** 先关闭会话再取消上下文
* **ETCD分布式锁:** 使用cancel上下文处理会话
* **ETCD分布式锁:** 初始化分布式锁时创建Mutex
* **ETCD分布式锁:** 删除AutoSyncInterval
* **ETCD分布式锁:** 创建会话失败无需关闭
* **打印错误日志:** 不需要格式化的使用log.Error
* **打印错误日志:** 改用log.Errorf和log.Error

### Code Refactoring

* 移除不使用的依赖
* **Consul配置源:** 获取Consul配置源逻辑调整到基础设施层
* **ETCD分布式锁:** 移动测试文件
* **ETCD分布式锁:** 删除ETCDLock内的客户端
* **pprof服务器:** 移动到基础设施层的server包
* **主函数和加锁逻辑:** 优化主函数和加锁逻辑
* **健康检查服务器:** 移动到基础设施层的server包
* **基础设施层:** 升级Go micro框架到4.11.0版本
* **服务上下文:** main函数初始化组件移入服务上下文
* **订单支付回调:** 调整订单支付回调功能

### Features

* **DTM分布式事务:** 新增分布式事务
* **ETCD分布式锁:** 新增tryLock
* **ETCD分布式锁:** 新增分布式锁的接口和ETCD实现类

### Performance Improvements

* **ETCD分布式锁:** 设置心跳和心跳超时
* **ETCD分布式锁:** 共享Session减少资源开销
* **商品服务proto文件:** 优化proto文件以支持go micro v4
* **商品服务客户端:** 移除商品服务客户端
* **订单服务Proto:** 修改订单服务的Proto文件

### BREAKING CHANGE


- 集成DTM分布式事务
- 新增事务管理器支持子事务屏障处理的方法
- 重写基础设施层里涉及go micro 2.9.1的方法


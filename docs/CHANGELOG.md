
<a name="v1.0.1"></a>
## v1.0.1 (2025-11-23)

### Bug Fixes

* **打印错误日志:** 改用log.Errorf和log.Error


<a name="v1.0.0"></a>
## v1.0.0 (2025-11-17)

### Bug Fixes

* 设置订单服务客户端超时时间
* 已支付成功的订单不再操作
* 已支付成功的订单不再操作
* **client:** 修复商品服务客户端Consul.watch异常的问题
* **config:** 修复配置错误问题
* **main函数:** 优化服务beforeStop逻辑

### Code Refactoring

* **all:** 修改go mod名称
* **infrastructure:** 调整初始化数据库及健康检查探针

### Features

* **config:** 增加健康检查探针地址
* **config:** 增加健康检查探针地址
* **consul register:** 修复WaitTime超时设置失效问题

### Performance Improvements

* **ProductClient:** 更改商品服务客户端初始化逻辑
* **proto:** 细化返回结果


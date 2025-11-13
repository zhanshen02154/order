
<a name="v1.0.0"></a>
## v1.0.0 (2025-11-17)

### Bug Fixes

* 设置订单服务客户端超时时间
* 已支付成功的订单不再操作
* **client:** 修复商品服务客户端Consul.watch异常的问题
* **config:** 修复配置错误问题
* **main函数:** 优化服务beforeStop逻辑

### Code Refactoring

* **all:** 修改go mod名称
* **infrastructure:** 调整初始化数据库及健康检查探针
* 调整订单服务的目录结构

### Features

* **config:** 增加健康检查探针地址
* **订单支付回调:** 新增订单支付回调API接口
* **consul register:** 修复WaitTime超时设置失效问题

### Performance Improvements

* **ProductClient:** 更改商品服务客户端初始化逻辑
* **proto:** 细化返回结果


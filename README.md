## 前言：
conf后面是不会更新了的
public/const也是不会更新的
```bash
$ git reset HEAD public/const.go
```

##  更新日志

<details>
<summary>20220424</summary>
<h3>容器化</h3>

- 布局Docker-composer
- 个人中心的正在进行
</details>

<details>
<summary>20220419</summary>
<h3>中间件</h3>

- 完成Token中间件判断
- 数量限制创建20个正在创建
- uid加入上下文
- 架构图书写
</details>

<details>
<summary>20220418</summary>
<h3>完成发送者角色</h3>

- 相关业务逻辑
- 使用了zstd压缩
- 完成数据库主从表的设计与实现
- 测试通过
</details>

<details>
<summary>20220417</summary>
<h3>创建表单逻辑完成</h3>

- 创建订单使用了SHA256
- 创建订单加入了有序集合并自动删除前3天的
- 送入RabbitMQ进行削峰缓解多并发下的数据库的写入
- Rabbit的消费者封装完毕
- 测试通过
</details>

<details>
<summary>20220416</summary>
<h3>首页</h3>

- 首页加入信息加入到了缓存
- 通行证变短
</details>

<details>
<summary>20220415</summary>
<h3>搞懂Token通行证</h3>

- 全新用户进行注册
- 12天以后的话会更新信息
- 其余会颁发新的Token
- 加入了缓存删除的随机值防止缓存雪崩
- openId设置了唯一索引加快查找
- 部署数据库/文档/运行
</details>
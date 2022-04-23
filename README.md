##  后端介绍

> 后端项目开发负责人：[HengY1Sky](https://github.com/HengY1Sky/)

借**软件工程大作业的机会**，想将自己所学所会运用起来，便在“**没有拿任何奖项，只想做好的**“驱动力下进行《暨数据》后端研发。

后端从设计到研发始终秉承着尽可能减小耗时，实现**高并发低延时**的想法，以Golang作为主要的后端开发语言，使用了Gin作为Web开发框架，加入了Redis作为缓存层，消息队列；Mysql设计数据表以及实现增删改查；MongoDb存储以及分析Json；并使用了RabbitMQ来实现异步削峰；同时使用了ZSTD无损压缩算法以及基础的对称与非对称加密等来实现后端系统。作为一个单机小团队项目，使用了Docker编排部署环境，尽可能满足需求。**更多亮点请看下面描述**

当然相比于企业等有组织有分工的体系，从容灾；日志；预警；备份等角度，本项目仍然存在很多不足。但是随着迭代会不断进行改进，我觉得我还是能继续优化的。

⚠️：本项目所有代码均是第一手源码（没有从以前的项目迁移过来）

##  快速部署

😊 前提准备：建议是Ubuntu20.4

1. 自己设置好Go环境/Docker环境：谷歌搜一下一大堆
2. 微信小程序个人中心拿到自己的appId/appSecret
3. 因为涉及邮箱通知：可以到QQ邮箱等设置里面拿到第三方SMTP密钥
4. 因为涉及七牛云第三方存储：到七牛云管理平台里面拿到accessKey/secretKey并创建对象桶
5. 安装`SCREEN`后台运行

😊 配置`Nginx反向代理`

> 因为涉及Websocket，在反向代理的时候进行了协议升级

```nginx
#PROXY-START/

location ^~ /
{
    proxy_pass http://127.0.0.1:8880;
    proxy_set_header Host $host;
    proxy_set_header X-Real-IP $remote_addr;
    proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
   	proxy_set_header Upgrade $http_upgrade;
    proxy_set_header Connection "upgrade";
}

#PROXY-END/
```

😊 配置`common/secret.go`

1. 根据提示将第三方的密钥设置上去
2. 自己设置用于整个项目的AES密钥以及偏移量

😊 配置`conf`文件夹

conf里面的连接命令是根据docker编排的账号密码进行设置的

*如果你不使用Docker编排，那么你需要自己安装`Mysql`;`Redis`;`MongoDB`;`RabbitMQ`;

然后在`conf`文件中与`common/secret.go`修改连接命令即可。

> **具体版本可以参考`docker-compose`中的版本来进行安装**

😊 编排`docker-compose`

1. 解压`durable.zip`为`durable`文件夹：为持久化外部挂载文件  `unzip -o -d ./ durable.zip`
2. 使用`docer-compose`命令直接启动环境：`docker-compose up -d`

😊 准备开始运行

1. 编译2个启动文件`go build main.go`；`go build rabbitConsume.go`
2. 打开2个后台`screen -S GinBackEnd ./main --config ./conf/prod/`；`screen -S rabbitMq ./rabbitConsume`

##  技术栈

- 主语言：Goalng
- 开发框架：Gin
- 数据库：Mysql、MongoDB
- 中间件：Redis、RabbitMQ
- 对象存储：七牛云对象存储
- 其他：压缩算法[ZSTD](https://github.com/klauspost/compress/tree/master/zstd)；WebSocket；布隆过滤器

##  亮点

- MongoDb操作的封装：实现了JSON直接到BSON然后直接存储到MongoDb中
- 使用NewTicker心跳进行对WebSocket维护；控制第三方邮件请求频率等
- 使用布隆过滤器：对一小时内的相同用户请求的过滤，防止邮件爆破
- 中间件实现了：流量统计；QPS限制；Token校验；用户信息加入上下文等功能
- 使用了ZSTD无损压缩算法：尽可能减小数据库存储压力
- 使用了RabbitMQ对高并发流量进行削峰：尽可能减小大量写操作造成死锁等问题
- 使用Gorm进行了数据库映射：实现了基本业务的增删改查
- 使用了WebSocket与Nginx配合进行服务升级：在填写表单页面能实时获取别人填写表格信息

##  展望

- Admin管理端实现：等我把手上全部忙完了来写（还有项目🐶）
- 容灾｜降级｜预警：等我以后发现了些好的第三方再来用上
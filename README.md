# go-mysql-syncer
- 一个基于go-mysql & cannel 二次开发的binlog解析工具，最终生成SQL语句使用kafka发送到消费者进行执行
- 目前 DDL DLC不能完美支持
- 多个mysql实例同时处理的功能
- 解决跨机房mysql自带同步缓慢的问题
- * 修改了go-mysql源码所以必须使用vendor下的包
- consumer 多实力分库消费，数据保证不丢失


# TODO
-mysql同步根据 表名拆分 topic
-支持 DDL DLC
-双A功能


## build
- make
- ./Producer
- ./Consumer

## config
### producer
- 在conf目录下修改 producer.toml 配置mysql地址，banyandb集群agent地址和库表，mysql主库可以添加多个同时处理，只要保证id 、data_path和ns table选项不重复即可
- 程序运行后会在conf目录下生成data_path配置的master.info文件，保存binlog文件名 pos号信息 
- * 如果同步遇到drop table、drop database命令或者程序异常中断之后有表结构更改（alter table）命令那么需要删除master.info文件并手动进行全量同步，否则可能出现解析表字段名异常 

### consumer
- 参考producer对应的配置修改 consumer.toml

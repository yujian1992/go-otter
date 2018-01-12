package gobinlog

import (
	"os"
	"path/filepath"
	"sync"
    _"time"
    _"fmt"
	"github.com/ngaut/log"
    "github.com/siddontang/go-mysql/client"
    cluster "github.com/bsm/sarama-cluster"
)

//Consumer 定义Consumer的结构
type Consumer struct {
	c *Config

	exitChan chan struct{}

	posLock sync.Mutex

	waitGroup WaitGroupWrapper
    poolSize  int
    
    conn *client.Conn

	IsRestart bool
	cluster   *cluster.Consumer
	sqlcfg    MysqlConfig
	schemaMap map[string]string
}

//NewConsumer 初始化
func NewConsumer(c *Config, mysqlcfg MysqlConfig) (*Consumer, error) {
	//日志目录确保存在
	dir := filepath.Dir(c.LogConfig.Path)
	exist, _ := PathExists(dir)

	if !exist {
		err := os.Mkdir(dir, os.ModePerm)

		if err != nil {
			return nil, err
		}
	}

	//配置日志
	log.SetHighlighting(c.LogConfig.Highlighting)
	log.SetLevel(log.StringToLogLevel(c.LogConfig.Level))
	log.SetFlags(log.Lshortfile | log.Ldate | log.Ltime)
	log.SetOutputByName(c.LogConfig.Path)

	if c.LogConfig.Type == LogTypeDay {
		log.SetRotateByDay()
	} else if c.LogConfig.Type == LogTypeHour {
		log.SetRotateByHour()
	}

	consumer := &Consumer{
		c:         c,
		exitChan:  make(chan struct{}),
		IsRestart: false,
		schemaMap: make(map[string]string, 0),
	}
	//var err error = nil
	clusters := make([]string, 0)
	for _, v := range c.ClusterConfig.Agents {
		clusters = append(clusters, v)
    }

    topics := make([]string, 0)
    for _, v := range c.ClusterConfig.Topics {
		topics = append(topics, v)
    }

    //消费者连kafka
    config := cluster.NewConfig()
	config.Consumer.Return.Errors = true
    config.Group.Return.Notifications = true
    //config.Group.Offsets.Retry.Max =0
    //fmt.Println(clusters,c.ClusterConfig.Group,topics)
    var err error = nil
    consumer.cluster,err = cluster.NewConsumer(clusters, string(c.ClusterConfig.Group), topics, config)
	if err != nil {
		log.Errorf("GetBanyanClient failed: %v", err)
		return nil, err
    }
    go func() {
		for ntf := range consumer.cluster.Notifications() {
			log.Infof("Rebalanced: %+v\n", ntf)
		}
	}()

    consumer.sqlcfg = mysqlcfg

    consumer.conn, err = client.Connect(consumer.sqlcfg.Addr, consumer.sqlcfg.User, consumer.sqlcfg.Password, "")
	if err != nil {
		log.Errorf("connect mysql server failed: %v conn: %v", err, consumer.sqlcfg)
		return nil,err
    }

	consumer.waitGroup.Wrap(func() { consumer.scanSchemaLoop() })

	return consumer, nil

}

//Close 关闭Consumer,释放资源
func (consumer *Consumer) Close() {

	close(consumer.exitChan)

	consumer.waitGroup.Wait()
    consumer.cluster.Close()
    err := consumer.conn.Close()
    if(err != nil) {
        log.Errorf("Mysql close err : %v",err)
    }
    log.Infof("Consumer safe close and exit id = %s", consumer.sqlcfg.Id)
}

func (consumer *Consumer) scanSchemaLoop() {
	for {
        select {
        case msg,ok:= <-consumer.cluster.Messages() :
			if ok {
                consumer.cluster.MarkOffset(msg, "")	// mark message as processed
                consumer.processingSQL(string(msg.Value))
                consumer.cluster.CommitOffsets()
            }
        case <-consumer.exitChan:
            log.Infof("save binlog position loop exit.")
            return
        }
    } 
}

func (consumer *Consumer) processingSQL(sql string) {

    log.Error("sql is",sql)
    result, err := consumer.conn.Execute(sql)
    if(err != nil) {
        log.Errorf("Mysql Execute err : %v",err)
    }
    log.Infof("Mysql result is : %v",result)
}

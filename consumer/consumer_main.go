package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	//_ "github.com/go-sql-driver/mysql"
    "gobinlog"
	"github.com/juju/errors"
	"github.com/ngaut/log"
)

var (
	configFile = flag.String("config", "./conf/consumer.toml", "go-consumer config file")
	binlogName = flag.String("binlog_name", "", "binlog file name")
	binlogPos  = flag.Int64("binlog_pos", 0, "binlog position")
)

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	flag.Parse()

	sc := make(chan os.Signal, 1)
	signal.Notify(sc,
		os.Kill,
		os.Interrupt,
		//syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT)

	var err error
	var config *gobinlog.Config
	config, err = gobinlog.NewConfigWithFile(*configFile)
	if err != nil {
		log.Fatalf("config load failed.detail=%s", errors.ErrorStack(err))
	}
	for _, v := range *config.MysqlConfig {
		r, err := gobinlog.NewConsumer(config, v)
		defer r.Close()

		if err != nil {
			fmt.Println("new Rail error.", err)
			log.Fatalf("new Rail error. detail:%v ,%s", err, v.Id)
		}

		fmt.Println("rail start succ. %s", v.Id)
	}

	signal := <-sc

    log.Errorf("program terminated! signal:%v", signal)
    return
}

package config

import (
	"io/ioutil"
	"encoding/json"
	"fmt"
)

type TopicInfo struct {
	Partitions int
}

type ProducerACL struct {
	Secret string
	Topic string
	Name string
}

type CircuitBreakerInfo struct {
	 BreakPeriod int // 熔断封锁时间
	 RecoverPeriod int // 熔断恢复时间
	 WinSize int // 滑动窗口大小
	 MinStats int // 最小统计样本
	 HealthRate float64 // 健康阀值
}

type ConsumerInfo struct {
	Topic string
	GroupId string
	RateLimit int
	Retries int
	Timeout int
	Concurrency int
	CircuiteBreakerInfo *CircuitBreakerInfo
}

type Config struct {
	// log配置
	Log_directory string
	Log_level int

	// Kafka地址
	Kafka_bootstrap_servers string
	Kafka_topics map[string]TopicInfo // topic信息

	// Producer配置
	Kafka_producer_retries int
	Kafka_producer_acl map[string]ProducerACL // acl访问权限

	// Consumer配置
	Kafka_consumer_list []ConsumerInfo

	// HTTP服务配置
	Http_server_port int
	Http_server_read_timeout int
	Http_server_write_timeout int
	Http_server_handler_channel_size int
}

// 解析并返回Config对象
func ParseConfig(path string) *Config {
	content, err := ioutil.ReadFile(path)
	if err != nil {
		return nil
	}

	dict := map[string]interface{} {}

	err = json.Unmarshal(content, &dict)
	if err != nil {
		return nil
	}

	config := Config{}

	config.Log_directory = dict["log.directory"].(string)
	config.Log_level = int(dict["log.level"].(float64))
	config.Kafka_bootstrap_servers = dict["kafka.bootstrap.servers"].(string)
	config.Kafka_producer_retries = int(dict["kafka.producer.retries"].(float64))

	config.Http_server_port = int(dict["http.server.port"].(float64))
	config.Http_server_read_timeout = int(dict["http.server.read.timeout"].(float64))
	config.Http_server_write_timeout = int(dict["http.server.write.timeout"].(float64))
	config.Http_server_handler_channel_size = int(dict["http.server.handler.channel.size"].(float64))

	config.Kafka_topics = map[string]TopicInfo{}

	topicsArr := dict["kafka.topics"].([]interface{})
	for _, value := range topicsArr {
		topicMap := value.(map[string]interface{})
		name := topicMap["name"].(string)
		partitions := int(topicMap["partitions"].(float64))
		config.Kafka_topics[name] = TopicInfo{Partitions: partitions}
	}

	config.Kafka_producer_acl = map[string]ProducerACL{}

	aclArr := dict["kafka.producer.acl"].([]interface{})
	for _, value := range aclArr {
		aclMap := value.(map[string]interface{})
		name := aclMap["name"].(string)
		secret := aclMap["secret"].(string)
		topic := aclMap["topic"].(string)
		config.Kafka_producer_acl[name] = ProducerACL{Name: name, Secret: secret, Topic: topic}
		// 检查acl涉及的topic是否配置
		if _, exists := config.Kafka_topics[topic]; !exists {
			fmt.Println("ACL中配置的topic: " + topic + " 不存在,请检查kafka.topics.")
			return nil
		}
	}

	consumerArr := dict["kafka.consumer.list"].([]interface{})
	for _, value := range consumerArr {
		item := value.(map[string]interface{})
		consumerInfo := ConsumerInfo{}
		consumerInfo.Topic = item["topic"].(string)
		consumerInfo.GroupId = item["groupId"].(string)
		consumerInfo.RateLimit = int(item["rateLimit"].(float64))
		consumerInfo.Retries = int(item["retries"].(float64))
		consumerInfo.Timeout = int(item["timeout"].(float64))
		consumerInfo.Concurrency = int(item["concurrency"].(float64))
		// 加载熔断器配置（可选）
		if circuitValue, exist := item["circuitBreaker"]; exist {
			circuitMap := circuitValue.(map[string]interface{})
			consumerInfo.CircuiteBreakerInfo = &CircuitBreakerInfo{}
			consumerInfo.CircuiteBreakerInfo.RecoverPeriod = int(circuitMap["recoverPeriod"].(float64))
			consumerInfo.CircuiteBreakerInfo.BreakPeriod = int(circuitMap["breakPeriod"].(float64))
			consumerInfo.CircuiteBreakerInfo.WinSize = int(circuitMap["winSize"].(float64))
			consumerInfo.CircuiteBreakerInfo.HealthRate = circuitMap["healthRate"].(float64)
			consumerInfo.CircuiteBreakerInfo.MinStats = int(circuitMap["minStats"].(float64))
			if consumerInfo.CircuiteBreakerInfo.WinSize <= 0 || consumerInfo.CircuiteBreakerInfo.MinStats <=0 || consumerInfo.CircuiteBreakerInfo.HealthRate <= 0 || consumerInfo.CircuiteBreakerInfo.HealthRate > 100 {
				fmt.Println("consumer配置的熔断器参数有误, 请检查一下.")
				return nil
			}
		}

		config.Kafka_consumer_list = append(config.Kafka_consumer_list, consumerInfo)
		// 检查acl涉及的topic是否配置
		if _, exists := config.Kafka_topics[consumerInfo.Topic]; !exists {
			fmt.Println("consumer中配置的topic: " + consumerInfo.Topic + " 不存在,请检查kafka.topics.")
			return nil
		}
	}
	return &config
}
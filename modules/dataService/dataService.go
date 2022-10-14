package dataService

import (
	cm "ThingsPanel-Go/modules/dataService/mqtt"
	"ThingsPanel-Go/modules/dataService/tcp"
	tphttp "ThingsPanel-Go/others/http"
	"ThingsPanel-Go/services"
	uuid "ThingsPanel-Go/utils"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/panjf2000/ants/v2"
	"github.com/spf13/viper"
)

func init() {
	loadConfig()
	MqttHttpHost := os.Getenv("MQTT_HTTP_HOST")
	if MqttHttpHost == "" {
		MqttHttpHost = viper.GetString("api.http_host")
	}
	resps, errs := tphttp.Post("http://"+MqttHttpHost+"/v1/accounts/root", "{\"password\":\""+viper.GetString("mqtt.pass")+"\"}")
	if errs != nil {
		log.Println("Response1:", errs.Error())
	} else {
		defer resps.Body.Close()
		if resps.StatusCode == 200 {
			body, errs := ioutil.ReadAll(resps.Body)
			if errs != nil {
				log.Println("Response2:", errs.Error())
			} else {
				log.Println("Response3: ", string(body))
			}
		} else {
			log.Println("Get failed with error:" + resps.Status)
		}
	}
	listenMQTT()
	listenTCP()
}

func loadConfig() {
	log.Println("read config")
	var err error
	envConfigFile := flag.String("config", "./modules/dataService/config.yml", "path of configuration file")
	flag.Parse()
	viper.SetConfigFile(*envConfigFile)
	if err = viper.ReadInConfig(); err != nil {
		fmt.Println("FAILURE", err)
		return
	}
	return
}

func listenMQTT() {
	var TSKVS services.TSKVService
	mqttHost := os.Getenv("TP_MQTT_HOST")
	if mqttHost == "" {
		mqttHost = viper.GetString("mqtt.broker")
	}
	broker := mqttHost
	uuid := uuid.GetUuid()
	clientid := viper.GetString(uuid)
	user := viper.GetString("mqtt.user")
	pass := viper.GetString("mqtt.pass")
	p, _ := ants.NewPool(500)
	p1, _ := ants.NewPool(500)
	cm.Listen(broker, user, pass, clientid, func(m mqtt.Message) {
		_ = p.Submit(func() {
			TSKVS.MsgProc(m.Payload())
		})
	}, func(m mqtt.Message) {
		_ = p1.Submit(func() {
			TSKVS.MsgProc(m.Payload())
		})
	})
}

func listenTCP() {
	tcpPort := viper.GetString("tcp.port")
	log.Printf("config of tcp port -- %s", tcpPort)
	tcp.Listen(tcpPort)
}

package cache

import (
	"fmt"
	"strconv"

	util "applytics.in/yin/src/helpers"
	redis "github.com/go-redis/redis/v7"
)

const (
	online = "-online"
)

//Config :- config for redis
type Config struct {
	Host string
	Port string
	// Password string
	client *redis.Client
}

//Init :- init cache
func (c *Config) Init() {
	c.client = redis.NewClient(&redis.Options{
		Addr: c.Host + ":" + c.Port,
		// Password:  c.Password,
		DB:        0,
		OnConnect: onConnect,
	})
}

func onConnect(conn *redis.Conn) error {
	fmt.Println("redis connected")
	return nil
}

// //AddIP :- add ip and sid to cache
// func (c *Config) AddIP(IP string, sID string) {
// 	err := c.client.Set(IP, sID, 0).Err()
// 	if err != nil {
// 		util.LogError("cannot save value for SID: "+sID+" of IP"+IP, err)
// 	}
// }

// //RemoveIP :- remove ip
// func (c *Config) RemoveIP(IP string) {
// 	err := c.client.Del(IP).Err()
// 	if err != nil {
// 		util.LogError("cannot delete for key: "+IP, err)
// 	}
// }

// //GetSid :- return sid
// func (c *Config) GetSid(IP string) *string {
// 	val, err := c.client.Get(IP).Result()
// 	if err != nil {
// 		util.LogError("cannot sid for ip: "+IP, err)
// 		return nil
// 	}
// 	return &val
// }

//UpdateOnlineCount :- cache appid in redis
func (c *Config) UpdateOnlineCount(appID string) {
	appIDKey := appID + online
	val, err := c.client.Get(appIDKey).Result()
	if val != "" {
		prevValue, err := strconv.Atoi(val)

		if err != nil {
			util.LogError("cannot convert app count to int for appID: "+appIDKey+" with value "+val, err)
			prevValue = 0
		}
		prevValue++

		err = c.client.Set(appIDKey, prevValue, 0).Err()

		if err != nil {
			util.LogError("cannot save value for appID: "+appIDKey+" with value "+val, err)
		}
		return
	}

	err = c.client.Set(appIDKey, 1, 0).Err()

	if err != nil {
		util.LogError("cannot save value for appID: "+appIDKey+" with value 1", err)
	}
}

//ReduceOnlineCount :- remove Aid from cache
func (c *Config) ReduceOnlineCount(appID string) {
	appIDKey := appID + online
	val, err := c.client.Get(appIDKey).Result()
	if val != "" {
		prevValue, err := strconv.Atoi(val)

		if err != nil {
			util.LogError("cannot convert app count to int for appID: "+appIDKey+" with value "+val, err)
			prevValue = 0
		}
		prevValue++

		err = c.client.Set(appIDKey, prevValue, 0).Err()

		if err != nil {
			util.LogError("cannot save value for appID: "+appIDKey+" with value "+val, err)
		}
		return
	}

	err = c.client.Set(appIDKey, 0, 0).Err()

	if err != nil {
		util.LogError("cannot save value for appID: "+appIDKey+" with value 1", err)
	}

}

//GetAppCount :- get app count
func (c *Config) GetAppCount(appID string) *int {
	val, err := c.client.Get(appID).Result()
	if err != nil {
		util.LogError("cannot count for appID: "+appID, err)
		return nil
	}
	count, err := strconv.Atoi(val)
	if err != nil {
		util.LogError("cannot convert count for appID: "+appID+"for value"+val, err)
		return nil
	}
	return &count
}

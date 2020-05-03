package cache

import (
	"fmt"
	"strconv"

	util "applytics.in/yin/src/helpers"
	redis "github.com/go-redis/redis/v7"
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

//AddIP :- add ip and sid to cache
func (c *Config) AddIP(IP string, sID string) {
	err := c.client.Set(IP, sID, 0).Err()
	if err != nil {
		util.LogError("cannot save value for SID: "+sID+" of IP"+IP, err)
	}
}

//RemoveIP :- remove ip
func (c *Config) RemoveIP(IP string) {
	err := c.client.Del(IP).Err()
	if err != nil {
		util.LogError("cannot delete for key: "+IP, err)
	}
}

//GetSid :- return sid
func (c *Config) GetSid(IP string) *string {
	val, err := c.client.Get(IP).Result()
	if err != nil {
		util.LogError("cannot sid for ip: "+IP, err)
		return nil
	}
	return &val
}

//AddAppID :- cache appid in redis
func (c *Config) AddAppID(appID string) {

	val, err := c.client.Get(appID).Result()
	if val != "" {
		prevValue, err := strconv.Atoi(val)

		if err != nil {
			util.LogError("cannot convert app count to int for appID: "+appID+" with value "+val, err)
			prevValue = 0
		}
		prevValue++

		err = c.client.Set(appID, prevValue, 0).Err()

		if err != nil {
			util.LogError("cannot save value for appID: "+appID+" with value "+val, err)
		}
		return
	}

	err = c.client.Set(appID, 1, 0).Err()

	if err != nil {
		util.LogError("cannot save value for appID: "+appID+" with value 1", err)
	}
}

//RemoveAppID :- remove Aid from cache
func (c *Config) RemoveAppID(appID string) {
	err := c.client.Del(appID).Err()
	if err != nil {
		util.LogError("cannot delete for key: "+appID, err)
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

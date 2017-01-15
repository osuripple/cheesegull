package redis

import (
	"fmt"
	"strconv"

	redis "gopkg.in/redis.v5"
)

func (i *impl) SendBeatmapRequest(b int) error {
	return i.Publish("beatmap_requests", fmt.Sprintf("%d", b)).Err()
}
func (i *impl) BeatmapRequestsChannel() (<-chan int, error) {
	ps, err := i.Subscribe("beatmap_requests")
	if err != nil {
		return nil, err
	}
	c := make(chan int, 10)
	go messageSubscriberInt(ps, c)
	return c, nil
}

func messageSubscriberInt(ps *redis.PubSub, c chan<- int) {
	for {
		msg, err := ps.ReceiveMessage()
		if err != nil {
			fmt.Println(err)
			continue
		}
		m, err := strconv.Atoi(msg.Payload)
		if err != nil {
			continue
		}
		c <- m
	}
}

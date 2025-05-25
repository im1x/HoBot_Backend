package actions

import (
	"HoBot_Backend/pkg/model"
	"HoBot_Backend/pkg/mongo"
	"HoBot_Backend/pkg/service/chat"
	"HoBot_Backend/pkg/service/vkplay"
	"context"
	"github.com/gofiber/fiber/v2/log"
	"go.mongodb.org/mongo-driver/bson"
	"strings"
	"time"
)

func FixWsIdForAllUsers() {
	time.Sleep(40 * time.Second)
	log.Info("fixWsIdForAllUsers: start")
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	// get all users
	var users []model.User
	cur, err := mongo.GetCollection(mongo.Users).Find(ctx, bson.M{})
	if err != nil {
		log.Error("fixWsIdForAllUsers: get users", err)
		return
	}
	defer cur.Close(ctx)

	err = cur.All(ctx, &users)
	if err != nil {
		log.Error("fixWsIdForAllUsers: decode users", err)
	}

	changed := 0
	for _, user := range users {
		if user.ChannelWS == "" {
			log.Info("fixWsIdForAllUsers: processing: ", user.Channel)
			channel, err := vkplay.GetChannelInfo(user.Channel)
			time.Sleep(2 * time.Second)
			if err != nil {
				log.Errorf("fixWsIdForAllUsers: get channel info %s, err: %s", user.Channel, err)
				log.Info("fixWsIdForAllUsers: --------------------")
				continue
			}

			newId := strings.Split(channel.Data.Channel.WebSocketChannels.Chat, ":")[1]

			// update user if DB
			user.ChannelWS = newId
			err = vkplay.InsertOrUpdateUser(context.Background(), user)
			if err != nil {
				log.Error("fixWsIdForAllUsers: update user", err)
				return
			}

			if user.Id == newId {
				log.Info("fixWsIdForAllUsers: doesn't need to update WS")
				log.Info("fixWsIdForAllUsers: --------------------")
				continue
			}

			// update ws
			err = chat.UpdateUserInWs(user.Id, newId)
			if err != nil {
				log.Error("fixWsIdForAllUsers: update user in ws", err)
				return
			}

			changed++
			log.Infof("fixWsIdForAllUsers: updated %s in ws, %s->%s", user.Channel, user.Id, newId)
			log.Info("fixWsIdForAllUsers: --------------------")
		}
	}
	log.Infof("fixWsIdForAllUsers: updated %d users", changed)
}

func GenerateNewWsConnectConfig() {
	log.Info("GenerateNewWsConnectConfig: start")
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	// get all users
	var users []model.User
	cur, err := mongo.GetCollection(mongo.Users).Find(ctx, bson.M{})
	if err != nil {
		log.Error("GenerateNewWsConnectConfig: get users", err)
		return
	}
	defer cur.Close(ctx)

	err = cur.All(ctx, &users)
	if err != nil {
		log.Error("GenerateNewWsConnectConfig: decode users", err)
	}

	cfg := vkplay.GetWsChannelsFromDB()
	cfg.ChannelsAutoJoin = []string{}

	for _, user := range users {
		if user.ChannelWS != "" {
			cfg.ChannelsAutoJoin = append(cfg.ChannelsAutoJoin, user.ChannelWS)
		}
	}

	log.Info("GenerateNewWsConnectConfig: new ws config length: ", len(cfg.ChannelsAutoJoin))
	err = vkplay.SaveWsChannelsToDB(cfg)
	if err != nil {
		log.Error("GenerateNewWsConnectConfig: save ws config", err)
		return
	}

	log.Info("GenerateNewWsConnectConfig: end")
}

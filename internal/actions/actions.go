package actions

import (
	"HoBot_Backend/internal/model"
	"HoBot_Backend/internal/mongodb"
	"HoBot_Backend/internal/service/user"
	"HoBot_Backend/internal/service/vkplay"
	"context"
	"encoding/csv"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2/log"
	"go.mongodb.org/mongo-driver/v2/bson"
)

func ChangeIds(ctx context.Context, db *mongodb.Client) {
	log.Info("ChangeIds: start")
	ctx, cancel := context.WithTimeout(ctx, 2*time.Minute)
	defer cancel()

	// read users.csv
	fileNewIds, err := os.Open("users.csv")
	if err != nil {
		log.Error("ChangeIds: error opening users.csv", err)
		return
	}
	defer fileNewIds.Close()

	newIdsReader := csv.NewReader(fileNewIds)
	newIdsRecords, err := newIdsReader.ReadAll()
	if err != nil {
		log.Error("ChangeIds: error reading users.csv", err)
		return
	}

	newUserIds := make(map[string]string)
	for i, record := range newIdsRecords {
		if i == 0 {
			continue
		}
		if len(record) > 0 {
			newUserIds[record[0]] = record[1]
		}
	}

	// get all users from DB
	var users []model.User
	cur, err := db.GetCollection(mongodb.Users).Find(ctx, bson.M{})
	if err != nil {
		log.Error("ChangeIds: error getting users from DB", err)
		return
	}
	defer cur.Close(ctx)

	err = cur.All(ctx, &users)
	if err != nil {
		log.Error("ChangeIds: error decoding users", err)
		return
	}

	var newIds, oldIds int
	for i, user := range users {
		if newId, ok := newUserIds[user.Id]; ok {
			user.UserId = newId
			newIds++
		} else {
			user.UserId = user.Id
			oldIds++
		}
		updateUser(ctx, db, user)
		log.Info("User updated ", i, "\\", len(users))
	}
	log.Info("ChangeIds: new ids: ", newIds)
	log.Info("ChangeIds: old ids: ", oldIds)
}

func updateUser(c context.Context, db *mongodb.Client, user model.User) {
	ctx, cancel := context.WithTimeout(c, 5*time.Second)
	defer cancel()

	_, err := db.GetCollection(mongodb.Users).UpdateOne(ctx, bson.M{"_id": user.Id}, bson.M{"$set": bson.M{"user_id": user.UserId}})
	if err != nil {
		log.Error("updateUser: update user", err)
		return
	}
}

func UpdateUserAvatarWsNick(c context.Context, db *mongodb.Client) {
	ctx, cancel := context.WithTimeout(c, 1*time.Hour)
	defer cancel()

	// get all users from DB
	var users []model.User
	cur, err := db.GetCollection(mongodb.Users).Find(ctx, bson.M{})
	if err != nil {
		log.Error("ChangeIds: error getting users from DB", err)
		return
	}
	defer cur.Close(ctx)

	err = cur.All(ctx, &users)
	if err != nil {
		log.Error("ChangeIds: error decoding users", err)
		return
	}

	var wsUpdated, nickUpdated int
	var usersError []model.User
	for i, user := range users {
		log.Info("Processing user:  ", user.Channel)
		channelInfo, err := vkplay.GetChannelInfo(user.Channel)
		if err != nil {
			log.Error("UpdateUserAvatarWsNick: error getting channel info", err)
			usersError = append(usersError, user)
			continue
		}

		user.AvatarURL = channelInfo.Data.Owner.AvatarURL + "&croped=1&mh=80&mw=80"

		if user.Nick != channelInfo.Data.Owner.Nick {
			log.Info("Nick updated: ", user.Nick, " -> ", channelInfo.Data.Owner.Nick)
			user.Nick = channelInfo.Data.Owner.Nick
			nickUpdated++
		}

		channelWs := strings.Split(channelInfo.Data.Channel.WebSocketChannels.Chat, ":")[1]
		if user.ChannelWS != channelWs {
			log.Info("WS updated: ", user.ChannelWS, " -> ", channelWs)
			user.ChannelWS = channelWs
			wsUpdated++
		}

		user.UserId = strconv.Itoa(channelInfo.Data.Owner.ID)

		updateUserAvatarWsNickUserId(ctx, db, user)
		log.Info("User updated ", i, "\\", len(users), " --------------------")
		time.Sleep(2 * time.Second)
	}

	log.Info("UpdateUserAvatarWsNick: ws updated: ", wsUpdated)
	log.Info("UpdateUserAvatarWsNick: nick updated: ", nickUpdated)
	log.Info("UpdateUserAvatarWsNick: users error: ", len(usersError))
	for _, user := range usersError {
		log.Info("User error: ", user.Channel)
	}
}

func updateUserAvatarWsNickUserId(c context.Context, db *mongodb.Client, user model.User) {
	ctx, cancel := context.WithTimeout(c, 5*time.Second)
	defer cancel()

	_, err := db.GetCollection(mongodb.Users).UpdateOne(ctx, bson.M{"_id": user.Id}, bson.M{"$set": bson.M{"avatar_url": user.AvatarURL, "nick": user.Nick, "channel_ws": user.ChannelWS, "user_id": user.UserId}})
	if err != nil {
		log.Error("updateUser: update user", err)
		return
	}
}

func RemoveNotExistChannels(c context.Context, db *mongodb.Client, userService *user.UserService) {
	ctx, cancel := context.WithTimeout(c, 2*time.Minute)
	defer cancel()

	var users []model.User
	cur, err := db.GetCollection(mongodb.Users).Find(ctx, bson.M{"user_id": bson.M{"$exists": false}})
	if err != nil {
		log.Error("RemoveNotExistChannels: get users", err)
		return
	}
	defer cur.Close(ctx)

	err = cur.All(ctx, &users)
	if err != nil {
		log.Error("RemoveNotExistChannels: decode", err)
	}

	log.Info("Found users: ", len(users))

	for _, user := range users {
		log.Info("Removing user: ", user.Channel)
		err = userService.WipeUser(ctx, user.Id)
		if err != nil {
			log.Error("RemoveNotExistChannels: wipe user", err)
			continue
		}
		log.Info("User removed: ", user.Channel)
	}

	log.Info("Done")
}

// Update ids in DB
func UpdateUserAndSettings(c context.Context, db *mongodb.Client) {
	ctx, cancel := context.WithTimeout(c, 1*time.Hour)
	defer cancel()

	// get all users from DB
	var users []model.User
	cur, err := db.GetCollection(mongodb.Users).Find(ctx, bson.M{})
	if err != nil {
		log.Error("ChangeIds: error getting users from DB", err)
		return
	}
	defer cur.Close(ctx)

	err = cur.All(ctx, &users)
	if err != nil {
		log.Error("ChangeIds: error decoding users", err)
		return
	}

	var usersIdUpdated, usersUpdated int
	var usersError []model.User
	var config model.Config

	for i, u := range users {
		user := u
		log.Info("Processing user:  ", user.Channel)
		channelInfo, err := vkplay.GetChannelInfo(user.Channel)
		if err != nil {
			log.Error("UpdateUserAvatarWsNick: error getting channel info", err)
			usersError = append(usersError, user)
			continue
		}

		var oldId string

		if user.Id != strconv.Itoa(channelInfo.Data.Channel.ID) {
			log.Info("Id updated: ", user.Id, " -> ", channelInfo.Data.Channel.ID)
			oldId = user.Id
			user.Id = strconv.Itoa(channelInfo.Data.Channel.ID)
		}

		user.AvatarURL = channelInfo.Data.Owner.AvatarURL + "&croped=1&mh=80&mw=80"

		if user.Nick != channelInfo.Data.Owner.Nick {
			log.Info("Nick updated: ", user.Nick, " -> ", channelInfo.Data.Owner.Nick)
			user.Nick = channelInfo.Data.Owner.Nick
		}

		channelWs := strings.Split(channelInfo.Data.Channel.WebSocketChannels.Chat, ":")[1]
		if user.ChannelWS != channelWs {
			log.Info("WS updated: ", user.ChannelWS, " -> ", channelWs)
			user.ChannelWS = channelWs
		}
		config.ChannelsAutoJoin = append(config.ChannelsAutoJoin, user.ChannelWS)

		if user.UserId != strconv.Itoa(channelInfo.Data.Owner.ID) {
			log.Info("UserId updated: ", user.UserId, " -> ", channelInfo.Data.Owner.ID)
			user.UserId = strconv.Itoa(channelInfo.Data.Owner.ID)
		}

		if oldId == "" {
			updateUserAvatarWsNickUserId(ctx, db, user)
			usersUpdated++
		} else {
			deleteUserAndInsertNew(ctx, db, oldId, user)
			deleteSettingsAndInsertNew(ctx, db, oldId, user.Id)
			usersIdUpdated++
		}

		log.Info("User updated ", i+1, "\\", len(users), " --------------------")
		time.Sleep(2 * time.Second)
	}

	err = saveWsChannels(ctx, db, config)
	if err != nil {
		log.Error("UpdateUserAvatarWsNick: error saving ws channels", err)
	}
	log.Info("Save ws channels: ", len(config.ChannelsAutoJoin))

	log.Info("Update User Id: ", usersIdUpdated)
	log.Info("Update User (without Id): ", usersUpdated)
	log.Info("Update User error: ", len(usersError))
	for _, user := range usersError {
		log.Info("User error: ", user.Channel)
	}
}

func deleteUserAndInsertNew(c context.Context, db *mongodb.Client, oldId string, user model.User) {
	ctx, cancel := context.WithTimeout(c, 5*time.Second)
	defer cancel()

	_, err := db.GetCollection(mongodb.Users).DeleteOne(ctx, bson.M{"_id": oldId})
	if err != nil {
		log.Error("deleteUserAndInsertNew: delete user", err)
		return
	}

	_, err = db.GetCollection(mongodb.Users).InsertOne(ctx, user)
	if err != nil {
		log.Error("deleteUserAndInsertNew: insert user", err)
		return
	}
}

func deleteSettingsAndInsertNew(c context.Context, db *mongodb.Client, oldId string, newId string) {
	ctx, cancel := context.WithTimeout(c, 5*time.Second)
	defer cancel()

	var userSetting model.UserSettings
	err := db.GetCollection(mongodb.UserSettings).FindOneAndDelete(ctx, bson.M{"_id": oldId}).Decode(&userSetting)
	if err != nil {
		log.Error("error deleting user setting", err)
		return
	}

	userSetting.Id = newId

	_, err = db.GetCollection(mongodb.UserSettings).InsertOne(ctx, userSetting)
	if err != nil {
		log.Error("deleteSettingsAndInsertNew: insert settings", err)
		return
	}
}

func saveWsChannels(ctx context.Context, db *mongodb.Client, config model.Config) error {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	config.Id = "ws"
	_, err := db.GetCollection(mongodb.Config).UpdateByID(ctx, "ws", bson.M{"$set": config})
	if err != nil {
		log.Error("Error while saving channels:", err)
		return err
	}
	return nil
}

/* func updateUser(ctx context.Context, db *mongodb.Client, id string, newId string) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	// update id in config ========================
	var config model.Config
	err := db.GetCollection(mongodb.Config).FindOne(ctx, bson.M{"_id": "ws"}).Decode(&config)
	if err != nil {
		log.Error("updateUser: get config", err)
		return
	}

	for i, v := range config.ChannelsAutoJoin {
		if v == id {
			config.ChannelsAutoJoin[i] = newId
			break
		}
	}

	_, err = db.GetCollection(mongodb.Config).UpdateOne(ctx, bson.M{"_id": "ws"}, bson.M{"$set": config})
	if err != nil {
		log.Error("updateUser: update config", err)
		return
	}

	log.Info("updateUser: config updated ")
	// ===============================================

	// update id in song requests ========================
	var songRequests []model.SongRequest
	cur, err := db.GetCollection(mongodb.SongRequests).Find(ctx, bson.M{"channel_id": id})
	if err != nil {
		log.Error("updateUser: get song requests", err)
		return
	}
	defer cur.Close(ctx)

	err = cur.All(ctx, &songRequests)
	if err != nil {
		log.Error("updateUser: decode song requests", err)
		return
	}

	for _, songRequest := range songRequests {
		if songRequest.ChannelId == id {
			songRequest.ChannelId = newId
			_, err = db.GetCollection(mongodb.SongRequests).UpdateOne(ctx, bson.M{"_id": songRequest.Id}, bson.M{"$set": songRequest})
			if err != nil {
				log.Error("updateUser: update song request", err)
				return
			}
		}

		log.Info("updateUser: song request updated ")
		// ===============================================

		// update id in song requests history ========================
		var songRequestsHistory []model.SongRequest
		cur, err := db.GetCollection(mongodb.SongRequestsHistory).Find(ctx, bson.M{"channel_id": id})
		if err != nil {
			log.Error("updateUser: get song requests history", err)
			return
		}
		defer cur.Close(ctx)

		err = cur.All(ctx, &songRequestsHistory)
		if err != nil {
			log.Error("updateUser: decode song requests history", err)
			return
		}

		for _, songRequestHistory := range songRequestsHistory {
			if songRequestHistory.ChannelId == id {
				songRequestHistory.ChannelId = newId
				_, err = db.GetCollection(mongodb.SongRequestsHistory).UpdateOne(ctx, bson.M{"_id": songRequestHistory.Id}, bson.M{"$set": songRequestHistory})
				if err != nil {
					log.Error("updateUser: update song requests history", err)
					return
				}
			}

			log.Info("updateUser: song requests history updated ")
			// ===============================================

			// update id in tokens ========================
			_, err = db.GetCollection(mongodb.Tokens).DeleteOne(ctx, bson.M{"user_id": id})
			if err != nil {
				log.Error("updateUser: delete tokens", err)
				return
			}

			log.Info("updateUser: tokens deleted ")
			// ===============================================

			// update id in user settings ========================
			var userSetting model.UserSettings
			err = db.GetCollection(mongodb.UserSettings).FindOne(ctx, bson.M{"_id": id}).Decode(&userSetting)
			if err != nil {
				log.Error("updateUser: get user settings", err)
				return
			}
			userSetting.Id = newId
			_, err = db.GetCollection(mongodb.UserSettings).UpdateOne(ctx, bson.M{"_id": id}, bson.M{"$set": userSetting})
			if err != nil {
				log.Error("updateUser: update user settings", err)
				return
			}

			log.Info("updateUser: user settings updated ")
			// ===============================================

			// update id in users ========================
			var user model.User
			err = db.GetCollection(mongodb.Users).FindOne(ctx, bson.M{"_id": id}).Decode(&user)
			if err != nil {
				log.Error("updateUser: get user", err)
				return
			}
			user.Id = newId
			_, err = db.GetCollection(mongodb.Users).UpdateOne(ctx, bson.M{"_id": id}, bson.M{"$set": user})
			if err != nil {
				log.Error("updateUser: update user", err)
				return
			}

			log.Info("updateUser: user updated ")
			// ===============================================

		}
	}

} */

/* import (
	"HoBot_Backend/internal/model"
	"HoBot_Backend/internal/mongo"
	"HoBot_Backend/internal/service/chat"
	"HoBot_Backend/internal/service/vkplay"
	"context"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2/log"
	"go.mongodb.org/mongo-driver/v2/bson"
)

func FixWsIdForAllUsers() {
	time.Sleep(40 * time.Second)
	log.Info("fixWsIdForAllUsers: start")
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	// get all users
	var users []model.User
	cur, err := mongodb.GetCollection(mongodb.Users).Find(ctx, bson.M{})
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
	cur, err := mongodb.GetCollection(mongodb.Users).Find(ctx, bson.M{})
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
*/

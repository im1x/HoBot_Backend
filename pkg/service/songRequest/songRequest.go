package songRequest

import (
	DB "HoBot_Backend/pkg/mongo"
	"HoBot_Backend/pkg/service/settings"
	"context"
	"errors"
	"github.com/gofiber/fiber/v2/log"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"time"
)

var VotesForSkip = make(map[string]*VotesForSkipSong)

func AddSongRequestToDB(songRequest SongRequest) (primitive.ObjectID, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	res, err := DB.GetCollection(DB.SongRequests).InsertOne(ctx, songRequest)
	if err != nil {
		log.Error("Error while inserting song request:", err)
		return primitive.NilObjectID, err
	}

	insertedID := res.InsertedID.(primitive.ObjectID)

	return insertedID, nil
}

func GetPlaylist(ctxReq context.Context, user string) ([]SongRequest, error) {
	var playlist []SongRequest
	ctx, cancel := context.WithTimeout(ctxReq, 3*time.Second)
	defer cancel()

	cursor, err := DB.GetCollection(DB.SongRequests).Find(ctx, bson.M{"channel_id": user})
	if err != nil {
		log.Error("Error while finding song request:", err)
		return nil, err
	}

	for cursor.Next(ctx) {
		var song SongRequest
		if err := cursor.Decode(&song); err != nil {
			log.Error("Error decoding song request:", err)
			continue
		}
		playlist = append(playlist, song)
	}

	if err := cursor.Err(); err != nil {
		log.Error("Error iterating over cursor:", err)
		return nil, err
	}

	return playlist, nil
}

func GetPlaylistHistory(ctxReq context.Context, user string) ([]SongRequest, error) {
	var playlistHistory []SongRequest
	ctx, cancel := context.WithTimeout(ctxReq, 3*time.Second)
	defer cancel()

	cursor, err := DB.GetCollection(DB.SongRequestsHistory).Find(ctx, bson.M{"channel_id": user})
	if err != nil {
		log.Error("Error while finding song requests history:", err)
		return nil, err
	}

	for cursor.Next(ctx) {
		var song SongRequest
		if err := cursor.Decode(&song); err != nil {
			log.Error("Error decoding song request history:", err)
			continue
		}
		playlistHistory = append(playlistHistory, song)
	}

	if err := cursor.Err(); err != nil {
		log.Error("Error iterating over cursor:", err)
		return nil, err
	}

	return playlistHistory, nil
}

func IsPlaylistFull(user string) bool {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	playlistCount, err := DB.GetCollection(DB.SongRequests).CountDocuments(ctx, bson.M{"channel_id": user})
	if err != nil {
		log.Error("Error while getting playlist count:", err)
		return false
	}

	if playlistCount >= 30 {
		return true
	}

	return false
}

func saveSongRequestToHistory(song SongRequest) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	// count user documents in history
	count, err := DB.GetCollection(DB.SongRequestsHistory).CountDocuments(ctx, bson.M{"channel_id": song.ChannelId})
	if err != nil {
		log.Error("Error while counting user documents in history:", err)
		return err
	}

	if count >= 3 {
		// delete oldest document
		filter := bson.M{"channel_id": song.ChannelId}
		opt := options.FindOneAndDelete().SetSort(bson.D{{"_id", 1}})
		result := DB.GetCollection(DB.SongRequestsHistory).FindOneAndDelete(ctx, filter, opt)
		if result.Err() != nil {
			log.Error("Error while deleting oldest document from history:", result.Err())
			return result.Err()
		}
	}

	_, err = DB.GetCollection(DB.SongRequestsHistory).InsertOne(ctx, song)
	if err != nil {
		log.Error("Error while inserting song to history:", err)
		return err
	}

	return nil
}

func SkipSong(channelId string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var song SongRequest

	filter := bson.M{"channel_id": channelId}
	opt := options.FindOneAndDelete().SetSort(bson.D{{"_id", 1}})
	err := DB.GetCollection(DB.SongRequests).FindOneAndDelete(ctx, filter, opt).Decode(&song)
	if err != nil {
		log.Error("Error while deleting song request:", err)
		return err
	}

	err = saveSongRequestToHistory(song)
	if err != nil {
		return err
	}

	/*if result.Err() != nil {
		log.Error("Error while deleting song request:", result.Err())
		return result.Err()
	}*/

	if settings.UsersSettings[channelId].SongRequests.IsUsersSkipAllowed {
		VotesForSkip[channelId] = nil
	}

	return nil
}

func RemoveAllSongs(channelId string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	_, err := DB.GetCollection(DB.SongRequests).DeleteMany(ctx, bson.M{"channel_id": channelId})
	if err != nil {
		log.Error("Error while deleting playlist:", err)
		return err
	}
	return nil
}

func RemoveSong(channelId, songId string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	objectId, err := primitive.ObjectIDFromHex(songId)
	if err != nil {
		return err
	}

	res, err := DB.GetCollection(DB.SongRequests).DeleteOne(ctx, bson.M{"channel_id": channelId, "_id": objectId})
	if err != nil || res.DeletedCount != 1 {
		log.Error("Error while deleting song:", err)
		return errors.New("error while deleting song")
	}
	return nil
}

func DeleteSongByYtId(channelId, ytId string) (SongRequest, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var song SongRequest
	err := DB.GetCollection(DB.SongRequests).FindOneAndDelete(ctx, bson.M{"channel_id": channelId, "yt_id": ytId}).Decode(&song)

	if err != nil {
		log.Error("Error while deleting song:", err)
		return SongRequest{}, errors.New("error while deleting song")
	}

	return song, nil
}

func GetCurrentSong(channelId string) (SongRequest, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var curSong SongRequest
	filter := bson.M{"channel_id": channelId}
	opt := options.FindOne().SetSort(bson.D{{"_id", 1}})
	res := DB.GetCollection(DB.SongRequests).FindOne(ctx, filter, opt)

	if res.Err() != nil {
		if errors.Is(res.Err(), mongo.ErrNoDocuments) {
			return SongRequest{}, nil
		}
		log.Error("Error while getting current song:", res.Err())
		return SongRequest{}, res.Err()
	}

	err := res.Decode(&curSong)
	if err != nil {
		log.Error("Error decoding current song:", err)
		return SongRequest{}, err
	}

	return curSong, nil
}

func CountSongsByUser(channelId string, userName string) (int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	res, err := DB.GetCollection(DB.SongRequests).CountDocuments(ctx, bson.M{"channel_id": channelId, "by": userName})
	if err != nil {
		log.Error("Error while counting songs by user:", err)
		return 0, err
	}
	return int(res), nil
}

func InitUsersSkipIfNeeded(userId string) {
	if VotesForSkip[userId] == nil {
		VotesForSkip[userId] = &VotesForSkipSong{
			AlreadyVoted: make(map[int]bool),
		}
	}
}

func VotesForSkipYes(channelId string, userId int) bool {
	InitUsersSkipIfNeeded(channelId)
	VotesForSkip[channelId].VoteYes(userId)

	if VotesForSkip[channelId].GetCount() >= settings.UsersSettings[channelId].SongRequests.UsersSkipValue {
		err := SkipSong(channelId)
		if err != nil {
			return false
		}
		return true
	}

	return false
}

func VotesForSkipNo(channelId string, userId int) {
	InitUsersSkipIfNeeded(channelId)
	VotesForSkip[channelId].VoteNo(userId)
}

package model

import "time"

type MovieKp struct {
	Id      int       `bson:"_id"`
	TitleEn string    `bson:"title_en"`
	TitleRu string    `bson:"title_ru"`
	Rating  int       `bson:"rating"`
	Date    time.Time `bson:"date"`
}

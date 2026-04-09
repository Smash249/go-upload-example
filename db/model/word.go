package model

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Word struct {
	ID                  primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	SeriesID            primitive.ObjectID `bson:"seriesId" json:"seriesId"`
	CategoryID          primitive.ObjectID `bson:"categoryId" json:"categoryId"`
	BaseText            string             `bson:"baseText" json:"baseText"`
	PosText             string             `bson:"posText" json:"posText"`
	PronText            string             `bson:"pronText" json:"pronText"`
	DefinitionText      string             `bson:"definitionText" json:"definitionText"`
	LearnerExamplesText string             `bson:"learnerExamplesText" json:"learnerExamplesText"`
	ChineseMeaning      string             `bson:"chineseMeaning" json:"chineseMeaning"`
	Image               string             `bson:"image" json:"image"`
	Sort                int                `bson:"sort" json:"sort"`
	IsActive            bool               `bson:"isActive" json:"isActive"`
	CreatedAt           time.Time          `bson:"createdAt" json:"createdAt"`
	UpdatedAt           time.Time          `bson:"updatedAt" json:"updatedAt"`
}

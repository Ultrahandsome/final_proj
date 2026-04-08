package data

import "go.mongodb.org/mongo-driver/bson/primitive"

// Comment is the internal representation for a comment
type Comment struct {
	ID               primitive.ObjectID `bson:"_id,omitempty" json:"_id,omitempty"`
	ConfidenceScore  float64            `bson:"confidenceScore" json:"confidenceScore"`
	IdentificationNo string             `bson:"identificationNumber" json:"identificationNumber"`
	ModeOfAttendance string             `bson:"modeOfAttendance" json:"modeOfAttendance"`
	TypeOfAttendance string             `bson:"typeOfAttendance" json:"typeOfAttendance"`
	NESBIndicator    string             `bson:"nesbIndicator" json:"nesbIndicator"`
	Citizenship      string             `bson:"citizenship" json:"citizenship"`
	StudyArea        string             `bson:"studyArea" json:"studyArea"`
	CourseLevel      string             `bson:"courseLevel" json:"courseLevel"`
	RawComment       string             `bson:"rawComment" json:"rawComment"`
	TrainCategory    string             `bson:"trainCategory" json:"trainCategory"`
	Category         string             `bson:"category" json:"category"`
	Deleted          bool               `bson:"deleted" json:"deleted"`
	LastUpdated      int64              `bson:"lastUpdated" json:"lastUpdated"`
	UUID             string             `bson:"uuid" json:"uuid"`
	SimilarComments  []string           `bson:"similarComments" json:"similarComments"`
	Keywords         []string           `bson:"keywords" json:"keywords"`
	UpdateHistory    []UpdateHistory    `bson:"updateHistory" json:"updateHistory"`
	CreateAt         int64              `bson:"createAt" json:"createAt"`
}

type UpdateHistory struct {
	User     string `bson:"user" json:"user"`
	Category string `bson:"category" json:"category"`
	Comment  string `bson:"comment" json:"comment"`
	Time     int64  `bson:"time" json:"time"`
}

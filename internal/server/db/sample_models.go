package db

import (
	"awans.org/aft/internal/model"
	"github.com/google/uuid"
)

func (db DB) AddSampleModels() {
	db.h.Insert(User)
	db.h.Insert(Profile)
	db.h.Insert(Post)
}

var User = model.Model{
	Type: "model",
	Id:   uuid.MustParse("887a91b8-3857-4b4d-a633-a6386a4fae25"),
	Name: "user",
	Attributes: map[string]model.Attribute{
		"firstName": model.Attribute{
			Type: model.String,
		},
		"lastName": model.Attribute{
			Type: model.String,
		},
		"age": model.Attribute{
			Type: model.Int,
		},
	},
	Relationships: map[string]model.Relationship{
		"posts": model.Relationship{
			TargetModel: "post",
			TargetRel:   "author",
			Type:        model.HasMany,
		},
		"profile": model.Relationship{
			TargetModel: "profile",
			TargetRel:   "user",
			Type:        model.HasOne,
		},
	},
}

var Profile = model.Model{
	Type: "model",
	Id:   uuid.MustParse("66783192-4111-4bd8-95dd-e7da460378df"),
	Name: "profile",
	Attributes: map[string]model.Attribute{
		"text": model.Attribute{
			Type: model.String,
		},
	},
	Relationships: map[string]model.Relationship{
		"user": model.Relationship{
			TargetModel: "user",
			TargetRel:   "profile",
			Type:        model.BelongsTo,
		},
	},
}

var Post = model.Model{
	Type: "model",
	Id:   uuid.MustParse("e25750c8-bb31-41fe-bdec-6bff1dceb2b4"),
	Name: "post",
	Attributes: map[string]model.Attribute{
		"text": model.Attribute{
			Type: model.String,
		},
	},
	Relationships: map[string]model.Relationship{
		"author": model.Relationship{
			TargetModel: "user",
			TargetRel:   "posts",
			Type:        model.BelongsTo,
		},
	},
}
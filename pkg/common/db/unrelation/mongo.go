// Copyright © 2023 OpenIM. All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package unrelation

import (
	"context"
	"fmt"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/x/bsonx"

	"github.com/OpenIMSDK/Open-IM-Server/pkg/common/config"
	"github.com/OpenIMSDK/Open-IM-Server/pkg/common/db/table/unrelation"
	"github.com/OpenIMSDK/Open-IM-Server/pkg/common/mw/specialerror"
	"github.com/OpenIMSDK/Open-IM-Server/pkg/errs"
	"github.com/OpenIMSDK/Open-IM-Server/pkg/utils"
)

type Mongo struct {
	db *mongo.Client
}

func NewMongo() (*Mongo, error) {
	specialerror.AddReplace(mongo.ErrNoDocuments, errs.ErrRecordNotFound)
	uri := "mongodb://sample.host:27017/?maxPoolSize=20&w=majority"
	if config.Config.Mongo.URI != "" {
		// example:
		// mongodb://$user:$password@mongo1.mongo:27017,mongo2.mongo:27017,mongo3.mongo:27017/$DBDatabase/?replicaSet=rs0&readPreference=secondary&authSource=admin&maxPoolSize=$DBMaxPoolSize
		uri = config.Config.Mongo.URI
	} else {
		// mongodb://mongodb1.example.com:27317,mongodb2.example.com:27017/?replicaSet=mySet&authSource=authDB
		mongodbHosts := ""
		for i, v := range config.Config.Mongo.Address {
			if i == len(config.Config.Mongo.Address)-1 {
				mongodbHosts += v
			} else {
				mongodbHosts += v + ","
			}
		}
		if config.Config.Mongo.Password != "" && config.Config.Mongo.Username != "" {
			uri = fmt.Sprintf("mongodb://%s:%s@%s/%s?maxPoolSize=%d&authSource=admin",
				config.Config.Mongo.Username, config.Config.Mongo.Password, mongodbHosts,
				config.Config.Mongo.Database, config.Config.Mongo.MaxPoolSize)
		} else {
			uri = fmt.Sprintf("mongodb://%s/%s/?maxPoolSize=%d&authSource=admin",
				mongodbHosts, config.Config.Mongo.Database,
				config.Config.Mongo.MaxPoolSize)
		}
	}
	fmt.Println("mongo:", uri)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*60)
	defer cancel()
	mongoClient, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		return nil, err
	}
	return &Mongo{db: mongoClient}, nil
}

func (m *Mongo) GetClient() *mongo.Client {
	return m.db
}

func (m *Mongo) GetDatabase() *mongo.Database {
	return m.db.Database(config.Config.Mongo.Database)
}

func (m *Mongo) CreateMsgIndex() error {
	return m.createMongoIndex(unrelation.Msg, true, "doc_id")
}

func (m *Mongo) CreateSuperGroupIndex() error {
	if err := m.createMongoIndex(unrelation.CSuperGroup, true, "group_id"); err != nil {
		return err
	}
	if err := m.createMongoIndex(unrelation.CUserToSuperGroup, true, "user_id"); err != nil {
		return err
	}
	return nil
}

func (m *Mongo) createMongoIndex(collection string, isUnique bool, keys ...string) error {
	db := m.db.Database(config.Config.Mongo.Database).Collection(collection)
	opts := options.CreateIndexes().SetMaxTime(10 * time.Second)
	indexView := db.Indexes()
	keysDoc := bsonx.Doc{}
	// create composite indexes
	for _, key := range keys {
		if strings.HasPrefix(key, "-") {
			keysDoc = keysDoc.Append(strings.TrimLeft(key, "-"), bsonx.Int32(-1))
		} else {
			keysDoc = keysDoc.Append(key, bsonx.Int32(1))
		}
	}
	// create index
	index := mongo.IndexModel{
		Keys: keysDoc,
	}
	if isUnique {
		index.Options = options.Index().SetUnique(true)
	}
	result, err := indexView.CreateOne(
		context.Background(),
		index,
		opts,
	)
	if err != nil {
		return utils.Wrap(err, result)
	}
	return nil
}

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

package rpcclient

import (
	"context"

	"google.golang.org/grpc"

	"github.com/OpenIMSDK/Open-IM-Server/pkg/common/config"
	"github.com/OpenIMSDK/Open-IM-Server/pkg/discoveryregistry"
	"github.com/OpenIMSDK/Open-IM-Server/pkg/proto/push"
)

type Push struct {
	conn   grpc.ClientConnInterface
	Client push.PushMsgServiceClient
	discov discoveryregistry.SvcDiscoveryRegistry
}

func NewPush(discov discoveryregistry.SvcDiscoveryRegistry) *Push {
	conn, err := discov.GetConn(context.Background(), config.Config.RPCRegisterName.OpenImPushName)
	if err != nil {
		panic(err)
	}
	return &Push{
		discov: discov,
		conn:   conn,
		Client: push.NewPushMsgServiceClient(conn),
	}
}

type PushRPCClient Push

func NewPushRPCClient(discov discoveryregistry.SvcDiscoveryRegistry) PushRPCClient {
	return PushRPCClient(*NewPush(discov))
}

func (p *PushRPCClient) DelUserPushToken(
	ctx context.Context,
	req *push.DelUserPushTokenReq,
) (*push.DelUserPushTokenResp, error) {
	return p.Client.DelUserPushToken(ctx, req)
}

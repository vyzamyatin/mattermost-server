// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See License.txt for license information.

package model

import (
	"encoding/json"
)

type PostCount struct {
	Count int64 `json:"count"`
}

func (o *PostCount) ToJson() string {
	b, _ := json.Marshal(o)
	return string(b)
}

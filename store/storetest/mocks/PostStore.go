// Code generated by mockery v1.0.0. DO NOT EDIT.

// Regenerate this file using `make store-mocks`.

package mocks

import (
	model "github.com/mattermost/mattermost-server/v5/model"
	mock "github.com/stretchr/testify/mock"
)

// PostStore is an autogenerated mock type for the PostStore type
type PostStore struct {
	mock.Mock
}

// AnalyticsPostCount provides a mock function with given fields: teamId, mustHaveFile, mustHaveHashtag
func (_m *PostStore) AnalyticsPostCount(teamId string, mustHaveFile bool, mustHaveHashtag bool) (int64, *model.AppError) {
	ret := _m.Called(teamId, mustHaveFile, mustHaveHashtag)

	var r0 int64
	if rf, ok := ret.Get(0).(func(string, bool, bool) int64); ok {
		r0 = rf(teamId, mustHaveFile, mustHaveHashtag)
	} else {
		r0 = ret.Get(0).(int64)
	}

	var r1 *model.AppError
	if rf, ok := ret.Get(1).(func(string, bool, bool) *model.AppError); ok {
		r1 = rf(teamId, mustHaveFile, mustHaveHashtag)
	} else {
		if ret.Get(1) != nil {
			r1 = ret.Get(1).(*model.AppError)
		}
	}

	return r0, r1
}

// AnalyticsPostCountsByDay provides a mock function with given fields: options
func (_m *PostStore) AnalyticsPostCountsByDay(options *model.AnalyticsPostCountsOptions) (model.AnalyticsRows, *model.AppError) {
	ret := _m.Called(options)

	var r0 model.AnalyticsRows
	if rf, ok := ret.Get(0).(func(*model.AnalyticsPostCountsOptions) model.AnalyticsRows); ok {
		r0 = rf(options)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(model.AnalyticsRows)
		}
	}

	var r1 *model.AppError
	if rf, ok := ret.Get(1).(func(*model.AnalyticsPostCountsOptions) *model.AppError); ok {
		r1 = rf(options)
	} else {
		if ret.Get(1) != nil {
			r1 = ret.Get(1).(*model.AppError)
		}
	}

	return r0, r1
}

// AnalyticsUserCountsWithPostsByDay provides a mock function with given fields: teamId
func (_m *PostStore) AnalyticsUserCountsWithPostsByDay(teamId string) (model.AnalyticsRows, *model.AppError) {
	ret := _m.Called(teamId)

	var r0 model.AnalyticsRows
	if rf, ok := ret.Get(0).(func(string) model.AnalyticsRows); ok {
		r0 = rf(teamId)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(model.AnalyticsRows)
		}
	}

	var r1 *model.AppError
	if rf, ok := ret.Get(1).(func(string) *model.AppError); ok {
		r1 = rf(teamId)
	} else {
		if ret.Get(1) != nil {
			r1 = ret.Get(1).(*model.AppError)
		}
	}

	return r0, r1
}

// ClearCaches provides a mock function with given fields:
func (_m *PostStore) ClearCaches() {
	_m.Called()
}

// Delete provides a mock function with given fields: postId, time, deleteByID
func (_m *PostStore) Delete(postId string, time int64, deleteByID string) *model.AppError {
	ret := _m.Called(postId, time, deleteByID)

	var r0 *model.AppError
	if rf, ok := ret.Get(0).(func(string, int64, string) *model.AppError); ok {
		r0 = rf(postId, time, deleteByID)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*model.AppError)
		}
	}

	return r0
}

// Get provides a mock function with given fields: id, skipFetchThreads
func (_m *PostStore) Get(id string, skipFetchThreads bool) (*model.PostList, *model.AppError) {
	ret := _m.Called(id, skipFetchThreads)

	var r0 *model.PostList
	if rf, ok := ret.Get(0).(func(string, bool) *model.PostList); ok {
		r0 = rf(id, skipFetchThreads)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*model.PostList)
		}
	}

	var r1 *model.AppError
	if rf, ok := ret.Get(1).(func(string, bool) *model.AppError); ok {
		r1 = rf(id, skipFetchThreads)
	} else {
		if ret.Get(1) != nil {
			r1 = ret.Get(1).(*model.AppError)
		}
	}

	return r0, r1
}

// GetDirectPostParentsForExportAfter provides a mock function with given fields: limit, afterId
func (_m *PostStore) GetDirectPostParentsForExportAfter(limit int, afterId string) ([]*model.DirectPostForExport, *model.AppError) {
	ret := _m.Called(limit, afterId)

	var r0 []*model.DirectPostForExport
	if rf, ok := ret.Get(0).(func(int, string) []*model.DirectPostForExport); ok {
		r0 = rf(limit, afterId)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*model.DirectPostForExport)
		}
	}

	var r1 *model.AppError
	if rf, ok := ret.Get(1).(func(int, string) *model.AppError); ok {
		r1 = rf(limit, afterId)
	} else {
		if ret.Get(1) != nil {
			r1 = ret.Get(1).(*model.AppError)
		}
	}

	return r0, r1
}

// GetEtag provides a mock function with given fields: channelId, allowFromCache
func (_m *PostStore) GetEtag(channelId string, allowFromCache bool) string {
	ret := _m.Called(channelId, allowFromCache)

	var r0 string
	if rf, ok := ret.Get(0).(func(string, bool) string); ok {
		r0 = rf(channelId, allowFromCache)
	} else {
		r0 = ret.Get(0).(string)
	}

	return r0
}

// GetFlaggedPosts provides a mock function with given fields: userId, offset, limit
func (_m *PostStore) GetFlaggedPosts(userId string, offset int, limit int) (*model.PostList, *model.AppError) {
	ret := _m.Called(userId, offset, limit)

	var r0 *model.PostList
	if rf, ok := ret.Get(0).(func(string, int, int) *model.PostList); ok {
		r0 = rf(userId, offset, limit)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*model.PostList)
		}
	}

	var r1 *model.AppError
	if rf, ok := ret.Get(1).(func(string, int, int) *model.AppError); ok {
		r1 = rf(userId, offset, limit)
	} else {
		if ret.Get(1) != nil {
			r1 = ret.Get(1).(*model.AppError)
		}
	}

	return r0, r1
}

// GetFlaggedPostsForChannel provides a mock function with given fields: userId, channelId, offset, limit
func (_m *PostStore) GetFlaggedPostsForChannel(userId string, channelId string, offset int, limit int) (*model.PostList, *model.AppError) {
	ret := _m.Called(userId, channelId, offset, limit)

	var r0 *model.PostList
	if rf, ok := ret.Get(0).(func(string, string, int, int) *model.PostList); ok {
		r0 = rf(userId, channelId, offset, limit)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*model.PostList)
		}
	}

	var r1 *model.AppError
	if rf, ok := ret.Get(1).(func(string, string, int, int) *model.AppError); ok {
		r1 = rf(userId, channelId, offset, limit)
	} else {
		if ret.Get(1) != nil {
			r1 = ret.Get(1).(*model.AppError)
		}
	}

	return r0, r1
}

// GetFlaggedPostsForTeam provides a mock function with given fields: userId, teamId, offset, limit
func (_m *PostStore) GetFlaggedPostsForTeam(userId string, teamId string, offset int, limit int) (*model.PostList, *model.AppError) {
	ret := _m.Called(userId, teamId, offset, limit)

	var r0 *model.PostList
	if rf, ok := ret.Get(0).(func(string, string, int, int) *model.PostList); ok {
		r0 = rf(userId, teamId, offset, limit)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*model.PostList)
		}
	}

	var r1 *model.AppError
	if rf, ok := ret.Get(1).(func(string, string, int, int) *model.AppError); ok {
		r1 = rf(userId, teamId, offset, limit)
	} else {
		if ret.Get(1) != nil {
			r1 = ret.Get(1).(*model.AppError)
		}
	}

	return r0, r1
}

// GetMaxPostSize provides a mock function with given fields:
func (_m *PostStore) GetMaxPostSize() int {
	ret := _m.Called()

	var r0 int
	if rf, ok := ret.Get(0).(func() int); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(int)
	}

	return r0
}

// GetOldest provides a mock function with given fields:
func (_m *PostStore) GetOldest() (*model.Post, *model.AppError) {
	ret := _m.Called()

	var r0 *model.Post
	if rf, ok := ret.Get(0).(func() *model.Post); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*model.Post)
		}
	}

	var r1 *model.AppError
	if rf, ok := ret.Get(1).(func() *model.AppError); ok {
		r1 = rf()
	} else {
		if ret.Get(1) != nil {
			r1 = ret.Get(1).(*model.AppError)
		}
	}

	return r0, r1
}

// GetParentsForExportAfter provides a mock function with given fields: limit, afterId
func (_m *PostStore) GetParentsForExportAfter(limit int, afterId string) ([]*model.PostForExport, *model.AppError) {
	ret := _m.Called(limit, afterId)

	var r0 []*model.PostForExport
	if rf, ok := ret.Get(0).(func(int, string) []*model.PostForExport); ok {
		r0 = rf(limit, afterId)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*model.PostForExport)
		}
	}

	var r1 *model.AppError
	if rf, ok := ret.Get(1).(func(int, string) *model.AppError); ok {
		r1 = rf(limit, afterId)
	} else {
		if ret.Get(1) != nil {
			r1 = ret.Get(1).(*model.AppError)
		}
	}

	return r0, r1
}

// GetPostAfterTime provides a mock function with given fields: channelId, time
func (_m *PostStore) GetPostAfterTime(channelId string, time int64) (*model.Post, *model.AppError) {
	ret := _m.Called(channelId, time)

	var r0 *model.Post
	if rf, ok := ret.Get(0).(func(string, int64) *model.Post); ok {
		r0 = rf(channelId, time)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*model.Post)
		}
	}

	var r1 *model.AppError
	if rf, ok := ret.Get(1).(func(string, int64) *model.AppError); ok {
		r1 = rf(channelId, time)
	} else {
		if ret.Get(1) != nil {
			r1 = ret.Get(1).(*model.AppError)
		}
	}

	return r0, r1
}

// GetPostIdAfterTime provides a mock function with given fields: channelId, time
func (_m *PostStore) GetPostIdAfterTime(channelId string, time int64) (string, *model.AppError) {
	ret := _m.Called(channelId, time)

	var r0 string
	if rf, ok := ret.Get(0).(func(string, int64) string); ok {
		r0 = rf(channelId, time)
	} else {
		r0 = ret.Get(0).(string)
	}

	var r1 *model.AppError
	if rf, ok := ret.Get(1).(func(string, int64) *model.AppError); ok {
		r1 = rf(channelId, time)
	} else {
		if ret.Get(1) != nil {
			r1 = ret.Get(1).(*model.AppError)
		}
	}

	return r0, r1
}

// GetPostIdBeforeTime provides a mock function with given fields: channelId, time
func (_m *PostStore) GetPostIdBeforeTime(channelId string, time int64) (string, *model.AppError) {
	ret := _m.Called(channelId, time)

	var r0 string
	if rf, ok := ret.Get(0).(func(string, int64) string); ok {
		r0 = rf(channelId, time)
	} else {
		r0 = ret.Get(0).(string)
	}

	var r1 *model.AppError
	if rf, ok := ret.Get(1).(func(string, int64) *model.AppError); ok {
		r1 = rf(channelId, time)
	} else {
		if ret.Get(1) != nil {
			r1 = ret.Get(1).(*model.AppError)
		}
	}

	return r0, r1
}

// GetPosts provides a mock function with given fields: options, allowFromCache
func (_m *PostStore) GetPosts(options model.GetPostsOptions, allowFromCache bool) (*model.PostList, *model.AppError) {
	ret := _m.Called(options, allowFromCache)

	var r0 *model.PostList
	if rf, ok := ret.Get(0).(func(model.GetPostsOptions, bool) *model.PostList); ok {
		r0 = rf(options, allowFromCache)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*model.PostList)
		}
	}

	var r1 *model.AppError
	if rf, ok := ret.Get(1).(func(model.GetPostsOptions, bool) *model.AppError); ok {
		r1 = rf(options, allowFromCache)
	} else {
		if ret.Get(1) != nil {
			r1 = ret.Get(1).(*model.AppError)
		}
	}

	return r0, r1
}

// GetPostsAfter provides a mock function with given fields: options
func (_m *PostStore) GetPostsAfter(options model.GetPostsOptions) (*model.PostList, *model.AppError) {
	ret := _m.Called(options)

	var r0 *model.PostList
	if rf, ok := ret.Get(0).(func(model.GetPostsOptions) *model.PostList); ok {
		r0 = rf(options)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*model.PostList)
		}
	}

	var r1 *model.AppError
	if rf, ok := ret.Get(1).(func(model.GetPostsOptions) *model.AppError); ok {
		r1 = rf(options)
	} else {
		if ret.Get(1) != nil {
			r1 = ret.Get(1).(*model.AppError)
		}
	}

	return r0, r1
}

func (_m *PostStore) GetChannelPostsUA(channelId string, after, before int64, desc bool, page int, perPage int) (*model.PostList, *model.AppError) {
	var r0 *model.PostList
	return r0, nil
}

func (_m *PostStore) CountChannelPostsUA(channelId string, after int64) (*model.PostCount, *model.AppError) {
	var r0 *model.PostCount
	return r0, nil
}

// GetPostsBatchForIndexing provides a mock function with given fields: startTime, endTime, limit
func (_m *PostStore) GetPostsBatchForIndexing(startTime int64, endTime int64, limit int) ([]*model.PostForIndexing, *model.AppError) {
	ret := _m.Called(startTime, endTime, limit)

	var r0 []*model.PostForIndexing
	if rf, ok := ret.Get(0).(func(int64, int64, int) []*model.PostForIndexing); ok {
		r0 = rf(startTime, endTime, limit)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*model.PostForIndexing)
		}
	}

	var r1 *model.AppError
	if rf, ok := ret.Get(1).(func(int64, int64, int) *model.AppError); ok {
		r1 = rf(startTime, endTime, limit)
	} else {
		if ret.Get(1) != nil {
			r1 = ret.Get(1).(*model.AppError)
		}
	}

	return r0, r1
}

// GetPostsBefore provides a mock function with given fields: options
func (_m *PostStore) GetPostsBefore(options model.GetPostsOptions) (*model.PostList, *model.AppError) {
	ret := _m.Called(options)

	var r0 *model.PostList
	if rf, ok := ret.Get(0).(func(model.GetPostsOptions) *model.PostList); ok {
		r0 = rf(options)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*model.PostList)
		}
	}

	var r1 *model.AppError
	if rf, ok := ret.Get(1).(func(model.GetPostsOptions) *model.AppError); ok {
		r1 = rf(options)
	} else {
		if ret.Get(1) != nil {
			r1 = ret.Get(1).(*model.AppError)
		}
	}

	return r0, r1
}

// GetPostsByIds provides a mock function with given fields: postIds
func (_m *PostStore) GetPostsByIds(postIds []string) ([]*model.Post, *model.AppError) {
	ret := _m.Called(postIds)

	var r0 []*model.Post
	if rf, ok := ret.Get(0).(func([]string) []*model.Post); ok {
		r0 = rf(postIds)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*model.Post)
		}
	}

	var r1 *model.AppError
	if rf, ok := ret.Get(1).(func([]string) *model.AppError); ok {
		r1 = rf(postIds)
	} else {
		if ret.Get(1) != nil {
			r1 = ret.Get(1).(*model.AppError)
		}
	}

	return r0, r1
}

// GetPostsCreatedAt provides a mock function with given fields: channelId, time
func (_m *PostStore) GetPostsCreatedAt(channelId string, time int64) ([]*model.Post, *model.AppError) {
	ret := _m.Called(channelId, time)

	var r0 []*model.Post
	if rf, ok := ret.Get(0).(func(string, int64) []*model.Post); ok {
		r0 = rf(channelId, time)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*model.Post)
		}
	}

	var r1 *model.AppError
	if rf, ok := ret.Get(1).(func(string, int64) *model.AppError); ok {
		r1 = rf(channelId, time)
	} else {
		if ret.Get(1) != nil {
			r1 = ret.Get(1).(*model.AppError)
		}
	}

	return r0, r1
}

// GetPostsSince provides a mock function with given fields: options, allowFromCache
func (_m *PostStore) GetPostsSince(options model.GetPostsSinceOptions, allowFromCache bool) (*model.PostList, *model.AppError) {
	ret := _m.Called(options, allowFromCache)

	var r0 *model.PostList
	if rf, ok := ret.Get(0).(func(model.GetPostsSinceOptions, bool) *model.PostList); ok {
		r0 = rf(options, allowFromCache)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*model.PostList)
		}
	}

	var r1 *model.AppError
	if rf, ok := ret.Get(1).(func(model.GetPostsSinceOptions, bool) *model.AppError); ok {
		r1 = rf(options, allowFromCache)
	} else {
		if ret.Get(1) != nil {
			r1 = ret.Get(1).(*model.AppError)
		}
	}

	return r0, r1
}

// GetRepliesForExport provides a mock function with given fields: parentId
func (_m *PostStore) GetRepliesForExport(parentId string) ([]*model.ReplyForExport, *model.AppError) {
	ret := _m.Called(parentId)

	var r0 []*model.ReplyForExport
	if rf, ok := ret.Get(0).(func(string) []*model.ReplyForExport); ok {
		r0 = rf(parentId)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*model.ReplyForExport)
		}
	}

	var r1 *model.AppError
	if rf, ok := ret.Get(1).(func(string) *model.AppError); ok {
		r1 = rf(parentId)
	} else {
		if ret.Get(1) != nil {
			r1 = ret.Get(1).(*model.AppError)
		}
	}

	return r0, r1
}

// GetSingle provides a mock function with given fields: id
func (_m *PostStore) GetSingle(id string) (*model.Post, *model.AppError) {
	ret := _m.Called(id)

	var r0 *model.Post
	if rf, ok := ret.Get(0).(func(string) *model.Post); ok {
		r0 = rf(id)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*model.Post)
		}
	}

	var r1 *model.AppError
	if rf, ok := ret.Get(1).(func(string) *model.AppError); ok {
		r1 = rf(id)
	} else {
		if ret.Get(1) != nil {
			r1 = ret.Get(1).(*model.AppError)
		}
	}

	return r0, r1
}

// InvalidateLastPostTimeCache provides a mock function with given fields: channelId
func (_m *PostStore) InvalidateLastPostTimeCache(channelId string) {
	_m.Called(channelId)
}

// Overwrite provides a mock function with given fields: post
func (_m *PostStore) Overwrite(post *model.Post) (*model.Post, *model.AppError) {
	ret := _m.Called(post)

	var r0 *model.Post
	if rf, ok := ret.Get(0).(func(*model.Post) *model.Post); ok {
		r0 = rf(post)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*model.Post)
		}
	}

	var r1 *model.AppError
	if rf, ok := ret.Get(1).(func(*model.Post) *model.AppError); ok {
		r1 = rf(post)
	} else {
		if ret.Get(1) != nil {
			r1 = ret.Get(1).(*model.AppError)
		}
	}

	return r0, r1
}

// OverwriteMultiple provides a mock function with given fields: posts
func (_m *PostStore) OverwriteMultiple(posts []*model.Post) ([]*model.Post, int, *model.AppError) {
	ret := _m.Called(posts)

	var r0 []*model.Post
	if rf, ok := ret.Get(0).(func([]*model.Post) []*model.Post); ok {
		r0 = rf(posts)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*model.Post)
		}
	}

	var r1 int
	if rf, ok := ret.Get(1).(func([]*model.Post) int); ok {
		r1 = rf(posts)
	} else {
		r1 = ret.Get(1).(int)
	}

	var r2 *model.AppError
	if rf, ok := ret.Get(2).(func([]*model.Post) *model.AppError); ok {
		r2 = rf(posts)
	} else {
		if ret.Get(2) != nil {
			r2 = ret.Get(2).(*model.AppError)
		}
	}

	return r0, r1, r2
}

// PermanentDeleteBatch provides a mock function with given fields: endTime, limit
func (_m *PostStore) PermanentDeleteBatch(endTime int64, limit int64) (int64, *model.AppError) {
	ret := _m.Called(endTime, limit)

	var r0 int64
	if rf, ok := ret.Get(0).(func(int64, int64) int64); ok {
		r0 = rf(endTime, limit)
	} else {
		r0 = ret.Get(0).(int64)
	}

	var r1 *model.AppError
	if rf, ok := ret.Get(1).(func(int64, int64) *model.AppError); ok {
		r1 = rf(endTime, limit)
	} else {
		if ret.Get(1) != nil {
			r1 = ret.Get(1).(*model.AppError)
		}
	}

	return r0, r1
}

// PermanentDeleteByChannel provides a mock function with given fields: channelId
func (_m *PostStore) PermanentDeleteByChannel(channelId string) *model.AppError {
	ret := _m.Called(channelId)

	var r0 *model.AppError
	if rf, ok := ret.Get(0).(func(string) *model.AppError); ok {
		r0 = rf(channelId)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*model.AppError)
		}
	}

	return r0
}

// PermanentDeleteByUser provides a mock function with given fields: userId
func (_m *PostStore) PermanentDeleteByUser(userId string) *model.AppError {
	ret := _m.Called(userId)

	var r0 *model.AppError
	if rf, ok := ret.Get(0).(func(string) *model.AppError); ok {
		r0 = rf(userId)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*model.AppError)
		}
	}

	return r0
}

// Save provides a mock function with given fields: post
func (_m *PostStore) Save(post *model.Post) (*model.Post, *model.AppError) {
	ret := _m.Called(post)

	var r0 *model.Post
	if rf, ok := ret.Get(0).(func(*model.Post) *model.Post); ok {
		r0 = rf(post)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*model.Post)
		}
	}

	var r1 *model.AppError
	if rf, ok := ret.Get(1).(func(*model.Post) *model.AppError); ok {
		r1 = rf(post)
	} else {
		if ret.Get(1) != nil {
			r1 = ret.Get(1).(*model.AppError)
		}
	}

	return r0, r1
}

// SaveMultiple provides a mock function with given fields: posts
func (_m *PostStore) SaveMultiple(posts []*model.Post) ([]*model.Post, int, *model.AppError) {
	ret := _m.Called(posts)

	var r0 []*model.Post
	if rf, ok := ret.Get(0).(func([]*model.Post) []*model.Post); ok {
		r0 = rf(posts)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*model.Post)
		}
	}

	var r1 int
	if rf, ok := ret.Get(1).(func([]*model.Post) int); ok {
		r1 = rf(posts)
	} else {
		r1 = ret.Get(1).(int)
	}

	var r2 *model.AppError
	if rf, ok := ret.Get(2).(func([]*model.Post) *model.AppError); ok {
		r2 = rf(posts)
	} else {
		if ret.Get(2) != nil {
			r2 = ret.Get(2).(*model.AppError)
		}
	}

	return r0, r1, r2
}

// Search provides a mock function with given fields: teamId, userId, params
func (_m *PostStore) Search(teamId string, userId string, params *model.SearchParams) (*model.PostList, *model.AppError) {
	ret := _m.Called(teamId, userId, params)

	var r0 *model.PostList
	if rf, ok := ret.Get(0).(func(string, string, *model.SearchParams) *model.PostList); ok {
		r0 = rf(teamId, userId, params)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*model.PostList)
		}
	}

	var r1 *model.AppError
	if rf, ok := ret.Get(1).(func(string, string, *model.SearchParams) *model.AppError); ok {
		r1 = rf(teamId, userId, params)
	} else {
		if ret.Get(1) != nil {
			r1 = ret.Get(1).(*model.AppError)
		}
	}

	return r0, r1
}

// SearchPostsInTeamForUser provides a mock function with given fields: paramsList, userId, teamId, isOrSearch, includeDeletedChannels, page, perPage
func (_m *PostStore) SearchPostsInTeamForUser(paramsList []*model.SearchParams, userId string, teamId string, isOrSearch bool, includeDeletedChannels bool, page int, perPage int) (*model.PostSearchResults, *model.AppError) {
	ret := _m.Called(paramsList, userId, teamId, isOrSearch, includeDeletedChannels, page, perPage)

	var r0 *model.PostSearchResults
	if rf, ok := ret.Get(0).(func([]*model.SearchParams, string, string, bool, bool, int, int) *model.PostSearchResults); ok {
		r0 = rf(paramsList, userId, teamId, isOrSearch, includeDeletedChannels, page, perPage)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*model.PostSearchResults)
		}
	}

	var r1 *model.AppError
	if rf, ok := ret.Get(1).(func([]*model.SearchParams, string, string, bool, bool, int, int) *model.AppError); ok {
		r1 = rf(paramsList, userId, teamId, isOrSearch, includeDeletedChannels, page, perPage)
	} else {
		if ret.Get(1) != nil {
			r1 = ret.Get(1).(*model.AppError)
		}
	}

	return r0, r1
}

// Update provides a mock function with given fields: newPost, oldPost
func (_m *PostStore) Update(newPost *model.Post, oldPost *model.Post) (*model.Post, *model.AppError) {
	ret := _m.Called(newPost, oldPost)

	var r0 *model.Post
	if rf, ok := ret.Get(0).(func(*model.Post, *model.Post) *model.Post); ok {
		r0 = rf(newPost, oldPost)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*model.Post)
		}
	}

	var r1 *model.AppError
	if rf, ok := ret.Get(1).(func(*model.Post, *model.Post) *model.AppError); ok {
		r1 = rf(newPost, oldPost)
	} else {
		if ret.Get(1) != nil {
			r1 = ret.Get(1).(*model.AppError)
		}
	}

	return r0, r1
}

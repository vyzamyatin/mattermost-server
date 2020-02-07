// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package storetest

import (
	"net/http"
	"sort"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/go-sql-driver/mysql"
	"github.com/lib/pq"
	"github.com/stretchr/testify/require"

	"github.com/mattermost/mattermost-server/v5/model"
	"github.com/mattermost/mattermost-server/v5/store"
	"github.com/stretchr/testify/assert"
)

func TestPluginStore(t *testing.T, ss store.Store, s SqlSupplier) {
	t.Run("SaveOrUpdate", func(t *testing.T) { testPluginSaveOrUpdate(t, ss, s) })
	t.Run("CompareAndSet", func(t *testing.T) { testPluginCompareAndSet(t, ss, s) })
	// t.Run("CompareAndDelete", func(t *testing.T) { testPluginCompareAndDelete(t, ss) })
	// t.Run("SetWithOptions", func(t *testing.T) { testPluginSetWithOptions(t, ss) })
	t.Run("Get", func(t *testing.T) { testPluginGet(t, ss) })
	t.Run("Delete", func(t *testing.T) { testPluginDelete(t, ss) })
	t.Run("DeleteAllForPlugin", func(t *testing.T) { testPluginDeleteAllForPlugin(t, ss) })
	t.Run("DeleteAllExpired", func(t *testing.T) { testPluginDeleteAllExpired(t, ss) })
	t.Run("List", func(t *testing.T) { testPluginList(t, ss) })
}

func setupKVs(t *testing.T, ss store.Store) (string, func()) {
	pluginId := model.NewId()
	otherPluginId := model.NewId()

	// otherKV is another key value for the current plugin, and used to verify other keys
	// aren't modified unintentionally.
	otherKV := &model.PluginKeyValue{
		PluginId: pluginId,
		Key:      model.NewId(),
		Value:    []byte(model.NewId()),
		ExpireAt: 0,
	}
	_, err := ss.Plugin().SaveOrUpdate(otherKV)
	require.Nil(t, err)

	// otherPluginKV is another key value for another plugin, and used to verify other keys
	// aren't modified unintentionally.
	otherPluginKV := &model.PluginKeyValue{
		PluginId: otherPluginId,
		Key:      model.NewId(),
		Value:    []byte(model.NewId()),
		ExpireAt: 0,
	}
	_, err = ss.Plugin().SaveOrUpdate(otherPluginKV)
	require.Nil(t, err)

	return pluginId, func() {
		actualOtherKV, err := ss.Plugin().Get(otherKV.PluginId, otherKV.Key)
		require.Nil(t, err, "failed to find other key value for same plugin")
		assert.Equal(t, otherKV, actualOtherKV)

		actualOtherPluginKV, err := ss.Plugin().Get(otherPluginKV.PluginId, otherPluginKV.Key)
		require.Nil(t, err, "failed to find other key value from different plugin")
		assert.Equal(t, otherPluginKV, actualOtherPluginKV)
	}
}

// raceyInserts stresses the given callback for setting a key by running it in parallel until
// a successful number of updates occur, all while also deleting the key repeatedly.
func raceyInserts(t *testing.T, ss store.Store, s SqlSupplier, callback func(kv *model.PluginKeyValue) error) {
	pluginId, tearDown := setupKVs(t, ss)
	defer tearDown()

	key := model.NewId()
	value := model.NewId()
	expireAt := int64(0)

	kv := &model.PluginKeyValue{
		PluginId: pluginId,
		Key:      key,
		Value:    []byte(value),
		ExpireAt: expireAt,
	}

	var wg sync.WaitGroup
	var count int32
	done := make(chan bool)

	// Repeatedly try to re-write the key value
	inserter := func() {
		defer wg.Done()
		for {
			select {
			case <-done:
				return
			default:
				err := callback(kv)
				if err != nil {
					// Duplicate isSerializationError to avoid circular import
					if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "40001" {
						continue
					}

					if mysqlErr, ok := err.(*mysql.MySQLError); ok && mysqlErr.Number == 1213 {
						continue
					}

					require.Nil(t, err, "expected nil or serialization error")
				}
				atomic.AddInt32(&count, 1)
			}
		}
	}

	// Repeatedly delete all key values
	deleter := func() {
		defer wg.Done()
		for {
			select {
			case <-done:
				return
			default:
				_, err := s.GetMaster().Exec("DELETE FROM PluginKeyValueStore WHERE PKey = :key", map[string]interface{}{"key": key})
				require.NoError(t, err)
			}
		}
	}

	// Spawn three competing inserters, and a task that consistently deletes records.
	wg.Add(1)
	go inserter()
	wg.Add(1)
	go inserter()
	wg.Add(1)
	go inserter()
	wg.Add(1)
	go deleter()

	assert.Eventually(t, func() bool {
		return atomic.LoadInt32(&count) >= 1000
	}, 15*time.Second, 100*time.Millisecond)

	close(done)
	wg.Wait()
}

func testPluginSaveOrUpdate(t *testing.T, ss store.Store, s SqlSupplier) {
	t.Run("invalid kv", func(t *testing.T) {
		_, tearDown := setupKVs(t, ss)
		defer tearDown()

		kv := &model.PluginKeyValue{
			PluginId: "",
			Key:      model.NewId(),
			Value:    []byte(model.NewId()),
			ExpireAt: 0,
		}

		kv, err := ss.Plugin().SaveOrUpdate(kv)
		require.NotNil(t, err)
		require.Equal(t, "model.plugin_key_value.is_valid.plugin_id.app_error", err.Id)
		assert.Nil(t, kv)
	})

	t.Run("new key", func(t *testing.T) {
		pluginId, tearDown := setupKVs(t, ss)
		defer tearDown()

		key := model.NewId()
		value := model.NewId()
		expireAt := int64(0)

		kv := &model.PluginKeyValue{
			PluginId: pluginId,
			Key:      key,
			Value:    []byte(value),
			ExpireAt: expireAt,
		}

		retKV, err := ss.Plugin().SaveOrUpdate(kv)
		require.Nil(t, err)
		assert.Equal(t, kv, retKV)
		// SaveOrUpdate returns the kv passed in, so test each field individually for
		// completeness. It should probably be changed to not bother doing that.
		assert.Equal(t, pluginId, kv.PluginId)
		assert.Equal(t, key, kv.Key)
		assert.Equal(t, []byte(value), retKV.Value)
		assert.Equal(t, expireAt, kv.ExpireAt)

		actualKV, err := ss.Plugin().Get(pluginId, key)
		require.Nil(t, err)
		assert.Equal(t, kv, actualKV)
	})

	t.Run("nil value for new key", func(t *testing.T) {
		pluginId, tearDown := setupKVs(t, ss)
		defer tearDown()

		key := model.NewId()
		var value []byte
		expireAt := int64(0)

		kv := &model.PluginKeyValue{
			PluginId: pluginId,
			Key:      key,
			Value:    value,
			ExpireAt: expireAt,
		}

		retKV, err := ss.Plugin().SaveOrUpdate(kv)
		require.Nil(t, err)

		require.Nil(t, err)
		assert.Equal(t, kv, retKV)
		// SaveOrUpdate returns the kv passed in, so test each field individually for
		// completeness. It should probably be changed to not bother doing that.
		assert.Equal(t, pluginId, kv.PluginId)
		assert.Equal(t, key, kv.Key)
		assert.Nil(t, retKV.Value)
		assert.Equal(t, expireAt, kv.ExpireAt)

		actualKV, err := ss.Plugin().Get(pluginId, key)
		require.NotNil(t, err)
		assert.Equal(t, err.StatusCode, http.StatusNotFound)
		assert.Nil(t, actualKV)
	})

	t.Run("existing key", func(t *testing.T) {
		pluginId, tearDown := setupKVs(t, ss)
		defer tearDown()

		key := model.NewId()
		value := model.NewId()
		expireAt := int64(0)

		kv := &model.PluginKeyValue{
			PluginId: pluginId,
			Key:      key,
			Value:    []byte(value),
			ExpireAt: expireAt,
		}

		_, err := ss.Plugin().SaveOrUpdate(kv)
		require.Nil(t, err)

		newValue := model.NewId()
		kv.Value = []byte(newValue)
		retKV, err := ss.Plugin().SaveOrUpdate(kv)

		require.Nil(t, err)
		assert.Equal(t, kv, retKV)
		// SaveOrUpdate returns the kv passed in, so test each field individually for
		// completeness. It should probably be changed to not bother doing that.
		assert.Equal(t, pluginId, kv.PluginId)
		assert.Equal(t, key, kv.Key)
		assert.Equal(t, []byte(newValue), retKV.Value)
		assert.Equal(t, expireAt, kv.ExpireAt)

		actualKV, err := ss.Plugin().Get(pluginId, key)
		require.Nil(t, err)
		assert.Equal(t, kv, actualKV)
	})

	t.Run("nil value for existing key", func(t *testing.T) {
		pluginId, tearDown := setupKVs(t, ss)
		defer tearDown()

		key := model.NewId()
		value := model.NewId()
		expireAt := int64(0)

		kv := &model.PluginKeyValue{
			PluginId: pluginId,
			Key:      key,
			Value:    []byte(value),
			ExpireAt: expireAt,
		}

		_, err := ss.Plugin().SaveOrUpdate(kv)
		require.Nil(t, err)

		kv.Value = nil
		retKV, err := ss.Plugin().SaveOrUpdate(kv)

		require.Nil(t, err)
		assert.Equal(t, kv, retKV)
		// SaveOrUpdate returns the kv passed in, so test each field individually for
		// completeness. It should probably be changed to not bother doing that.
		assert.Equal(t, pluginId, kv.PluginId)
		assert.Equal(t, key, kv.Key)
		assert.Nil(t, retKV.Value)
		assert.Equal(t, expireAt, kv.ExpireAt)

		actualKV, err := ss.Plugin().Get(pluginId, key)
		require.NotNil(t, err)
		assert.Equal(t, err.StatusCode, http.StatusNotFound)
		assert.Nil(t, actualKV)
	})

	t.Run("racey inserts", func(t *testing.T) {
		raceyInserts(t, ss, s, func(kv *model.PluginKeyValue) error {
			_, err := ss.Plugin().SaveOrUpdate(kv)
			return err
		})
	})
}

func testPluginCompareAndSet(t *testing.T, ss store.Store, s SqlSupplier) {
	t.Run("invalid kv", func(t *testing.T) {
		_, tearDown := setupKVs(t, ss)
		defer tearDown()

		kv := &model.PluginKeyValue{
			PluginId: "",
			Key:      model.NewId(),
			Value:    []byte(model.NewId()),
			ExpireAt: 0,
		}

		ok, err := ss.Plugin().CompareAndSet(kv, nil)
		require.NotNil(t, err)
		assert.Equal(t, "model.plugin_key_value.is_valid.plugin_id.app_error", err.Id)
		assert.False(t, ok)
	})

	// Note: CompareAndSet with a nil new value is explicitly tested in CompareAndDelete.

	t.Run("permutations", func(t *testing.T) {
		// The plugin id and key are signal values, and rewritten during the test case execution.
		pluginId := model.NewId()
		key := model.NewId()
		value := model.NewId()
		expireAt := int64(0)

		testCases := []struct {
			Description     string
			KV              *model.PluginKeyValue
			OldValue        []byte
			ExpectedSuccess bool
			ExpectedFound   bool
		}{
			{
				Description: "set non-existent key should succeed given nil old value",
				KV: &model.PluginKeyValue{
					PluginId: pluginId,
					Key:      model.NewId(),
					Value:    []byte(model.NewId()),
					ExpireAt: 0,
				},
				OldValue:        nil,
				ExpectedSuccess: true,
				ExpectedFound:   true,
			},
			{
				Description: "set non-existent key with nil value should succeed given nil old value",
				KV: &model.PluginKeyValue{
					PluginId: pluginId,
					Key:      model.NewId(),
					Value:    nil,
					ExpireAt: 0,
				},
				OldValue:        nil,
				ExpectedSuccess: false,
				ExpectedFound:   true,
			},
			{
				Description: "set non-existent key with future expiry should succeed given nil old value",
				KV: &model.PluginKeyValue{
					PluginId: pluginId,
					Key:      model.NewId(),
					Value:    []byte(model.NewId()),
					ExpireAt: model.GetMillis() + 15*1000,
				},
				OldValue:        nil,
				ExpectedSuccess: true,
				ExpectedFound:   true,
			},
			{
				Description: "set non-existent key should fail given non-nil old value",
				KV: &model.PluginKeyValue{
					PluginId: pluginId,
					Key:      model.NewId(),
					Value:    []byte(model.NewId()),
					ExpireAt: 0,
				},
				OldValue:        []byte(model.NewId()),
				ExpectedSuccess: false,
				ExpectedFound:   false,
			},
			{
				Description: "set non-existent key with nil value should fail given non-nil old value",
				KV: &model.PluginKeyValue{
					PluginId: pluginId,
					Key:      model.NewId(),
					Value:    nil,
					ExpireAt: 0,
				},
				OldValue:        []byte(model.NewId()),
				ExpectedSuccess: false,
				ExpectedFound:   false,
			},
			{
				Description: "set existing key with same value should fail given nil old value",
				KV: &model.PluginKeyValue{
					PluginId: pluginId,
					Key:      key,
					Value:    []byte(value),
					ExpireAt: 0,
				},
				OldValue:        nil,
				ExpectedSuccess: false,
				ExpectedFound:   true,
			},
			{
				Description: "set existing key with different value should fail given nil old value",
				KV: &model.PluginKeyValue{
					PluginId: pluginId,
					Key:      key,
					Value:    []byte(model.NewId()),
					ExpireAt: 0,
				},
				OldValue:        nil,
				ExpectedSuccess: false,
				ExpectedFound:   true,
			},
			{
				Description: "set existing key with nil value should fail given different old value",
				KV: &model.PluginKeyValue{
					PluginId: pluginId,
					Key:      key,
					Value:    nil,
					ExpireAt: 0,
				},
				OldValue:        []byte(model.NewId()),
				ExpectedSuccess: false,
				ExpectedFound:   true,
			},
			{
				Description: "set existing key with same value should fail given different old value",
				KV: &model.PluginKeyValue{
					PluginId: pluginId,
					Key:      key,
					Value:    []byte(value),
					ExpireAt: 0,
				},
				OldValue:        []byte(model.NewId()),
				ExpectedSuccess: false,
				ExpectedFound:   true,
			},
			{
				Description: "set existing key with different value should fail given different old value",
				KV: &model.PluginKeyValue{
					PluginId: pluginId,
					Key:      key,
					Value:    []byte(model.NewId()),
					ExpireAt: 0,
				},
				OldValue:        []byte(model.NewId()),
				ExpectedSuccess: false,
				ExpectedFound:   true,
			},
			{
				Description: "set existing key with nil value should succeed, deleting, given same old value",
				KV: &model.PluginKeyValue{
					PluginId: pluginId,
					Key:      key,
					Value:    nil,
					ExpireAt: 0,
				},
				OldValue:        []byte(value),
				ExpectedSuccess: true,
				ExpectedFound:   false,
			},
			{
				Description: "set existing key with same value should succeed given same old value",
				KV: &model.PluginKeyValue{
					PluginId: pluginId,
					Key:      key,
					Value:    []byte(value),
					ExpireAt: 0,
				},
				OldValue:        []byte(value),
				ExpectedSuccess: true,
				ExpectedFound:   true,
			},
			{
				Description: "set existing key with different value should succeed given same old value",
				KV: &model.PluginKeyValue{
					PluginId: pluginId,
					Key:      key,
					Value:    []byte(model.NewId()),
					ExpireAt: 0,
				},
				OldValue:        []byte(value),
				ExpectedSuccess: true,
				ExpectedFound:   true,
			},
			{
				Description: "set existing key with same value and future expiry should succeed given same old value",
				KV: &model.PluginKeyValue{
					PluginId: pluginId,
					Key:      key,
					Value:    []byte(value),
					ExpireAt: model.GetMillis() + 15*1000,
				},
				OldValue:        []byte(value),
				ExpectedSuccess: true,
				ExpectedFound:   true,
			},
			{
				Description: "set existing key with different value and future expiry should succeed given same old value",
				KV: &model.PluginKeyValue{
					PluginId: pluginId,
					Key:      key,
					Value:    []byte(model.NewId()),
					ExpireAt: model.GetMillis() + 15*1000,
				},
				OldValue:        []byte(value),
				ExpectedSuccess: true,
				ExpectedFound:   true,
			},
			{
				Description: "set existing key with same value and past expiry should succeed given same old value",
				KV: &model.PluginKeyValue{
					PluginId: pluginId,
					Key:      key,
					Value:    []byte(value),
					ExpireAt: model.GetMillis() - 15*1000,
				},
				OldValue:        []byte(value),
				ExpectedSuccess: true,
				ExpectedFound:   false,
			},
			{
				Description: "set existing key with different value and past expiry should succeed given same old value",
				KV: &model.PluginKeyValue{
					PluginId: pluginId,
					Key:      key,
					Value:    []byte(model.NewId()),
					ExpireAt: model.GetMillis() - 15*1000,
				},
				OldValue:        []byte(value),
				ExpectedSuccess: true,
				ExpectedFound:   false,
			},
		}

		for _, testCase := range testCases {
			t.Run(testCase.Description, func(t *testing.T) {
				// Isolate the tests by always using a new plugin id / key each time.
				actualPluginId, tearDown := setupKVs(t, ss)
				defer tearDown()

				actualKey := model.NewId()

				kv := &model.PluginKeyValue{
					PluginId: actualPluginId,
					Key:      actualKey,
					Value:    []byte(value),
					ExpireAt: expireAt,
				}
				_, err := ss.Plugin().SaveOrUpdate(kv)
				require.Nil(t, err)

				testCase.KV.PluginId = actualPluginId
				if testCase.KV.Key == key {
					testCase.KV.Key = actualKey
				}

				ok, err := ss.Plugin().CompareAndSet(testCase.KV, testCase.OldValue)
				require.Nil(t, err)
				assert.Equal(t, ok, testCase.ExpectedSuccess)

				if testCase.ExpectedSuccess {
					if testCase.ExpectedFound {
						actualKV, err := ss.Plugin().Get(testCase.KV.PluginId, testCase.KV.Key)
						require.Nil(t, err)
						assert.Equal(t, testCase.KV, actualKV)
					} else {
						actualKV, err := ss.Plugin().Get(testCase.KV.PluginId, testCase.KV.Key)
						assert.Equal(t, err.StatusCode, http.StatusNotFound)
						assert.Nil(t, actualKV)
					}
				} else {
					actualKV, err := ss.Plugin().Get(testCase.KV.PluginId, testCase.KV.Key)

					if testCase.KV.Key == actualKey {
						// If attempting to update the existing key, assume the
						// value is unchanged
						require.Nil(t, err)
						assert.Equal(t, kv, actualKV)
					} else {
						// If attempting to update any other key, assume the
						// value is never saved.
						assert.Equal(t, err.StatusCode, http.StatusNotFound)
						assert.Nil(t, actualKV)
					}
				}
			})
		}
	})

	t.Run("racey inserts", func(t *testing.T) {
		raceyInserts(t, ss, s, func(kv *model.PluginKeyValue) error {
			_, err := ss.Plugin().CompareAndSet(kv, nil)
			return err
		})
	})
}

func testPluginGet(t *testing.T, ss store.Store) {
	t.Run("no matching key value", func(t *testing.T) {
		pluginId := model.NewId()
		key := model.NewId()

		kv, err := ss.Plugin().Get(pluginId, key)
		require.NotNil(t, err)
		assert.Equal(t, err.StatusCode, http.StatusNotFound)
		assert.Nil(t, kv)
	})

	t.Run("no-matching key value for plugin id", func(t *testing.T) {
		pluginId := model.NewId()
		key := model.NewId()
		value := model.NewId()
		expireAt := int64(0)

		kv := &model.PluginKeyValue{
			PluginId: pluginId,
			Key:      key,
			Value:    []byte(value),
			ExpireAt: expireAt,
		}

		_, err := ss.Plugin().SaveOrUpdate(kv)
		require.Nil(t, err)

		kv, err = ss.Plugin().Get(model.NewId(), key)
		require.NotNil(t, err)
		assert.Equal(t, err.StatusCode, http.StatusNotFound)
		assert.Nil(t, kv)
	})

	t.Run("no-matching key value for key", func(t *testing.T) {
		pluginId := model.NewId()
		key := model.NewId()
		value := model.NewId()
		expireAt := int64(0)

		kv := &model.PluginKeyValue{
			PluginId: pluginId,
			Key:      key,
			Value:    []byte(value),
			ExpireAt: expireAt,
		}

		_, err := ss.Plugin().SaveOrUpdate(kv)
		require.Nil(t, err)

		kv, err = ss.Plugin().Get(pluginId, model.NewId())
		require.NotNil(t, err)
		assert.Equal(t, err.StatusCode, http.StatusNotFound)
		assert.Nil(t, kv)
	})

	t.Run("old expired key value", func(t *testing.T) {
		pluginId := model.NewId()
		key := model.NewId()
		value := model.NewId()
		expireAt := int64(1)

		kv := &model.PluginKeyValue{
			PluginId: pluginId,
			Key:      key,
			Value:    []byte(value),
			ExpireAt: expireAt,
		}

		_, err := ss.Plugin().SaveOrUpdate(kv)
		require.Nil(t, err)

		kv, err = ss.Plugin().Get(pluginId, model.NewId())
		require.NotNil(t, err)
		assert.Equal(t, err.StatusCode, http.StatusNotFound)
		assert.Nil(t, kv)
	})

	t.Run("recently expired key value", func(t *testing.T) {
		pluginId := model.NewId()
		key := model.NewId()
		value := model.NewId()
		expireAt := model.GetMillis() - 15*1000

		kv := &model.PluginKeyValue{
			PluginId: pluginId,
			Key:      key,
			Value:    []byte(value),
			ExpireAt: expireAt,
		}

		_, err := ss.Plugin().SaveOrUpdate(kv)
		require.Nil(t, err)

		kv, err = ss.Plugin().Get(pluginId, model.NewId())
		require.NotNil(t, err)
		assert.Equal(t, err.StatusCode, http.StatusNotFound)
		assert.Nil(t, kv)
	})

	t.Run("matching key value, non-expiring", func(t *testing.T) {
		pluginId := model.NewId()
		key := model.NewId()
		value := model.NewId()
		expireAt := int64(0)

		kv := &model.PluginKeyValue{
			PluginId: pluginId,
			Key:      key,
			Value:    []byte(value),
			ExpireAt: expireAt,
		}

		_, err := ss.Plugin().SaveOrUpdate(kv)
		require.Nil(t, err)

		actualKV, err := ss.Plugin().Get(pluginId, key)
		require.Nil(t, err)
		require.Equal(t, kv, actualKV)
	})

	t.Run("matching key value, not yet expired", func(t *testing.T) {
		pluginId := model.NewId()
		key := model.NewId()
		value := model.NewId()
		expireAt := model.GetMillis() + 15*1000

		kv := &model.PluginKeyValue{
			PluginId: pluginId,
			Key:      key,
			Value:    []byte(value),
			ExpireAt: expireAt,
		}

		_, err := ss.Plugin().SaveOrUpdate(kv)
		require.Nil(t, err)

		actualKV, err := ss.Plugin().Get(pluginId, key)
		require.Nil(t, err)
		require.Equal(t, kv, actualKV)
	})
}

func testPluginDelete(t *testing.T, ss store.Store) {
	t.Run("no matching key value", func(t *testing.T) {
		pluginId, tearDown := setupKVs(t, ss)
		defer tearDown()

		key := model.NewId()

		err := ss.Plugin().Delete(pluginId, key)
		require.Nil(t, err)

		kv, err := ss.Plugin().Get(pluginId, key)
		require.NotNil(t, err)
		assert.Equal(t, err.StatusCode, http.StatusNotFound)
		assert.Nil(t, kv)
	})

	testCases := []struct {
		description string
		expireAt    int64
	}{
		{
			"expired key value",
			model.GetMillis() - 15*1000,
		},
		{
			"never expiring value",
			0,
		},
		{
			"not yet expired value",
			model.GetMillis() + 15*1000,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.description, func(t *testing.T) {
			pluginId, tearDown := setupKVs(t, ss)
			defer tearDown()

			key := model.NewId()
			value := model.NewId()
			expireAt := testCase.expireAt

			kv := &model.PluginKeyValue{
				PluginId: pluginId,
				Key:      key,
				Value:    []byte(value),
				ExpireAt: expireAt,
			}

			_, err := ss.Plugin().SaveOrUpdate(kv)
			require.Nil(t, err)

			err = ss.Plugin().Delete(pluginId, key)
			require.Nil(t, err)

			kv, err = ss.Plugin().Get(pluginId, key)
			require.NotNil(t, err)
			assert.Equal(t, err.StatusCode, http.StatusNotFound)
			assert.Nil(t, kv)
		})
	}
}

func testPluginDeleteAllForPlugin(t *testing.T, ss store.Store) {
	setupKVsForDeleteAll := func(t *testing.T) (string, func()) {
		pluginId := model.NewId()
		otherPluginId := model.NewId()

		// otherPluginKV is another key value for another plugin, and used to verify other keys
		// aren't modified unintentionally.
		otherPluginKV := &model.PluginKeyValue{
			PluginId: otherPluginId,
			Key:      model.NewId(),
			Value:    []byte(model.NewId()),
			ExpireAt: 0,
		}
		_, err := ss.Plugin().SaveOrUpdate(otherPluginKV)
		require.Nil(t, err)

		return pluginId, func() {
			actualOtherPluginKV, err := ss.Plugin().Get(otherPluginKV.PluginId, otherPluginKV.Key)
			require.Nil(t, err, "failed to find other key value from different plugin")
			assert.Equal(t, otherPluginKV, actualOtherPluginKV)
		}
	}

	t.Run("no keys to delete", func(t *testing.T) {
		pluginId, tearDown := setupKVsForDeleteAll(t)
		defer tearDown()

		err := ss.Plugin().DeleteAllForPlugin(pluginId)
		require.Nil(t, err)
	})

	t.Run("2 keys to delete", func(t *testing.T) {
		pluginId, tearDown := setupKVsForDeleteAll(t)
		defer tearDown()

		kv := &model.PluginKeyValue{
			PluginId: pluginId,
			Key:      model.NewId(),
			Value:    []byte(model.NewId()),
			ExpireAt: 0,
		}
		_, err := ss.Plugin().SaveOrUpdate(kv)
		require.Nil(t, err)

		kv2 := &model.PluginKeyValue{
			PluginId: pluginId,
			Key:      model.NewId(),
			Value:    []byte(model.NewId()),
			ExpireAt: 0,
		}
		_, err = ss.Plugin().SaveOrUpdate(kv2)
		require.Nil(t, err)

		err = ss.Plugin().DeleteAllForPlugin(pluginId)
		require.Nil(t, err)

		_, err = ss.Plugin().Get(kv.PluginId, kv.Key)
		require.NotNil(t, err)
		assert.Equal(t, err.StatusCode, http.StatusNotFound)

		_, err = ss.Plugin().Get(kv.PluginId, kv2.Key)
		require.NotNil(t, err)
		assert.Equal(t, err.StatusCode, http.StatusNotFound)
	})
}

func testPluginDeleteAllExpired(t *testing.T, ss store.Store) {
	t.Run("no keys", func(t *testing.T) {
		err := ss.Plugin().DeleteAllExpired()
		require.Nil(t, err)
	})

	t.Run("no expiring keys to delete", func(t *testing.T) {
		pluginIdA := model.NewId()
		pluginIdB := model.NewId()

		kvA1 := &model.PluginKeyValue{
			PluginId: pluginIdA,
			Key:      model.NewId(),
			Value:    []byte(model.NewId()),
			ExpireAt: 0,
		}
		_, err := ss.Plugin().SaveOrUpdate(kvA1)
		require.Nil(t, err)

		kvA2 := &model.PluginKeyValue{
			PluginId: pluginIdA,
			Key:      model.NewId(),
			Value:    []byte(model.NewId()),
			ExpireAt: 0,
		}
		_, err = ss.Plugin().SaveOrUpdate(kvA2)
		require.Nil(t, err)

		kvB1 := &model.PluginKeyValue{
			PluginId: pluginIdB,
			Key:      model.NewId(),
			Value:    []byte(model.NewId()),
			ExpireAt: 0,
		}
		_, err = ss.Plugin().SaveOrUpdate(kvB1)
		require.Nil(t, err)

		kvB2 := &model.PluginKeyValue{
			PluginId: pluginIdB,
			Key:      model.NewId(),
			Value:    []byte(model.NewId()),
			ExpireAt: 0,
		}
		_, err = ss.Plugin().SaveOrUpdate(kvB2)
		require.Nil(t, err)

		err = ss.Plugin().DeleteAllExpired()
		require.Nil(t, err)

		actualKVA1, err := ss.Plugin().Get(pluginIdA, kvA1.Key)
		require.Nil(t, err)
		assert.Equal(t, kvA1, actualKVA1)

		actualKVA2, err := ss.Plugin().Get(pluginIdA, kvA2.Key)
		require.Nil(t, err)
		assert.Equal(t, kvA2, actualKVA2)

		actualKVB1, err := ss.Plugin().Get(pluginIdB, kvB1.Key)
		require.Nil(t, err)
		assert.Equal(t, kvB1, actualKVB1)

		actualKVB2, err := ss.Plugin().Get(pluginIdB, kvB2.Key)
		require.Nil(t, err)
		assert.Equal(t, kvB2, actualKVB2)
	})

	t.Run("no expired keys to delete", func(t *testing.T) {
		pluginIdA := model.NewId()
		pluginIdB := model.NewId()

		kvA1 := &model.PluginKeyValue{
			PluginId: pluginIdA,
			Key:      model.NewId(),
			Value:    []byte(model.NewId()),
			ExpireAt: model.GetMillis() + 15*1000,
		}
		_, err := ss.Plugin().SaveOrUpdate(kvA1)
		require.Nil(t, err)

		kvA2 := &model.PluginKeyValue{
			PluginId: pluginIdA,
			Key:      model.NewId(),
			Value:    []byte(model.NewId()),
			ExpireAt: model.GetMillis() + 15*1000,
		}
		_, err = ss.Plugin().SaveOrUpdate(kvA2)
		require.Nil(t, err)

		kvB1 := &model.PluginKeyValue{
			PluginId: pluginIdB,
			Key:      model.NewId(),
			Value:    []byte(model.NewId()),
			ExpireAt: model.GetMillis() + 15*1000,
		}
		_, err = ss.Plugin().SaveOrUpdate(kvB1)
		require.Nil(t, err)

		kvB2 := &model.PluginKeyValue{
			PluginId: pluginIdB,
			Key:      model.NewId(),
			Value:    []byte(model.NewId()),
			ExpireAt: model.GetMillis() + 15*1000,
		}
		_, err = ss.Plugin().SaveOrUpdate(kvB2)
		require.Nil(t, err)

		err = ss.Plugin().DeleteAllExpired()
		require.Nil(t, err)

		actualKVA1, err := ss.Plugin().Get(pluginIdA, kvA1.Key)
		require.Nil(t, err)
		assert.Equal(t, kvA1, actualKVA1)

		actualKVA2, err := ss.Plugin().Get(pluginIdA, kvA2.Key)
		require.Nil(t, err)
		assert.Equal(t, kvA2, actualKVA2)

		actualKVB1, err := ss.Plugin().Get(pluginIdB, kvB1.Key)
		require.Nil(t, err)
		assert.Equal(t, kvB1, actualKVB1)

		actualKVB2, err := ss.Plugin().Get(pluginIdB, kvB2.Key)
		require.Nil(t, err)
		assert.Equal(t, kvB2, actualKVB2)
	})

	t.Run("some expired keys to delete", func(t *testing.T) {
		pluginIdA := model.NewId()
		pluginIdB := model.NewId()

		kvA1 := &model.PluginKeyValue{
			PluginId: pluginIdA,
			Key:      model.NewId(),
			Value:    []byte(model.NewId()),
			ExpireAt: model.GetMillis() + 15*1000,
		}
		_, err := ss.Plugin().SaveOrUpdate(kvA1)
		require.Nil(t, err)

		expiredKVA2 := &model.PluginKeyValue{
			PluginId: pluginIdA,
			Key:      model.NewId(),
			Value:    []byte(model.NewId()),
			ExpireAt: model.GetMillis() - 15*1000,
		}
		_, err = ss.Plugin().SaveOrUpdate(expiredKVA2)
		require.Nil(t, err)

		kvB1 := &model.PluginKeyValue{
			PluginId: pluginIdB,
			Key:      model.NewId(),
			Value:    []byte(model.NewId()),
			ExpireAt: model.GetMillis() + 15*1000,
		}
		_, err = ss.Plugin().SaveOrUpdate(kvB1)
		require.Nil(t, err)

		expiredKVB2 := &model.PluginKeyValue{
			PluginId: pluginIdB,
			Key:      model.NewId(),
			Value:    []byte(model.NewId()),
			ExpireAt: model.GetMillis() - 15*1000,
		}
		_, err = ss.Plugin().SaveOrUpdate(expiredKVB2)
		require.Nil(t, err)

		err = ss.Plugin().DeleteAllExpired()
		require.Nil(t, err)

		actualKVA1, err := ss.Plugin().Get(pluginIdA, kvA1.Key)
		require.Nil(t, err)
		assert.Equal(t, kvA1, actualKVA1)

		actualKVA2, err := ss.Plugin().Get(pluginIdA, expiredKVA2.Key)
		require.NotNil(t, err)
		assert.Equal(t, err.StatusCode, http.StatusNotFound)
		assert.Nil(t, actualKVA2)

		actualKVB1, err := ss.Plugin().Get(pluginIdB, kvB1.Key)
		require.Nil(t, err)
		assert.Equal(t, kvB1, actualKVB1)

		actualKVB2, err := ss.Plugin().Get(pluginIdB, expiredKVB2.Key)
		require.NotNil(t, err)
		assert.Equal(t, err.StatusCode, http.StatusNotFound)
		assert.Nil(t, actualKVB2)
	})
}

func testPluginList(t *testing.T, ss store.Store) {
	t.Run("no key values", func(t *testing.T) {
		_, tearDown := setupKVs(t, ss)
		defer tearDown()

		// Ignore the pluginId setup by setupKVs
		pluginId := model.NewId()
		keys, err := ss.Plugin().List(pluginId, 0, 100)
		require.Nil(t, err)
		assert.Empty(t, keys)
	})

	t.Run("single key", func(t *testing.T) {
		_, tearDown := setupKVs(t, ss)
		defer tearDown()

		// Ignore the pluginId setup by setupKVs
		pluginId := model.NewId()

		kv := &model.PluginKeyValue{
			PluginId: pluginId,
			Key:      model.NewId(),
			Value:    []byte(model.NewId()),
			ExpireAt: 0,
		}
		_, err := ss.Plugin().SaveOrUpdate(kv)
		require.Nil(t, err)

		keys, err := ss.Plugin().List(pluginId, 0, 100)
		require.Nil(t, err)
		require.Len(t, keys, 1)
		assert.Equal(t, kv.Key, keys[0])
	})

	t.Run("multiple keys", func(t *testing.T) {
		_, tearDown := setupKVs(t, ss)
		defer tearDown()

		// Ignore the pluginId setup by setupKVs
		pluginId := model.NewId()

		var keys []string
		for i := 0; i < 150; i++ {
			key := model.NewId()
			kv := &model.PluginKeyValue{
				PluginId: pluginId,
				Key:      key,
				Value:    []byte(model.NewId()),
				ExpireAt: 0,
			}
			_, err := ss.Plugin().SaveOrUpdate(kv)
			require.Nil(t, err)

			keys = append(keys, key)
		}
		sort.Strings(keys)

		keys1, err := ss.Plugin().List(pluginId, 0, 100)
		require.Nil(t, err)
		require.Len(t, keys1, 100)

		keys2, err := ss.Plugin().List(pluginId, 100, 100)
		require.Nil(t, err)
		require.Len(t, keys2, 50)

		actualKeys := append(keys1, keys2...)
		sort.Strings(actualKeys)

		assert.Equal(t, keys, actualKeys)
	})

	t.Run("multiple keys, some expiring", func(t *testing.T) {
		_, tearDown := setupKVs(t, ss)
		defer tearDown()

		// Ignore the pluginId setup by setupKVs
		pluginId := model.NewId()

		var keys []string
		var expiredKeys []string
		now := model.GetMillis()
		for i := 0; i < 150; i++ {
			key := model.NewId()
			var expireAt int64

			if i%10 == 0 {
				// Expire keys 0, 10, 20, ...
				expireAt = 1

			} else if (i+5)%10 == 0 {
				// Mark for expiry keys 5, 15, 25, ...
				expireAt = now + 5*60*1000
			}

			kv := &model.PluginKeyValue{
				PluginId: pluginId,
				Key:      key,
				Value:    []byte(model.NewId()),
				ExpireAt: expireAt,
			}
			_, err := ss.Plugin().SaveOrUpdate(kv)
			require.Nil(t, err)

			if expireAt == 0 || expireAt > now {
				keys = append(keys, key)
			} else {
				expiredKeys = append(expiredKeys, key)
			}
		}
		sort.Strings(keys)

		keys1, err := ss.Plugin().List(pluginId, 0, 100)
		require.Nil(t, err)
		require.Len(t, keys1, 100)

		keys2, err := ss.Plugin().List(pluginId, 100, 100)
		require.Nil(t, err)
		require.Len(t, keys2, 35)

		actualKeys := append(keys1, keys2...)
		sort.Strings(actualKeys)

		assert.Equal(t, keys, actualKeys)
	})
}

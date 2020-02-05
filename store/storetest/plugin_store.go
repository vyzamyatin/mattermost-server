// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package storetest

import (
	"net/http"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/mattermost/mattermost-server/v5/model"
	"github.com/mattermost/mattermost-server/v5/store"
	"github.com/stretchr/testify/assert"
)

func TestPluginStore(t *testing.T, ss store.Store, s SqlSupplier) {
	t.Run("SaveOrUpdate", func(t *testing.T) { testPluginSaveOrUpdate(t, ss, s) })
	t.Run("CompareAndSet", func(t *testing.T) { testPluginCompareAndSet(t, ss) }) // TODO
	// t.Run("CompareAndDelete", func(t *testing.T) { testPluginCompareAndDelete(t, ss) })
	// t.Run("SetWithOptions", func(t *testing.T) { testPluginSetWithOptions(t, ss) })
	t.Run("PluginGet", func(t *testing.T) { testPluginGet(t, ss) })
	t.Run("PluginDelete", func(t *testing.T) { testPluginDelete(t, ss) })
	t.Run("PluginDeleteAllForPlugin", func(t *testing.T) { testPluginDeleteAllForPlugin(t, ss) })
	t.Run("PluginDeleteAllExpired", func(t *testing.T) { testPluginDeleteAllExpired(t, ss) })
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

		// Repeatedly try to write the key value
		inserter := func() {
			defer wg.Done()
			for {
				select {
				case <-done:
					return
				default:
					_, err := ss.Plugin().SaveOrUpdate(kv)
					require.Nil(t, err)
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
			// Require 1000 successful inserts.
			return atomic.LoadInt32(&count) >= 1000
		}, 15*time.Second, 100*time.Millisecond)

		close(done)
		wg.Wait()
	})
}

func testPluginCompareAndSet(t *testing.T, ss store.Store) {
	kv := &model.PluginKeyValue{
		PluginId: model.NewId(),
		Key:      model.NewId(),
		Value:    []byte(model.NewId()),
		ExpireAt: 0,
	}
	defer func() {
		_ = ss.Plugin().Delete(kv.PluginId, kv.Key)
	}()

	t.Run("set non-existent key should succeed given nil old value", func(t *testing.T) {
		ok, err := ss.Plugin().CompareAndSet(kv, nil)
		require.Nil(t, err)
		assert.True(t, ok)
	})

	t.Run("set existing key without old value should fail without error because is a automatically handled race condition", func(t *testing.T) {
		_, err := ss.Plugin().SaveOrUpdate(kv)
		require.Nil(t, err)

		kvNew := &model.PluginKeyValue{
			PluginId: kv.PluginId,
			Key:      kv.Key,
			Value:    []byte(model.NewId()),
			ExpireAt: 0,
		}

		ok, err := ss.Plugin().CompareAndSet(kvNew, nil)
		require.Nil(t, err)
		assert.False(t, ok)
	})

	t.Run("set existing key with new value should succeed given same old value", func(t *testing.T) {
		_, err := ss.Plugin().SaveOrUpdate(kv)
		require.Nil(t, err)

		kvNew := &model.PluginKeyValue{
			PluginId: kv.PluginId,
			Key:      kv.Key,
			Value:    []byte(model.NewId()),
			ExpireAt: 0,
		}

		ok, err := ss.Plugin().CompareAndSet(kvNew, kv.Value)
		require.Nil(t, err)
		assert.True(t, ok)
	})

	t.Run("set existing key with new value should fail given different old value", func(t *testing.T) {
		_, err := ss.Plugin().SaveOrUpdate(kv)
		require.Nil(t, err)

		kvNew := &model.PluginKeyValue{
			PluginId: kv.PluginId,
			Key:      kv.Key,
			Value:    []byte(model.NewId()),
			ExpireAt: 0,
		}

		ok, err := ss.Plugin().CompareAndSet(kvNew, []byte(model.NewId()))
		require.Nil(t, err)
		assert.False(t, ok)
	})

	t.Run("set existing key with same value should succeed given same old value", func(t *testing.T) {
		_, err := ss.Plugin().SaveOrUpdate(kv)
		require.Nil(t, err)

		ok, err := ss.Plugin().CompareAndSet(kv, kv.Value)
		require.Nil(t, err)
		assert.True(t, ok)
	})

	t.Run("set existing key with same value should fail given different old value", func(t *testing.T) {
		_, err := ss.Plugin().SaveOrUpdate(kv)
		require.Nil(t, err)

		ok, err := ss.Plugin().CompareAndSet(kv, []byte(model.NewId()))
		require.Nil(t, err)
		assert.False(t, ok)
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
		assert.Empty(keys)
	})

	t.Run("single key", func(t *testing.T) {
		pluginId, tearDown := setupKVs(t, ss)
		defer tearDown()

		keys, err := ss.Plugin().List(pluginId, 0, 100)
		require.Nil(t, err)
		assert.Len(t, keys, 1)

		kv, err := ss.Plugin().Get(pluginId, keys[0])
		require.Nil(t, err)
		require.Equal(t, keys[0], kv.Key)
		require.Equal(t, pluginId, kv.PluginId)
	})

	// t.Run("multiple keys", func(t *testing.T) {
	// 	_, tearDown := setupKVs(t, ss)
	// 	defer tearDown()

	// 	// Ignore the pluginId setup by setupKVs
	// 	pluginId := model.NewId()

	// 	keys, err := ss.Plugin().List(pluginId, 0, 100)
	// 	require.Nil(t, err)
	// 	assert.Len(t, keys, 1)

	// 	kv, err := ss.Plugin().Get(pluginId, keys[0])
	// 	require.Nil(t, err)
	// 	require.Equal(t, keys[0], kv.Key)
	// 	require.Equal(t, pluginId, kv.PluginId)
	// })

	// testCases := []struct {
	// 	description string
	// 	expireAt    int64
	// }{
	// 	{
	// 		"expired key value",
	// 		model.GetMillis() - 15*1000,
	// 	},
	// 	{
	// 		"never expiring value",
	// 		0,
	// 	},
	// 	{
	// 		"not yet expired value",
	// 		model.GetMillis() + 15*1000,
	// 	},
	// }

	// for _, testCase := range testCases {
	// 	t.Run(testCase.description, func(t *testing.T) {
	// 		pluginId, tearDown := setupKVs(t, ss)
	// 		defer tearDown()

	// 		key := model.NewId()
	// 		value := model.NewId()
	// 		expireAt := testCase.expireAt

	// 		kv := &model.PluginKeyValue{
	// 			PluginId: pluginId,
	// 			Key:      key,
	// 			Value:    []byte(value),
	// 			ExpireAt: expireAt,
	// 		}

	// 		_, err := ss.Plugin().SaveOrUpdate(kv)
	// 		require.Nil(t, err)

	// 		err = ss.Plugin().Delete(pluginId, key)
	// 		require.Nil(t, err)

	// 		kv, err = ss.Plugin().Get(pluginId, key)
	// 		require.NotNil(t, err)
	// 		assert.Equal(t, err.StatusCode, http.StatusNotFound)
	// 		assert.Nil(t, kv)
	// 	})
	// }
}

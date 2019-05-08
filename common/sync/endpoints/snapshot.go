/*
 * Copyright (c) 2019. Abstrium SAS <team (at) pydio.com>
 * This file is part of Pydio Cells.
 *
 * Pydio Cells is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Affero General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * Pydio Cells is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU Affero General Public License for more details.
 *
 * You should have received a copy of the GNU Affero General Public License
 * along with Pydio Cells.  If not, see <http://www.gnu.org/licenses/>.
 *
 * The latest code can be found at <https://pydio.com>.
 */

package endpoints

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/etcd-io/bbolt"
	"github.com/golang/protobuf/proto"
	"github.com/micro/go-micro/errors"

	"github.com/pydio/cells/common/config"
	"github.com/pydio/cells/common/proto/tree"
	"github.com/pydio/cells/common/sync/model"
)

var (
	bucketName        = []byte("snapshot")
	captureBucketName = []byte("capture")
)

type Snapshot struct {
	db         *bbolt.DB
	name       string
	empty      bool
	folderPath string
}

func NewSnapshot(name, syncUuid string) (*Snapshot, error) {
	s := &Snapshot{name: name}
	options := bbolt.DefaultOptions
	options.Timeout = 5 * time.Second
	appDir := config.ApplicationDataDir()
	s.folderPath = filepath.Join(appDir, "sync", syncUuid)
	os.MkdirAll(s.folderPath, 0755)
	p := filepath.Join(s.folderPath, "snapshot-"+name)
	if _, err := os.Stat(p); err != nil {
		s.empty = true
	}
	db, err := bbolt.Open(p, 0644, options)
	if err != nil {
		return nil, err
	}
	s.db = db
	return s, nil
}

func (s *Snapshot) CreateNode(ctx context.Context, node *tree.Node, updateIfExists bool) (err error) {
	return s.db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket(bucketName)
		if b == nil {
			return fmt.Errorf("cannot find root bucket")
		}
		// Create parents if necessary
		dir := strings.Trim(path.Dir(node.Path), "/")
		if dir != "" && dir != "." {
			parts := strings.Split(strings.Trim(path.Dir(node.Path), "/"), "/")
			for i := 0; i < len(parts); i++ {
				pKey := strings.Join(parts[:i+1], "/")
				if ex := b.Get([]byte(pKey)); ex == nil {
					b.Put([]byte(pKey), s.marshal(&tree.Node{Path: pKey, Type: tree.NodeType_COLLECTION, Etag: "-1"}))
				}
			}
		}
		return b.Put([]byte(node.Path), s.marshal(node))
	})
}

func (s *Snapshot) UpdateNode(ctx context.Context, node *tree.Node) (err error) {
	return s.CreateNode(ctx, node, true)
}

func (s *Snapshot) DeleteNode(ctx context.Context, path string) (err error) {
	return s.db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket(bucketName)
		if b == nil {
			return fmt.Errorf("cannot find root bucket")
		}
		if d := b.Get([]byte(path)); d != nil {
			b.Delete([]byte(path))
			// Delete children
			c := b.Cursor()
			prefix := []byte(strings.TrimRight(path, "/") + "/")
			var children [][]byte
			for k, _ := c.Seek(prefix); k != nil && bytes.HasPrefix(k, prefix); k, _ = c.Next() {
				children = append(children, k)
			}
			for _, k := range children {
				b.Delete(k)
			}
		}
		return nil
	})
}

func (s *Snapshot) MoveNode(ctx context.Context, oldPath string, newPath string) (err error) {
	return s.db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket(bucketName)
		if b == nil {
			return fmt.Errorf("cannot find root bucket")
		}
		var moves [][]byte
		if d := b.Get([]byte(oldPath)); d != nil {
			moves = append(moves, []byte(oldPath))
			// Delete children
			c := b.Cursor()
			prefix := []byte(oldPath + "/")
			for k, _ := c.Seek(prefix); k != nil && (string(k) == oldPath || bytes.HasPrefix(k, prefix)); k, _ = c.Next() {
				moves = append(moves, k)
			}
		}
		for _, m := range moves {
			renamed := path.Join(newPath, strings.TrimPrefix(string(m), oldPath))
			if node, e := s.unmarshal(b.Get(m)); e == nil {
				node.Path = renamed
				b.Delete(m)
				b.Put([]byte(renamed), s.marshal(node))
			}
		}
		return nil
	})
}

func (s *Snapshot) IsEmpty() bool {
	return s.empty
}

func (s *Snapshot) Close(delete ...bool) {
	s.db.Close()
	if len(delete) > 0 && delete[0] && s.folderPath != "" {
		os.RemoveAll(s.folderPath)
	}
}

func (s *Snapshot) Capture(ctx context.Context, source model.PathSyncSource, paths ...string) error {
	// Capture in temporary bucket
	e := s.db.Update(func(tx *bbolt.Tx) error {
		var capture *bbolt.Bucket
		var e error
		if b := tx.Bucket(captureBucketName); b != nil {
			if e = tx.DeleteBucket(captureBucketName); e != nil {
				return e
			}
		}
		if capture, e = tx.CreateBucket(captureBucketName); e != nil {
			return e
		}
		if len(paths) == 0 {
			return source.Walk(func(path string, node *tree.Node, err error) {
				capture.Put([]byte(path), s.marshal(node))
			}, "/")
		} else {
			for _, p := range paths {
				e := source.Walk(func(path string, node *tree.Node, err error) {
					capture.Put([]byte(path), s.marshal(node))
				}, p)
				if e != nil {
					return e
				}
			}
			return nil
		}
	})
	if e != nil {
		return e
	}
	// Now copy all to original bucket
	if e = s.db.Update(func(tx *bbolt.Tx) error {
		var clear *bbolt.Bucket
		var e error
		if b := tx.Bucket(bucketName); b != nil {
			if e = tx.DeleteBucket(bucketName); e != nil {
				return e
			}
		}
		if clear, e = tx.CreateBucket(bucketName); e != nil {
			return e
		}
		if captured := tx.Bucket(captureBucketName); captured != nil {
			if e := captured.ForEach(func(k, v []byte) error {
				return clear.Put(k, v)
			}); e != nil {
				return e
			}
		}
		return tx.DeleteBucket(captureBucketName)
	}); e == nil {
		s.empty = false
	}
	return e
}

func (s *Snapshot) LoadNode(ctx context.Context, path string, leaf ...bool) (node *tree.Node, err error) {
	err = s.db.View(func(tx *bbolt.Tx) error {
		if b := tx.Bucket(bucketName); b != nil {
			value := b.Get([]byte(path))
			if value != nil {
				if node, err = s.unmarshal(value); err != nil {
					return err
				}
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	} else if node == nil {
		err = errors.NotFound("not.found", "node not found in snapshot %s", path)
	}
	return
}

func (s *Snapshot) GetEndpointInfo() model.EndpointInfo {
	return model.EndpointInfo{
		URI: "snapshot://" + s.name,
		RequiresNormalization: false,
		RequiresFoldersRescan: false,
	}
}

func (s *Snapshot) Walk(walknFc model.WalkNodesFunc, root string) (err error) {
	root = strings.Trim(root, "/") + "/"
	err = s.db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket(bucketName)
		if b == nil {
			return nil
		}
		return b.ForEach(func(k, v []byte) error {
			key := string(k)
			if root != "/" && !strings.HasPrefix(key, root) {
				return nil
			}
			if node, e := s.unmarshal(v); e == nil {
				walknFc(key, node, nil)
			}
			return nil
		})
	})
	return err
}

func (s *Snapshot) Watch(recursivePath string, connectionInfo chan model.WatchConnectionInfo) (*model.WatchObject, error) {
	return nil, fmt.Errorf("not.implemented")
}

func (s *Snapshot) ComputeChecksum(node *tree.Node) error {
	return fmt.Errorf("not.implemented")
}

func (s *Snapshot) marshal(node *tree.Node) []byte {
	store := node.Clone()
	store.MetaStore = nil
	data, _ := proto.Marshal(node)
	return data
}

func (s *Snapshot) unmarshal(value []byte) (*tree.Node, error) {
	var n tree.Node
	if e := proto.Unmarshal(value, &n); e != nil {
		return nil, e
	} else {
		return &n, nil
	}
}

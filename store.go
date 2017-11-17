package main

import (
	"encoding/json"
	"strings"
	"time"

	"github.com/boltdb/bolt"
)

type Repo struct {
	ImportRoot string `json:"import_root"`
	VCS        string `json:"vcs"`
	VCSRoot    string `json:"vcs_root"`
	Suffix     string `json:"suffix"`
}

func NewStore(file string) *Store { return &Store{dbfile: file} }

type Store struct {
	dbfile string
}

func (s *Store) Initialize(seed map[string]Repo) error {
	db, err := bolt.Open(s.dbfile, 0600, &bolt.Options{Timeout: 20 * time.Second})
	if err != nil {
		return err
	}
	defer db.Close()

	return db.Update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists([]byte("imports"))
		if err != nil {
			return err
		}
		for k, v := range seed {
			d := b.Get([]byte(k))
			if d != nil {
				continue
			}
			d, err = json.Marshal(v)
			if err != nil {
				return err
			}
			if err = b.Put([]byte(k), d); err != nil {
				return err
			}
		}
		return nil
	})
}

func (s *Store) Get(host string, path string) (Repo, error) {
	repo := Repo{}
	db, err := bolt.Open(s.dbfile, 0600, &bolt.Options{Timeout: 20 * time.Second})
	if err != nil {
		return repo, err
	}
	defer db.Close()

	err = db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("imports"))
		for _, key := range getImports(host, path) {
			if v := b.Get([]byte(key)); v != nil {
				if err := json.Unmarshal(v, &repo); err != nil {
					return err
				}
				return nil
			}
		}
		return errNotFound
	})

	if err != nil {
		return repo, err
	}
	return repo, nil
}

func getImports(host string, path string) []string {
	var paths []string
	for segs := strings.Split(strings.TrimSuffix(path, "/"), "/"); len(segs) > 1; segs = segs[:len(segs)-1] {
		paths = append(paths, host+strings.Join(segs, "/"))
	}
	return paths
}

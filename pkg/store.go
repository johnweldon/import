package pkg

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"path"
	"strings"
	"time"

	"github.com/boltdb/bolt"
)

const (
	bucketName = "imports"
)

var (
	errConflict = errors.New("conflict")
	errNotFound = errors.New("not found")
	errNilValue = errors.New("nil value")
)

// Repo represents the mapping from import path to VCS
type Repo struct {
	ImportRoot string `json:"import_root"`
	VCS        string `json:"vcs"`
	VCSRoot    string `json:"vcs_root"`
	Suffix     string `json:"suffix"`
}

func (r *Repo) Valid() error {
	switch {
	case r == nil:
		return errNilValue
	case r.ImportRoot == "":
		return errors.New("missing import_root")
	case r.VCSRoot == "":
		return errors.New("missing vcs_root")
	case r.VCS == "":
		return errors.New("missing vcs")
	default:
		host, pth := split(r.ImportRoot)
		if len(strings.Split(host, ".")) < 2 {
			return fmt.Errorf("invalid host %q", host)
		}
		if len(pth) < 2 {
			return fmt.Errorf("invalid import path %q", pth)
		}
	}
	return nil
}

func (r *Repo) String() string {
	if err := r.Valid(); err != nil {
		return err.Error()
	}
	return fmt.Sprintf("Name: %-20s VCS: %s (%s)", r.ImportRoot, r.VCSRoot, r.VCS)
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
		b, err := tx.CreateBucketIfNotExists([]byte(bucketName))
		if err != nil {
			return err
		}
		for k, v := range seed {
			if err := v.Valid(); err != nil {
				return err
			}
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

func (s *Store) Create(r Repo) error {
	if err := r.Valid(); err != nil {
		return err
	}

	db, err := bolt.Open(s.dbfile, 0600, &bolt.Options{Timeout: 20 * time.Second})
	if err != nil {
		return err
	}
	defer db.Close()

	return db.Update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists([]byte(bucketName))
		if err != nil {
			return err
		}
		k := clean(r.ImportRoot)
		d := b.Get([]byte(k))
		if d != nil {
			return errConflict
		}
		d, err = json.Marshal(r)
		if err != nil {
			return err
		}
		if err = b.Put([]byte(k), d); err != nil {
			return err
		}
		return nil
	})
}

func (s *Store) Read(name string) (Repo, error) {
	repo := Repo{}
	db, err := bolt.Open(s.dbfile, 0600, &bolt.Options{Timeout: 20 * time.Second})
	if err != nil {
		return repo, err
	}
	defer db.Close()

	err = db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucketName))
		for _, key := range getImports(name) {
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

func (s *Store) Update(r Repo) error {
	if err := r.Valid(); err != nil {
		return err
	}

	db, err := bolt.Open(s.dbfile, 0600, &bolt.Options{Timeout: 20 * time.Second})
	if err != nil {
		return err
	}
	defer db.Close()

	return db.Update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists([]byte(bucketName))
		if err != nil {
			return err
		}
		k := clean(r.ImportRoot)
		d := b.Get([]byte(k))
		if d == nil {
			return errNotFound
		}
		d, err = json.Marshal(r)
		if err != nil {
			return err
		}
		if err = b.Put([]byte(k), d); err != nil {
			return err
		}
		return nil
	})
}

func (s *Store) List(prefix string) ([]Repo, error) {
	db, err := bolt.Open(s.dbfile, 0600, &bolt.Options{Timeout: 20 * time.Second})
	if err != nil {
		return nil, err
	}
	defer db.Close()

	var res []Repo
	err = db.View(func(tx *bolt.Tx) error {
		c := tx.Bucket([]byte(bucketName)).Cursor()
		pfx := []byte(prefix)
		for k, v := c.Seek(pfx); k != nil && bytes.HasPrefix(k, pfx); k, v = c.Next() {
			var repo Repo
			if err := json.Unmarshal(v, &repo); err != nil {
				log.Printf("ERROR: unmarshaling %q: %v", string(k), err)
			} else {
				res = append(res, repo)
			}
		}
		return nil
	})
	return res, err
}

func (s *Store) Delete(name string) error {
	log.Printf("DELETING: %q", name)
	db, err := bolt.Open(s.dbfile, 0600, &bolt.Options{Timeout: 20 * time.Second})
	if err != nil {
		return err
	}
	defer db.Close()

	return db.Update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists([]byte(bucketName))
		if err != nil {
			return err
		}
		k := clean(name)
		return b.Delete([]byte(k))
	})
}

func getImports(name string) []string {
	host, pth := split(name)
	var paths []string
	for segs := strings.Split(strings.TrimSuffix(pth, "/"), "/"); len(segs) > 1; segs = segs[:len(segs)-1] {
		paths = append(paths, host+strings.Join(segs, "/"))
	}
	return paths
}

func clean(name string) string {
	host, pth := split(name)
	return host + pth
}

func split(name string) (host string, pth string) {
	segs := strings.SplitN(name, "/", 2)
	if len(segs) > 0 {
		host = segs[0]
	}
	if len(segs) > 1 {
		pth = path.Clean("/" + segs[1])
	}
	return
}

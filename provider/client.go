package provider

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"path"
	"time"

	client "github.com/Virtomize/uii-go-api"
	"github.com/boltdb/bolt"
)

var (
	ErrBucketNotFound    = errors.New("bucket not found")
	ErrStoragePathNotSet = errors.New("storage path not set")
	ErrClientInit        = errors.New("client not initialised")
)

// IUiiClient is an interface for abstracting the interactions with the UII service - used for testing
type IUiiClient interface {
	Build(filePath string, args client.BuildArgs, opts client.BuildOpts) error
	OperatingSystems() ([]client.OS, error)
}

// ITimeProvider is an interface for injecting custom time providers - used for testing
type ITimeProvider interface {
	Now() time.Time
}

type clientWithStorage struct {
	VirtomizeClient IUiiClient
	StorageFolder   string
	TimeProvider    ITimeProvider
}

// defaultTimeProvider is an implementation of ITimeProvider using local time
type defaultTimeProvider struct {
}

// Now returns the current time
func (p defaultTimeProvider) Now() time.Time {
	return time.Now()
}

// CreateIso creates a new iso resource
func (s *clientWithStorage) CreateIso(iso Iso) (StoredIso, error) {
	if s.StorageFolder == "" {
		log.Fatal(ErrStoragePathNotSet)
		return StoredIso{}, ErrStoragePathNotSet
	}

	db, err := setupDB(path.Join(s.StorageFolder, "my.db"))
	if err != nil {
		log.Fatal(err)
		return StoredIso{}, err
	}
	defer db.Close()

	localPath, err := s.createIsoFileWithUii(iso)
	if err != nil {
		return StoredIso{}, err
	}

	creationTime := time.Now()
	if s.TimeProvider != nil {
		creationTime = s.TimeProvider.Now()
	}

	id, err := addIso(db, StoredIso{
		ID:           iso.Name,
		Iso:          iso,
		LocalPath:    localPath,
		CreationTime: creationTime,
	})
	if err != nil {
		return StoredIso{}, err
	}

	return readIso(db, id)
}

// ReadIso reads a ISO resource
func (s *clientWithStorage) ReadIso(isoID string) (StoredIso, error) {
	db, err := setupDB(path.Join(s.StorageFolder, "my.db"))
	if err != nil {
		log.Fatal(err)
		return StoredIso{}, err
	}
	defer db.Close()

	iso, err := readIso(db, isoID)
	if err != nil {
		return StoredIso{}, err
	}

	if s.isExpired(iso) {
		err = s.refreshIso(isoID, db)
		if err != nil {
			return StoredIso{}, err
		}
	}

	return iso, err
}

func (s *clientWithStorage) ReadDistributions() ([]client.OS, error) {
	if s.VirtomizeClient != nil {
		return s.VirtomizeClient.OperatingSystems()
	}

	return nil, ErrClientInit
}

// DeleteIso reads a ISO resource
func (s *clientWithStorage) DeleteIso(isoID string) error {
	db, err := setupDB(path.Join(s.StorageFolder, "my.db"))
	if err != nil {
		log.Panic(err)
		return err
	}
	defer db.Close()

	oldIso, err := readIso(db, isoID)
	if err != nil {
		log.Panic(err)
		return err
	}
	_ = os.Remove(oldIso.LocalPath)
	return deleteIso(db, isoID)
}

// UpdateIso updates a ISO resource
func (s *clientWithStorage) UpdateIso(id string, iso Iso) error {
	db, err := setupDB(path.Join(s.StorageFolder, "my.db"))
	if err != nil {
		log.Panic(err)
		return err
	}
	defer db.Close()

	oldIso, err := readIso(db, id)
	if err != nil {
		log.Panic(err)
		return err
	}

	if requiresNewIsoFile(iso, oldIso) {
		// refresh iso and re-read, as path potentially updated
		err = s.refreshIso(id, db)
		if err != nil {
			log.Panic(err)
			return err
		}

		oldIso, err = readIso(db, id)
		if err != nil {
			log.Panic(err)
			return err
		}
	}

	return updateIso(db, id, StoredIso{id, iso, oldIso.LocalPath, time.Now()})
}

func setupDB(dbPath string) (*bolt.DB, error) {
	db, err := bolt.Open(dbPath, 0600, nil)
	if err != nil {
		return nil, fmt.Errorf("could not open db, %w", err)
	}
	err = db.Update(func(tx *bolt.Tx) error {
		root, err := tx.CreateBucketIfNotExists([]byte("DB"))
		if err != nil {
			return fmt.Errorf("could not create root bucket: %w", err)
		}
		_, err = root.CreateBucketIfNotExists([]byte("ISOS"))
		if err != nil {
			return fmt.Errorf("could not create weight bucket: %w", err)
		}

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("could not set up buckets, %w", err)
	}

	return db, nil
}

func addIso(db *bolt.DB, iso StoredIso) (string, error) {
	err := updateIso(db, iso.Name, iso)
	return iso.Name, err
}

func updateIso(db *bolt.DB, isoKey string, iso StoredIso) error {
	entryBytes, err := json.Marshal(StoredIso{
		ID:           isoKey,
		Iso:          iso.Iso,
		LocalPath:    iso.LocalPath,
		CreationTime: iso.CreationTime,
	})
	if err != nil {
		return fmt.Errorf("could marshal iso: %w", err)
	}

	err = db.Update(func(tx *bolt.Tx) error {
		err := tx.Bucket([]byte("DB")).Bucket([]byte("ISOS")).Put([]byte(isoKey), entryBytes)
		if err != nil {
			return fmt.Errorf("could not insert iso: %w", err)
		}
		return nil
	})

	return err
}

func readIso(db *bolt.DB, isoKey string) (isoData StoredIso, err error) {
	err = db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("DB")).Bucket([]byte("ISOS"))
		if b == nil {
			return ErrBucketNotFound
		}
		rawData := b.Get([]byte(isoKey))
		marshalErr := json.Unmarshal(rawData, &isoData)
		return marshalErr
	})
	return isoData, err
}

func deleteIso(db *bolt.DB, isoKey string) error {
	return db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("DB")).Bucket([]byte("ISOS"))
		if b == nil {
			return ErrBucketNotFound
		}
		return b.Delete([]byte(isoKey))
	})
}

func requiresNewIsoFile(_ Iso, _ StoredIso) bool {
	return true
}

func (s *clientWithStorage) createIsoFileWithUii(iso Iso) (string, error) {
	networks := []client.NetworkArgs{}
	for _, net := range iso.Networks {
		networks = append(networks, client.NetworkArgs{
			DHCP:       net.DHCP,
			Domain:     net.Domain,
			MAC:        net.MAC,
			IPNet:      net.IPNet,
			Gateway:    net.Gateway,
			DNS:        net.DNS,
			NoInternet: net.NoInternet,
		})
	}

	localPath := path.Join(s.StorageFolder, iso.Name+".iso")
	err := s.VirtomizeClient.Build(localPath, client.BuildArgs{
		Distribution: iso.Distribution,
		Version:      iso.Version,
		Hostname:     iso.HostName,
		Networks:     networks,
	}, client.BuildOpts{
		Locale:          iso.Optionals.Locale,
		Keyboard:        iso.Optionals.Keyboard,
		Password:        iso.Optionals.Password,
		SSHPasswordAuth: iso.Optionals.SSHPasswordAuth,
		SSHKeys:         iso.Optionals.SSHKeys,
		Timezone:        iso.Optionals.Timezone,
		Arch:            iso.Optionals.Arch,
		Packages:        iso.Optionals.Packages,
	})
	return localPath, err
}

// refreshIso recreates an Iso by reading the data from the db and requesting a new iso file from UII
func (s *clientWithStorage) refreshIso(isoID string, db *bolt.DB) error {
	iso, err := readIso(db, isoID)
	if err != nil {
		return err
	}

	// remove old file and update to new local path - just in case the path changes
	_ = os.Remove(iso.LocalPath)
	localPath, err := s.createIsoFileWithUii(iso.Iso)

	if err != nil {
		log.Fatal(err)
		return err
	}

	return updateIso(db, isoID, StoredIso{isoID, iso.Iso, localPath, s.TimeProvider.Now()})
}

func (s *clientWithStorage) isExpired(iso StoredIso) bool {
	return iso.CreationTime.Before(s.TimeProvider.Now().Add(-48 * time.Hour))
}

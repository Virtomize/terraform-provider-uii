package main

import (
	"encoding/json"
	"fmt"
	client "github.com/Virtomize/uii-go-api"
	"github.com/boltdb/bolt"
	"log"
	"os"
)

type clientWithStorage struct {
	VirtomizeClient *client.UIIClient
	StorageFolder   string
}

func (s *clientWithStorage) CreateIso(iso Iso) (StoredIso, error) {
	db, err := setupDB(s.StorageFolder + "my.db")
	if err != nil {
		log.Fatal(err)
		return StoredIso{}, err
	}
	defer db.Close()

	var networks []client.NetworkArgs
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

	localPath := "C:/Tools/Terraform/Isos/" + iso.Name + ".iso"
	err = s.VirtomizeClient.Build(localPath, client.BuildArgs{
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
	if err != nil {
		return StoredIso{}, err
	}

	id, err := addIso(db, StoredIso{
		Id:        iso.Name,
		Iso:       iso,
		LocalPath: localPath,
	})
	if err != nil {
		return StoredIso{}, err
	}

	return readIso(db, id)
}

func (s *clientWithStorage) ReadIso(isoId string) (StoredIso, error) {
	db, err := setupDB(s.StorageFolder + "my.db")
	if err != nil {
		log.Fatal(err)
		return StoredIso{}, err
	}
	defer db.Close()

	iso, err := readIso(db, isoId)
	if err != nil {
		return StoredIso{}, err
	}

	return iso, err
}

func (s *clientWithStorage) DeleteIso(isoId string) error {
	db, err := setupDB(s.StorageFolder + "my.db")
	if err != nil {
		log.Fatal(err)
		return err
	}
	defer db.Close()

	oldIso, err := readIso(db, isoId)
	if err != nil {
		log.Fatal(err)
		return err
	}
	_ = os.Remove(oldIso.LocalPath)
	return deleteIso(db, isoId)
}

func (s *clientWithStorage) UpdateIso(id string, iso Iso) error {
	db, err := setupDB(s.StorageFolder + "my.db")
	if err != nil {
		log.Fatal(err)
		return err
	}
	defer db.Close()

	// todo: generate new ISO file

	oldIso, err := readIso(db, id)
	if err != nil {
		log.Fatal(err)
		return err
	}
	return updateIso(db, id, StoredIso{id, iso, oldIso.LocalPath})
}

func setupDB(dbPath string) (*bolt.DB, error) {
	db, err := bolt.Open(dbPath, 0600, nil)
	if err != nil {
		return nil, fmt.Errorf("could not open db, %v", err)
	}
	err = db.Update(func(tx *bolt.Tx) error {
		root, err := tx.CreateBucketIfNotExists([]byte("DB"))
		if err != nil {
			return fmt.Errorf("could not create root bucket: %v", err)
		}
		_, err = root.CreateBucketIfNotExists([]byte("ISOS"))
		if err != nil {
			return fmt.Errorf("could not create weight bucket: %v", err)
		}

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("could not set up buckets, %v", err)
	}

	return db, nil
}

func addIso(db *bolt.DB, iso StoredIso) (string, error) {
	err := updateIso(db, iso.Name, iso)
	return iso.Name, err
}

func updateIso(db *bolt.DB, isoKey string, iso StoredIso) error {
	entryBytes, err := json.Marshal(StoredIso{
		Id:        isoKey,
		Iso:       iso.Iso,
		LocalPath: iso.LocalPath,
	})
	if err != nil {
		return fmt.Errorf("could marshal iso: %v", err)
	}

	err = db.Update(func(tx *bolt.Tx) error {
		err := tx.Bucket([]byte("DB")).Bucket([]byte("ISOS")).Put([]byte(isoKey), entryBytes)
		if err != nil {
			return fmt.Errorf("could not insert iso: %v", err)
		}
		return nil
	})

	return err
}

func readIso(db *bolt.DB, isoKey string) (isoData StoredIso, err error) {
	err = db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("DB")).Bucket([]byte("ISOS"))
		if b == nil {
			return fmt.Errorf("bucket not found")
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
			return fmt.Errorf("bucket not found")
		}
		return b.Delete([]byte(isoKey))
	})
}

package main

import (
	"bytes"
	"fmt"
	stdioutil "io/ioutil"
	"log"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/goccy/go-yaml"
	"github.com/gosimple/slug"
	"github.com/jinzhu/gorm"
	"github.com/k0kubun/pp"
	"github.com/karrick/godirwalk"
	"github.com/onionltd/oniontree-tools/pkg/types/service"

	"github.com/qorpress/qorpress/pkg/config/db"
	"github.com/qorpress/qorpress-contrib/oniontree/models"
)

var (
	truncate bool
	displayHelp     bool
	dirname string
	debugMode  = true
	debugMode2 = true
	isTruncate = true
	DB         *gorm.DB
	tables     = []interface{}{
		&models.Tag{},
		&models.Service{},
		&models.PublicKey{},
		&models.URL{},
	}
)

func main() {

	pflag.StringVarP(&dirname, "dirname", "d", "./data/tagged", "directory with onion yaml files.")
	pflag.BoolVarP(&truncate, "truncate", "t", false, "truncate tables")
	pflag.BoolVarP(&displayHelp, "help", "h", false, "help info")
	pflag.Parse()
	if displayHelp {
		pflag.PrintDefaults()
		os.Exit(1)
	}

	DB = db.DB

	if truncate {
		TruncateTables(tables...)
	}

	// getWorkTree(db)
	// os.Exit(1)
	dirWalkServices(DB, dirname)
}

func dirWalkServices(DB *gorm.DB, dirname string) {
	err := godirwalk.Walk(dirname, &godirwalk.Options{
		Callback: func(osPathname string, de *godirwalk.Dirent) error {
			if !de.IsDir() {
				parts := strings.Split(osPathname, "/")
				if debugMode {
					fmt.Printf("Type:%s osPathname:%s tag:%s\n", de.ModeType(), osPathname, parts[1])
				}
				bytes, err := stdioutil.ReadFile(osPathname)
				if err != nil {
					return err
				}
				t := service.Service{}
				yaml.Unmarshal(bytes, &t)
				if debugMode {
					pp.Println(t)
				}

				// add service
				m := &Service{
					Name:        t.Name,
					Description: t.Description,
					Slug:        slug.Make(t.Name),
				}

				if err := DB.Create(m).Error; err != nil {
					fmt.Println(err)
					os.Exit(1)
				}

				// add public keys
				for _, publicKey := range t.PublicKeys {
					pubKey := &PublicKey{
						UID:         publicKey.ID,
						UserID:      publicKey.UserID,
						Fingerprint: publicKey.Fingerprint,
						Description: publicKey.Description,
						Value:       publicKey.Value,
					}
					if _, err := createOrUpdatePublicKey(DB, m, pubKey); err != nil {
						fmt.Println(err)
						os.Exit(1)
					}
				}

				// add urls
				for _, url := range t.URLs {
					var urlExists URL
					u := &URL{Name: url}
					if DB.Where("name = ?", url).First(&urlExists).RecordNotFound() {
						DB.Create(&u)
						if debugMode {
							pp.Println(u)
						}
					}
					if _, err := createOrUpdateURL(DB, m, u); err != nil {
						fmt.Println(err)
						os.Exit(1)
					}

				}

				// add tags
				// check if tag already exists
				tag := &Tag{Name: parts[1]}
				var tagExists Tag
				if DB.Where("name = ?", parts[1]).First(&tagExists).RecordNotFound() {
					DB.Create(&tag)
					if debugMode {
						pp.Println(tag)
					}
				}

				if _, err := createOrUpdateTag(DB, m, tag); err != nil {
					fmt.Println(err)
					os.Exit(1)
				}

			}
			return nil
		},
		Unsorted: true, // (optional) set true for faster yet non-deterministic enumeration (see godoc)
	})
	if err != nil {
		log.Fatal(err)
	}
}

func createOrUpdateTag(DB *gorm.DB, svc *Service, tag *Tag) (bool, error) {
	var existingSvc Service
	if DB.Where("slug = ?", svc.Slug).First(&existingSvc).RecordNotFound() {
		err := DB.Create(svc).Error
		return err == nil, err
	}
	var existingTag Tag
	if DB.Where("name = ?", tag.Name).First(&existingTag).RecordNotFound() {
		err := DB.Create(tag).Error
		return err == nil, err
	}
	svc.ID = existingSvc.ID
	svc.CreatedAt = existingSvc.CreatedAt
	svc.Tags = append(svc.Tags, &existingTag)
	return false, DB.Save(svc).Error
}

func findPublicKeyByUID(DB *gorm.DB, uid string) *PublicKey {
	pubKey := &PublicKey{}
	if err := DB.Where(&PublicKey{UID: uid}).First(pubKey).Error; err != nil {
		log.Fatalf("can't find public_key with uid = %q, got err %v", uid, err)
	}
	return pubKey
}

func createOrUpdatePublicKey(DB *gorm.DB, svc *Service, pubKey *PublicKey) (bool, error) {
	var existingSvc Service
	if DB.Where("slug = ?", svc.Slug).First(&existingSvc).RecordNotFound() {
		err := DB.Create(svc).Error
		return err == nil, err
	}
	var existingPublicKey PublicKey
	if DB.Where("uid = ?", pubKey.UID).First(&existingPublicKey).RecordNotFound() {
		err := DB.Create(pubKey).Error
		return err == nil, err
	}
	svc.ID = existingSvc.ID
	svc.CreatedAt = existingSvc.CreatedAt
	svc.PublicKeys = append(svc.PublicKeys, &existingPublicKey)
	return false, DB.Save(svc).Error
}

func createOrUpdateURL(DB *gorm.DB, svc *Service, url *URL) (bool, error) {
	var existingSvc Service
	if DB.Where("slug = ?", svc.Slug).First(&existingSvc).RecordNotFound() {
		err := DB.Create(svc).Error
		return err == nil, err
	}
	var existingURL URL
	if DB.Where("name = ?", url.Name).First(&existingURL).RecordNotFound() {
		err := DB.Create(url).Error
		return err == nil, err
	}
	svc.ID = existingSvc.ID
	svc.CreatedAt = existingSvc.CreatedAt
	svc.URLs = append(svc.URLs, &existingURL)
	return false, DB.Save(svc).Error
}

func truncateTables(DB *gorm.DB, tables ...interface{}) {
	for _, table := range tables {
		if debugMode {
			pp.Println(table)
		}
		if err := DB.DropTableIfExists(table).Error; err != nil {
			panic(err)
		}
		DB.AutoMigrate(table)
	}
}

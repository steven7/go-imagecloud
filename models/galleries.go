package models

import "github.com/jinzhu/gorm"


// Gallery represents the galleries table in our DB
// and is mostly a container resource composed of images.
type Gallery struct {
	gorm.Model
	UserID uint     `gorm:"not_null;index"`
	Title  string   `gorm:"not_null"`
	Images []Image  `gorm:"-"`
}

func NewGalleryService(db *gorm.DB) GalleryService {
	return &galleryService{
		GalleryDB: &galleryValidator{
			GalleryDB: &galleryGorm{
				db: db,
			},
		},
	}
}

type GalleryService interface {
	GalleryDB
}

type galleryService struct {
	GalleryDB
}

// OAuthDB is used to interact with the galleries database.
//
// For pretty much all single gallery queries:
// If the gallery is found, we will return a nil error
// If the gallery is not found, we will return ErrNotFound
// If there is another error, we will return an error with
// more information about what went wrong. This may not be
// an error generated by the models package.
type GalleryDB interface {
	ByID(id uint) (*Gallery, error)
	ByUserID(userID uint) ([]Gallery, error)
	Create(gallery *Gallery) error
	Update(gallery *Gallery) error
	Delete(id uint) error
}

type galleryValidator struct {
	GalleryDB
}


var _ GalleryDB = &galleryGorm{}

type galleryGorm struct {
	db *gorm.DB
}

func (gg *galleryGorm) ByID(id uint) (*Gallery, error) {
	var gallery Gallery
	db := gg.db.Where("id = ?", id)
	err := first(db, &gallery)
	if err != nil {
		return nil, err
	}
	return &gallery, nil
}

func (gg *galleryGorm) ByUserID(userID uint) ([]Gallery, error) {
	var galleries []Gallery
	// We build this query *exactly* the same way we build
	// a query for a single user
	db := gg.db.Where("user_id = ?", userID)
	// The real difference is in using Find instead of First
	// and passing in a slice instead of a single gallery as
	// the argument
	if err := db.Find(&galleries).Error; err != nil {
		return nil, err
	}
	return galleries, nil
}

func (gg *galleryGorm) Create(gallery *Gallery) error {
	return gg.db.Create(gallery).Error
}

func (gg *galleryGorm) Update(gallery *Gallery) error {
	return gg.db.Save(gallery).Error
}

func (gg *galleryGorm) Delete(id uint) error {
	gallery := Gallery{Model: gorm.Model{ID: id}}
	return gg.db.Delete(&gallery).Error
}


func (gv *galleryValidator) Create(gallery *Gallery) error {
	err := runGalleryValFns(gallery,
		gv.userIDRequired,
		gv.titleRequired)
	if err != nil {
		return err
	}
	return gv.GalleryDB.Create(gallery)
}

func (gv *galleryValidator) userIDRequired(g *Gallery) error {
	if g.UserID <= 0 {
		return ErrUserIDRequired
	}
	return nil
}

func (gv *galleryValidator) titleRequired(g *Gallery) error {
	if g.Title == "" {
		return ErrTitleRequired
	}
	return nil
}

func (gv *galleryValidator) Update(gallery *Gallery) error {
	err := runGalleryValFns(gallery,
		gv.userIDRequired,
		gv.titleRequired)
	if err != nil {
		return err
	}
	return gv.GalleryDB.Update(gallery)
}

func (gv *galleryValidator) nonZeroID(gallery *Gallery) error {
	if gallery.ID <= 0 {
		return ErrIDInvalid
	}
	return nil
}

func (gv *galleryValidator) Delete(id uint) error {
	var gallery Gallery
	gallery.ID = id
	if err := runGalleryValFns(&gallery, gv.nonZeroID); err != nil {
		return err
	}
	return gv.GalleryDB.Delete(gallery.ID)
}


type galleryValFn func(*Gallery) error

func runGalleryValFns(gallery *Gallery, fns ...galleryValFn) error {
	for _, fn := range fns {
		if err := fn(gallery); err != nil {
			return err
		}
	}
	return nil
}

func (g *Gallery) ImagesSplitN(n int) [][]Image {
	ret := make([][]Image, n)
	for i := 0; i < n; i++ {
		ret[i] = make([]Image, 0)
	}
	for i, img := range g.Images {
		bucket := i % n
		ret[bucket] = append(ret[bucket], img)
	}
	return ret
}

/*
func (g *Gallery) ImagesSplitN(n int) [][]string {
	// Create out 2D slice
	ret := make([][]string, n)
	// Create the inner slices - we need N of them, and we will
	// start them with a size of 0.
	for i := 0; i < n; i++ {
		ret[i] = make([]string, 0)
	}
	// Iterate over our images, using the index % n to determine
	// which of the slices in ret to add the image to.
	for i, img := range g.Images {
		// % is the remainder operator in Go
		// eg:
		//    0%3 = 0
		//    1%3 = 1
		//    2%3 = 2
		//    3%3 = 0
		//    4%3 = 1
		//    5%3 = 2
		bucket := i % n
		ret[bucket] = append(ret[bucket], img)
	}
	return ret
}
*/

/*
package models

import (
	"github.com/jinzhu/gorm"
)

// Gallery represents the galleries table in our DB
// and is mostly a container resource composed of images.
type Gallery struct {
	gorm.Model
	UserID uint   `gorm:"not_null;index"`
	Title  string `gorm:"not_null"`
}

func NewGalleryService(db *gorm.DB) GalleryService {
	return &galleryService{
		OAuthDB: &galleryValidator{
			OAuthDB: &galleryGorm{
				db: db,
			},
		},
	}
}

type GalleryService interface {
	OAuthDB
}

type galleryService struct {
	OAuthDB
}

// OAuthDB is used to interact with the galleries database.
//
// For pretty much all single gallery queries:
// If the gallery is found, we will return a nil error
// If the gallery is not found, we will return ErrNotFound
// If there is another error, we will return an error with
// more information about what went wrong. This may not be
// an error generated by the models package.
type OAuthDB interface {
	ByID(id uint) (*Gallery, error)
	Create(gallery *Gallery) error
}

type galleryValidator struct {
	OAuthDB
}

var _ OAuthDB = &galleryGorm{}

type galleryGorm struct {
	db *gorm.DB
}

func (gg *galleryGorm) Create(gallery *Gallery) error {
	return gg.db.Create(gallery).Error
}

func (gg *galleryGorm) ByID(id uint) (*Gallery, error) {
	var gallery Gallery
	db := gg.db.Where("id = ?", id)
	err := first(db, &gallery)
	if err != nil {
		return nil, err
	}
	return &gallery, nil
}

type galleryValFn func(*Gallery) error

func runGalleryValFns(gallery *Gallery, fns ...galleryValFn) error {
	for _, fn := range fns {
		if err := fn(gallery); err != nil {
			return err
		}
	}
	return nil
}

const (
	ErrUserIDRequired modelError = "models: user ID is required"
)

func (gv *galleryValidator) userIDRequired(g *Gallery) error {
	if g.UserID <= 0 {
		return ErrUserIDRequired
	}
	return nil
}

const (
	ErrTitleRequired modelError = "models: title is required"
)

func (gv *galleryValidator) titleRequired(g *Gallery) error {
	if g.Title == "" {
		return ErrTitleRequired
	}
	return nil
}

func (gv *galleryValidator) Create(gallery *Gallery) error {
	err := runGalleryValFns(gallery,
		gv.userIDRequired,
		gv.titleRequired)
	if err != nil {
		return err
	}
	return gv.OAuthDB.Create(gallery)
}
*/
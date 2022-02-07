package bot

import (
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"

	"stargazer/SLIIT-Notifications/helpers"
	"stargazer/SLIIT-Notifications/keyreader"

	"github.com/PuerkitoBio/goquery"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type SLIITSyncable struct {
	site   string
	title  string
	user   *SLIITUser
	client *http.Client
	id     primitive.ObjectID
	db     *mongo.Database
}

type Section struct {
	Hash    string `bson:"h" json:"h,omitempty"`
	Section string `bson:"s" json:"s,omitempty"`
}

type SyncResult struct {
	h *SLIITHistory
	e error
}

var ErrLogin = errors.New("login error")

func NewSLIITSyncable(title string, user *SLIITUser, site string, client *http.Client, db *mongo.Database, id primitive.ObjectID) *SLIITSyncable {
	return &SLIITSyncable{
		site:   site,
		title:  title,
		user:   user,
		client: client,
		id:     id,
		db:     db,
	}
}

func (s *SLIITSyncable) Sync(ctx context.Context) (*SLIITHistory, error) {
	log.Println(s.site)
	result := make(chan SyncResult)

	go func() {
		doc, err := helpers.DoGetGoQuery(s.client, s.site)

		if err != nil {
			result <- SyncResult{nil, err}
			return
		}

		ok := assertLogin(doc, s.user.Username)

		if !ok {
			result <- SyncResult{nil, ErrLogin}
			return
		}

		// Get from database
		var old_history SLIITHistory
		k := keyreader.NewReader(old_history, "bson")

		findOpts := options.FindOne()
		findOpts.SetSort(bson.M{"_id": -1})

		findOpts.SetProjection(bson.M{
			k.Get("ID"):       1,
			k.Get("SiteID"):   1,
			k.Get("Sections"): 1,
		})

		filter := bson.M{
			k.Get("SiteID"): s.id,
		}

		res := s.db.Collection("history").FindOne(ctx, filter, findOpts)

		if err := res.Decode(&old_history); err != nil {
			if err == mongo.ErrNoDocuments {
				log.Printf("History dosen't exist. Adding one for %s\n", s.title)
				var sections []Section

				doc.Find(".section.main").Each(func(i int, sect *goquery.Selection) {
					sect_name := sect.AttrOr("id", "unknown")

					if sect_name == "unknown" {
						return
					}

					t_hash := sha256.New()
					if _, wErr := t_hash.Write([]byte(sect.Text())); wErr != nil {
						log.Panic(wErr)
					}

					h := fmt.Sprintf("%x", t_hash.Sum(nil))

					sections = append(sections, Section{
						Hash:    h,
						Section: sect_name,
					})
				})

				source, souErr := doc.Html()

				if souErr != nil {
					result <- SyncResult{nil, souErr}
					return
				}

				compressed, cErr := helpers.CompressString(source)

				if cErr != nil {
					result <- SyncResult{nil, cErr}
					return
				}

				_id := primitive.NewObjectID()

				new_history := SLIITHistory{
					ID:        _id,
					SiteID:    s.id,
					Sections:  sections,
					HTML:      compressed,
					AddedTime: time.Now(),
					LastID:    _id,
				}

				if _, iErr := s.db.Collection("history").InsertOne(ctx, new_history); iErr != nil {
					result <- SyncResult{nil, iErr}
					return
				}

				log.Printf("Completed %s", s.id)

				result <- SyncResult{nil, nil}
				return
			} else {
				result <- SyncResult{nil, err}
				return
			}
		}
		// Compare the stuff
		log.Printf("Comparing pages %s\n", s.title)

		section_map := make(map[string]*Section)
		sections := old_history.Sections
		changed_sections := 0

		for i := 0; i < len(sections); i++ {
			section_map[sections[i].Section] = &sections[i]
		}

		doc.Find(".section.main").Each(func(i int, sect *goquery.Selection) {
			sect_name := sect.AttrOr("id", "unknown")

			if sect_name == "unknown" {
				return
			}

			t_hash := sha256.New()
			if _, wErr := t_hash.Write([]byte(sect.Text())); wErr != nil {
				log.Panic(wErr)
			}

			h := fmt.Sprintf("%x", t_hash.Sum(nil))

			if old, ok := section_map[sect_name]; ok {
				if strings.Compare(h, old.Hash) != 0 {
					log.Printf("%s from %s Changed.", s.id, sect_name)

					changed_sections++
					old.Hash = h
				}
			} else {
				sections = append(sections, Section{Hash: h, Section: sect_name})
			}
		})

		if changed_sections > 0 {
			source, souErr := doc.Html()

			if souErr != nil {
				result <- SyncResult{nil, souErr}
				return
			}

			compressed, cErr := helpers.CompressString(source)

			if cErr != nil {
				result <- SyncResult{nil, cErr}
				return
			}

			new_history := SLIITHistory{
				ID:        primitive.NewObjectID(),
				SiteID:    s.id,
				Sections:  sections,
				HTML:      compressed,
				AddedTime: time.Now(),
				LastID:    old_history.ID,
			}

			if _, iErr := s.db.Collection("history").InsertOne(ctx, new_history); iErr != nil {
				result <- SyncResult{nil, iErr}
				return
			}

			log.Println("Completed")
			result <- SyncResult{&new_history, nil}
			return
		}
		result <- SyncResult{nil, nil}
	}()

	select {
	case r := <-result:
		return r.h, r.e
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

func (s *SLIITSyncable) Login(ctx context.Context) error {
	fErr := make(chan error)

	go func() {
		resp, err := s.client.PostForm(
			"https://courseweb.sliit.lk/login/index.php",
			url.Values{
				"username": {s.user.Username},
				"password": {s.user.Password},
			},
		)

		if err != nil {
			fErr <- err
			return
		}

		defer resp.Body.Close()

		doc, qerr := goquery.NewDocumentFromReader(resp.Body)

		if qerr != nil {
			fErr <- qerr
			return
		}

		logged := assertLogin(doc, s.user.Username)

		if !logged {
			fErr <- ErrLogin
			return
		}
		fErr <- nil
	}()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case e := <-fErr:
		return e
	}
}

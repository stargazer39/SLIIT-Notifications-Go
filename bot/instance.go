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
	Hash    string `bson:"h,omitempty" json:"h,omitempty"`
	Section string `bson:"s,omitempty" json:"s,omitempty"`
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

func (s *SLIITSyncable) Sync() (*SLIITHistory, error) {
	log.Println(s.site)
	doc, err := helpers.DoGetGoQuery(s.client, s.site)

	if err != nil {
		return nil, err
	}

	ok := assertLogin(doc, s.user.Username)

	if !ok {
		return nil, ErrLogin
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

	res := s.db.Collection("history").FindOne(context.TODO(), SLIITHistory{SiteID: s.id}, findOpts)

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
				return nil, souErr
			}

			compressed, cErr := helpers.CompressString(source)

			if cErr != nil {
				return nil, cErr
			}

			new_history := SLIITHistory{
				ID:        primitive.NewObjectID(),
				SiteID:    s.id,
				Sections:  sections,
				HTML:      compressed,
				AddedTime: time.Now(),
			}

			if _, iErr := s.db.Collection("history").InsertOne(context.TODO(), new_history); iErr != nil {
				return nil, iErr
			}

			log.Printf("Completed %s", s.id)

			return nil, nil
		} else {
			return nil, err
		}
	}
	// Compare the stuff
	log.Printf("Comparing pages %s\n", s.title)

	section_map := make(map[string]*Section)
	sections := old_history.Sections
	changed_sections := []string{}

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

				changed_sections = append(changed_sections, sect_name)
				old.Hash = h
			}
		}
	})

	if len(changed_sections) > 0 {
		source, souErr := doc.Html()

		if souErr != nil {
			return nil, souErr
		}

		compressed, cErr := helpers.CompressString(source)

		if cErr != nil {
			return nil, cErr
		}

		new_history := SLIITHistory{
			ID:        primitive.NewObjectID(),
			SiteID:    s.id,
			Sections:  sections,
			HTML:      compressed,
			AddedTime: time.Now(),
		}

		if _, iErr := s.db.Collection("history").InsertOne(context.TODO(), new_history); iErr != nil {
			return nil, iErr
		}

		log.Println("Completed")
		return &new_history, nil
	}

	return nil, nil
}

func (s *SLIITSyncable) Login() error {
	resp, err := s.client.PostForm(
		"https://courseweb.sliit.lk/login/index.php",
		url.Values{
			"username": {s.user.Username},
			"password": {s.user.Password},
		},
	)

	if err != nil {
		return err
	}

	defer resp.Body.Close()

	doc, qerr := goquery.NewDocumentFromReader(resp.Body)

	if qerr != nil {
		return qerr
	}

	logged := assertLogin(doc, s.user.Username)

	if !logged {
		return ErrLogin
	}
	return nil
}

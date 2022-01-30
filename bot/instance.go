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

	"stargazer/SLIIT-Notifications/helpers"

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

var LoginErr = errors.New("login error")

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

func (s *SLIITSyncable) Sync() error {
	log.Println(s.site)
	doc, err := helpers.DoGetGoQuery(s.client, s.site)

	if err != nil {
		return err
	}

	ok, err := assertLogin(doc, s.user.Username)
	if err != nil {
		return err
	}

	if !ok {
		return LoginErr
	}
	// os.WriteFile(fmt.Sprintf(".cache/%s.txt", u), []byte(doc.Text()), 0644)
	// Get from database
	var old_history SLIITHistory

	findOpts := options.FindOne()
	findOpts.SetSort(bson.M{"_id": 1})

	res := s.db.Collection("history").FindOne(context.TODO(), SLIITHistory{SiteID: s.id}, findOpts)

	if err := res.Decode(&old_history); err != nil {
		if err == mongo.ErrNoDocuments {
			log.Printf("History dosen't exist. Adding one for %s\n", s.title)
			var sections []Section

			doc.Find(".section.main").Each(func(i int, sect *goquery.Selection) {
				t_hash := sha256.New()
				t_hash.Write([]byte(sect.Text()))
				sections = append(sections, Section{
					Hash:    fmt.Sprintf("%x", t_hash.Sum(nil)),
					Section: sect.AttrOr("id", "unknown"),
				})
			})

			new_history := SLIITHistory{
				ID:       primitive.NewObjectID(),
				SiteID:   s.id,
				Sections: sections,
			}

			if _, iErr := s.db.Collection("history").InsertOne(context.TODO(), new_history); iErr != nil {
				return iErr
			}
			log.Println("Completed")
			log.Println(s.id)
			return nil
		} else {
			return err
		}
	}

	// Compare the stuff
	log.Printf("Comparing pages %s\n", s.title)

	section_map := make(map[string]string)
	old_sections := old_history.Sections

	for i := 0; i < len(old_sections); i++ {
		section_map[old_sections[i].Section] = old_sections[i].Hash
	}

	doc.Find(".section.main").Each(func(i int, sect *goquery.Selection) {
		t_hash := sha256.New()
		t_hash.Write([]byte(sect.Text()))

		h := fmt.Sprintf("%x", t_hash.Sum(nil))
		id := sect.AttrOr("id", "unknown")

		if lHash, ok := section_map[id]; ok {
			if strings.Compare(h, lHash) != 0 {
				log.Printf("%s from %s Changed.", s.id, id)
			}
		}
	})
	// fileName := fmt.Sprintf(".cache/%s.json", url.PathEscape(s.title))

	/* if _, err := os.Stat(fileName); err == nil {
		objects := []Section{}
		objmap := make(map[string]string)

		f, fErr := ioutil.ReadFile(fileName)

		if fErr != nil {
			return fErr
		}

		if err := json.Unmarshal(f, &objects); err != nil {
			return err
		}

		for _, v := range objects {
			objmap[v.Section] = v.Hash
		}

		doc.Find(".section.main").Each(func(i int, sect *goquery.Selection) {
			t_hash := sha256.New()
			t_hash.Write([]byte(sect.Text()))

			h := fmt.Sprintf("%x", t_hash.Sum(nil))
			id := sect.AttrOr("id", "unknown")

			if lHash, ok := objmap[id]; ok {
				if strings.Compare(h, lHash) != 0 {
					log.Printf("%s from %s Changed.", s.id, id)
				}
			}

		})
		// log.Println(objects)
	} */

	/* f, fErr := os.Create(fileName)

	if fErr != nil {
		return fErr
	}

	defer f.Close()

	hashes := []Section{}

	doc.Find(".section.main").Each(func(i int, sect *goquery.Selection) {
		t_hash := sha256.New()
		t_hash.Write([]byte(sect.Text()))
		hashes = append(hashes, Section{
			Hash:    fmt.Sprintf("%x", t_hash.Sum(nil)),
			Section: sect.AttrOr("id", "unknown"),
		})
	})

	jBytes, jErr := json.MarshalIndent(hashes, "", "   ")

	if jErr != nil {
		return jErr
	}

	f.Write(jBytes) */
	return nil
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

	logged, aerr := assertLogin(doc, s.user.Username)

	if aerr != nil {
		return aerr
	}

	if !logged {
		return LoginErr
	}
	return nil
}

/*
	pointer to SLIITUser
	cookiejar
*/

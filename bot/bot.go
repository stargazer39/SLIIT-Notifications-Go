package bot

import (
	"context"
	"log"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/PuerkitoBio/goquery"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type SLIITBot struct {
	db        *mongo.Database
	users     []SLIITUser
	exit_chan chan bool
}

type SLIITUser struct {
	ID       primitive.ObjectID `bson:"_id" json:"id,omitempty"`
	Username string             `bson:"username" json:"username,omitempty"`
	Password string             `bson:"password" json:"password,omitempty"`
}

type SLIITSite struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	Name      string             `bson:"name,omitempty" json:"name,omitempty"`
	UserID    primitive.ObjectID `bson:"uid,omitempty" json:"uid,omitempty"`
	URL       string             `bson:"url,omitempty" json:"url,omitempty"`
	AddedTime time.Time          `bson:"added_time,omitempty" json:"added_time,omitempty"`
}

type SLIITHistory struct {
	ID       primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	SiteID   primitive.ObjectID `bson:"sid,omitempty" json:"sid,omitempty"`
	Sections []Section          `bson:"sects,omitempty" json:"sects,omitempty"`
}

func NewBot(ctx context.Context, db *mongo.Database) *SLIITBot {
	// db.Collection("test").InsertOne(ctx, bson.M{
	// 	"testdoc": "hello_world",
	// })

	return &SLIITBot{
		db:        db,
		users:     []SLIITUser{},
		exit_chan: make(chan bool),
	}
}

func (s *SLIITBot) Start() error {
	var users []SLIITUser

	cur, curErr := s.db.Collection("user").Find(context.TODO(), bson.D{})

	if curErr != nil {
		return curErr
	}

	if err := cur.All(context.TODO(), &users); err != nil {
		return err
	}

	syncables, sErr := s.generateSyncables(users)

	length := len(syncables)
	done := make(chan bool, 8)

out:
	for {
		var wg sync.WaitGroup
		t1 := time.Now()

		for i := 0; i < length; i++ {
			s := &syncables[i]
			done <- true
			wg.Add(1)

			go func() {
				if err := s.Sync(); err != nil {
					if err == ErrLogin {
						rErr := s.Login()
						if rErr == ErrLogin {
							log.Println("ReLogin faild.", rErr)
							return
						}

						if rErr != nil {
							log.Println(rErr)
						}
						return
					}

					log.Println(err)
				}
				wg.Done()
				<-done
			}()
		}

		wg.Wait()
		log.Println(time.Since(t1))

		sleeper := time.NewTimer(time.Second * 5)
		//close(done)
		select {
		case <-s.exit_chan:
			break out
		case <-sleeper.C:
		}
	}

	return sErr
}

func (s *SLIITBot) generateSyncables(users []SLIITUser) ([]SLIITSyncable, error) {
	var syncable []SLIITSyncable

	for i := 0; i < len(users); i++ {
		currentUser := &users[i]

		jar, err := cookiejar.New(nil)

		if err != nil {
			log.Println(err)
			continue
		}

		var client = &http.Client{
			Jar:     jar,
			Timeout: 30 * time.Second,
		}

		resp, err := client.PostForm(
			"https://courseweb.sliit.lk/login/index.php",
			url.Values{
				"username": {currentUser.Username},
				"password": {currentUser.Password},
			},
		)

		if err != nil {
			log.Println(err)
			continue
		}

		defer resp.Body.Close()

		doc, qerr := goquery.NewDocumentFromReader(resp.Body)

		if qerr != nil {
			log.Println(qerr)
			continue
		}

		logged := assertLogin(doc, currentUser.Username)

		if !logged {
			log.Printf("Loggin faild for user - %s\n", currentUser.Username)
			continue
		}

		// Get already existing sites
		var sites []SLIITSite

		cur, curErr := s.db.Collection("sites").Find(context.TODO(), bson.D{})

		if curErr != nil {
			return nil, curErr
		}

		if err := cur.All(context.TODO(), &sites); err != nil {
			return nil, err
		}

		// Map them
		var site_map = make(map[string]*SLIITSite)

		for i := range sites {
			site_map[sites[i].Name] = &sites[i]
		}

		selec := doc.Find("a[title='My courses'] ~ ul > li a")

		selec.Each(func(i int, sel *goquery.Selection) {
			url, exists := sel.Attr("href")

			if exists {
				log.Println(url)
				name := sel.Text()
				site, ok := site_map[name]
				var mId primitive.ObjectID

				if !ok {
					res, inErr := s.db.Collection("sites").InsertOne(context.TODO(), SLIITSite{
						ID:        primitive.NewObjectID(),
						Name:      name,
						UserID:    currentUser.ID,
						URL:       url,
						AddedTime: time.Now(),
					})

					if inErr != nil {
						log.Println(inErr)
						return
					}

					mId = res.InsertedID.(primitive.ObjectID)
				} else {
					mId = site.ID
				}

				s := NewSLIITSyncable(name, currentUser, url, client, s.db, mId)
				syncable = append(syncable, *s)
			}
		})
	}
	return syncable, nil
}

func (s *SLIITBot) Stop() {
	s.exit_chan <- true
}

func assertLogin(doc *goquery.Document, username string) bool {
	selec := doc.Find("span.usertext")
	usertext := selec.Text()

	// log.Printf("Logged in as %s\n", usertext)
	logged := strings.Contains(strings.ToLower(usertext), strings.ToLower(username))

	return logged
}

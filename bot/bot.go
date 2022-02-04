package bot

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"stargazer/SLIIT-Notifications/keyreader"
	"strings"
	"sync"
	"time"

	"github.com/PuerkitoBio/goquery"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type SLIITBot struct {
	db               *mongo.Database
	users            []SLIITUser
	exit_chan        chan bool
	finish_chan      chan error
	stop             bool
	change_listeners []func(s SLIITHistory, uid primitive.ObjectID)
	interval         time.Duration
}

type SLIITUser struct {
	ID       primitive.ObjectID `bson:"_id" json:"id,omitempty"`
	Username string             `bson:"username" json:"username,omitempty"`
	Password string             `bson:"password" json:"password,omitempty"`
	DegreeID string             `bson:"degree_id" json:"degree_id,omitempty"`
	Disabled bool               `bson:"disabled" json:"disabled,omitempty"`
}

type SLIITSite struct {
	ID        primitive.ObjectID `bson:"_id" json:"id,omitempty"`
	Name      string             `bson:"name" json:"name,omitempty"`
	UserID    primitive.ObjectID `bson:"uid" json:"uid,omitempty"`
	URL       string             `bson:"url" json:"url,omitempty"`
	AddedTime time.Time          `bson:"added_time" json:"added_time,omitempty"`
	Disabled  bool               `bson:"disabled" json:"disabled,omitempty"`
}

type SLIITHistory struct {
	ID        primitive.ObjectID `bson:"_id" json:"id,omitempty"`
	SiteID    primitive.ObjectID `bson:"sid" json:"sid,omitempty"`
	Sections  []Section          `bson:"sects" json:"sects,omitempty"`
	HTML      string             `bson:"html" json:"html,omitempty"`
	AddedTime time.Time          `bson:"added_time" json:"added_time,omitempty"`
	LastID    primitive.ObjectID `bson:"last_id" json:"last_id,omitempty"`
}

var (
	SLIITUserK    = keyreader.NewReader(SLIITUser{}, "bson")
	SLIITSiteK    = keyreader.NewReader(SLIITSite{}, "bson")
	SLIITHistoryK = keyreader.NewReader(SLIITHistory{}, "bson")
)

func NewBot(db *mongo.Database, interval time.Duration) *SLIITBot {
	// db.Collection("test").InsertOne(ctx, bson.M{
	// 	"testdoc": "hello_world",
	// })

	return &SLIITBot{
		db:          db,
		users:       []SLIITUser{},
		interval:    interval,
		exit_chan:   make(chan bool),
		finish_chan: make(chan error),
		stop:        false,
	}
}

func (s *SLIITBot) Start(ctx context.Context) error {
	defer close(s.finish_chan)

	var users []SLIITUser

	cur, curErr := s.db.Collection("user").Find(context.TODO(), bson.D{})

	if curErr != nil {
		return fmt.Errorf("user collection read error: %w", curErr)
	}

	if err := cur.All(context.TODO(), &users); err != nil {
		return fmt.Errorf("user collection read error: %w", err)
	}

	syncables, sErr := s.generateSyncables(users)

	if sErr != nil {
		return fmt.Errorf("syncable generation error: %w", sErr)
	}

	length := len(syncables)
	done := make(chan bool, 8)

out:
	for {
		var wg sync.WaitGroup
		t1 := time.Now()

		for i := 0; i < length; i++ {
			if s.stop {
				break
			}

			syn := &syncables[i]
			done <- true
			wg.Add(1)

			go func() {
				history, sErr := syn.Sync()

				if sErr != nil {
					if sErr == ErrLogin {
						rErr := syn.Login()
						if rErr == ErrLogin {
							log.Println("ReLogin faild.", rErr)
							return
						}

						if rErr != nil {
							log.Println(rErr)
						}
						return
					}
					log.Println(sErr)
				}

				if history != nil {
					for j := 0; j < len(s.change_listeners); j++ {
						s.change_listeners[j](*history, syn.user.ID)
					}
				}

				wg.Done()
				<-done
			}()
		}

		wg.Wait()
		log.Println(time.Since(t1))

		sleeper := time.NewTimer(s.interval)
		//close(done)
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-sleeper.C:
		case <-s.exit_chan:
			break out
		}
	}

	return nil
}

func (s *SLIITBot) RegisterChangeListener(h func(his SLIITHistory, uid primitive.ObjectID)) {
	s.change_listeners = append(s.change_listeners, h)
}

func (s *SLIITBot) generateSyncables(users []SLIITUser) ([]SLIITSyncable, error) {
	var syncable []SLIITSyncable

	// Get already existing sites
	var sites []SLIITSite

	cur, curErr := s.db.Collection("sites").Find(context.TODO(), bson.D{})

	if curErr != nil {
		return nil, fmt.Errorf("generateSyncables error: %w", curErr)
	}

	if err := cur.All(context.TODO(), &sites); err != nil {
		return nil, fmt.Errorf("generateSyncables error: %w", err)
	}

	// Map them
	var site_map = make(map[string]*SLIITSite)

	for i := range sites {
		site_map[sites[i].Name] = &sites[i]
	}

	for i := 0; i < len(users); i++ {
		currentUser := &users[i]

		// Continue if disabled
		if currentUser.Disabled {
			continue
		}

		// Create http client
		jar, err := cookiejar.New(nil)

		if err != nil {
			log.Println(err)
			continue
		}

		var client = &http.Client{
			Jar:     jar,
			Timeout: 30 * time.Second,
		}

		// Try to login
		r_count := 0
	retry:
		doc, doErr := doPostFormGetDoc(client, currentUser)

		if doErr != nil {
			r_count++

			if r_count > 5 {
				return nil, fmt.Errorf("generateSyncables error: %w", doErr)
			}
			time.Sleep(time.Second * 10)
			goto retry
		}

		// Check if logged in
		logged := assertLogin(doc, currentUser.Username)

		if !logged {
			log.Printf("Loggin faild for user - %s\n", currentUser.Username)
			continue
		}

		selec := doc.Find("a[title='My courses'] ~ ul > li a")

		selec.Each(func(i int, sel *goquery.Selection) {
			url, exists := sel.Attr("href")

			if exists {
				log.Println(url)
				name := sel.Text()
				// Get the existing site
				site, ok := site_map[name]
				var mId primitive.ObjectID

				// If site dosen't exist add it
				if !ok {
					res, inErr := s.db.Collection("sites").InsertOne(context.TODO(), SLIITSite{
						ID:        primitive.NewObjectID(),
						Name:      name,
						UserID:    currentUser.ID,
						URL:       url,
						AddedTime: time.Now(),
						Disabled:  false,
					})

					if inErr != nil {
						log.Println(inErr)
						return
					}

					mId = res.InsertedID.(primitive.ObjectID)
				} else {
					mId = site.ID
					// If site disabled, return
					if site.Disabled {
						return
					}
				}
				// New syncable
				s := NewSLIITSyncable(name, currentUser, url, client, s.db, mId)
				syncable = append(syncable, *s)
			}
		})
	}
	return syncable, nil
}

func (s *SLIITBot) Stop(ctx context.Context) error {
	close(s.exit_chan)

	s.stop = true
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-s.finish_chan:
		return nil
	}
}

func assertLogin(doc *goquery.Document, username string) bool {
	selec := doc.Find("span.usertext")
	usertext := selec.Text()

	// log.Printf("Logged in as %s\n", usertext)
	logged := strings.Contains(strings.ToLower(usertext), strings.ToLower(username))

	return logged
}

func doPostFormGetDoc(client *http.Client, user *SLIITUser) (*goquery.Document, error) {
	resp, err := client.PostForm(
		"https://courseweb.sliit.lk/login/index.php",
		url.Values{
			"username": {user.Username},
			"password": {user.Password},
		},
	)

	if err != nil {
		return nil, fmt.Errorf("loginAndGetDoc: %w", err)
	}

	defer resp.Body.Close()

	doc, qErr := goquery.NewDocumentFromReader(resp.Body)

	if qErr != nil {
		return nil, fmt.Errorf("loginAndGetDoc: %w", qErr)
	}

	return doc, nil
}

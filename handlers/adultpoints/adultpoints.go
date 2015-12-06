package adultpoints

import (
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"github.com/sha1sum/golang_groupme_bot/bot"
	"fmt"
	"os"
	"time"
	"strings"
"math/rand"
	"strconv"
)

type Handler struct{}

type (
	user struct {
		ID       bson.ObjectId `bson:"_id"`
		UserID   string        `bson:"userID"`
		Name     string `bson:"name"`
		Points   int           `bson:"points"`
		Requests []request    `bson:"requests"`
	}

	request struct {
		Reference    string `bson:"reference"`
		RequestedOn  time.Time `bson:"requestedOn"`
		Approved     bool `bson:"approved"`
		ApprovedByID string `bson:"approvedByID"`
		ApprovedOn   time.Time `bson:"approvedOn"`
		Reason       string `bson:"reason"`
	}
)

var DB string

func (handler Handler) Handle(term string, c chan *bot.OutgoingMessage, message bot.IncomingMessage) {
	uri := os.Getenv("MONGOLAB_URI")
	if uri == "" {
		fmt.Println("no connection string provided")
		os.Exit(1)
	}
	DB = os.Getenv("MONGOLAB_DB")
	if uri == "" {
		fmt.Println("no database provided")
		os.Exit(1)
	}
	sess, err := mgo.Dial(uri)
	if err != nil {
		fmt.Printf("Can't connect to mongo, go error %v\n", err)
		os.Exit(1)
	}
	defer sess.Close()

	result := pointProcess(term, sess, message)
	if result != nil {
		c <- result
	}
}

func pointProcess(term string, sess *mgo.Session, message bot.IncomingMessage) *bot.OutgoingMessage {
	words := strings.Split(term, " ")
	switch strings.ToLower(words[0]) {
	case "adultme":
		return requestPoint(words[1:], sess, message)
	case "award":
		return awardPoint(words[1:], sess, message)
	case "adults":
		return listAdults(sess)
	default:
		return nil
	}
}

func requestPoint(words []string, sess *mgo.Session, message bot.IncomingMessage) *bot.OutgoingMessage {
	args := strings.Join(words, " ")
	col := sess.DB(DB).C("groupmeUsers")
	var cu user
	fmt.Println(message.UserID)
	err := col.Find(bson.M{"userID": message.UserID}).One(&cu)
	if err != nil {
		col.Insert(user{ID: bson.NewObjectId(), UserID: message.UserID, Name: message.Name, Points: 0})
	}
	_ = col.Find(bson.M{"userID": message.UserID}).One(&cu)
	requests := cu.Requests
	reference := strconv.Itoa(rand.Intn(500))
	found := true
	for found {
		notFound := col.Find(bson.M{"requests": bson.M{"$elemMatch": bson.M{"reference": reference}}})
		if notFound == nil {
			reference = strconv.Itoa(rand.Intn(500))
			continue
		}
		found = false
	}
	cu.Requests = append(requests, request{
		Reference: reference,
		RequestedOn: time.Now(),
		Approved: false,
		Reason: args,
	})
	col.Update(bson.M{"_id": cu.ID}, cu)
	t := message.Name + " has requested an adult point " + args + "."
	t += " To approve the point, just type \"!award " + reference + "\"."
	return &bot.OutgoingMessage{Message: t}
}

func awardPoint(words []string, sess *mgo.Session, message bot.IncomingMessage) *bot.OutgoingMessage {
	args := strings.Join(words, " ")
	col := sess.DB(DB).C("groupmeUsers")
	var cu user
	err := col.Find(bson.M{"requests": bson.M{"$elemMatch": bson.M{"reference": args, "approved": false}}}).One(&cu)
	if err != nil {
		return &bot.OutgoingMessage{Message: "Couldn't find an unapproved request with reference \"" + args + "\"."}
	}
	var ri int
	requests := cu.Requests
	for i, v := range requests {
		if v.Reference == args {
			ri = i
			break
		}
	}
	if cu.UserID == message.UserID {
		t := "Stop trying to be slick! You can't approve your own requests!"
		t += " Just for that, I'm revoking the request!"
		cu.Requests = append(requests[:ri], requests[ri+1:]...)
		col.Update(bson.M{"_id": cu.ID}, cu)
		return &bot.OutgoingMessage{Message: t}
	}
	requests[ri].Approved = true
	requests[ri].ApprovedOn = time.Now()
	requests[ri].ApprovedByID = message.UserID
	cu.Requests = requests
	cu.Points += 1
	col.Update(bson.M{"_id": cu.ID}, cu)
	return &bot.OutgoingMessage{
		Message: "Congratulations " + cu.Name + ", you've just been given an adult point " + requests[ri].Reason + "!",
	}
}

func listAdults(sess *mgo.Session) *bot.OutgoingMessage {
	var results []user
	_ = sess.DB(DB).C("groupmeUsers").Find(nil).Sort("-points").All(&results)
	board := ""
	total := 0
	for _, v := range results {
		board += v.Name + ": " + strconv.Itoa(v.Points) + "\n"
		total += v.Points
	}
	board += "\nTOTAL: " + strconv.Itoa(total)
	return &bot.OutgoingMessage{Message: board}
}
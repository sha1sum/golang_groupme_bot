package adultpoints

import (
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"github.com/sha1sum/golang_groupme_bot/bot"
	"fmt"
	"os"
	"time"
	"strings"
	"strconv"
)

type Handler struct{}

type (
	user struct {
		ID       bson.ObjectId `bson:"_id"`
		UserID   string        `bson:"userID"`
		Created time `bson:"created"`
		Name     string `bson:"name"`
		Points   int           `bson:"points"`
		Requests []request    `bson:"requests"`
	}

	request struct {
		Reference    string `bson:"reference"`
		RequestedOn  time.Time `bson:"requestedOn"`
		Approved     bool `bson:"approved"`
		Reason       string `bson:"reason"`
		Approvals	[]approval `bson:"approvals"`
		Rejections []rejection `bson:"rejections"`
	}

	approval struct {
		ApprovedByID string `bson:"approvedByID"`
		ApprovedOn   time.Time `bson:"approvedOn"`
	}

	rejection struct {
		RejectedByID string `bson:"rejectedByID"`
		RejectedOn time.Time `bson:"rejectedOn"`
	}
)

var DB string

func (handler Handler) Handle(term string, c chan []*bot.OutgoingMessage, message bot.IncomingMessage) {
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

	results := pointProcess(term, sess, message)
	if results != nil {
		c <- results
	}
}

func pointProcess(term string, sess *mgo.Session, message bot.IncomingMessage) []*bot.OutgoingMessage {
	words := strings.Split(term, " ")
	switch strings.ToLower(words[0]) {
	case "adultme":
		return requestPoint(words[1:], sess, message)
	case "award":
		return awardPoint(words[1:2], sess, message)
	case "reject":
		return rejectPoint(words[1:2], sess, message)
	case "adults":
		return listAdults(sess)
	default:
		return nil
	}
}

func requestPoint(words []string, sess *mgo.Session, message bot.IncomingMessage) []*bot.OutgoingMessage {
	args := strings.Join(words, " ")
	col := sess.DB(DB).C("groupmeUsers")
	var cu user
	fmt.Println(message.UserID)
	err := col.Find(bson.M{"userID": message.UserID}).One(&cu)
	if err != nil {
		col.Insert(user{ID: bson.NewObjectId(), UserID: message.UserID, Name: message.Name, Points: 0, Created: time.Now()})
	}
	_ = col.Find(bson.M{"userID": message.UserID}).One(&cu)
	requests := cu.Requests
	reference := determineReference(cu, sess)
	cu.Requests = append(requests, request{
		Reference: reference,
		RequestedOn: time.Now(),
		Approved: false,
		Reason: args,
	})
	col.Update(bson.M{"_id": cu.ID}, cu)
	t := message.Name + " has requested an adult point \"" + args + "\"."
	t += " To approve the point, just type \"!award " + reference + "\", or to reject it, use \"!reject " + reference + "\"."
	return []*bot.OutgoingMessage{&bot.OutgoingMessage{Message: t}}
}

func determineReference(cu user, sess *mgo.Session) string {
	var results []user
	_ = sess.DB(DB).C("groupmeUsers").Find(nil).Sort("created").All(&results)
	ui := 0
	for i, v := range results {
		if v.UserID == cu.UserID {
			ui = i
			break
		}
	}
	return strconv.Itoa(ui + 1) + strconv.Itoa(len(cu.Requests) + 1)
}

func awardPoint(words []string, sess *mgo.Session, message bot.IncomingMessage) []*bot.OutgoingMessage {
	args := strings.Join(words, " ")
	col := sess.DB(DB).C("groupmeUsers")
	var cu user
	err := col.Find(bson.M{"requests": bson.M{"$elemMatch": bson.M{"reference": args}}}).One(&cu)
	if err != nil {
		return []*bot.OutgoingMessage{&bot.OutgoingMessage{Message: "Couldn't find a request with reference \"" + args + "\"."}}
	}
	var ri int
	requests := cu.Requests
	for i, v := range requests {
		if v.Reference == args {
			ri = i
			break
		}
	}
	if cu.UserID == message.UserID || message.SenderType == "bot" {
		t := "Stop trying to be slick! You can't approve your own requests!"
		t += " Just for that, I'm revoking the request!"
		cu.Requests = append(requests[:ri], requests[ri+1:]...)
		col.Update(bson.M{"_id": cu.ID}, cu)
		return []*bot.OutgoingMessage{&bot.OutgoingMessage{Message: t}}
	}
	currentlyApproved := len(requests[ri].Approvals) >= len(requests[ri].Rejections) && len(requests[ri].Approvals) != 0
	for _, v := range requests[ri].Approvals {
		if v.ApprovedByID == message.UserID {
			return []*bot.OutgoingMessage{&bot.OutgoingMessage{Message: "You've already approved that request (dumbass)."}}
		}
	}
	for i, v := range requests[ri].Rejections {
		if v.RejectedByID == message.UserID {
			requests[ri].Rejections = append(requests[ri].Rejections[:i], requests[ri].Rejections[i+1:]...)
			addApproval(col, message.UserID, &cu, ri)
			return []*bot.OutgoingMessage{
				&bot.OutgoingMessage{Message: "Your previous rejection has been switched to an approval (make up your damn mind)."},
				announcePointChange(true, col, &cu, &requests[ri], currentlyApproved, message),
			}
		}
	}
	addApproval(col, message.UserID, &cu, ri)
	cu.Requests = requests
	return []*bot.OutgoingMessage{announcePointChange(true, col, &cu, &requests[ri], currentlyApproved, message)}
}

func addApproval(col *mgo.Collection, approvedByID string, cu *user, ri int) {
	cu.Requests[ri].Approvals = append(cu.Requests[ri].Approvals, approval{ApprovedByID: approvedByID, ApprovedOn: time.Now()})
	col.Update(bson.M{"_id": cu.ID}, cu)
}

func rejectPoint(words []string, sess *mgo.Session, message bot.IncomingMessage) []*bot.OutgoingMessage {
	args := strings.Join(words, " ")
	col := sess.DB(DB).C("groupmeUsers")
	var cu user
	err := col.Find(bson.M{"requests": bson.M{"$elemMatch": bson.M{"reference": args}}}).One(&cu)
	if err != nil {
		return []*bot.OutgoingMessage{&bot.OutgoingMessage{Message: "Couldn't find a request with reference \"" + args + "\"."}}
	}
	var ri int
	requests := cu.Requests
	for i, v := range requests {
		if v.Reference == args {
			ri = i
			break
		}
	}
	currentlyApproved := len(requests[ri].Approvals) >= len(requests[ri].Rejections)
	if cu.UserID == message.UserID || message.SenderType == "bot" {
		t := "Uhhh, okay. If you really want to reject your own request, whatever. Wish granted."
		addRejection(col, message.UserID, &cu, ri)
		return []*bot.OutgoingMessage{
			&bot.OutgoingMessage{Message: t},
			announcePointChange(false, col, &cu, &requests[ri], currentlyApproved, message),
		}
	}
	for _, v := range requests[ri].Rejections {
		if v.RejectedByID == message.UserID {
			return []*bot.OutgoingMessage{&bot.OutgoingMessage{Message: "You've already rejected that request (dumbass)."}}
		}
	}
	for i, v := range requests[ri].Approvals {
		if v.ApprovedByID == message.UserID {
			requests[ri].Approvals = append(requests[ri].Approvals[:i], requests[ri].Approvals[i+1:]...)
			addRejection(col, message.UserID, &cu, ri)
			return []*bot.OutgoingMessage{
				&bot.OutgoingMessage{Message: "Your previous approval has been switched to a rejection (make up your damn mind)."},
				announcePointChange(false, col, &cu, &requests[ri], currentlyApproved, message),
			}
		}
	}
	addRejection(col, message.UserID, &cu, ri)
	cu.Requests = requests
	return []*bot.OutgoingMessage{announcePointChange(false, col, &cu, &requests[ri], currentlyApproved, message)}
}

func addRejection(col *mgo.Collection, rejectedByID string, cu *user, ri int) {
	cu.Requests[ri].Rejections = append(cu.Requests[ri].Rejections, rejection{RejectedByID: rejectedByID, RejectedOn: time.Now()})
	col.Update(bson.M{"_id": cu.ID}, cu)
}

func announcePointChange(approving bool, col *mgo.Collection, cu *user, req *request, currentlyApproved bool, message bot.IncomingMessage) *bot.OutgoingMessage {
	switch {
	case len(req.Rejections) == 0 && len(req.Approvals) == 1:
		cu.Points += 1
		col.Update(bson.M{"_id": cu.ID}, cu)
		return &bot.OutgoingMessage{Message: cu.Name + ", you just got your approval \"" + req.Reason + "\"! You've been given a point (for now)."}
	case len(req.Approvals) >= len(req.Rejections) && currentlyApproved && !approving:
		col.Update(bson.M{"_id": cu.ID}, cu)
		return &bot.OutgoingMessage{Message: cu.Name + ", don't worry about that hater " + message.Name + ". You still have your point \"" + req.Reason + "\"."}
	case len(req.Approvals) >= len(req.Rejections) && currentlyApproved:
		return &bot.OutgoingMessage{Message: cu.Name + " is stackin' up approvals \"" + req.Reason + "\"!"}
	case len(req.Approvals) >= len(req.Rejections) && !currentlyApproved:
		cu.Points += 1
		col.Update(bson.M{"_id": cu.ID}, cu)
		return &bot.OutgoingMessage{Message: "Things are lookin' up, " + cu.Name + "! You just got a point \"" + req.Reason + "\"."}
	case currentlyApproved:
		cu.Points -= 1
		col.Update(bson.M{"_id": cu.ID}, cu)
		return &bot.OutgoingMessage{Message: message.Name + " doesn't seem to believe you deserve your point, " + cu.Name + "... Your point \"" + req.Reason + "\" just got revoked :("}
	default:
		return &bot.OutgoingMessage{Message: "The prospects for you getting a point \"" + req.Reason + "\" aren't looking great, " + cu.Name + "... You just got another rejection from " + message.Name}
	}
}

func listAdults(sess *mgo.Session) []*bot.OutgoingMessage {
	var results []user
	_ = sess.DB(DB).C("groupmeUsers").Find(nil).Sort("-points").All(&results)
	board := ""
	total := 0
	for _, v := range results {
		board += v.Name + ": " + strconv.Itoa(v.Points) + "\n"
		total += v.Points
	}
	board += "\nTOTAL: " + strconv.Itoa(total)
	return []*bot.OutgoingMessage{&bot.OutgoingMessage{Message: board}}
}
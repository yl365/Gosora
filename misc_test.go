package main

import (
	"bytes"
	"database/sql"
	"fmt"
	"io/ioutil"
	"net/http/httptest"
	"os"
	"runtime/debug"
	"strconv"
	"strings"
	"testing"
	"time"

	c "github.com/Azareal/Gosora/common"
	"github.com/Azareal/Gosora/common/gauth"
	"github.com/Azareal/Gosora/common/phrases"
	"github.com/pkg/errors"
)

func miscinit(t *testing.T) {
	if err := gloinit(); err != nil {
		t.Fatal(err)
	}
}

func recordMustExist(t *testing.T, err error, errmsg string, args ...interface{}) {
	if err == ErrNoRows {
		debug.PrintStack()
		t.Errorf(errmsg, args...)
	} else if err != nil {
		debug.PrintStack()
		t.Fatal(err)
	}
}

func recordMustNotExist(t *testing.T, err error, errmsg string, args ...interface{}) {
	if err == nil {
		debug.PrintStack()
		t.Errorf(errmsg, args...)
	} else if err != ErrNoRows {
		debug.PrintStack()
		t.Fatal(err)
	}
}

func TestUserStore(t *testing.T) {
	miscinit(t)
	if !c.PluginsInited {
		c.InitPlugins()
	}

	var err error
	uc := c.NewMemoryUserCache(c.Config.UserCacheCapacity)
	c.Users, err = c.NewDefaultUserStore(uc)
	expectNilErr(t, err)
	uc.Flush()
	userStoreTest(t, 2)
	c.Users, err = c.NewDefaultUserStore(nil)
	expectNilErr(t, err)
	userStoreTest(t, 5)
}
func userStoreTest(t *testing.T, newUserID int) {
	uc := c.Users.GetCache()
	// Go doesn't have short-circuiting, so this'll allow us to do one liner tests
	cacheLength := func(uc c.UserCache) int {
		if uc == nil {
			return 0
		}
		return uc.Length()
	}
	isCacheLengthZero := func(uc c.UserCache) bool {
		return cacheLength(uc) == 0
	}
	expectf(t, isCacheLengthZero(uc), "The initial ucache length should be zero, not %d", cacheLength(uc))

	_, err := c.Users.Get(-1)
	recordMustNotExist(t, err, "UID #-1 shouldn't exist")
	expectf(t, isCacheLengthZero(uc), "We found %d items in the user cache and it's supposed to be empty", cacheLength(uc))

	_, err = c.Users.Get(0)
	recordMustNotExist(t, err, "UID #0 shouldn't exist")
	expectf(t, isCacheLengthZero(uc), "We found %d items in the user cache and it's supposed to be empty", cacheLength(uc))

	user, err := c.Users.Get(1)
	recordMustExist(t, err, "Couldn't find UID #1")

	expectW := func(cond, expec bool, prefix, suffix string) {
		midfix := "should not be"
		if expec {
			midfix = "should be"
		}
		expect(t, cond, prefix+" "+midfix+" "+suffix)
	}

	// TODO: Add email checks too? Do them separately?
	expectUser := func(u *c.User, uid int, name string, group int, super, admin, mod, banned bool) {
		expectf(t, u.ID == uid, "u.ID should be %d. Got '%d' instead.", uid, u.ID)
		expectf(t, u.Name == name, "u.Name should be '%s', not '%s'", name, u.Name)
		expectW(u.Group == group, true, u.Name, "in group"+strconv.Itoa(group))
		expectW(u.IsSuperAdmin == super, super, u.Name, "a super admin")
		expectW(u.IsAdmin == admin, admin, u.Name, "an admin")
		expectW(u.IsSuperMod == mod, mod, u.Name, "a super mod")
		expectW(u.IsMod == mod, mod, u.Name, "a mod")
		expectW(u.IsBanned == banned, banned, u.Name, "banned")
	}
	expectUser(user, 1, "Admin", 1, true, true, true, false)

	user, err = c.Users.GetByName("Admin")
	recordMustExist(t, err, "Couldn't find user 'Admin'")
	expectUser(user, 1, "Admin", 1, true, true, true, false)
	us, err := c.Users.BulkGetByName([]string{"Admin"})
	recordMustExist(t, err, "Couldn't find user 'Admin'")
	expectf(t, len(us) == 1, "len(us) should be 1, not %d", len(us))
	expectUser(us[0], 1, "Admin", 1, true, true, true, false)

	_, err = c.Users.Get(newUserID)
	recordMustNotExist(t, err, fmt.Sprintf("UID #%d shouldn't exist", newUserID))

	// TODO: GetByName tests for newUserID

	if uc != nil {
		expectIntToBeX(t, uc.Length(), 1, "User cache length should be 1, not %d")
		_, err = uc.Get(-1)
		recordMustNotExist(t, err, "UID #-1 shouldn't exist, even in the cache")
		_, err = uc.Get(0)
		recordMustNotExist(t, err, "UID #0 shouldn't exist, even in the cache")
		user, err = uc.Get(1)
		recordMustExist(t, err, "Couldn't find UID #1 in the cache")

		expectf(t, user.ID == 1, "user.ID does not match the requested UID. Got '%d' instead.", user.ID)
		expectf(t, user.Name == "Admin", "user.Name should be 'Admin', not '%s'", user.Name)

		_, err = uc.Get(newUserID)
		recordMustNotExist(t, err, "UID #%d shouldn't exist, even in the cache", newUserID)
		uc.Flush()
		expectIntToBeX(t, uc.Length(), 0, "User cache length should be 0, not %d")
	}

	// TODO: Lock onto the specific error type. Is this even possible without sacrificing the detailed information in the error message?
	var userList map[int]*c.User
	userList, _ = c.Users.BulkGetMap([]int{-1})
	expectf(t, len(userList) == 0, "The userList length should be 0, not %d", len(userList))
	expectf(t, isCacheLengthZero(uc), "User cache length should be 0, not %d", cacheLength(uc))

	userList, _ = c.Users.BulkGetMap([]int{0})
	expectf(t, len(userList) == 0, "The userList length should be 0, not %d", len(userList))
	expectf(t, isCacheLengthZero(uc), "User cache length should be 0, not %d", cacheLength(uc))

	userList, _ = c.Users.BulkGetMap([]int{1})
	expectf(t, len(userList) == 1, "Returned map should have one result (UID #1), not %d", len(userList))

	user, ok := userList[1]
	if !ok {
		t.Error("We couldn't find UID #1 in the returned map")
		t.Error("userList", userList)
		return
	}
	expectf(t, user.ID == 1, "user.ID does not match the requested UID. Got '%d' instead.", user.ID)

	if uc != nil {
		expectIntToBeX(t, uc.Length(), 1, "User cache length should be 1, not %d")
		user, err = uc.Get(1)
		recordMustExist(t, err, "Couldn't find UID #1 in the cache")

		expectf(t, user.ID == 1, "user.ID does not match the requested UID. Got '%d' instead.", user.ID)
		uc.Flush()
	}

	expect(t, !c.Users.Exists(-1), "UID #-1 shouldn't exist")
	expect(t, !c.Users.Exists(0), "UID #0 shouldn't exist")
	expect(t, c.Users.Exists(1), "UID #1 should exist")
	expectf(t, !c.Users.Exists(newUserID), "UID #%d shouldn't exist", newUserID)

	expectf(t, isCacheLengthZero(uc), "User cache length should be 0, not %d", cacheLength(uc))
	expectIntToBeX(t, c.Users.Count(), 1, "The number of users should be one, not %d")

	awaitingActivation := 5
	// TODO: Write tests for the registration validators
	uid, err := c.Users.Create("Sam", "ReallyBadPassword", "sam@localhost.loc", awaitingActivation, false)
	expectNilErr(t, err)
	expectf(t, uid == newUserID, "The UID of the new user should be %d not %d", newUserID, uid)
	expectf(t, c.Users.Exists(newUserID), "UID #%d should exist", newUserID)

	user, err = c.Users.Get(newUserID)
	recordMustExist(t, err, "Couldn't find UID #%d", newUserID)
	expectUser(user, newUserID, "Sam", 5, false, false, false, false)

	if uc != nil {
		expectIntToBeX(t, uc.Length(), 1, "User cache length should be 1, not %d")
		user, err = uc.Get(newUserID)
		recordMustExist(t, err, "Couldn't find UID #%d in the cache", newUserID)
		expectf(t, user.ID == newUserID, "user.ID does not match the requested UID. Got '%d' instead.", user.ID)
	}

	userList, _ = c.Users.BulkGetMap([]int{1, uid})
	expectf(t, len(userList) == 2, "Returned map should have two results, not %d", len(userList))

	if uc != nil {
		expectIntToBeX(t, uc.Length(), 2, "User cache length should be 2, not %d")
		user, err = uc.Get(1)
		recordMustExist(t, err, "Couldn't find UID #%d in the cache", 1)
		expectf(t, user.ID == 1, "user.ID does not match the requested UID. Got '%d' instead.", user.ID)
		user, err = uc.Get(newUserID)
		recordMustExist(t, err, "Couldn't find UID #%d in the cache", newUserID)
		expectf(t, user.ID == newUserID, "user.ID does not match the requested UID. Got '%d' instead.", user.ID)
		uc.Flush()
	}

	user, err = c.Users.Get(newUserID)
	recordMustExist(t, err, "Couldn't find UID #%d", newUserID)
	expectUser(user, newUserID, "Sam", 5, false, false, false, false)

	if uc != nil {
		expectIntToBeX(t, uc.Length(), 1, "User cache length should be 1, not %d")
		user, err = uc.Get(newUserID)
		recordMustExist(t, err, "Couldn't find UID #%d in the cache", newUserID)
		expectf(t, user.ID == newUserID, "user.ID does not match the requested UID. Got '%d' instead.", user.ID)
	}

	err = user.Activate()
	expectNilErr(t, err)
	expectIntToBeX(t, user.Group, 5, "Sam should still be in group 5 in this copy")

	// ? - What if we change the caching mechanism so it isn't hard purged and reloaded? We'll deal with that when we come to it, but for now, this is a sign of a cache bug
	afterUserFlush := func(uid int) {
		if uc != nil {
			expectIntToBeX(t, uc.Length(), 0, "User cache length should be 0, not %d")
			_, err = uc.Get(uid)
			recordMustNotExist(t, err, "UID #%d shouldn't be in the cache", uid)
		}
	}
	afterUserFlush(newUserID)

	user, err = c.Users.Get(newUserID)
	recordMustExist(t, err, "Couldn't find UID #%d", newUserID)
	expectUser(user, newUserID, "Sam", c.Config.DefaultGroup, false, false, false, false)

	// Permanent ban
	duration, _ := time.ParseDuration("0")

	// TODO: Attempt a double ban, double activation, and double unban
	err = user.Ban(duration, 1)
	expectNilErr(t, err)
	expectf(t, user.Group == c.Config.DefaultGroup, "Sam should be in group %d, not %d", c.Config.DefaultGroup, user.Group)
	afterUserFlush(newUserID)

	user, err = c.Users.Get(newUserID)
	recordMustExist(t, err, "Couldn't find UID #%d", newUserID)
	expectUser(user, newUserID, "Sam", c.BanGroup, false, false, false, true)

	// TODO: Do tests against the scheduled updates table and the task system to make sure the ban exists there and gets revoked when it should

	err = user.Unban()
	expectNilErr(t, err)
	expectIntToBeX(t, user.Group, c.BanGroup, "Sam should still be in the ban group in this copy")
	afterUserFlush(newUserID)

	user, err = c.Users.Get(newUserID)
	recordMustExist(t, err, "Couldn't find UID #%d", newUserID)
	expectUser(user, newUserID, "Sam", c.Config.DefaultGroup, false, false, false, false)

	reportsForumID := 1 // TODO: Use the constant in common?
	generalForumID := 2
	dummyResponseRecorder := httptest.NewRecorder()
	bytesBuffer := bytes.NewBuffer([]byte(""))
	dummyRequest1 := httptest.NewRequest("", "/forum/"+strconv.Itoa(reportsForumID), bytesBuffer)
	dummyRequest2 := httptest.NewRequest("", "/forum/"+strconv.Itoa(generalForumID), bytesBuffer)
	var user2 *c.User

	changeGroupTest := func(oldGroup, newGroup int) {
		err = user.ChangeGroup(newGroup)
		expectNilErr(t, err)
		// ! I don't think ChangeGroup should be changing the value of user... Investigate this.
		expect(t, oldGroup == user.Group, "Someone's mutated this pointer elsewhere")

		user, err = c.Users.Get(newUserID)
		recordMustExist(t, err, "Couldn't find UID #%d", newUserID)
		user2 = c.BlankUser()
		*user2 = *user
	}

	changeGroupTest2 := func(rank string, firstShouldBe, secondShouldBe bool) {
		head, err := c.UserCheck(dummyResponseRecorder, dummyRequest1, user)
		if err != nil {
			t.Fatal(err)
		}
		head2, err := c.UserCheck(dummyResponseRecorder, dummyRequest2, user2)
		if err != nil {
			t.Fatal(err)
		}
		ferr := c.ForumUserCheck(head, dummyResponseRecorder, dummyRequest1, user, reportsForumID)
		expect(t, ferr == nil, "There shouldn't be any errors in forumUserCheck")
		expect(t, user.Perms.ViewTopic == firstShouldBe, rank+" should be able to access the reports forum")
		ferr = c.ForumUserCheck(head2, dummyResponseRecorder, dummyRequest2, user2, generalForumID)
		expect(t, ferr == nil, "There shouldn't be any errors in forumUserCheck")
		expect(t, user2.Perms.ViewTopic == secondShouldBe, "Sam should be able to access the general forum")
	}

	changeGroupTest(c.Config.DefaultGroup, 1)
	expectUser(user, newUserID, "Sam", 1, false, true, true, false)
	changeGroupTest2("Admins", true, true)

	changeGroupTest(1, 2)
	expectUser(user, newUserID, "Sam", 2, false, false, true, false)
	changeGroupTest2("Mods", true, true)

	changeGroupTest(2, 3)
	expectUser(user, newUserID, "Sam", 3, false, false, false, false)
	changeGroupTest2("Members", false, true)
	expect(t, user.Perms.ViewTopic != user2.Perms.ViewTopic, "user.Perms.ViewTopic and user2.Perms.ViewTopic should never match")

	changeGroupTest(3, 4)
	expectUser(user, newUserID, "Sam", 4, false, false, false, true)
	changeGroupTest2("Members", false, true)

	changeGroupTest(4, 5)
	expectUser(user, newUserID, "Sam", 5, false, false, false, false)
	changeGroupTest2("Members", false, true)

	changeGroupTest(5, 6)
	expectUser(user, newUserID, "Sam", 6, false, false, false, false)
	changeGroupTest2("Members", false, true)

	err = user.ChangeGroup(c.Config.DefaultGroup)
	expectNilErr(t, err)
	expect(t, user.Group == 6, "Someone's mutated this pointer elsewhere")

	err = user.Delete()
	expectNilErr(t, err)
	expectf(t, !c.Users.Exists(newUserID), "UID #%d should no longer exist", newUserID)
	afterUserFlush(newUserID)

	_, err = c.Users.Get(newUserID)
	recordMustNotExist(t, err, "UID #%d shouldn't exist", newUserID)

	// And a unicode test, even though I doubt it'll fail
	uid, err = c.Users.Create("サム", "😀😀😀", "sam@localhost.loc", awaitingActivation, false)
	expectNilErr(t, err)
	expectf(t, uid == newUserID+1, "The UID of the new user should be %d", newUserID+1)
	expectf(t, c.Users.Exists(newUserID+1), "UID #%d should exist", newUserID+1)

	user, err = c.Users.Get(newUserID + 1)
	recordMustExist(t, err, "Couldn't find UID #%d", newUserID+1)
	expectUser(user, newUserID+1, "サム", 5, false, false, false, false)

	err = user.Delete()
	expectNilErr(t, err)
	expectf(t, !c.Users.Exists(newUserID+1), "UID #%d should no longer exist", newUserID+1)

	// MySQL utf8mb4 username test
	uid, err = c.Users.Create("😀😀😀", "😀😀😀", "sam@localhost.loc", awaitingActivation, false)
	expectNilErr(t, err)
	expectf(t, uid == newUserID+2, "The UID of the new user should be %d", newUserID+2)
	expectf(t, c.Users.Exists(newUserID+2), "UID #%d should exist", newUserID+2)

	user, err = c.Users.Get(newUserID + 2)
	recordMustExist(t, err, "Couldn't find UID #%d", newUserID+1)
	expectUser(user, newUserID+2, "😀😀😀", 5, false, false, false, false)

	err = user.Delete()
	expectNilErr(t, err)
	expectf(t, !c.Users.Exists(newUserID+2), "UID #%d should no longer exist", newUserID+2)

	// TODO: Add unicode login tests somewhere? Probably with the rest of the auth tests
	// TODO: Add tests for the Cache* methods
}

// TODO: Add an error message to this?
func expectNilErr(t *testing.T, item error) {
	if item != nil {
		debug.PrintStack()
		t.Fatal(item)
	}
}

func expectIntToBeX(t *testing.T, item, expect int, errmsg string) {
	if item != expect {
		debug.PrintStack()
		t.Fatalf(errmsg, item)
	}
}

func expect(t *testing.T, item bool, errmsg string) {
	if !item {
		debug.PrintStack()
		t.Fatal(errmsg)
	}
}

func expectf(t *testing.T, item bool, errmsg string, args ...interface{}) {
	if !item {
		debug.PrintStack()
		t.Fatalf(errmsg, args...)
	}
}

func TestPermsMiddleware(t *testing.T) {
	miscinit(t)
	if !c.PluginsInited {
		c.InitPlugins()
	}

	dummyResponseRecorder := httptest.NewRecorder()
	bytesBuffer := bytes.NewBuffer([]byte(""))
	dummyRequest := httptest.NewRequest("", "/forum/1", bytesBuffer)
	user := c.BlankUser()

	ferr := c.SuperModOnly(dummyResponseRecorder, dummyRequest, user)
	expect(t, ferr != nil, "Blank users shouldn't be supermods")

	user.IsSuperMod = false
	ferr = c.SuperModOnly(dummyResponseRecorder, dummyRequest, user)
	expect(t, ferr != nil, "Non-supermods shouldn't be allowed through supermod gates")

	user.IsSuperMod = true
	ferr = c.SuperModOnly(dummyResponseRecorder, dummyRequest, user)
	expect(t, ferr == nil, "Supermods should be allowed through supermod gates")

	// TODO: Loop over the Control Panel routes and make sure only supermods can get in

	user = c.BlankUser()

	ferr = c.MemberOnly(dummyResponseRecorder, dummyRequest, user)
	expect(t, ferr != nil, "Blank users shouldn't be considered loggedin")

	user.Loggedin = false
	ferr = c.MemberOnly(dummyResponseRecorder, dummyRequest, user)
	expect(t, ferr != nil, "Guests shouldn't be able to access member areas")

	user.Loggedin = true
	ferr = c.MemberOnly(dummyResponseRecorder, dummyRequest, user)
	expect(t, ferr == nil, "Logged in users should be able to access member areas")

	// TODO: Loop over the /user/ routes and make sure only members can access the ones other than /user/username

	// TODO: Write tests for AdminOnly()

	user = c.BlankUser()

	ferr = c.SuperAdminOnly(dummyResponseRecorder, dummyRequest, user)
	expect(t, ferr != nil, "Blank users shouldn't be considered super admins")

	user.IsSuperAdmin = false
	ferr = c.SuperAdminOnly(dummyResponseRecorder, dummyRequest, user)
	expect(t, ferr != nil, "Non-super admins shouldn't be allowed through the super admin gate")

	user.IsSuperAdmin = true
	ferr = c.SuperAdminOnly(dummyResponseRecorder, dummyRequest, user)
	expect(t, ferr == nil, "Super admins should be allowed through super admin gates")

	// TODO: Make sure only super admins can access the backups route

	//dummyResponseRecorder = httptest.NewRecorder()
	//bytesBuffer = bytes.NewBuffer([]byte(""))
	//dummyRequest = httptest.NewRequest("", "/panel/backups/", bytesBuffer)
}

func TestTopicStore(t *testing.T) {
	miscinit(t)
	if !c.PluginsInited {
		c.InitPlugins()
	}

	var err error
	tcache := c.NewMemoryTopicCache(c.Config.TopicCacheCapacity)
	c.Topics, err = c.NewDefaultTopicStore(tcache)
	expectNilErr(t, err)
	c.Config.DisablePostIP = false
	topicStoreTest(t, 2, "::1")
	c.Config.DisablePostIP = true
	topicStoreTest(t, 3, "")

	c.Topics, err = c.NewDefaultTopicStore(nil)
	expectNilErr(t, err)
	c.Config.DisablePostIP = false
	topicStoreTest(t, 4, "::1")
	c.Config.DisablePostIP = true
	topicStoreTest(t, 5, "")
}
func topicStoreTest(t *testing.T, newID int, ip string) {
	var topic *c.Topic
	var err error

	_, err = c.Topics.Get(-1)
	recordMustNotExist(t, err, "TID #-1 shouldn't exist")
	_, err = c.Topics.Get(0)
	recordMustNotExist(t, err, "TID #0 shouldn't exist")

	topic, err = c.Topics.Get(1)
	recordMustExist(t, err, "Couldn't find TID #1")
	expectf(t, topic.ID == 1, "topic.ID does not match the requested TID. Got '%d' instead.", topic.ID)

	// TODO: Add BulkGetMap() to the TopicStore

	expect(t, !c.Topics.Exists(-1), "TID #-1 shouldn't exist")
	expect(t, !c.Topics.Exists(0), "TID #0 shouldn't exist")
	expect(t, c.Topics.Exists(1), "TID #1 should exist")

	count := c.Topics.Count()
	expectf(t, count == 1, "Global count for topics should be 1, not %d", count)

	//Create(fid int, topicName string, content string, uid int, ip string) (tid int, err error)
	tid, err := c.Topics.Create(2, "Test Topic", "Topic Content", 1, ip)
	expectNilErr(t, err)
	expectf(t, tid == newID, "TID for the new topic should be %d, not %d", newID, tid)
	expectf(t, c.Topics.Exists(newID), "TID #%d should exist", newID)

	count = c.Topics.Count()
	expectf(t, count == 2, "Global count for topics should be 2, not %d", count)

	iFrag := func(cond bool) string {
		if !cond {
			return "n't"
		}
		return ""
	}

	testTopic := func(tid int, title, content string, createdBy int, ip string, parentID int, isClosed, sticky bool) {
		topic, err = c.Topics.Get(tid)
		recordMustExist(t, err, fmt.Sprintf("Couldn't find TID #%d", tid))
		expectf(t, topic.ID == tid, "topic.ID does not match the requested TID. Got '%d' instead.", topic.ID)
		expectf(t, topic.GetID() == tid, "topic.ID does not match the requested TID. Got '%d' instead.", topic.GetID())
		expectf(t, topic.Title == title, "The topic's name should be '%s', not %s", title, topic.Title)
		expectf(t, topic.Content == content, "The topic's body should be '%s', not %s", content, topic.Content)
		expectf(t, topic.CreatedBy == createdBy, "The topic's creator should be %d, not %d", createdBy, topic.CreatedBy)
		expectf(t, topic.IP == ip, "The topic's IP should be '%s', not %s", ip, topic.IP)
		expectf(t, topic.ParentID == parentID, "The topic's parent forum should be %d, not %d", parentID, topic.ParentID)
		expectf(t, topic.IsClosed == isClosed, "This topic should%s be locked", iFrag(topic.IsClosed))
		expectf(t, topic.Sticky == sticky, "This topic should%s be sticky", iFrag(topic.Sticky))
		expectf(t, topic.GetTable() == "topics", "The topic's table should be 'topics', not %s", topic.GetTable())
	}

	tc := c.Topics.GetCache()
	shouldNotBeIn := func(tid int) {
		if tc != nil {
			_, err = tc.Get(tid)
			recordMustNotExist(t, err, "Topic cache should be empty")
		}
	}
	if tc != nil {
		_, err = tc.Get(newID)
		expectNilErr(t, err)
	}

	testTopic(newID, "Test Topic", "Topic Content", 1, ip, 2, false, false)

	expectNilErr(t, topic.Lock())
	shouldNotBeIn(newID)
	testTopic(newID, "Test Topic", "Topic Content", 1, ip, 2, true, false)

	expectNilErr(t, topic.Unlock())
	shouldNotBeIn(newID)
	testTopic(newID, "Test Topic", "Topic Content", 1, ip, 2, false, false)

	expectNilErr(t, topic.Stick())
	shouldNotBeIn(newID)
	testTopic(newID, "Test Topic", "Topic Content", 1, ip, 2, false, true)

	expectNilErr(t, topic.Unstick())
	shouldNotBeIn(newID)
	testTopic(newID, "Test Topic", "Topic Content", 1, ip, 2, false, false)

	expectNilErr(t, topic.MoveTo(1))
	shouldNotBeIn(newID)
	testTopic(newID, "Test Topic", "Topic Content", 1, ip, 1, false, false)
	// TODO: Add more tests for more *Topic methods

	expectNilErr(t, topic.Delete())
	shouldNotBeIn(newID)

	_, err = c.Topics.Get(newID)
	recordMustNotExist(t, err, fmt.Sprintf("TID #%d shouldn't exist", newID))
	expectf(t, !c.Topics.Exists(newID), "TID #%d shouldn't exist", newID)

	// TODO: Test topic creation and retrieving that created topic plus reload and inspecting the cache
}

func TestForumStore(t *testing.T) {
	miscinit(t)
	if !c.PluginsInited {
		c.InitPlugins()
	}
	// TODO: Test ForumStore.Reload

	fcache, ok := c.Forums.(c.ForumCache)
	expect(t, ok, "Unable to cast ForumStore to ForumCache")
	expect(t, c.Forums.Count() == 2, "The forumstore global count should be 2")
	expect(t, fcache.Length() == 2, "The forum cache length should be 2")

	_, err := c.Forums.Get(-1)
	recordMustNotExist(t, err, "FID #-1 shouldn't exist")
	_, err = c.Forums.Get(0)
	recordMustNotExist(t, err, "FID #0 shouldn't exist")

	forum, err := c.Forums.Get(1)
	recordMustExist(t, err, "Couldn't find FID #1")
	expectf(t, forum.ID == 1, "forum.ID doesn't not match the requested FID. Got '%d' instead.'", forum.ID)
	// TODO: Check the preset and forum permissions
	expectf(t, forum.Name == "Reports", "FID #0 is named '%s' and not 'Reports'", forum.Name)
	expectf(t, !forum.Active, "The reports forum shouldn't be active")
	expectDesc := "All the reports go here"
	expectf(t, forum.Desc == expectDesc, "The forum description should be '%s' not '%s'", expectDesc, forum.Desc)
	forum, err = c.Forums.BypassGet(1)
	recordMustExist(t, err, "Couldn't find FID #1")

	forum, err = c.Forums.Get(2)
	recordMustExist(t, err, "Couldn't find FID #2")
	forum, err = c.Forums.BypassGet(2)
	recordMustExist(t, err, "Couldn't find FID #2")

	expectf(t, forum.ID == 2, "The FID should be 2 not %d", forum.ID)
	expectf(t, forum.Name == "General", "The name of the forum should be 'General' not '%s'", forum.Name)
	expectf(t, forum.Active, "The general forum should be active")
	expectDesc = "A place for general discussions which don't fit elsewhere"
	expectf(t, forum.Desc == expectDesc, "The forum description should be '%s' not '%s'", expectDesc, forum.Desc)

	// Forum reload test, kind of hacky but gets the job done
	/*
		CacheGet(id int) (*Forum, error)
		CacheSet(forum *Forum) error
	*/
	expect(t, ok, "ForumCache should be available")
	forum.Name = "nanana"
	fcache.CacheSet(forum)
	forum, err = c.Forums.Get(2)
	recordMustExist(t, err, "Couldn't find FID #2")
	expectf(t, forum.Name == "nanana", "The faux name should be nanana not %s", forum.Name)
	expectNilErr(t, c.Forums.Reload(2))
	forum, err = c.Forums.Get(2)
	recordMustExist(t, err, "Couldn't find FID #2")
	expectf(t, forum.Name == "General", "The proper name should be 2 not %s", forum.Name)

	expect(t, !c.Forums.Exists(-1), "FID #-1 shouldn't exist")
	expect(t, !c.Forums.Exists(0), "FID #0 shouldn't exist")
	expect(t, c.Forums.Exists(1), "FID #1 should exist")
	expect(t, c.Forums.Exists(2), "FID #2 should exist")
	expect(t, !c.Forums.Exists(3), "FID #3 shouldn't exist")

	_, err = c.Forums.Create("", "", true, "all")
	expect(t, err != nil, "A forum shouldn't be successfully created, if it has a blank name")

	fid, err := c.Forums.Create("Test Forum", "", true, "all")
	expectNilErr(t, err)
	expect(t, fid == 3, "The first forum we create should have an ID of 3")
	expect(t, c.Forums.Exists(3), "FID #2 should exist")

	expect(t, c.Forums.Count() == 3, "The forumstore global count should be 3")
	expect(t, fcache.Length() == 3, "The forum cache length should be 3")

	forum, err = c.Forums.Get(3)
	recordMustExist(t, err, "Couldn't find FID #3")
	forum, err = c.Forums.BypassGet(3)
	recordMustExist(t, err, "Couldn't find FID #3")

	expectf(t, forum.ID == 3, "The FID should be 3 not %d", forum.ID)
	expectf(t, forum.Name == "Test Forum", "The name of the forum should be 'Test Forum' not '%s'", forum.Name)
	expectf(t, forum.Active, "The test forum should be active")
	expectf(t, forum.Desc == "", "The forum description should be blank not '%s'", forum.Desc)

	// TODO: More forum creation tests

	expectNilErr(t, c.Forums.Delete(3))
	expect(t, forum.ID == 3, "forum pointer shenanigans")
	expect(t, c.Forums.Count() == 2, "The forumstore global count should be 2")
	expect(t, fcache.Length() == 2, "The forum cache length should be 2")
	expect(t, !c.Forums.Exists(3), "FID #3 shouldn't exist after being deleted")
	_, err = c.Forums.Get(3)
	recordMustNotExist(t, err, "FID #3 shouldn't exist after being deleted")
	_, err = c.Forums.BypassGet(3)
	recordMustNotExist(t, err, "FID #3 shouldn't exist after being deleted")

	expect(t, c.Forums.Delete(c.ReportForumID) != nil, "The reports forum shouldn't be deletable")
	expectf(t, c.Forums.Exists(c.ReportForumID), "FID #%d should still exist", c.ReportForumID)
	_, err = c.Forums.Get(c.ReportForumID)
	expectf(t, err == nil, "FID #%d should still exist", c.ReportForumID)
	_, err = c.Forums.BypassGet(c.ReportForumID)
	expectf(t, err == nil, "FID #%d should still exist", c.ReportForumID)

	eforums := map[int]bool{1: true, 2: true}
	{
		forums, err := c.Forums.GetAll()
		expectNilErr(t, err)
		found := make(map[int]*c.Forum)
		for _, forum := range forums {
			_, ok := eforums[forum.ID]
			expectf(t, ok, "unknown forum #%d in forums", forum.ID)
			found[forum.ID] = forum
		}
		for fid, _ := range eforums {
			_, ok := found[fid]
			expectf(t, ok, "unable to find expected forum #%d in forums", fid)
		}
	}

	{
		fids, err := c.Forums.GetAllIDs()
		expectNilErr(t, err)
		found := make(map[int]bool)
		for _, fid := range fids {
			_, ok := eforums[fid]
			expectf(t, ok, "unknown fid #%d in fids", fid)
			found[fid] = true
		}
		for fid, _ := range eforums {
			_, ok := found[fid]
			expectf(t, ok, "unable to find expected fid #%d in fids", fid)
		}
	}

	vforums := map[int]bool{2: true}
	{
		forums, err := c.Forums.GetAllVisible()
		expectNilErr(t, err)
		found := make(map[int]*c.Forum)
		for _, forum := range forums {
			_, ok := vforums[forum.ID]
			expectf(t, ok, "unknown forum #%d in forums", forum.ID)
			found[forum.ID] = forum
		}
		for fid, _ := range vforums {
			_, ok := found[fid]
			expectf(t, ok, "unable to find expected forum #%d in forums", fid)
		}
	}

	{
		fids, err := c.Forums.GetAllVisibleIDs()
		expectNilErr(t, err)
		found := make(map[int]bool)
		for _, fid := range fids {
			_, ok := vforums[fid]
			expectf(t, ok, "unknown fid #%d in fids", fid)
			found[fid] = true
		}
		for fid, _ := range vforums {
			_, ok := found[fid]
			expectf(t, ok, "unable to find expected fid #%d in fids", fid)
		}
	}

	// TODO: Test forum update
	// TODO: Other forumstore stuff and forumcache?
}

// TODO: Implement this
func TestForumPermsStore(t *testing.T) {
	miscinit(t)
	if !c.PluginsInited {
		c.InitPlugins()
	}

	f := func(fid, gid int, msg string, inv ...bool) {
		fp, err := c.FPStore.Get(fid, gid)
		expectNilErr(t, err)
		vt := fp.ViewTopic
		if len(inv) > 0 && inv[0] == true {
			vt = !vt
		}
		expect(t, vt, msg)
	}

	initialState := func() {
		f(1, 1, "admins should be able to see reports")
		f(1, 2, "mods should be able to see reports")
		f(1, 3, "members should not be able to see reports", true)
		f(1, 4, "banned users should not be able to see reports", true)
		f(2, 1, "admins should be able to see general")
		f(2, 3, "members should be able to see general")
		f(2, 6, "guests should be able to see general")
	}
	initialState()

	expectNilErr(t, c.FPStore.Reload(1))
	initialState()
	expectNilErr(t, c.FPStore.Reload(2))
	initialState()
}

// TODO: Test the group permissions
// TODO: Test group.CanSee for forum presets + group perms
func TestGroupStore(t *testing.T) {
	miscinit(t)
	if !c.PluginsInited {
		c.InitPlugins()
	}

	_, err := c.Groups.Get(-1)
	recordMustNotExist(t, err, "GID #-1 shouldn't exist")

	// TODO: Refactor the group store to remove GID #0
	g, err := c.Groups.Get(0)
	recordMustExist(t, err, "Couldn't find GID #0")

	expectf(t, g.ID == 0, "g.ID doesn't not match the requested GID. Got '%d' instead.", g.ID)
	expectf(t, g.Name == "Unknown", "GID #0 is named '%s' and not 'Unknown'", g.Name)

	g, err = c.Groups.Get(1)
	recordMustExist(t, err, "Couldn't find GID #1")
	expectf(t, g.ID == 1, "g.ID doesn't not match the requested GID. Got '%d' instead.'", g.ID)
	expect(t, len(g.CanSee) > 0, "g.CanSee should not be zero")

	expect(t, !c.Groups.Exists(-1), "GID #-1 shouldn't exist")
	// 0 aka Unknown, for system posts and other oddities
	expect(t, c.Groups.Exists(0), "GID #0 should exist")
	expect(t, c.Groups.Exists(1), "GID #1 should exist")

	isAdmin := true
	isMod := true
	isBanned := false
	gid, err := c.Groups.Create("Testing", "Test", isAdmin, isMod, isBanned)
	expectNilErr(t, err)
	expect(t, c.Groups.Exists(gid), "The group we just made doesn't exist")

	g, err = c.Groups.Get(gid)
	expectNilErr(t, err)
	expect(t, g.ID == gid, "The group ID should match the requested ID")
	expect(t, g.IsAdmin, "This should be an admin group")
	expect(t, g.IsMod, "This should be a mod group")
	expect(t, !g.IsBanned, "This shouldn't be a ban group")
	expect(t, len(g.CanSee) == 0, "g.CanSee should be empty")

	isAdmin = false
	isMod = true
	isBanned = true
	gid, err = c.Groups.Create("Testing 2", "Test", isAdmin, isMod, isBanned)
	expectNilErr(t, err)
	expect(t, c.Groups.Exists(gid), "The group we just made doesn't exist")

	g, err = c.Groups.Get(gid)
	expectNilErr(t, err)
	expect(t, g.ID == gid, "The group ID should match the requested ID")
	expect(t, !g.IsAdmin, "This should not be an admin group")
	expect(t, g.IsMod, "This should be a mod group")
	expect(t, !g.IsBanned, "This shouldn't be a ban group")

	// TODO: Make sure this pointer doesn't change once we refactor the group store to stop updating the pointer
	err = g.ChangeRank(false, false, true)
	expectNilErr(t, err)

	g, err = c.Groups.Get(gid)
	expectNilErr(t, err)
	expect(t, g.ID == gid, "The group ID should match the requested ID")
	expect(t, !g.IsAdmin, "This shouldn't be an admin group")
	expect(t, !g.IsMod, "This shouldn't be a mod group")
	expect(t, g.IsBanned, "This should be a ban group")

	err = g.ChangeRank(true, true, true)
	expectNilErr(t, err)

	g, err = c.Groups.Get(gid)
	expectNilErr(t, err)
	expect(t, g.ID == gid, "The group ID should match the requested ID")
	expect(t, g.IsAdmin, "This should be an admin group")
	expect(t, g.IsMod, "This should be a mod group")
	expect(t, !g.IsBanned, "This shouldn't be a ban group")
	expect(t, len(g.CanSee) == 0, "len(g.CanSee) should be 0")

	err = g.ChangeRank(false, true, true)
	expectNilErr(t, err)

	forum, err := c.Forums.Get(2)
	expectNilErr(t, err)
	forumPerms, err := c.FPStore.GetCopy(2, gid)
	if err == sql.ErrNoRows {
		forumPerms = *c.BlankForumPerms()
	} else if err != nil {
		expectNilErr(t, err)
	}
	forumPerms.ViewTopic = true

	err = forum.SetPerms(&forumPerms, "custom", gid)
	expectNilErr(t, err)

	g, err = c.Groups.Get(gid)
	expectNilErr(t, err)
	expect(t, g.ID == gid, "The group ID should match the requested ID")
	expect(t, !g.IsAdmin, "This shouldn't be an admin group")
	expect(t, g.IsMod, "This should be a mod group")
	expect(t, !g.IsBanned, "This shouldn't be a ban group")
	expect(t, g.CanSee != nil, "g.CanSee must not be nil")
	expect(t, len(g.CanSee) == 1, "len(g.CanSee) should not be one")
	expect(t, g.CanSee[0] == 2, "g.CanSee[0] should be 2")
	canSee := g.CanSee

	// Make sure the data is static
	c.Groups.Reload(gid)

	g, err = c.Groups.Get(gid)
	expectNilErr(t, err)
	expect(t, g.ID == gid, "The group ID should match the requested ID")
	expect(t, !g.IsAdmin, "This shouldn't be an admin group")
	expect(t, g.IsMod, "This should be a mod group")
	expect(t, !g.IsBanned, "This shouldn't be a ban group")

	// TODO: Don't enforce a specific order here
	canSeeTest := func(a, b []int) bool {
		if (a == nil) != (b == nil) {
			return false
		}
		if len(a) != len(b) {
			return false
		}
		for i := range a {
			if a[i] != b[i] {
				return false
			}
		}
		return true
	}

	expect(t, canSeeTest(g.CanSee, canSee), "g.CanSee is not being reused")

	// TODO: Test group deletion
	// TODO: Test group reload
	// TODO: Test group cache set
}

func TestGroupPromotions(t *testing.T) {
	miscinit(t)
	if !c.PluginsInited {
		c.InitPlugins()
	}

	_, err := c.GroupPromotions.Get(-1)
	recordMustNotExist(t, err, "GP #-1 shouldn't exist")
	_, err = c.GroupPromotions.Get(0)
	recordMustNotExist(t, err, "GP #0 shouldn't exist")
	_, err = c.GroupPromotions.Get(1)
	recordMustNotExist(t, err, "GP #1 shouldn't exist")
	expectNilErr(t, c.GroupPromotions.Delete(1))

	//GetByGroup(gid int) (gps []*GroupPromotion, err error)

	testPromo := func(exid, from, to, level, posts, registeredFor int, shouldFail bool) {
		gpid, err := c.GroupPromotions.Create(from, to, false, level, posts, registeredFor)
		expectf(t, gpid == exid, "gpid should be %d not %d", exid, gpid)
		//fmt.Println("gpid:", gpid)
		gp, err := c.GroupPromotions.Get(gpid)
		expectNilErr(t, err)
		expectf(t, gp.ID == gpid, "gp.ID should be %d not %d", gpid, gp.ID)
		expectf(t, gp.From == from, "gp.From should be %d not %d", from, gp.From)
		expectf(t, gp.To == to, "gp.To should be %d not %d", to, gp.To)
		expect(t, !gp.TwoWay, "gp.TwoWay should be false not true")
		expectf(t, gp.Level == level, "gp.Level should be %d not %d", level, gp.Level)
		expectf(t, gp.Posts == posts, "gp.Posts should be %d not %d", posts, gp.Posts)
		expectf(t, gp.MinTime == 0, "gp.MinTime should be %d not %d", 0, gp.MinTime)
		expectf(t, gp.RegisteredFor == registeredFor, "gp.RegisteredFor should be %d not %d", registeredFor, gp.RegisteredFor)

		uid, err := c.Users.Create("Lord_"+strconv.Itoa(gpid), "I_Rule", "", from, false)
		expectNilErr(t, err)
		u, err := c.Users.Get(uid)
		expectNilErr(t, err)
		expectf(t, u.ID == uid, "u.ID should be %d not %d", uid, u.ID)
		expectf(t, u.Group == from, "u.Group should be %d not %d", from, u.Group)
		err = c.GroupPromotions.PromoteIfEligible(u, u.Level, u.Posts, u.CreatedAt)
		expectNilErr(t, err)
		u.CacheRemove()
		u, err = c.Users.Get(uid)
		expectNilErr(t, err)
		expectf(t, u.ID == uid, "u.ID should be %d not %d", uid, u.ID)
		if shouldFail {
			expectf(t, u.Group == from, "u.Group should be (from-group) %d not %d", from, u.Group)
		} else {
			expectf(t, u.Group == to, "u.Group should be (to-group)%d not %d", to, u.Group)
		}

		expectNilErr(t, c.GroupPromotions.Delete(gpid))
		_, err = c.GroupPromotions.Get(gpid)
		recordMustNotExist(t, err, fmt.Sprintf("GP #%d should no longer exist", gpid))
	}
	testPromo(1, 1, 2, 0, 0, 0, false)
	testPromo(2, 1, 2, 5, 5, 0, true)
	testPromo(3, 1, 2, 0, 0, 1, true)
}

func TestReplyStore(t *testing.T) {
	miscinit(t)
	if !c.PluginsInited {
		c.InitPlugins()
	}
	_, err := c.Rstore.Get(-1)
	recordMustNotExist(t, err, "RID #-1 shouldn't exist")
	_, err = c.Rstore.Get(0)
	recordMustNotExist(t, err, "RID #0 shouldn't exist")

	c.Config.DisablePostIP = false
	testReplyStore(t, 2, 1, "::1")
	c.Config.DisablePostIP = true
	testReplyStore(t, 5, 3, "")
}

func testReplyStore(t *testing.T, newID, newPostCount int, ip string) {
	replyTest2 := func(r *c.Reply, err error, rid, parentID, createdBy int, content, ip string) {
		expectNilErr(t, err)
		expectf(t, r.ID == rid, "RID #%d has the wrong ID. It should be %d not %d", rid, rid, r.ID)
		expectf(t, r.ParentID == parentID, "The parent topic of RID #%d should be %d not %d", rid, parentID, r.ParentID)
		expectf(t, r.CreatedBy == createdBy, "The creator of RID #%d should be %d not %d", rid, createdBy, r.CreatedBy)
		expectf(t, r.Content == content, "The contents of RID #%d should be '%s' not %s", rid, content, r.Content)
		expectf(t, r.IP == ip, "The IP of RID#%d should be '%s' not %s", rid, ip, r.IP)
	}

	replyTest := func(rid, parentID, createdBy int, content, ip string) {
		r, err := c.Rstore.Get(rid)
		replyTest2(r, err, rid, parentID, createdBy, content, ip)
		r, err = c.Rstore.GetCache().Get(rid)
		replyTest2(r, err, rid, parentID, createdBy, content, ip)
	}
	replyTest(1, 1, 1, "A reply!", "::1")

	// ! This is hard to do deterministically as the system may pre-load certain items but let's give it a try:
	//_, err = c.Rstore.GetCache().Get(1)
	//recordMustNotExist(t, err, "RID #1 shouldn't be in the cache")

	_, err := c.Rstore.Get(newID)
	recordMustNotExist(t, err, "RID #2 shouldn't exist")

	topic, err := c.Topics.Get(1)
	expectNilErr(t, err)
	expectf(t, topic.PostCount == newPostCount, "TID #%d's post count should be %d, not %d", topic.ID, newPostCount, topic.PostCount)

	_, err = c.Rstore.GetCache().Get(newID)
	recordMustNotExist(t, err, "RID #%d shouldn't be in the cache", newID)

	rid, err := c.Rstore.Create(topic, "Fofofo", ip, 1)
	expectNilErr(t, err)
	expectf(t, rid == newID, "The next reply ID should be %d not %d", newID, rid)
	expectf(t, topic.PostCount == newPostCount, "The old TID #1 in memory's post count should be %d, not %d", newPostCount+1, topic.PostCount)
	// TODO: Test the reply count on the topic

	replyTest(newID, 1, 1, "Fofofo", ip)

	topic, err = c.Topics.Get(1)
	expectNilErr(t, err)
	expectf(t, topic.PostCount == newPostCount+1, "TID #1's post count should be %d, not %d", newPostCount+1, topic.PostCount)

	err = topic.CreateActionReply("destroy", ip, 1)
	expectNilErr(t, err)
	expectf(t, topic.PostCount == newPostCount+1, "The old TID #1 in memory's post count should be %d, not %d", newPostCount+1, topic.PostCount)
	replyTest(newID+1, 1, 1, "", ip)
	// TODO: Check the actionType field of the reply, this might not be loaded by TopicStore, maybe we should add it there?

	topic, err = c.Topics.Get(1)
	expectNilErr(t, err)
	expectf(t, topic.PostCount == newPostCount+2, "TID #1's post count should be %d, not %d", newPostCount+2, topic.PostCount)

	// TODO: Expand upon this
	rid, err = c.Rstore.Create(topic, "hiii", ip, 1)
	expectNilErr(t, err)
	replyTest(rid, topic.ID, 1, "hiii", ip)

	reply, err := c.Rstore.Get(rid)
	expectNilErr(t, err)
	expectNilErr(t, reply.SetPost("huuu"))
	expectf(t, reply.Content == "hiii", "topic.Content should be hiii, not %s", reply.Content)
	reply, err = c.Rstore.Get(rid)
	expectNilErr(t, err)
	expectf(t, reply.Content == "huuu", "topic.Content should be huuu, not %s", reply.Content)
	expectNilErr(t, reply.Delete())
	// No pointer shenanigans x.x
	// TODO: Log reply.ID and rid in cases of pointer shenanigans?
	expect(t, reply.ID == rid, "pointer shenanigans")

	_, err = c.Rstore.GetCache().Get(rid)
	recordMustNotExist(t, err, fmt.Sprintf("RID #%d shouldn't be in the cache", rid))
	_, err = c.Rstore.Get(rid)
	recordMustNotExist(t, err, fmt.Sprintf("RID #%d shouldn't exist", rid))

	// TODO: Write a test for this
	//(topic *TopicUser) Replies(offset int, pFrag int, user *User) (rlist []*ReplyUser, ogdesc string, err error)

	// TODO: Add tests for *Reply
	// TODO: Add tests for ReplyCache
}

func TestLikes(t *testing.T) {
	miscinit(t)
	if !c.PluginsInited {
		c.InitPlugins()
	}

	uid := 1
	ids, err := c.Likes.BulkExists([]int{}, uid, "replies")
	//recordMustNotExist(t, err, "no likes should be found")
	expectNilErr(t, err)
	expect(t, len(ids) == 0, "len ids should be 0")

	topic, err := c.Topics.Get(1)
	expectNilErr(t, err)
	rid, err := c.Rstore.Create(topic, "hiii", "", uid)
	expectNilErr(t, err)
	r, err := c.Rstore.Get(rid)
	expectNilErr(t, err)
	expectNilErr(t, r.Like(uid))
	ids, err = c.Likes.BulkExists([]int{rid}, uid, "replies")
	expectNilErr(t, err)
	expectf(t, len(ids) == 1, "ids should be %d not %d", 1, len(ids))

	rid2, err := c.Rstore.Create(topic, "hi 2 u 2", "", uid)
	expectNilErr(t, err)
	r2, err := c.Rstore.Get(rid2)
	expectNilErr(t, err)
	expectNilErr(t, r2.Like(uid))
	ids, err = c.Likes.BulkExists([]int{rid, rid2}, uid, "replies")
	expectNilErr(t, err)
	expectf(t, len(ids) == 2, "ids should be %d not %d", 2, len(ids))

	expectNilErr(t, r.Unlike(uid))
	ids, err = c.Likes.BulkExists([]int{rid2}, uid, "replies")
	expectNilErr(t, err)
	expectf(t, len(ids) == 1, "ids should be %d not %d", 1, len(ids))
	expectNilErr(t, r2.Unlike(uid))
	ids, err = c.Likes.BulkExists([]int{}, uid, "replies")
	//recordMustNotExist(t, err, "no likes should be found")
	expectNilErr(t, err)
	expect(t, len(ids) == 0, "len ids should be 0")

	//BulkExists(ids []int, sentBy int, targetType string) (eids []int, err error)

	expectNilErr(t, topic.Like(1, uid))
	expectNilErr(t, topic.Unlike(uid))
}

func TestAttachments(t *testing.T) {
	miscinit(t)
	if !c.PluginsInited {
		c.InitPlugins()
	}

	filename := "n0-48.png"
	srcFile := "./test_data/" + filename
	destFile := "./attachs/" + filename

	expect(t, c.Attachments.Count() == 0, "the number of attachments should be 0")
	expect(t, c.Attachments.CountIn("topics", 1) == 0, "the number of attachments in topic 1 should be 0")
	expectf(t, c.Attachments.CountInPath(filename) == 0, "the number of attachments with path '%s' should be 0", filename)
	_, err := c.Attachments.FGet(1)
	if err != nil && err != sql.ErrNoRows {
		t.Error(err)
	}
	expect(t, err == sql.ErrNoRows, ".FGet should have no results")
	_, err = c.Attachments.Get(1)
	if err != nil && err != sql.ErrNoRows {
		t.Error(err)
	}
	expect(t, err == sql.ErrNoRows, ".Get should have no results")
	_, err = c.Attachments.MiniGetList("topics", 1)
	if err != nil && err != sql.ErrNoRows {
		t.Error(err)
	}
	expect(t, err == sql.ErrNoRows, ".MiniGetList should have no results")
	_, err = c.Attachments.BulkMiniGetList("topics", []int{1})
	if err != nil && err != sql.ErrNoRows {
		t.Error(err)
	}
	expect(t, err == sql.ErrNoRows, ".BulkMiniGetList should have no results")

	simUpload := func() {
		// Sim an upload, try a proper upload through the proper pathway later on
		_, err = os.Stat(destFile)
		if err != nil && !os.IsNotExist(err) {
			expectNilErr(t, err)
		} else if err == nil {
			err := os.Remove(destFile)
			expectNilErr(t, err)
		}

		input, err := ioutil.ReadFile(srcFile)
		expectNilErr(t, err)
		err = ioutil.WriteFile(destFile, input, 0644)
		expectNilErr(t, err)
	}
	simUpload()

	tid, err := c.Topics.Create(2, "Attach Test", "Filler Body", 1, "")
	expectNilErr(t, err)
	aid, err := c.Attachments.Add(2, "forums", tid, "topics", 1, filename, "")
	expectNilErr(t, err)
	expectf(t, aid == 1, "aid should be 1 not %d", aid)
	expectNilErr(t, c.Attachments.AddLinked("topics", tid))
	expect(t, c.Attachments.Count() == 1, "the number of attachments should be 1")
	expectf(t, c.Attachments.CountIn("topics", tid) == 1, "the number of attachments in topic %d should be 1", tid)
	expectf(t, c.Attachments.CountInPath(filename) == 1, "the number of attachments with path '%s' should be 1", filename)

	e := func(a *c.MiniAttachment, aid, sid, oid, uploadedBy int, path, extra, ext string) {
		expectf(t, a.ID == aid, "ID should be %d not %d", aid, a.ID)
		expectf(t, a.SectionID == sid, "SectionID should be %d not %d", sid, a.SectionID)
		expectf(t, a.OriginID == oid, "OriginID should be %d not %d", oid, a.OriginID)
		expectf(t, a.UploadedBy == uploadedBy, "UploadedBy should be %d not %d", uploadedBy, a.UploadedBy)
		expectf(t, a.Path == path, "Path should be %s not %s", path, a.Path)
		expectf(t, a.Extra == extra, "Extra should be %s not %s", extra, a.Extra)
		expect(t, a.Image, "Image should be true")
		expectf(t, a.Ext == ext, "Ext should be %s not %s", ext, a.Ext)
	}
	e2 := func(a *c.Attachment, aid, sid, oid, uploadedBy int, path, extra, ext string) {
		expectf(t, a.ID == aid, "ID should be %d not %d", aid, a.ID)
		expectf(t, a.SectionID == sid, "SectionID should be %d not %d", sid, a.SectionID)
		expectf(t, a.OriginID == oid, "OriginID should be %d not %d", oid, a.OriginID)
		expectf(t, a.UploadedBy == uploadedBy, "UploadedBy should be %d not %d", uploadedBy, a.UploadedBy)
		expectf(t, a.Path == path, "Path should be %s not %s", path, a.Path)
		expectf(t, a.Extra == extra, "Extra should be %s not %s", extra, a.Extra)
		expect(t, a.Image, "Image should be true")
		expectf(t, a.Ext == ext, "Ext should be %s not %s", ext, a.Ext)
	}

	f2 := func(aid, sid, oid int, extra string, topic bool) {
		var tbl string
		if topic {
			tbl = "topics"
		} else {
			tbl = "replies"
		}
		fa, err := c.Attachments.FGet(aid)
		expectNilErr(t, err)
		e2(fa, aid, sid, oid, 1, filename, extra, "png")

		a, err := c.Attachments.Get(aid)
		expectNilErr(t, err)
		e(a, aid, sid, oid, 1, filename, extra, "png")

		alist, err := c.Attachments.MiniGetList(tbl, oid)
		expectNilErr(t, err)
		expectf(t, len(alist) == 1, "len(alist) should be 1 not %d", len(alist))
		a = alist[0]
		e(a, aid, sid, oid, 1, filename, extra, "png")

		amap, err := c.Attachments.BulkMiniGetList(tbl, []int{oid})
		expectNilErr(t, err)
		expectf(t, len(amap) == 1, "len(amap) should be 1 not %d", len(amap))
		alist, ok := amap[oid]
		if !ok {
			t.Logf("key %d not found in amap", oid)
		}
		expectf(t, len(alist) == 1, "len(alist) should be 1 not %d", len(alist))
		a = alist[0]
		e(a, aid, sid, oid, 1, filename, extra, "png")
	}

	topic, err := c.Topics.Get(tid)
	expectNilErr(t, err)
	expectf(t, topic.AttachCount == 1, "topic.AttachCount should be 1 not %d", topic.AttachCount)
	f2(aid, 2, tid, "", true)
	expectNilErr(t, topic.MoveTo(1))
	f2(aid, 1, tid, "", true)
	expectNilErr(t, c.Attachments.MoveTo(2, tid, "topics"))
	f2(aid, 2, tid, "", true)

	// TODO: ShowAttachment test

	deleteTest := func(aid, oid int, topic bool) {
		var tbl string
		if topic {
			tbl = "topics"
		} else {
			tbl = "replies"
		}
		//expectNilErr(t, c.Attachments.Delete(aid))
		expectNilErr(t, c.DeleteAttachment(aid))
		expect(t, c.Attachments.Count() == 0, "the number of attachments should be 0")
		expectf(t, c.Attachments.CountIn(tbl, oid) == 0, "the number of attachments in topic %d should be 0", tid)
		expectf(t, c.Attachments.CountInPath(filename) == 0, "the number of attachments with path '%s' should be 0", filename)
		_, err = c.Attachments.FGet(aid)
		if err != nil && err != sql.ErrNoRows {
			t.Error(err)
		}
		expect(t, err == sql.ErrNoRows, ".FGet should have no results")
		_, err = c.Attachments.Get(aid)
		if err != nil && err != sql.ErrNoRows {
			t.Error(err)
		}
		expect(t, err == sql.ErrNoRows, ".Get should have no results")
		_, err = c.Attachments.MiniGetList(tbl, oid)
		if err != nil && err != sql.ErrNoRows {
			t.Error(err)
		}
		expect(t, err == sql.ErrNoRows, ".MiniGetList should have no results")
		_, err = c.Attachments.BulkMiniGetList(tbl, []int{oid})
		if err != nil && err != sql.ErrNoRows {
			t.Error(err)
		}
		expect(t, err == sql.ErrNoRows, ".BulkMiniGetList should have no results")
	}
	deleteTest(aid, tid, true)
	topic, err = c.Topics.Get(tid)
	expectNilErr(t, err)
	expectf(t, topic.AttachCount == 0, "topic.AttachCount should be 0 not %d", topic.AttachCount)

	simUpload()
	rid, err := c.Rstore.Create(topic, "Reply Filler", "", 1)
	expectNilErr(t, err)
	aid, err = c.Attachments.Add(2, "forums", rid, "replies", 1, filename, strconv.Itoa(topic.ID))
	expectNilErr(t, err)
	expectf(t, aid == 2, "aid should be 2 not %d", aid)
	expectNilErr(t, c.Attachments.AddLinked("replies", rid))
	r, err := c.Rstore.Get(rid)
	expectNilErr(t, err)
	expectf(t, r.AttachCount == 1, "r.AttachCount should be 1 not %d", r.AttachCount)
	f2(aid, 2, rid, strconv.Itoa(topic.ID), false)
	expectNilErr(t, c.Attachments.MoveTo(1, rid, "replies"))
	f2(aid, 1, rid, strconv.Itoa(topic.ID), false)
	deleteTest(aid, rid, false)
	r, err = c.Rstore.Get(rid)
	expectNilErr(t, err)
	expectf(t, r.AttachCount == 0, "r.AttachCount should be 0 not %d", r.AttachCount)

	// TODO: Path overlap tests
}

func TestPolls(t *testing.T) {
	miscinit(t)
	if !c.PluginsInited {
		c.InitPlugins()
	}

	shouldNotExist := func(id int) {
		expectf(t, !c.Polls.Exists(id), "poll %d should not exist", id)
		_, err := c.Polls.Get(id)
		recordMustNotExist(t, err, fmt.Sprintf("poll %d shouldn't exist", id))
	}
	shouldNotExist(-1)
	shouldNotExist(0)
	shouldNotExist(1)

	tid, err := c.Topics.Create(2, "Poll Test", "Filler Body", 1, "")
	expectNilErr(t, err)
	topic, err := c.Topics.Get(tid)
	/*Options      map[int]string
		Results      map[int]int  // map[optionIndex]points
		QuickOptions []PollOption // TODO: Fix up the template transpiler so we don't need to use this hack anymore
	}*/
	pollType := 0 // Basic single choice
	pid, err := c.Polls.Create(topic, pollType, map[int]string{0: "item 1", 1: "item 2", 2: "item 3"})
	expectNilErr(t, err)
	expectf(t, pid == 1, "poll id should be 1 not %d", pid)
	expect(t, c.Polls.Exists(1), "poll 1 should exist")

	testPoll := func(p *c.Poll, id, parentID int, parentTable string, ptype int, antiCheat bool, voteCount int) {
		ef := expectf
		ef(t, p.ID == id, "p.ID should be %d not %d", id, p.ID)
		ef(t, p.ParentID == parentID, "p.ParentID should be %d not %d", parentID, p.ParentID)
		ef(t, p.ParentTable == parentTable, "p.ParentID should be %s not %s", parentTable, p.ParentTable)
		ef(t, p.Type == ptype, "p.ParentID should be %d not %d", ptype, p.Type)
		s := "false"
		if p.AntiCheat {
			s = "true"
		}
		ef(t, p.AntiCheat == antiCheat, "p.AntiCheat should be ", s)
		// TODO: More fields
		ef(t, p.VoteCount == voteCount, "p.VoteCount should be %d not %d", voteCount, p.VoteCount)
	}

	p, err := c.Polls.Get(1)
	expectNilErr(t, err)
	testPoll(p, 1, tid, "topics", 0, false, 0)

	expectNilErr(t, p.CastVote(0, 1, ""))
	expectNilErr(t, c.Polls.Reload(p.ID))
	p, err = c.Polls.Get(1)
	expectNilErr(t, err)
	testPoll(p, 1, tid, "topics", 0, false, 1)

	expectNilErr(t, p.Delete())
	expect(t, !c.Polls.Exists(1), "poll 1 should no longer exist")
	_, err = c.Polls.Get(1)
	recordMustNotExist(t, err, "poll 1 should no longer exist")
}

func TestSearch(t *testing.T) {
	miscinit(t)
	if !c.PluginsInited {
		c.InitPlugins()
	}

	title := "search"
	body := "bab bab bab bab"
	q := "search"
	tid, err := c.Topics.Create(2, title, body, 1, "")
	expectNilErr(t, err)

	tids, err := c.RepliesSearch.Query(q, []int{2})
	fmt.Printf("tids: %+v\n", tids)
	expectNilErr(t, err)
	expectf(t, len(tids) == 1, "len(tids) should be 1 not %d", len(tids))

	topic, err := c.Topics.Get(tids[0])
	expectNilErr(t, err)
	expectf(t, topic.ID == tid, "topic.ID should be %d not %d", tid, topic.ID)
	expectf(t, topic.Title == title, "topic.Title should be %s not %s", title, topic.Title)

	tids, err = c.RepliesSearch.Query(q, []int{1, 2})
	fmt.Printf("tids: %+v\n", tids)
	expectNilErr(t, err)
	expectf(t, len(tids) == 1, "len(tids) should be 1 not %d", len(tids))

	q = "bab"
	tids, err = c.RepliesSearch.Query(q, []int{1, 2})
	fmt.Printf("tids: %+v\n", tids)
	expectNilErr(t, err)
	expectf(t, len(tids) == 1, "len(tids) should be 1 not %d", len(tids))
}

func TestProfileReplyStore(t *testing.T) {
	miscinit(t)
	if !c.PluginsInited {
		c.InitPlugins()
	}

	_, err := c.Prstore.Get(-1)
	recordMustNotExist(t, err, "PRID #-1 shouldn't exist")
	_, err = c.Prstore.Get(0)
	recordMustNotExist(t, err, "PRID #0 shouldn't exist")
	_, err = c.Prstore.Get(1)
	recordMustNotExist(t, err, "PRID #1 shouldn't exist")

	c.Config.DisablePostIP = false
	testProfileReplyStore(t, 1, "::1")
	c.Config.DisablePostIP = true
	testProfileReplyStore(t, 2, "")
}
func testProfileReplyStore(t *testing.T, newID int, ip string) {
	// ? - Commented this one out as strong constraints like this put an unreasonable load on the database, we only want errors if a delete which should succeed fails
	//profileReply := c.BlankProfileReply(1)
	//err = profileReply.Delete()
	//expect(t,err != nil,"You shouldn't be able to delete profile replies which don't exist")

	profileID := 1
	prid, err := c.Prstore.Create(profileID, "Haha", 1, ip)
	expectNilErr(t, err)
	expectf(t, prid == newID, "The first profile reply should have an ID of %d", newID)

	pr, err := c.Prstore.Get(newID)
	expectNilErr(t, err)
	expectf(t, pr.ID == newID, "The profile reply should have an ID of %d not %d", newID, pr.ID)
	expectf(t, pr.ParentID == 1, "The parent ID of the profile reply should be 1 not %d", pr.ParentID)
	expectf(t, pr.Content == "Haha", "The profile reply's contents should be 'Haha' not '%s'", pr.Content)
	expectf(t, pr.CreatedBy == 1, "The profile reply's creator should be 1 not %d", pr.CreatedBy)
	expectf(t, pr.IP == ip, "The profile reply's IP should be '%s' not '%s'", ip, pr.IP)

	err = pr.Delete()
	expectNilErr(t, err)
	_, err = c.Prstore.Get(newID)
	expectf(t, err != nil, "PRID #%d shouldn't exist after being deleted", newID)

	// TODO: Test pr.SetBody() and pr.Creator()
}

func TestConvos(t *testing.T) {
	miscinit(t)
	if !c.PluginsInited {
		c.InitPlugins()
	}

	sf := func(i interface{}, e error) error {
		return e
	}
	mf := func(e error, msg string, exists bool) {
		if !exists {
			recordMustNotExist(t, e, msg)
		} else {
			recordMustExist(t, e, msg)
		}
	}
	gu := func(uid, offset int, exists bool) {
		s := ""
		if !exists {
			s = " not"
		}
		mf(sf(c.Convos.GetUser(uid, offset)), fmt.Sprintf("convo getuser %d %d should%s exist", uid, offset, s), exists)
	}
	gue := func(uid, offset int, exists bool) {
		s := ""
		if !exists {
			s = " not"
		}
		mf(sf(c.Convos.GetUserExtra(uid, offset)), fmt.Sprintf("convo getuserextra %d %d should%s exist", uid, offset, s), exists)
	}

	expect(t, c.Convos.GetUserCount(-1) == 0, "getusercount should be zero")
	expect(t, c.Convos.GetUserCount(0) == 0, "getusercount should be zero")
	mf(sf(c.Convos.Get(-1)), "convo -1 should not exist", false)
	mf(sf(c.Convos.Get(0)), "convo 0 should not exist", false)
	gu(-1, -1, false)
	gu(-1, 0, false)
	gu(0, 0, false)
	gue(-1, -1, false)
	gue(-1, 0, false)
	gue(0, 0, false)

	nf := func(cid int, count int) {
		ex := count > 0
		s := ""
		if !ex {
			s = " not"
		}
		mf(sf(c.Convos.Get(cid)), fmt.Sprintf("convo %d should%s exist", cid, s), ex)
		gu(1, 0, ex)
		gu(1, 5, false) // invariant may change in future tests

		expectf(t, c.Convos.GetUserCount(1) == count, "getusercount should be %d", count)
		gue(1, 0, ex)
		gue(1, 5, false) // invariant may change in future tests
		expectf(t, c.Convos.Count() == count, "convos count should be %d", count)
	}
	nf(1, 0)

	awaitingActivation := 5
	uid, err := c.Users.Create("Saturn", "ReallyBadPassword", "", awaitingActivation, false)
	expectNilErr(t, err)

	cid, err := c.Convos.Create("hehe", 1, []int{uid})
	expectNilErr(t, err)
	expect(t, cid == 1, "cid should be 1")
	expect(t, c.Convos.Count() == 1, "convos count should be 1")

	co, err := c.Convos.Get(cid)
	expectNilErr(t, err)
	expect(t, co.ID == 1, "co.ID should be 1")
	expect(t, co.CreatedBy == 1, "co.CreatedBy should be 1")
	// TODO: CreatedAt test
	expect(t, co.LastReplyBy == 1, "co.LastReplyBy should be 1")
	// TODO: LastReplyAt test
	expectIntToBeX(t, co.PostsCount(), 1, "postscount should be 1, not %d")
	expect(t, co.Has(uid), "saturn should be in the conversation")
	expect(t, !co.Has(9999), "uid 9999 should not be in the conversation")
	uids, err := co.Uids()
	expectNilErr(t, err)
	expectIntToBeX(t, len(uids), 2, "uids length should be 2, not %d")
	expectf(t, uids[0] == uid, "uids[0] should be %d, not %d", uid, uids[0])
	expectf(t, uids[1] == 1, "uids[1] should be %d, not %d", 1, uids[1])
	nf(cid, 1)

	expectNilErr(t, c.Convos.Delete(cid))
	expectIntToBeX(t, co.PostsCount(), 0, "postscount should be 0, not %d")
	expect(t, !co.Has(uid), "saturn should not be in a deleted conversation")
	uids, err = co.Uids()
	expectNilErr(t, err)
	expectIntToBeX(t, len(uids), 0, "uids length should be 0, not %d")
	nf(cid, 0)

	// TODO: More tests

	// Block tests

	ok, err := c.UserBlocks.IsBlockedBy(1, 1)
	expectNilErr(t, err)
	expect(t, !ok, "there shouldn't be any blocks")
	ok, err = c.UserBlocks.BulkIsBlockedBy([]int{1}, 1)
	expectNilErr(t, err)
	expect(t, !ok, "there shouldn't be any blocks")
	bf := func(blocker, offset, perPage, expectLen, blockee int) {
		l, err := c.UserBlocks.BlockedByOffset(blocker, offset, perPage)
		expectNilErr(t, err)
		expectf(t, len(l) == expectLen, "there should be %d users blocked by %d not %d", expectLen, blocker, len(l))
		if len(l) > 0 {
			expectf(t, l[0] == blockee, "blocked uid should be %d not %d", blockee, l[0])
		}
	}
	nbf := func(blocker, blockee int) {
		ok, err := c.UserBlocks.IsBlockedBy(1, 2)
		expectNilErr(t, err)
		expect(t, !ok, "there shouldn't be any blocks")
		ok, err = c.UserBlocks.BulkIsBlockedBy([]int{1}, 2)
		expectNilErr(t, err)
		expect(t, !ok, "there shouldn't be any blocks")
		expectIntToBeX(t, c.UserBlocks.BlockedByCount(1), 0, "blockedbycount for 1 should be 1, not %d")
		bf(1, 0, 1, 0, 0)
		bf(1, 0, 15, 0, 0)
		bf(1, 1, 15, 0, 0)
		bf(1, 5, 15, 0, 0)
	}
	nbf(1, 2)

	expectNilErr(t, c.UserBlocks.Add(1, 2))
	ok, err = c.UserBlocks.IsBlockedBy(1, 2)
	expectNilErr(t, err)
	expect(t, ok, "2 should be blocked by 1")
	expectIntToBeX(t, c.UserBlocks.BlockedByCount(1), 1, "blockedbycount for 1 should be 1, not %d")
	bf(1, 0, 1, 1, 2)
	bf(1, 0, 15, 1, 2)
	bf(1, 1, 15, 0, 0)
	bf(1, 5, 15, 0, 0)

	// Double add test
	expectNilErr(t, c.UserBlocks.Add(1, 2))
	ok, err = c.UserBlocks.IsBlockedBy(1, 2)
	expectNilErr(t, err)
	expect(t, ok, "2 should be blocked by 1")
	//expectIntToBeX(t, c.UserBlocks.BlockedByCount(1), 1, "blockedbycount for 1 should be 1, not %d") // todo: fix this
	//bf(1, 0, 1, 1, 2) // todo: fix this
	//bf(1, 0, 15, 1, 2) // todo: fix this
	//bf(1, 1, 15, 0, 0) // todo: fix this
	bf(1, 5, 15, 0, 0)

	expectNilErr(t, c.UserBlocks.Remove(1, 2))
	nbf(1, 2)
	// Double remove test
	expectNilErr(t, c.UserBlocks.Remove(1, 2))
	nbf(1, 2)

	// TODO: Self-block test

	// TODO: More Block tests
}

func TestActivityStream(t *testing.T) {
	miscinit(t)

	expect(t, c.Activity.Count() == 0, "activity stream count should be 0")

	_, err := c.Activity.Get(-1)
	recordMustNotExist(t, err, "activity item -1 shouldn't exist")
	_, err = c.Activity.Get(0)
	recordMustNotExist(t, err, "activity item 0 shouldn't exist")
	_, err = c.Activity.Get(1)
	recordMustNotExist(t, err, "activity item 1 shouldn't exist")

	a := c.Alert{ActorID: 1, TargetUserID: 1, Event: "like", ElementType: "topic", ElementID: 1}
	id, err := c.Activity.Add(a)
	expectNilErr(t, err)
	expect(t, id == 1, "new activity item id should be 1")

	expect(t, c.Activity.Count() == 1, "activity stream count should be 1")
	alert, err := c.Activity.Get(1)
	expectNilErr(t, err)
	expect(t, alert.ActorID == 1, "alert actorid should be 1")
	expect(t, alert.TargetUserID == 1, "alert targetuserid should be 1")
	expect(t, alert.Event == "like", "alert event type should be like")
	expect(t, alert.ElementType == "topic", "alert element type should be topic")
	expect(t, alert.ElementID == 1, "alert element id should be 1")
}

func TestLogs(t *testing.T) {
	miscinit(t)
	gTests := func(s c.LogStore, phrase string) {
		expect(t, s.Count() == 0, "There shouldn't be any "+phrase)
		logs, err := s.GetOffset(0, 25)
		expectNilErr(t, err)
		expect(t, len(logs) == 0, "The log slice should be empty")
	}
	gTests(c.ModLogs, "modlogs")
	gTests(c.AdminLogs, "adminlogs")

	gTests2 := func(s c.LogStore, phrase string) {
		err := s.Create("something", 0, "bumblefly", "::1", 1)
		expectNilErr(t, err)
		count := s.Count()
		expectf(t, count == 1, "store.Count should return one, not %d", count)
		logs, err := s.GetOffset(0, 25)
		recordMustExist(t, err, "We should have at-least one "+phrase)
		expect(t, len(logs) == 1, "The length of the log slice should be one")

		l := logs[0]
		expect(t, l.Action == "something", "l.Action is not something")
		expect(t, l.ElementID == 0, "l.ElementID is not 0")
		expect(t, l.ElementType == "bumblefly", "l.ElementType is not bumblefly")
		expect(t, l.IP == "::1", "l.IP is not ::1")
		expect(t, l.ActorID == 1, "l.ActorID is not 1")
		// TODO: Add a test for log.DoneAt? Maybe throw in some dates and times which are clearly impossible but which may occur due to timezone bugs?
	}
	gTests2(c.ModLogs, "modlog")
	gTests2(c.AdminLogs, "adminlog")
}

// TODO: Add tests for registration logs

func TestPluginManager(t *testing.T) {
	miscinit(t)
	if !c.PluginsInited {
		c.InitPlugins()
	}

	_, ok := c.Plugins["fairy-dust"]
	expect(t, !ok, "Plugin fairy-dust shouldn't exist")
	pl, ok := c.Plugins["bbcode"]
	expect(t, ok, "Plugin bbcode should exist")
	expect(t, !pl.Installable, "Plugin bbcode shouldn't be installable")
	expect(t, !pl.Installed, "Plugin bbcode shouldn't be 'installed'")
	expect(t, !pl.Active, "Plugin bbcode shouldn't be active")
	active, err := pl.BypassActive()
	expectNilErr(t, err)
	expect(t, !active, "Plugin bbcode shouldn't be active in the database either")
	hasPlugin, err := pl.InDatabase()
	expectNilErr(t, err)
	expect(t, !hasPlugin, "Plugin bbcode shouldn't exist in the database")
	// TODO: Add some test cases for SetActive and SetInstalled before calling AddToDatabase

	expectNilErr(t, pl.AddToDatabase(true, false))
	expect(t, !pl.Installable, "Plugin bbcode shouldn't be installable")
	expect(t, !pl.Installed, "Plugin bbcode shouldn't be 'installed'")
	expect(t, pl.Active, "Plugin bbcode should be active")
	active, err = pl.BypassActive()
	expectNilErr(t, err)
	expect(t, active, "Plugin bbcode should be active in the database too")
	hasPlugin, err = pl.InDatabase()
	expectNilErr(t, err)
	expect(t, hasPlugin, "Plugin bbcode should exist in the database")
	expect(t, pl.Init != nil, "Plugin bbcode should have an init function")
	expectNilErr(t, pl.Init(pl))

	expectNilErr(t, pl.SetActive(true))
	expect(t, !pl.Installable, "Plugin bbcode shouldn't be installable")
	expect(t, !pl.Installed, "Plugin bbcode shouldn't be 'installed'")
	expect(t, pl.Active, "Plugin bbcode should still be active")
	active, err = pl.BypassActive()
	expectNilErr(t, err)
	expect(t, active, "Plugin bbcode should still be active in the database too")
	hasPlugin, err = pl.InDatabase()
	expectNilErr(t, err)
	expect(t, hasPlugin, "Plugin bbcode should still exist in the database")

	expectNilErr(t, pl.SetActive(false))
	expect(t, !pl.Installable, "Plugin bbcode shouldn't be installable")
	expect(t, !pl.Installed, "Plugin bbcode shouldn't be 'installed'")
	expect(t, !pl.Active, "Plugin bbcode shouldn't be active")
	active, err = pl.BypassActive()
	expectNilErr(t, err)
	expect(t, !active, "Plugin bbcode shouldn't be active in the database")
	hasPlugin, err = pl.InDatabase()
	expectNilErr(t, err)
	expect(t, hasPlugin, "Plugin bbcode should still exist in the database")
	expect(t, pl.Deactivate != nil, "Plugin bbcode should have an init function")
	pl.Deactivate(pl) // Returns nothing

	// Not installable, should not be mutated
	expect(t, pl.SetInstalled(true) == c.ErrPluginNotInstallable, "Plugin was set as installed despite not being installable")
	expect(t, !pl.Installable, "Plugin bbcode shouldn't be installable")
	expect(t, !pl.Installed, "Plugin bbcode shouldn't be 'installed'")
	expect(t, !pl.Active, "Plugin bbcode shouldn't be active")
	active, err = pl.BypassActive()
	expectNilErr(t, err)
	expect(t, !active, "Plugin bbcode shouldn't be active in the database either")
	hasPlugin, err = pl.InDatabase()
	expectNilErr(t, err)
	expect(t, hasPlugin, "Plugin bbcode should still exist in the database")

	expect(t, pl.SetInstalled(false) == c.ErrPluginNotInstallable, "Plugin was set as not installed despite not being installable")
	expect(t, !pl.Installable, "Plugin bbcode shouldn't be installable")
	expect(t, !pl.Installed, "Plugin bbcode shouldn't be 'installed'")
	expect(t, !pl.Active, "Plugin bbcode shouldn't be active")
	active, err = pl.BypassActive()
	expectNilErr(t, err)
	expect(t, !active, "Plugin bbcode shouldn't be active in the database either")
	hasPlugin, err = pl.InDatabase()
	expectNilErr(t, err)
	expect(t, hasPlugin, "Plugin bbcode should still exist in the database")

	// This isn't really installable, but we want to get a few tests done before getting plugins which are stateful
	pl.Installable = true
	expectNilErr(t, pl.SetInstalled(true))
	expect(t, pl.Installable, "Plugin bbcode should be installable")
	expect(t, pl.Installed, "Plugin bbcode should be 'installed'")
	expect(t, !pl.Active, "Plugin bbcode shouldn't be active")
	active, err = pl.BypassActive()
	expectNilErr(t, err)
	expect(t, !active, "Plugin bbcode shouldn't be active in the database either")
	hasPlugin, err = pl.InDatabase()
	expectNilErr(t, err)
	expect(t, hasPlugin, "Plugin bbcode should still exist in the database")

	expectNilErr(t, pl.SetInstalled(false))
	expect(t, pl.Installable, "Plugin bbcode should be installable")
	expect(t, !pl.Installed, "Plugin bbcode shouldn't be 'installed'")
	expect(t, !pl.Active, "Plugin bbcode shouldn't be active")
	active, err = pl.BypassActive()
	expectNilErr(t, err)
	expect(t, !active, "Plugin bbcode shouldn't be active in the database either")
	hasPlugin, err = pl.InDatabase()
	expectNilErr(t, err)
	expect(t, hasPlugin, "Plugin bbcode should still exist in the database")

	// Bugs sometimes arise when we try to delete a hook when there are multiple, so test for that
	// TODO: Do a finer grained test for that case...? A bigger test might catch more odd cases with multiple plugins
	pl2, ok := c.Plugins["markdown"]
	expect(t, ok, "Plugin markdown should exist")
	expect(t, !pl2.Installable, "Plugin markdown shouldn't be installable")
	expect(t, !pl2.Installed, "Plugin markdown shouldn't be 'installed'")
	expect(t, !pl2.Active, "Plugin markdown shouldn't be active")
	active, err = pl2.BypassActive()
	expectNilErr(t, err)
	expect(t, !active, "Plugin markdown shouldn't be active in the database either")
	hasPlugin, err = pl2.InDatabase()
	expectNilErr(t, err)
	expect(t, !hasPlugin, "Plugin markdown shouldn't exist in the database")

	expectNilErr(t, pl2.AddToDatabase(true, false))
	expectNilErr(t, pl2.Init(pl2))
	expectNilErr(t, pl.SetActive(true))
	expectNilErr(t, pl.Init(pl))
	pl2.Deactivate(pl2)
	expectNilErr(t, pl2.SetActive(false))
	pl.Deactivate(pl)
	expectNilErr(t, pl.SetActive(false))

	// Hook tests
	ht := func() *c.HookTable {
		return c.GetHookTable()
	}
	expect(t, ht().Sshook("haha", "ho") == "ho", "Sshook shouldn't have anything bound to it yet")
	handle := func(in string) (out string) {
		return in + "hi"
	}
	pl.AddHook("haha", handle)
	expect(t, ht().Sshook("haha", "ho") == "hohi", "Sshook didn't give hohi")
	pl.RemoveHook("haha", handle)
	expect(t, ht().Sshook("haha", "ho") == "ho", "Sshook shouldn't have anything bound to it anymore")

	/*expect(t, ht().Hook("haha", "ho") == "ho", "Hook shouldn't have anything bound to it yet")
	handle2 := func(inI interface{}) (out interface{}) {
		return inI.(string) + "hi"
	}
	pl.AddHook("hehe", handle2)
	expect(t, ht().Hook("hehe", "ho").(string) == "hohi", "Hook didn't give hohi")
	pl.RemoveHook("hehe", handle2)
	expect(t, ht().Hook("hehe", "ho").(string) == "ho", "Hook shouldn't have anything bound to it anymore")*/

	// TODO: Add tests for more hook types
}

func TestPhrases(t *testing.T) {
	getPhrase := phrases.GetPermPhrase
	tp := func(name, expects string) {
		res := getPhrase(name)
		expect(t, res == expects, "Not the expected phrase, got '"+res+"' instead")
	}
	tp("BanUsers", "Can ban users")
	tp("NoSuchPerm", "{lang.perms[NoSuchPerm]}")
	tp("ViewTopic", "Can view topics")
	tp("NoSuchPerm", "{lang.perms[NoSuchPerm]}")

	// TODO: Cover the other phrase types, also try switching between languages to see if anything strange happens
}

func TestMetaStore(t *testing.T) {
	m, err := c.Meta.Get("magic")
	expect(t, m == "", "meta var magic should be empty")
	recordMustNotExist(t, err, "meta var magic should not exist")

	err = c.Meta.Set("magic", "lol")
	expectNilErr(t, err)

	m, err = c.Meta.Get("magic")
	expectNilErr(t, err)
	expect(t, m == "lol", "meta var magic should be lol")

	err = c.Meta.Set("magic", "wha")
	expectNilErr(t, err)

	m, err = c.Meta.Get("magic")
	expectNilErr(t, err)
	expect(t, m == "wha", "meta var magic should be wha")

	m, err = c.Meta.Get("giggle")
	expect(t, m == "", "meta var giggle should be empty")
	recordMustNotExist(t, err, "meta var giggle should not exist")
}

func TestPages(t *testing.T) {
	expect(t, c.Pages.Count() == 0, "Page count should be 0")
	_, err := c.Pages.Get(1)
	recordMustNotExist(t, err, "Page 1 should not exist yet")
	expectNilErr(t, c.Pages.Delete(-1))
	expectNilErr(t, c.Pages.Delete(0))
	expectNilErr(t, c.Pages.Delete(1))
	_, err = c.Pages.Get(1)
	recordMustNotExist(t, err, "Page 1 should not exist yet")
	//err = c.Pages.Reload(1)
	//recordMustNotExist(t,err,"Page 1 should not exist yet")

	ipage := c.BlankCustomPage()
	ipage.Name = "test"
	ipage.Title = "Test"
	ipage.Body = "A test page"
	pid, err := ipage.Create()
	expectNilErr(t, err)
	expect(t, pid == 1, "The first page should have an ID of 1")
	expect(t, c.Pages.Count() == 1, "Page count should be 1")

	page, err := c.Pages.Get(1)
	expectNilErr(t, err)
	expect(t, page.Name == ipage.Name, "The page name should be "+ipage.Name)
	expect(t, page.Title == ipage.Title, "The page title should be "+ipage.Title)
	expect(t, page.Body == ipage.Body, "The page body should be "+ipage.Body)

	opage, err := c.Pages.Get(1)
	expectNilErr(t, err)
	opage.Name = "t"
	opage.Title = "T"
	opage.Body = "testing"
	expectNilErr(t, opage.Commit())

	page, err = c.Pages.Get(1)
	expectNilErr(t, err)
	expect(t, page.Name == opage.Name, "The page name should be "+opage.Name)
	expect(t, page.Title == opage.Title, "The page title should be "+opage.Title)
	expect(t, page.Body == opage.Body, "The page body should be "+opage.Body)

	err = c.Pages.Delete(1)
	expectNilErr(t, err)
	expect(t, c.Pages.Count() == 0, "Page count should be 0")
	_, err = c.Pages.Get(1)
	recordMustNotExist(t, err, "Page 1 should not exist")
	//err = c.Pages.Reload(1)
	//recordMustNotExist(t,err,"Page 1 should not exist")

	// TODO: More tests
}

func TestWordFilters(t *testing.T) {
	// TODO: Test the word filters and their store
	expect(t, c.WordFilters.Length() == 0, "Word filter list should be empty")
	expect(t, c.WordFilters.EstCount() == 0, "Word filter list should be empty")
	expect(t, c.WordFilters.Count() == 0, "Word filter list should be empty")
	filters, err := c.WordFilters.GetAll()
	expectNilErr(t, err) // TODO: Slightly confusing that we don't get ErrNoRow here
	expect(t, len(filters) == 0, "Word filter map should be empty")
	// TODO: Add a test for ParseMessage relating to word filters
	_, err = c.WordFilters.Get(1)
	recordMustNotExist(t, err, "filter 1 should not exist")

	wfid, err := c.WordFilters.Create("imbecile", "lovely")
	expectNilErr(t, err)
	expect(t, wfid == 1, "The first word filter should have an ID of 1")
	expect(t, c.WordFilters.Length() == 1, "Word filter list should not be empty")
	expect(t, c.WordFilters.EstCount() == 1, "Word filter list should not be empty")
	expect(t, c.WordFilters.Count() == 1, "Word filter list should not be empty")

	filters, err = c.WordFilters.GetAll()
	expectNilErr(t, err)
	expect(t, len(filters) == 1, "Word filter map should not be empty")
	filter := filters[1]
	expect(t, filter.ID == 1, "Word filter ID should be 1")
	expect(t, filter.Find == "imbecile", "Word filter needle should be imbecile")
	expect(t, filter.Replace == "lovely", "Word filter replacement should be lovely")

	filter, err = c.WordFilters.Get(1)
	expectNilErr(t, err)
	expect(t, filter.ID == 1, "Word filter ID should be 1")
	expect(t, filter.Find == "imbecile", "Word filter needle should be imbecile")
	expect(t, filter.Replace == "lovely", "Word filter replacement should be lovely")

	// Update
	expectNilErr(t, c.WordFilters.Update(1, "b", "a"))

	expect(t, c.WordFilters.Length() == 1, "Word filter list should not be empty")
	expect(t, c.WordFilters.EstCount() == 1, "Word filter list should not be empty")
	expect(t, c.WordFilters.Count() == 1, "Word filter list should not be empty")

	filters, err = c.WordFilters.GetAll()
	expectNilErr(t, err)
	expect(t, len(filters) == 1, "Word filter map should not be empty")
	filter = filters[1]
	expect(t, filter.ID == 1, "Word filter ID should be 1")
	expect(t, filter.Find == "b", "Word filter needle should be b")
	expect(t, filter.Replace == "a", "Word filter replacement should be a")

	filter, err = c.WordFilters.Get(1)
	expectNilErr(t, err)
	expect(t, filter.ID == 1, "Word filter ID should be 1")
	expect(t, filter.Find == "b", "Word filter needle should be imbecile")
	expect(t, filter.Replace == "a", "Word filter replacement should be a")

	// TODO: Add a test for ParseMessage relating to word filters

	err = c.WordFilters.Delete(1)
	expectNilErr(t, err)

	expect(t, c.WordFilters.Length() == 0, "Word filter list should be empty")
	expect(t, c.WordFilters.EstCount() == 0, "Word filter list should be empty")
	expect(t, c.WordFilters.Count() == 0, "Word filter list should be empty")
	filters, err = c.WordFilters.GetAll()
	expectNilErr(t, err) // TODO: Slightly confusing that we don't get ErrNoRow here
	expect(t, len(filters) == 0, "Word filter map should be empty")
	_, err = c.WordFilters.Get(1)
	recordMustNotExist(t, err, "filter 1 should not exist")

	// TODO: Any more tests we could do?
}

func TestMFAStore(t *testing.T) {
	_, err := c.MFAstore.Get(-1)
	recordMustNotExist(t, err, "mfa uid -1 should not exist")
	_, err = c.MFAstore.Get(0)
	recordMustNotExist(t, err, "mfa uid 0 should not exist")
	_, err = c.MFAstore.Get(1)
	recordMustNotExist(t, err, "mfa uid 1 should not exist")

	secret, err := c.GenerateGAuthSecret()
	expectNilErr(t, err)
	expectNilErr(t, c.MFAstore.Create(secret, 1))
	_, err = c.MFAstore.Get(0)
	recordMustNotExist(t, err, "mfa uid 0 should not exist")
	var scratches []string
	it, err := c.MFAstore.Get(1)
	test := func(j int) {
		expectNilErr(t, err)
		expectf(t, it.UID == 1, "UID should be 1 not %d", it.UID)
		expectf(t, it.Secret == secret, "Secret should be '%s' not %s", secret, it.Secret)
		expectf(t, len(it.Scratch) == 8, "Scratch should be 8 not %d", len(it.Scratch))
		for i, scratch := range it.Scratch {
			expectf(t, scratch != "", "scratch %d should not be empty", i)
			if scratches != nil {
				if j == i {
					expectf(t, scratches[i] != scratch, "scratches[%d] should not be %s", i, scratches[i])
				} else {
					expectf(t, scratches[i] == scratch, "scratches[%d] should be %s not %s", i, scratches[i], scratch)
				}
			}
		}
		scratches = make([]string, 8)
		copy(scratches, it.Scratch)
	}
	test(0)
	for i := 0; i < len(scratches); i++ {
		expectNilErr(t, it.BurnScratch(i))
		it, err = c.MFAstore.Get(1)
		test(i)
	}
	token, err := gauth.GetTOTPToken(secret)
	expectNilErr(t, err)
	expectNilErr(t, c.Auth.ValidateMFAToken(token, 1))
	expectNilErr(t, it.Delete())
	_, err = c.MFAstore.Get(-1)
	recordMustNotExist(t, err, "mfa uid -1 should not exist")
	_, err = c.MFAstore.Get(0)
	recordMustNotExist(t, err, "mfa uid 0 should not exist")
	_, err = c.MFAstore.Get(1)
	recordMustNotExist(t, err, "mfa uid 1 should not exist")
}

// TODO: Expand upon the valid characters which can go in URLs?
func TestSlugs(t *testing.T) {
	l := &MEPairList{nil}
	c.Config.BuildSlugs = true // Flip this switch, otherwise all the tests will fail

	l.Add("Unknown", "unknown")
	l.Add("Unknown2", "unknown2")
	l.Add("Unknown ", "unknown")
	l.Add("Unknown 2", "unknown-2")
	l.Add("Unknown  2", "unknown-2")
	l.Add("Admin Alice", "admin-alice")
	l.Add("Admin_Alice", "adminalice")
	l.Add("Admin_Alice-", "adminalice")
	l.Add("-Admin_Alice-", "adminalice")
	l.Add("-Admin@Alice-", "adminalice")
	l.Add("-Admin😀Alice-", "adminalice")
	l.Add("u", "u")
	l.Add("", "untitled")
	l.Add(" ", "untitled")
	l.Add("-", "untitled")
	l.Add("--", "untitled")
	l.Add("é", "é")
	l.Add("-é-", "é")
	l.Add("-你好-", "untitled")
	l.Add("-こにちは-", "untitled")

	for _, item := range l.Items {
		t.Log("Testing string '" + item.Msg + "'")
		res := c.NameToSlug(item.Msg)
		if res != item.Expects {
			t.Error("Bad output:", "'"+res+"'")
			t.Error("Expected:", item.Expects)
		}
	}
}

func TestWidgets(t *testing.T) {
	_, err := c.Widgets.Get(1)
	recordMustNotExist(t, err, "There shouldn't be any widgets by default")
	widgets := c.Docks.RightSidebar.Items
	expectf(t, len(widgets) == 0, "RightSidebar should have 0 items, not %d", len(widgets))

	widget := &c.Widget{Position: 0, Side: "rightSidebar", Type: "simple", Enabled: true, Location: "global"}
	ewidget := &c.WidgetEdit{widget, map[string]string{"Name": "Test", "Text": "Testing"}}
	wid, err := ewidget.Create()
	expectNilErr(t, err)
	expect(t, wid == 1, "wid should be 1")

	wtest := func(w, w2 *c.Widget) {
		expect(t, w.Position == w2.Position, "wrong position")
		expect(t, w.Side == w2.Side, "wrong side")
		expect(t, w.Type == w2.Type, "wrong type")
		expect(t, w2.Enabled, "not enabled")
		expect(t, w.Location == w2.Location, "wrong location")
	}

	// TODO: Do a test for the widget body
	widget2, err := c.Widgets.Get(1)
	expectNilErr(t, err)
	wtest(widget, widget2)

	widgets = c.Docks.RightSidebar.Items
	expectf(t, len(widgets) == 1, "RightSidebar should have 1 item, not %d", len(widgets))
	wtest(widget, widgets[0])

	widget2.Enabled = false
	ewidget = &c.WidgetEdit{widget2, map[string]string{"Name": "Test", "Text": "Testing"}}
	err = ewidget.Commit()
	expectNilErr(t, err)

	widget2, err = c.Widgets.Get(1)
	expectNilErr(t, err)
	expect(t, widget2.Position == widget.Position, "wrong position")
	expect(t, widget2.Side == widget.Side, "wrong side")
	expect(t, widget2.Type == widget.Type, "wrong type")
	expect(t, !widget2.Enabled, "should not be enabled")
	expect(t, widget2.Location == widget.Location, "wrong location")

	widgets = c.Docks.RightSidebar.Items
	expectf(t, len(widgets) == 1, "RightSidebar should have 1 item, not %d", len(widgets))
	expect(t, widgets[0].Position == widget.Position, "wrong position")
	expect(t, widgets[0].Side == widget.Side, "wrong side")
	expect(t, widgets[0].Type == widget.Type, "wrong type")
	expect(t, !widgets[0].Enabled, "should not be enabled")
	expect(t, widgets[0].Location == widget.Location, "wrong location")

	err = widget2.Delete()
	expectNilErr(t, err)

	_, err = c.Widgets.Get(1)
	recordMustNotExist(t, err, "There shouldn't be any widgets anymore")
	widgets = c.Docks.RightSidebar.Items
	expectf(t, len(widgets) == 0, "RightSidebar should have 0 items, not %d", len(widgets))
}

func TestUtils(t *testing.T) {
	email := "test@example.com"
	cemail := c.CanonEmail(email)
	expectf(t, cemail == email, "%s should be %s", cemail, email)
	email = "test.test@example.com"
	cemail = c.CanonEmail(email)
	expectf(t, cemail == email, "%s should be %s", cemail, email)

	ee := func(email, eemail string) {
		cemail := c.CanonEmail(email)
		expectf(t, cemail == eemail, "%s should be %s", cemail, eemail)
	}
	ee("", "")
	ee("ddd", "ddd")
	ee("test.test@gmail.com", "testtest@gmail.com")
	ee("TEST.test@gmail.com", "testtest@gmail.com")
	ee("test.TEST.test@gmail.com", "testtesttest@gmail.com")
	ee("test..TEST.test@gmail.com", "testtesttest@gmail.com")

	email = "TEST.test@example.com"
	lowEmail := strings.ToLower(email)
	cemail = c.CanonEmail(email)
	expectf(t, cemail == lowEmail, "%s should be %s", cemail, lowEmail)

	email = "test.TEST.test@example.com"
	lowEmail = strings.ToLower(email)
	cemail = c.CanonEmail(email)
	expectf(t, cemail == lowEmail, "%s should be %s", cemail, lowEmail)

	// TODO: More utils.go tests
}

func TestWeakPassword(t *testing.T) {
	/*weakPass := func(password, username, email string) func(error,string,...interface{}) {
		err := c.WeakPassword(password, username, email)
		return func(expectErr error, m string, p ...interface{}) {
			m = fmt.Sprintf("pass=%s, user=%s, email=%s ", password, username, email) + m
			expect(t, err == expectErr, fmt.Sprintf(m,p...))
		}
	}*/
	nilErrStr := func(e error) error {
		if e == nil {
			e = errors.New("nil")
		}
		return e
	}
	weakPass := func(password, username, email string) func(error) {
		err := c.WeakPassword(password, username, email)
		e := nilErrStr(err)
		m := fmt.Sprintf("pass=%s, user=%s, email=%s ", password, username, email)
		return func(expectErr error) {
			ee := nilErrStr(expectErr)
			expect(t, err == expectErr, m+fmt.Sprintf("err should be '%s' not '%s'", ee, e))
		}
	}

	//weakPass("test", "test", "test@example.com")(c.ErrWeakPasswordContains,"err should be ErrWeakPasswordContains not '%s'")
	weakPass("", "draw", "test@example.com")(c.ErrWeakPasswordNone)
	weakPass("test", "draw", "test@example.com")(c.ErrWeakPasswordShort)
	weakPass("testtest", "draw", "test@example.com")(c.ErrWeakPasswordContains)
	weakPass("testdraw", "draw", "test@example.com")(c.ErrWeakPasswordNameInPass)
	weakPass("test@example.com", "draw", "test@example.com")(c.ErrWeakPasswordEmailInPass)
	weakPass("meet@example.com2", "draw", "")(c.ErrWeakPasswordNoUpper)
	weakPass("Meet@example.com2", "draw", "")(nil)
	weakPass("test2", "draw", "test@example.com")(c.ErrWeakPasswordShort)
	weakPass("test22222222", "draw", "test@example.com")(c.ErrWeakPasswordContains)
	weakPass("superman", "draw", "test@example.com")(c.ErrWeakPasswordCommon)
	weakPass("Superman", "draw", "test@example.com")(c.ErrWeakPasswordCommon)
	weakPass("Superma2", "draw", "test@example.com")(nil)
	weakPass("superman2", "draw", "test@example.com")(c.ErrWeakPasswordCommon)
	weakPass("Superman2", "draw", "test@example.com")(c.ErrWeakPasswordCommon)
	weakPass("superman22", "draw", "test@example.com")(c.ErrWeakPasswordNoUpper)
	weakPass("K\\@<^s}1", "draw", "test@example.com")(nil)
	weakPass("K\\@<^s}r", "draw", "test@example.com")(c.ErrWeakPasswordNoNumbers)
	weakPass("k\\@<^s}1", "draw", "test@example.com")(c.ErrWeakPasswordNoUpper)
	weakPass("aaaaaaaa", "draw", "test@example.com")(c.ErrWeakPasswordNoUpper)
	weakPass("aA1aA1aA1", "draw", "test@example.com")(c.ErrWeakPasswordUniqueChars)
	weakPass("abababab", "draw", "test@example.com")(c.ErrWeakPasswordNoUpper)
	weakPass("11111111111111111111", "draw", "test@example.com")(c.ErrWeakPasswordNoUpper)
	weakPass("aaaaaaaaaaAAAAAAAAAA", "draw", "test@example.com")(c.ErrWeakPasswordUniqueChars)
	weakPass("-:u/nMxb,A!n=B;H\\sjM", "draw", "test@example.com")(nil)
}

func TestAuth(t *testing.T) {
	// bcrypt likes doing stupid things, so this test will probably fail
	realPassword := "Madame Cassandra's Mystic Orb"
	t.Logf("Set realPassword to '%s'", realPassword)
	t.Log("Hashing the real password with bcrypt")
	hashedPassword, _, err := c.BcryptGeneratePassword(realPassword)
	if err != nil {
		t.Error(err)
	}
	passwordTest(t, realPassword, hashedPassword)
	// TODO: Peek at the prefix to verify this is a bcrypt hash

	t.Log("Hashing the real password")
	hashedPassword2, _, err := c.GeneratePassword(realPassword)
	if err != nil {
		t.Error(err)
	}
	passwordTest(t, realPassword, hashedPassword2)
	// TODO: Peek at the prefix to verify this is a bcrypt hash

	_, err, _ = c.Auth.Authenticate("None", "password")
	errmsg := "Username None shouldn't exist"
	if err != nil {
		errmsg += "\n" + err.Error()
	}
	expect(t, err == c.ErrNoUserByName, errmsg)

	uid, err, _ := c.Auth.Authenticate("Admin", "password")
	expectNilErr(t, err)
	expectf(t, uid == 1, "Default admin uid should be 1 not %d", uid)

	_, err, _ = c.Auth.Authenticate("Sam", "ReallyBadPassword")
	errmsg = "Username Sam shouldn't exist"
	if err != nil {
		errmsg += "\n" + err.Error()
	}
	expect(t, err == c.ErrNoUserByName, errmsg)

	admin, err := c.Users.Get(1)
	expectNilErr(t, err)
	// TODO: Move this into the user store tests to provide better coverage? E.g. To see if the installer and the user creator initialise the field differently
	expect(t, admin.Session == "", "Admin session should be blank")

	session, err := c.Auth.CreateSession(1)
	expectNilErr(t, err)
	expect(t, session != "", "Admin session shouldn't be blank")
	// TODO: Test the actual length set in the setting in addition to this "too short" test
	// TODO: We might be able to push up this minimum requirement
	expect(t, len(session) > 10, "Admin session shouldn't be too short")
	expect(t, admin.Session != session, "Old session should not match new one")
	admin, err = c.Users.Get(1)
	expectNilErr(t, err)
	expect(t, admin.Session == session, "Sessions should match")

	// TODO: Create a user with a unicode password and see if we can login as them
	// TODO: Tests for SessionCheck, GetCookies, and ForceLogout
	// TODO: Tests for MFA Verification
}

// TODO: Vary the salts? Keep in mind that some algorithms store the salt in the hash therefore the salt string may be blank
func passwordTest(t *testing.T, realPassword, hashedPassword string) {
	if len(hashedPassword) < 10 {
		t.Error("Hash too short")
	}
	salt := ""
	password := realPassword
	t.Logf("Testing password '%s'", password)
	t.Logf("Testing salt '%s'", salt)
	err := c.CheckPassword(hashedPassword, password, salt)
	if err == c.ErrMismatchedHashAndPassword {
		t.Error("The two don't match")
	} else if err == c.ErrPasswordTooLong {
		t.Error("CheckPassword thinks the password is too long")
	} else if err != nil {
		t.Error(err)
	}

	password = "hahaha"
	t.Logf("Testing password '%s'", password)
	t.Logf("Testing salt '%s'", salt)
	err = c.CheckPassword(hashedPassword, password, salt)
	if err == c.ErrPasswordTooLong {
		t.Error("CheckPassword thinks the password is too long")
	} else if err == nil {
		t.Error("The two shouldn't match!")
	}

	password = "Madame Cassandra's Mystic"
	t.Logf("Testing password '%s'", password)
	t.Logf("Testing salt '%s'", salt)
	err = c.CheckPassword(hashedPassword, password, salt)
	expect(t, err != c.ErrPasswordTooLong, "CheckPassword thinks the password is too long")
	expect(t, err != nil, "The two shouldn't match!")
}

func TestUserPrivacy(t *testing.T) {
	pu, u := c.BlankUser(), &c.GuestUser
	pu.ID = 1

	var msg string
	test := func(expects bool, level int) {
		pu.Privacy.ShowComments = level
		val := c.PrivacyCommentsShow(pu, u)
		var bit string
		if !expects {
			bit = " not"
			val = !val
		}
		expectf(t, val, "%s should%s be able to see comments on level %d", msg, bit, level)
	}
	// 0 = default, 1 = public, 2 = registered, 3 = friends, 4 = self, 5 = disabled

	msg = "guest users"
	test(true, 0)
	test(true, 1)
	test(false, 2)
	test(false, 3)
	test(false, 4)
	test(false, 5)

	u = c.BlankUser()
	msg = "blank users"
	test(true, 0)
	test(true, 1)
	test(false, 2)
	//test(false,3)
	test(false, 4)
	test(false, 5)

	u.Loggedin = true
	msg = "registered users"
	test(true, 0)
	test(true, 1)
	test(true, 2)
	test(false, 3)
	test(false, 4)
	test(false, 5)

	u.IsBanned = true
	msg = "banned users"
	test(true, 0)
	test(true, 1)
	test(true, 2)
	test(false, 3)
	test(false, 4)
	test(false, 5)
	u.IsBanned = false

	u.IsMod = true
	msg = "mods"
	test(true, 0)
	test(true, 1)
	test(true, 2)
	test(false, 3)
	test(false, 4)
	test(false, 5)
	u.IsMod = false

	u.IsSuperMod = true
	msg = "super mods"
	test(true, 0)
	test(true, 1)
	test(true, 2)
	test(false, 3)
	test(false, 4)
	test(false, 5)
	u.IsSuperMod = false

	u.IsAdmin = true
	msg = "admins"
	test(true, 0)
	test(true, 1)
	test(true, 2)
	test(false, 3)
	test(false, 4)
	test(false, 5)
	u.IsAdmin = false

	u.IsSuperAdmin = true
	msg = "super admins"
	test(true, 0)
	test(true, 1)
	test(true, 2)
	test(false, 3)
	test(false, 4)
	test(false, 5)
	u.IsSuperAdmin = false

	u.ID = 1
	test(true, 0)
	test(true, 1)
	test(true, 2)
	test(true, 3)
	test(true, 4)
	test(false, 5)
}

type METri struct {
	Name    string // Optional, this is here for tests involving invisible characters so we know what's going in
	Msg     string
	Expects string
}

type METriList struct {
	Items []METri
}

func (l *METriList) Add(args ...string) {
	if len(args) < 2 {
		panic("need 2 or more args")
	}
	if len(args) > 2 {
		l.Items = append(l.Items, METri{args[0], args[1], args[2]})
	} else {
		l.Items = append(l.Items, METri{"", args[0], args[1]})
	}
}

type CountTest struct {
	Name    string
	Msg     string
	Expects int
}

type CountTestList struct {
	Items []CountTest
}

func (l *CountTestList) Add(name, msg string, expects int) {
	l.Items = append(l.Items, CountTest{name, msg, expects})
}

func TestWordCount(t *testing.T) {
	l := &CountTestList{nil}
	l.Add("blank", "", 0)
	l.Add("single-letter", "h", 1)
	l.Add("single-kana", "お", 1)
	l.Add("single-letter-words", "h h", 2)
	l.Add("two-letter", "h", 1)
	l.Add("two-kana", "おは", 1)
	l.Add("two-letter-words", "hh hh", 2)
	l.Add("", "h,h", 2)
	l.Add("", "h,,h", 2)
	l.Add("", "h, h", 2)
	l.Add("", "  h, h", 2)
	l.Add("", "h, h  ", 2)
	l.Add("", "  h, h  ", 2)
	l.Add("", "h,  h", 2)
	l.Add("", "h\nh", 2)
	l.Add("", "h\"h", 2)
	l.Add("", "h[r]h", 3)
	l.Add("", "お,お", 2)
	l.Add("", "お、お", 2)
	l.Add("", "お\nお", 2)
	l.Add("", "お”お", 2)
	l.Add("", "お「あ」お", 3)

	for _, item := range l.Items {
		res := c.WordCount(item.Msg)
		if res != item.Expects {
			if item.Name != "" {
				t.Error("Name: ", item.Name)
			}
			t.Error("Testing string '" + item.Msg + "'")
			t.Error("Bad output:", res)
			t.Error("Expected:", item.Expects)
		}
	}
}

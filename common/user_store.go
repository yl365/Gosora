package common

import (
	"database/sql"
	"errors"
	"strconv"

	qgen "github.com/Azareal/Gosora/query_gen"
	"golang.org/x/crypto/bcrypt"
)

// TODO: Add the watchdog goroutine
// TODO: Add some sort of update method
var Users UserStore
var ErrAccountExists = errors.New("this username is already in use")
var ErrLongUsername = errors.New("this username is too long")
var ErrSomeUsersNotFound = errors.New("Unable to find some users")

type UserStore interface {
	DirtyGet(id int) *User
	Get(id int) (*User, error)
	Getn(id int) *User
	GetByName(name string) (*User, error)
	BulkGetByName(names []string) (list []*User, err error)
	RawBulkGetByNameForConvo(f func(int, string, int, bool, int, int) error, names []string) error
	Exists(id int) bool
	SearchOffset(name, email string, gid, offset, perPage int) (users []*User, err error)
	GetOffset(offset, perPage int) ([]*User, error)
	Each(f func(*User) error) error
	//BulkGet(ids []int) ([]*User, error)
	BulkGetMap(ids []int) (map[int]*User, error)
	BypassGet(id int) (*User, error)
	Create(name, password, email string, group int, active bool) (int, error)
	Reload(id int) error
	Count() int
	CountSearch(name, email string, gid int) int

	SetCache(cache UserCache)
	GetCache() UserCache
}

type DefaultUserStore struct {
	cache UserCache

	get          *sql.Stmt
	getByName    *sql.Stmt
	searchOffset *sql.Stmt
	getOffset    *sql.Stmt
	getAll       *sql.Stmt
	exists       *sql.Stmt
	register     *sql.Stmt
	nameExists   *sql.Stmt

	count       *sql.Stmt
	countSearch *sql.Stmt
}

// NewDefaultUserStore gives you a new instance of DefaultUserStore
func NewDefaultUserStore(cache UserCache) (*DefaultUserStore, error) {
	acc := qgen.NewAcc()
	if cache == nil {
		cache = NewNullUserCache()
	}
	u := "users"
	allCols := "uid,name,group,active,is_super_admin,session,email,avatar,message,level,score,posts,liked,last_ip,temp_group,createdAt,enable_embeds,profile_comments,who_can_convo"
	// TODO: Add an admin version of registerStmt with more flexibility?
	return &DefaultUserStore{
		cache: cache,

		get:          acc.Select(u).Columns("name,group,active,is_super_admin,session,email,avatar,message,level,score,posts,liked,last_ip,temp_group,createdAt,enable_embeds,profile_comments,who_can_convo").Where("uid=?").Prepare(),
		getByName:    acc.Select(u).Columns(allCols).Where("name=?").Prepare(),
		searchOffset: acc.Select(u).Columns(allCols).Where("(name=? OR ?='') AND (email=? OR ?='') AND (group=? OR ?=0)").Orderby("uid ASC").Limit("?,?").Prepare(),
		getOffset:    acc.Select(u).Columns(allCols).Orderby("uid ASC").Limit("?,?").Prepare(),
		getAll:       acc.Select(u).Columns(allCols).Prepare(),

		exists:     acc.Exists(u, "uid").Prepare(),
		register:   acc.Insert(u).Columns("name,email,password,salt,group,is_super_admin,session,active,message,createdAt,lastActiveAt,lastLiked,oldestItemLikedCreatedAt").Fields("?,?,?,?,?,0,'',?,'',UTC_TIMESTAMP(),UTC_TIMESTAMP(),UTC_TIMESTAMP(),UTC_TIMESTAMP()").Prepare(), // TODO: Implement user_count on users_groups here
		nameExists: acc.Exists(u, "name").Prepare(),

		count:       acc.Count(u).Prepare(),
		countSearch: acc.Count(u).Where("(name LIKE ('%'+?+'%') OR ?='') AND (email=? OR ?='') AND (group=? OR ?=0)").Prepare(),
	}, acc.FirstError()
}

func (s *DefaultUserStore) DirtyGet(id int) *User {
	user, err := s.Get(id)
	if err == nil {
		return user
	}
	/*if s.OutOfBounds(id) {
		return BlankUser()
	}*/
	return BlankUser()
}

func (s *DefaultUserStore) scanUser(r *sql.Row, u *User) (embeds int, err error) {
	e := r.Scan(&u.Name, &u.Group, &u.Active, &u.IsSuperAdmin, &u.Session, &u.Email, &u.RawAvatar, &u.Message, &u.Level, &u.Score, &u.Posts, &u.Liked, &u.LastIP, &u.TempGroup, &u.CreatedAt, &embeds, &u.Privacy.ShowComments, &u.Privacy.AllowMessage)
	return embeds, e
}

// TODO: Log weird cache errors? Not just here but in every *Cache?
func (s *DefaultUserStore) Get(id int) (*User, error) {
	u, err := s.cache.Get(id)
	if err == nil {
		//log.Print("cached user")
		//log.Print(string(debug.Stack()))
		//log.Println("")
		return u, nil
	}
	//log.Print("uncached user")

	u = &User{ID: id, Loggedin: true}
	embeds, err := s.scanUser(s.get.QueryRow(id), u)
	if err == nil {
		if embeds != -1 {
			u.ParseSettings = DefaultParseSettings.CopyPtr()
			u.ParseSettings.NoEmbed = embeds == 0
		}
		u.Init()
		s.cache.Set(u)
	}
	return u, err
}

func (s *DefaultUserStore) Getn(id int) *User {
	u := s.cache.Getn(id)
	if u != nil {
		return u
	}

	u = &User{ID: id, Loggedin: true}
	embeds, err := s.scanUser(s.get.QueryRow(id), u)
	if err != nil {
		return nil
	}
	if embeds != -1 {
		u.ParseSettings = DefaultParseSettings.CopyPtr()
		u.ParseSettings.NoEmbed = embeds == 0
	}
	u.Init()
	s.cache.Set(u)
	return u
}

// TODO: Log weird cache errors? Not just here but in every *Cache?
// ! This bypasses the cache, use frugally
func (s *DefaultUserStore) GetByName(name string) (*User, error) {
	u := &User{Loggedin: true}
	var embeds int
	err := s.getByName.QueryRow(name).Scan(&u.ID, &u.Name, &u.Group, &u.Active, &u.IsSuperAdmin, &u.Session, &u.Email, &u.RawAvatar, &u.Message, &u.Level, &u.Score, &u.Posts, &u.Liked, &u.LastIP, &u.TempGroup, &u.CreatedAt, &embeds, &u.Privacy.ShowComments, &u.Privacy.AllowMessage)
	if err != nil {
		return nil, err
	}
	if embeds != -1 {
		u.ParseSettings = DefaultParseSettings.CopyPtr()
		u.ParseSettings.NoEmbed = embeds == 0
	}
	u.Init()
	s.cache.Set(u)
	return u, nil
}

// TODO: Optimise the query to avoid preparing it on the spot? Maybe, use knowledge of the most common IN() parameter counts?
// ! This bypasses the cache, use frugally
func (s *DefaultUserStore) BulkGetByName(names []string) (list []*User, err error) {
	if len(names) == 0 {
		return list, nil
	} else if len(names) == 1 {
		user, err := s.GetByName(names[0])
		if err != nil {
			return list, err
		}
		return []*User{user}, nil
	}

	idList, q := inqbuildstr(names)
	rows, err := qgen.NewAcc().Select("users").Columns("uid,name,group,active,is_super_admin,session,email,avatar,message,level,score,posts,liked,last_ip,temp_group,createdAt,enable_embeds,profile_comments,who_can_convo").Where("name IN(" + q + ")").Query(idList...)
	if err != nil {
		return list, err
	}
	defer rows.Close()

	var embeds int
	for rows.Next() {
		u := &User{Loggedin: true}
		err := rows.Scan(&u.ID, &u.Name, &u.Group, &u.Active, &u.IsSuperAdmin, &u.Session, &u.Email, &u.RawAvatar, &u.Message, &u.Level, &u.Score, &u.Posts, &u.Liked, &u.LastIP, &u.TempGroup, &u.CreatedAt, &embeds, &u.Privacy.ShowComments, &u.Privacy.AllowMessage)
		if err != nil {
			return list, err
		}
		if embeds != -1 {
			u.ParseSettings = DefaultParseSettings.CopyPtr()
			u.ParseSettings.NoEmbed = embeds == 0
		}
		u.Init()
		s.cache.Set(u)
		list[u.ID] = u
	}
	if err = rows.Err(); err != nil {
		return list, err
	}

	// Did we miss any users?
	if len(names) > len(list) {
		return list, ErrSomeUsersNotFound
	}
	return list, err
}

// Special case function for efficiency
func (s *DefaultUserStore) RawBulkGetByNameForConvo(f func(int, string, int, bool, int, int) error, names []string) error {
	idList, q := inqbuildstr(names)
	rows, e := qgen.NewAcc().Select("users").Columns("uid,name,group,is_super_admin,temp_group,who_can_convo").Where("name IN(" + q + ")").Query(idList...)
	if e != nil {
		return e
	}
	defer rows.Close()
	for rows.Next() {
		var name string
		var id, group, temp_group, who_can_convo int
		var super_admin bool
		if e = rows.Scan(&id, &name, &group, &super_admin, &temp_group, &who_can_convo); e != nil {
			return e
		}
		if e = f(id, name, group, super_admin, temp_group, who_can_convo); e != nil {
			return e
		}
	}
	return rows.Err()
}

// TODO: Optimise this, so we don't wind up hitting the database every-time for small gaps
// TODO: Make this a little more consistent with DefaultGroupStore's GetRange method
func (s *DefaultUserStore) GetOffset(offset, perPage int) (users []*User, err error) {
	rows, err := s.getOffset.Query(offset, perPage)
	if err != nil {
		return users, err
	}
	defer rows.Close()

	var embeds int
	for rows.Next() {
		u := &User{Loggedin: true}
		err := rows.Scan(&u.ID, &u.Name, &u.Group, &u.Active, &u.IsSuperAdmin, &u.Session, &u.Email, &u.RawAvatar, &u.Message, &u.Level, &u.Score, &u.Posts, &u.Liked, &u.LastIP, &u.TempGroup, &u.CreatedAt, &embeds, &u.Privacy.ShowComments, &u.Privacy.AllowMessage)
		if err != nil {
			return nil, err
		}
		if embeds != -1 {
			u.ParseSettings = DefaultParseSettings.CopyPtr()
			u.ParseSettings.NoEmbed = embeds == 0
		}
		u.Init()
		s.cache.Set(u)
		users = append(users, u)
	}
	return users, rows.Err()
}
func (s *DefaultUserStore) SearchOffset(name, email string, gid, offset, perPage int) (users []*User, err error) {
	rows, err := s.searchOffset.Query(name, name, email, email, gid, gid, offset, perPage)
	if err != nil {
		return users, err
	}
	defer rows.Close()

	var embeds int
	for rows.Next() {
		u := &User{Loggedin: true}
		err := rows.Scan(&u.ID, &u.Name, &u.Group, &u.Active, &u.IsSuperAdmin, &u.Session, &u.Email, &u.RawAvatar, &u.Message, &u.Level, &u.Score, &u.Posts, &u.Liked, &u.LastIP, &u.TempGroup, &u.CreatedAt, &embeds, &u.Privacy.ShowComments, &u.Privacy.AllowMessage)
		if err != nil {
			return nil, err
		}
		if embeds != -1 {
			u.ParseSettings = DefaultParseSettings.CopyPtr()
			u.ParseSettings.NoEmbed = embeds == 0
		}
		u.Init()
		s.cache.Set(u)
		users = append(users, u)
	}
	return users, rows.Err()
}
func (s *DefaultUserStore) Each(f func(*User) error) error {
	rows, e := s.getAll.Query()
	if e != nil {
		return e
	}
	defer rows.Close()
	var embeds int
	for rows.Next() {
		u := new(User)
		if e := rows.Scan(&u.ID, &u.Name, &u.Group, &u.Active, &u.IsSuperAdmin, &u.Session, &u.Email, &u.RawAvatar, &u.Message, &u.Level, &u.Score, &u.Posts, &u.Liked, &u.LastIP, &u.TempGroup, &u.CreatedAt, &embeds, &u.Privacy.ShowComments, &u.Privacy.AllowMessage); e != nil {
			return e
		}
		if embeds != -1 {
			u.ParseSettings = DefaultParseSettings.CopyPtr()
			u.ParseSettings.NoEmbed = embeds == 0
		}
		u.Init()
		if e := f(u); e != nil {
			return e
		}
	}
	return rows.Err()
}

// TODO: Optimise the query to avoid preparing it on the spot? Maybe, use knowledge of the most common IN() parameter counts?
// TODO: ID of 0 should always error?
func (s *DefaultUserStore) BulkGetMap(ids []int) (list map[int]*User, err error) {
	idCount := len(ids)
	list = make(map[int]*User)
	if idCount == 0 {
		return list, nil
	}

	var stillHere []int
	sliceList := s.cache.BulkGet(ids)
	if len(sliceList) > 0 {
		for i, sliceItem := range sliceList {
			if sliceItem != nil {
				list[sliceItem.ID] = sliceItem
			} else {
				stillHere = append(stillHere, ids[i])
			}
		}
		ids = stillHere
	}

	// If every user is in the cache, then return immediately
	if len(ids) == 0 {
		return list, nil
	} else if len(ids) == 1 {
		user, err := s.Get(ids[0])
		if err != nil {
			return list, err
		}
		list[user.ID] = user
		return list, nil
	}

	idList, q := inqbuild(ids)
	rows, err := qgen.NewAcc().Select("users").Columns("uid,name,group,active,is_super_admin,session,email,avatar,message,level,score,posts,liked,last_ip,temp_group,createdAt,enable_embeds,profile_comments,who_can_convo").Where("uid IN(" + q + ")").Query(idList...)
	if err != nil {
		return list, err
	}
	defer rows.Close()

	var embeds int
	for rows.Next() {
		u := &User{Loggedin: true}
		err := rows.Scan(&u.ID, &u.Name, &u.Group, &u.Active, &u.IsSuperAdmin, &u.Session, &u.Email, &u.RawAvatar, &u.Message, &u.Level, &u.Score, &u.Posts, &u.Liked, &u.LastIP, &u.TempGroup, &u.CreatedAt, &embeds, &u.Privacy.ShowComments, &u.Privacy.AllowMessage)
		if err != nil {
			return list, err
		}
		if embeds != -1 {
			u.ParseSettings = DefaultParseSettings.CopyPtr()
			u.ParseSettings.NoEmbed = embeds == 0
		}
		u.Init()
		s.cache.Set(u)
		list[u.ID] = u
	}
	if err = rows.Err(); err != nil {
		return list, err
	}

	// Did we miss any users?
	if idCount > len(list) {
		var sidList string
		for _, id := range ids {
			_, ok := list[id]
			if !ok {
				sidList += strconv.Itoa(id) + ","
			}
		}
		if sidList != "" {
			sidList = sidList[0 : len(sidList)-1]
			err = errors.New("Unable to find users with the following IDs: " + sidList)
		}
	}

	return list, err
}

func (s *DefaultUserStore) BypassGet(id int) (*User, error) {
	u := &User{ID: id, Loggedin: true}
	embeds, err := s.scanUser(s.get.QueryRow(id), u)
	if err == nil {
		if embeds != -1 {
			u.ParseSettings = DefaultParseSettings.CopyPtr()
			u.ParseSettings.NoEmbed = embeds == 0
		}
		u.Init()
	}
	return u, err
}

func (s *DefaultUserStore) Reload(id int) error {
	u, err := s.BypassGet(id)
	if err != nil {
		s.cache.Remove(id)
		return err
	}
	_ = s.cache.Set(u)
	TopicListThaw.Thaw()
	return nil
}

func (s *DefaultUserStore) Exists(id int) bool {
	err := s.exists.QueryRow(id).Scan(&id)
	if err != nil && err != ErrNoRows {
		LogError(err)
	}
	return err != ErrNoRows
}

// TODO: Change active to a bool?
// TODO: Use unique keys for the usernames
func (s *DefaultUserStore) Create(name, password, email string, group int, active bool) (int, error) {
	// TODO: Strip spaces?

	// ? This number might be a little screwy with Unicode, but it's the only consistent thing we have, as Unicode characters can be any number of bytes in theory?
	if len(name) > Config.MaxUsernameLength {
		return 0, ErrLongUsername
	}

	// Is this name already taken..?
	err := s.nameExists.QueryRow(name).Scan(&name)
	if err != ErrNoRows {
		return 0, ErrAccountExists
	}
	salt, err := GenerateSafeString(SaltLength)
	if err != nil {
		return 0, err
	}
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password+salt), bcrypt.DefaultCost)
	if err != nil {
		return 0, err
	}

	res, err := s.register.Exec(name, email, string(hashedPassword), salt, group, active)
	if err != nil {
		return 0, err
	}
	lastID, err := res.LastInsertId()
	return int(lastID), err
}

// Count returns the total number of users registered on the forums
func (s *DefaultUserStore) Count() (count int) {
	return Countf(s.count)
}

func (s *DefaultUserStore) CountSearch(name, email string, gid int) (count int) {
	return Countf(s.countSearch, name, name, email, email, gid, gid)
}

func (s *DefaultUserStore) SetCache(cache UserCache) {
	s.cache = cache
}

// TODO: We're temporarily doing this so that you can do ucache != nil in getTopicUser. Refactor it.
func (s *DefaultUserStore) GetCache() UserCache {
	_, ok := s.cache.(*NullUserCache)
	if ok {
		return nil
	}
	return s.cache
}

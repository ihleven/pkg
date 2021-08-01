package fs

import (
	"fmt"
	"os/user"

	"github.com/ihleven/cloud11-api/auth"
	"github.com/ihleven/cloud11-api/drive"
)

var uid = map[string]uint32{"matt": 501}

var gid = map[string]uint32{"matt": 20}

func GetUidGidByUsername(username string) (uint32, uint32) {
	return uid[username], gid[username]
}

func GetUidGidByAccount(account *auth.Account) (uint32, uint32) {
	return uid[account.Username], gid[account.Username]
}

var cachedUsers = make(map[uint32]*drive.User)
var cachedGroups = map[uint32]*drive.Group{}

func GetUserByID(uid uint32) *drive.User {
	userPtr, ok := cachedUsers[uid]
	if !ok {
		usr, err := user.LookupId(fmt.Sprint(uid))
		if err != nil {
			switch err.(type) {
			case user.UnknownUserIdError:
				userPtr = &drive.User{Username: "unknown"}
			default:
				userPtr = &drive.User{Username: "unknown"}
			}
			return nil
		}
		cachedUsers[uid] = &drive.User{Uid: usr.Uid, Gid: usr.Gid, Username: usr.Username, Name: usr.Name, HomeDir: usr.HomeDir}
		userPtr = cachedUsers[uid]
	}
	return userPtr
}

func GetGroupByID(gid uint32) *drive.Group {

	grpPtr, ok := cachedGroups[gid]
	if !ok {
		userGrp, err := user.LookupGroupId(fmt.Sprint(gid))
		if err != nil {
			fmt.Printf(" * GetGroupByID(%s) => error looking up group: %v\n", string(gid), err)
			return nil
		}
		cachedGroups[gid] = &drive.Group{userGrp.Gid, userGrp.Name}
		grpPtr = cachedGroups[gid]
	}
	return grpPtr
}

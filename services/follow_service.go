package services

import (
	"douyin/daos"
	"douyin/models"
	"douyin/tools"
	"errors"
	"strconv"
	"strings"
	"sync"
)

var (
	followService     FollowService
	followServiceOnce sync.Once
	//管理redis中关注数的hash名
	cacheHashRead  = "follow_hash_one" //写入使用的变量
	cacheHashWrite = "follow_hash_two" //读出使用的变量
)

type FollowService interface {
	Action(userId string, toUserId string, actionType string) error
	FollowListCdRedis(userId string) error
	FollowerListCdRedis(userId string) error
}
type FollowServiceImpl struct {
	followDao daos.FollowDao
}

func GetFollowService() FollowService {
	followServiceOnce.Do(func() {
		followService = &FollowServiceImpl{
			followDao: daos.GetFollowDao(),
		}
	})
	return followService
}

func (f *FollowServiceImpl) Action(userId string, toUserId string, actionType string) error {
	//关注发起者:userId,被关注者:toUserId
	uid, err := strconv.Atoi(userId)
	if err != nil {
		return err
	}
	tuid, err := strconv.Atoi(toUserId)
	if err != nil {
		return err
	}
	follow := &models.Follow{FollowerId: uid, FollowId: tuid}
	switch actionType {
	case "1":
		{
			_, err = f.followDao.FindFollow(toUserId, userId)
			if err == nil {
				return errors.New("follow exist")

			}
			err = f.followDao.AddFollow(follow)
			if err != nil {
				return err
			}
			//清除redis的userId_toUserId关注操作缓存
			//...待补充
		}
	case "2":
		err = f.followDao.DelFollow(follow)
		if err != nil {
			return err
		}
		//清除redis的userId_toUserId取关操作缓存
		//...待补充
	}
	return nil
}

//@author cwh
//将用户的关注列表缓存进redis（无缓存的情况下）
func (f *FollowServiceImpl) FollowListCdRedis(userId string) error {
	followKey := GetFollowKey(userId)
	if !tools.RedisKeyExists(followKey) {
		return tools.RedisKeyFlush(followKey)
	}
	id, _ := strconv.Atoi(userId)
	follows, err := f.followDao.UserFollow(id)
	if err != nil {
		return err
	}
	for _, value := range follows {
		//sadd userId followId
		_ = tools.RedisDoKV("SADD", followKey, value.FollowId)
	}
	//设置半小时有效期
	_ = tools.RedisDoKV("EXPIRE", followKey, 1800)
	return nil
}

//@author cwh
//将用户的粉丝列表缓存进redis（无缓存的情况下）
func (f *FollowServiceImpl) FollowerListCdRedis(userId string) error {
	followerKey := GetFollowerKey(userId)
	if !tools.RedisKeyExists(followerKey) {
		return tools.RedisKeyFlush(followerKey)
	}
	id, _ := strconv.Atoi(userId)
	follower, err := f.followDao.UserFollow(id)
	if err != nil {
		return err
	}
	_ = tools.RedisDoKV("SADD", GetFollowWrite(), userId)
	for _, value := range follower {
		//sadd userId followId
		_ = tools.RedisDoKV("SADD", followerKey, value.FollowId)
	}
	//设置半小时有效期
	_ = tools.RedisDoKV("EXPIRE", followerKey, 1800)
	return nil
}

func ReHashKey() {
	cacheHashWrite, cacheHashRead = cacheHashRead, cacheHashWrite
}

func GetFollowWrite() string {
	return cacheHashWrite
}

func GetFollowRead() string {
	return cacheHashRead
}

func GetFollowKey(userId string) string {
	return strings.Join([]string{userId, "follow"}, "_")
}

func GetFollowerKey(userId string) string {
	return strings.Join([]string{userId, "follow"}, "_")
}

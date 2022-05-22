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
	RedisAction(userId, toUserId, actionType string) error
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
//关注操作时，对redis的操作
func (f *FollowServiceImpl) RedisAction(userId, toUserId, actionType string) error {
	//查询对应缓存，没有便加载
	err := f.followListCdRedis(userId)
	_ = f.followerListCdRedis(toUserId)
	if err != nil {
		return errors.New("缓存错误！")
	}

	var action string
	var add int
	if actionType == "1" {
		//关注操作，关注列表添加touserId，粉丝列表添加userId，粉丝数+1
		action = "SADD"
		add = 1
	} else if actionType == "2" {
		//取关操作，关注列表删除toUserId，粉丝列表删除userId，粉丝数-1
		action = "SREM"
		add = -1
	} else {
		return errors.New("操作类型错误！")
	}
	//用户关注列表更新
	if tools.RedisDoKV(action, getFollowKey(userId), toUserId) != nil {
		return errors.New("系统错误！，请稍后重试")
	}
	//被关注者，粉丝列表更新
	followerKey := getFollowerKey(toUserId)
	_ = tools.RedisDoKV(action, followerKey, userId)
	_ = tools.RedisDoHash("HINCRBY", cacheHashWrite, followerKey, add)
	return nil
}

//@author cwh
//将用户的关注列表缓存进redis（无缓存的情况下）
func (f *FollowServiceImpl) followListCdRedis(userId string) error {
	followKey := getFollowKey(userId)
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
func (f *FollowServiceImpl) followerListCdRedis(userId string) error {
	followerKey := getFollowerKey(userId)
	if !tools.RedisKeyExists(followerKey) {
		return tools.RedisKeyFlush(followerKey)
	}
	id, _ := strconv.Atoi(userId)
	follower, err := f.followDao.UserFollow(id)
	if err != nil {
		return err
	}
	_ = tools.RedisDoHash("HSET", cacheHashWrite, followerKey, len(follower))
	for _, value := range follower {
		//sadd userId followId
		_ = tools.RedisDoKV("SADD", followerKey, value.FollowId)
	}
	//设置半小时有效期
	_ = tools.RedisDoKV("EXPIRE", followerKey, 1800)
	return nil
}

func reHashKey() {
	cacheHashWrite, cacheHashRead = cacheHashRead, cacheHashWrite
}

func getFollowKey(userId string) string {
	return strings.Join([]string{userId, "follow"}, "_")
}

func getFollowerKey(userId string) string {
	return strings.Join([]string{userId, "follower"}, "_")
}

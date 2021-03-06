package services

import (
	"douyin/daos"
	"douyin/models"
	"douyin/tools"
	"errors"
	"fmt"
	"github.com/gomodule/redigo/redis"
	"strconv"
	"strings"
	"sync"
)

var (
	followService     FollowService
	followServiceOnce sync.Once
	//管理redis中关注数的hash名
	followHash = "follow_hash"
	//管理redis中粉丝数的hash名
	followerHash = "follower_hash"
)

type FollowService interface {
	Action(userId string, toUserId string, actionType string) error
	RedisAction(userId, toUserId, actionType string) error
	UserFollowInfo(find *models.User, userId string) *models.UserMessage
	UserFollowList(userId string) ([]models.UserMessage, error)
	UserFollowerList(userId string) ([]models.UserMessage, error)
	FollowUpdate()
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

// RedisAction @author cwh
//关注操作时，对redis的操作
func (f *FollowServiceImpl) RedisAction(userId, toUserId, actionType string) error {
	//查询对应缓存，没有便加载
	err := f.followListCdRedis(userId)
	_ = f.followerListCdRedis(toUserId)
	if err != nil {
		return errors.New("缓存错误！")
	}

	//关注列表名
	followKey := getFollowKey(userId)
	//粉丝列表名
	followerKey := getFollowerKey(toUserId)

	//查询是否关注了
	do, err := tools.RedisDo("sismember", followKey, toUserId)
	exists := do.(int64)

	var action string
	var add int
	if actionType == "1" && exists == 0 {
		//关注操作，关注列表添加touserId，粉丝列表添加userId，粉丝数+1
		action = "SADD"
		add = 1
	} else if actionType == "2" && exists == 1 {
		//取关操作，关注列表删除toUserId，粉丝列表删除userId，粉丝数-1
		action = "SREM"
		add = -1
	} else {
		return errors.New("操作类型错误！")
	}
	//用户关注列表更新
	if tools.RedisDoKV(action, followKey, toUserId) != nil {
		return errors.New("系统错误！，请稍后重试")
	}
	//被关注者，粉丝列表更新
	_ = tools.RedisDoKV(action, followerKey, userId)
	//关注者关注数+1，被关注者粉丝数+1
	fmt.Printf("userId = %s, add = %d", userId, add)
	_ = tools.RedisDoHash("HINCRBY", followHash, userId, add)
	_ = tools.RedisDoHash("HINCRBY", followerHash, toUserId, add)
	return nil
}

//@author cwh
//将用户的关注列表缓存进redis（无缓存的情况下）
func (f *FollowServiceImpl) followListCdRedis(userId string) error {
	followKey := getFollowKey(userId)
	if tools.RedisKeyExists(followKey) {
		return tools.RedisKeyFlush(followKey)
	}
	id, _ := strconv.Atoi(userId)
	follows, err := f.followDao.UserFollow(id)
	if err != nil {
		return err
	}
	_ = tools.RedisDoHash("HSET", followHash, userId, len(follows))
	fmt.Println("hset", userId, len(follows))
	for _, value := range follows {
		//sadd userId followId
		_ = tools.RedisDoKV("SADD", followKey, value)
	}
	//设置半小时有效期
	_ = tools.RedisDoKV("EXPIRE", followKey, 1800)
	return nil
}

//@author cwh
//将用户的粉丝列表缓存进redis（无缓存的情况下）
func (f *FollowServiceImpl) followerListCdRedis(userId string) error {
	followerKey := getFollowerKey(userId)
	if tools.RedisKeyExists(followerKey) {
		return tools.RedisKeyFlush(followerKey)
	}
	id, _ := strconv.Atoi(userId)
	follower, err := f.followDao.UserFollower(id)
	if err != nil {
		return err
	}
	_ = tools.RedisDoHash("HSET", followerHash, userId, len(follower))
	fmt.Println("hset", userId, len(follower))
	for _, value := range follower {
		//sadd userId followId
		_ = tools.RedisDoKV("SADD", followerKey, value)
	}
	//设置半小时有效期
	_ = tools.RedisDoKV("EXPIRE", followerKey, 1800)
	return nil
}

// UserFollowInfo @author cwh
//更新user的关注信息，返回前端需求的格式
func (f *FollowServiceImpl) UserFollowInfo(find *models.User, userId string) *models.UserMessage {
	//刷新自己的关注列表
	_ = f.followListCdRedis(userId)
	//构建返回
	res := models.UserMessage{
		Id:         find.Id,
		Name:       find.Name,
		Avatar:     find.Avatar,
		Signature:  find.Signature,
		BackGround: find.BackGround,
	}
	//查询是否关注
	do, _ := tools.RedisDo("sismember", getFollowKey(userId), find.Id)
	if do == 1 {
		res.IsFollow = true
	}
	f.setMessageCount(find.Id, &res)
	return &res
}

//@author cwh
//用户信息关注数和粉丝数的查询
func (f *FollowServiceImpl) setMessageCount(userId int64, message *models.UserMessage) {
	//查询对方的关注数
	do, err := tools.RedisDo("hget", followHash, userId)
	if do != nil {
		//有缓存时更新，当0是真实值，更不更新一样
		message.FollowCount, _ = redis.Int64(do, err)
	}
	//查询对方粉丝数
	do, _ = tools.RedisDo("hget", followerHash, userId)
	if do != nil {
		message.FollowerCount, _ = redis.Int64(do, err)
	}
}

// UserFollowList @author cwh
//查询用户的关注列表信息
func (f *FollowServiceImpl) UserFollowList(userId string) ([]models.UserMessage, error) {
	//缓存处理，是否加载
	err := f.followListCdRedis(userId)
	if err != nil {
		return nil, err
	}
	//查询缓存的关注列表
	do, _ := tools.RedisDo("smembers", getFollowKey(userId))
	ids, _ := redis.Ints(do, nil)
	//查询对应的user
	finds, err := daos.GetUserDao().FindListByIds(ids)
	if err != nil {
		return nil, err
	}
	//将user包装成userMessage
	res := make([]models.UserMessage, len(finds))
	for i, find := range finds {
		message := models.UserMessage{
			Id:         find.Id,
			Name:       find.Name,
			IsFollow:   true,
			Avatar:     find.Avatar,
			Signature:  find.Signature,
			BackGround: find.BackGround,
		}
		f.setMessageCount(find.Id, &message)
		res[i] = message
	}
	return res, nil
}

// UserFollowerList @author cwh
//查询用户的粉丝列表信息
func (f *FollowServiceImpl) UserFollowerList(userId string) ([]models.UserMessage, error) {
	//缓存处理，是否加载
	err := f.followerListCdRedis(userId)
	if err != nil {
		return nil, err
	}
	//查询缓存的粉丝列表
	do, _ := tools.RedisDo("smembers", getFollowerKey(userId))
	ids, _ := redis.Ints(do, nil)
	//查询对应的user
	finds, err := daos.GetUserDao().FindListByIds(ids)
	if err != nil {
		return nil, err
	}
	//将user包装成userMessage
	res := make([]models.UserMessage, len(finds))
	for i, find := range finds {
		res[i] = *f.UserFollowInfo(&find, userId)
	}
	return res, nil
}

func getFollowKey(userId string) string {
	return strings.Join([]string{userId, "follow"}, "_")
}

func getFollowerKey(userId string) string {
	return strings.Join([]string{userId, "follower"}, "_")
}

// FollowUpdate
// @author zia
// @Description: 定时缓存写回mysql
// @return error
func (f *FollowServiceImpl) FollowUpdate() {
	follow, follower := GetHashRead()
	//获取关注数量所有键值对
	followMap, err := tools.GetAllKV(follow)
	if err != nil {
		return
	}
	for k, v := range followMap {
		userId, followCount, err := ParseIdAndCount(k, v)
		if err != nil {
			continue
		}
		//更新对应用户关注数
		err = daos.GetFollowDao().UpdateUserFollowCount(userId, followCount)
		if err != nil {
			continue
		}
	}
	//获取粉丝数量所有键值对
	followerMap, err := tools.GetAllKV(follower)
	if err != nil {
		return
	}
	for k, v := range followerMap {
		userId, followerCount, err := ParseIdAndCount(k, v)
		if err != nil {
			continue
		}
		//更新对应用户粉丝数
		err = daos.GetFollowDao().UpdateUserFollowerCount(userId, followerCount)
		if err != nil {
			continue
		}
	}
	return
}

func ParseIdAndCount(k string, v string) (userId int, count int, err error) {
	userId, err = strconv.Atoi(k)
	if err != nil {
		return
	}
	count, err = strconv.Atoi(v)
	if err != nil {
		return
	}
	return
}

func GetHashRead() (string, string) {
	return followHash, followerHash
}

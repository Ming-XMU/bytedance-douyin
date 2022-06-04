package tools

import (
	"douyin/models"
	"errors"
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"github.com/gomodule/redigo/redis"
	"github.com/google/uuid"
	"time"
)

const (
	LoginTokenKey string = "login_tokens:"
	PrivateKey    string = "douyin"
	TokenUserHash string = "token_user_hash"
)

//  userStdClaims
//  @Description: 定义payload结构体,保存TokenKey
type userStdClaims struct {
	jwt.StandardClaims
	TokenKey string
}

// LoginUser 登录状态
type LoginUser struct {
	IssuedAt  int64
	ExpiresAt int64
	UserId    int64
	Name      string
	TokenKey  string
	Status    int //登录状态 0:在线 | 1:被挤下线
}

// CreateToken
// @author zia
// @Description: 创建token,把登录对象设置存储进redis中
// @param loginUser
// @return string
// @return error
func CreateToken(user *models.User) (string, error) {
	uid, err := uuid.NewUUID()
	if err != nil {
		return "", err
	}
	tokenKey := LoginTokenKey + uid.String()

	//查找是否存在映射
	do, err := RedisDo("HGET", TokenUserHash, user.Id)
	if err != nil {
		return "", err
	}
	//如果存在映射
	if do != nil {
		t, err := redis.String(do, err)
		if err != nil {
			return "", err
		}
		//获取在线登录用户设置成下线状态
		loginUser, err := RedisTokenKeyValue(t)
		if err != nil {
			return "", err
		}

		loginUser.Status = 1
		//三分钟后下线状态移除tokenKey,转为下线用户tokenKey失效
		err = RedisCacheTokenKey(t, loginUser, 180)
		if err != nil {
			return "", err
		}
	}
	//设置新的用户id与tokenkey的映射hash
	err = RedisDoHash("HSET", TokenUserHash, user.Id, tokenKey)
	if err != nil {
		return "", err
	}
	fmt.Println("key:" + tokenKey)
	issuedTime := time.Now().Unix()
	//设置过期时间
	t := fmt.Sprintf("%dm", DefaultExpirationTime/60)
	am, err := time.ParseDuration(t)
	if err != nil {
		return "", err
	}
	expireTime := time.Now().Add(am).Unix()
	loginUser := &LoginUser{
		TokenKey:  tokenKey,
		IssuedAt:  issuedTime,
		ExpiresAt: expireTime,
		Name:      user.Name,
		UserId:    user.Id,
		Status:    0,
	}
	err = RedisCacheTokenKey(tokenKey, loginUser, DefaultExpirationTime)
	if err != nil {
		return "", err
	}
	stdClaims := jwt.StandardClaims{}
	uClaims := userStdClaims{
		StandardClaims: stdClaims,
		TokenKey:       tokenKey,
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, uClaims)
	tokenString, err := token.SignedString([]byte(PrivateKey))
	return tokenString, nil
}

// JwtParseTokenKey
// @author zia
// @Description: 解析payload的内容,得到用户信息
// @param token
// @return *models.User
// @return error
func JwtParseTokenKey(token string) (string, error) {
	if token == "" {
		return "", errors.New("no token is found")
	}
	claims := userStdClaims{}
	_, err := jwt.ParseWithClaims(token, &claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(PrivateKey), nil
	})
	if err != nil {
		//log.Fatalln("token is invalid")
		return "", err
	}
	return claims.TokenKey, err
}

// RefreshToken
// @author zia
// @Description: 刷新token
// @param loginUser
// @return error
func RefreshToken(loginUser *LoginUser) error {
	loginUser.IssuedAt = time.Now().Unix()
	t := fmt.Sprintf("%dm", DefaultExpirationTime/60)
	am, err := time.ParseDuration(t)
	if err != nil {
		return err
	}
	loginUser.ExpiresAt = time.Now().Add(am).Unix()
	tokenKey := loginUser.TokenKey
	err = RedisCacheTokenKey(tokenKey, loginUser, DefaultExpirationTime)
	if err != nil {
		return err
	}
	return nil
}

// VeifyToken
// @author zia
// @Description: Token验证 | 刷新
// @param loginUser
// @return (token有效：loginUser , nil | token无效: nil, err)
func VeifyToken(token string) (*LoginUser, error) {
	tokenKey, err := JwtParseTokenKey(token)
	if err != nil {
		return nil, err
	}
	//判断tokenKey是否存在
	exist, err := RedisCheckKey(tokenKey)
	if exist == false || err != nil {
		return nil, errors.New("token is not exist")
	}
	//获取loginUser登录信息
	loginUser, err := RedisTokenKeyValue(tokenKey)
	if err != nil {
		return nil, err
	}
	fmt.Printf("loginUserStatus = %d", loginUser.Status)
	//判断loginUser登录状态
	if loginUser.Status == 1 {
		return nil, errors.New("your account is already logged in elsewhere")
	}
	//过期时间刷新
	expireTime := loginUser.ExpiresAt
	curTime := time.Now()
	//验证token是否失效
	if expireTime < curTime.Unix() {
		return nil, errors.New("token valid")
	}
	// token有效且相差不足20分钟，自动刷新
	am, _ := time.ParseDuration("20m")
	if expireTime <= curTime.Add(am).Unix() {
		err = RefreshToken(loginUser)
		if err != nil {
			return nil, err
		}
	}
	return loginUser, nil
}

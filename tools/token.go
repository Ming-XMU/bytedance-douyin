package tools

import (
	"douyin/models"
	"errors"
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"github.com/google/uuid"
	"time"
)

const (
	LoginTokenKey string = "login_tokens:"
	PrivateKey    string = "douyin"
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
}

// CreateToken
// @author zia
// @Description: 创建token,把登录对象设置存储进redis中
// @param loginUser
// @return string
// @return error
func CreateToken(user *models.User) (string, error) {
	uuid, err := uuid.NewUUID()
	if err != nil {
		return "", err
	}
	tokenKey := LoginTokenKey + uuid.String()
	fmt.Println(tokenKey)
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
// @return (nil token有效 | err token 过期,不存在)
func VeifyToken(token string) error {
	tokenKey, err := JwtParseTokenKey(token)
	if err != nil {
		return err
	}
	loginUser, err := RedisTokenKeyValue(tokenKey)
	if err != nil {
		return err
	}
	expireTime := loginUser.ExpiresAt
	curTime := time.Now()
	//验证token是否失效
	if expireTime < curTime.Unix() {
		return errors.New("token valid")
	}
	// 相差不足20分钟，自动刷新
	am, _ := time.ParseDuration("20m")
	if expireTime <= curTime.Add(am).Unix() {
		err := RefreshToken(loginUser)
		if err != nil {
			return err
		}
	}
	return nil
}

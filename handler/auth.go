package handler

import (
	"context"
	"crypto/rsa"
	"fmt"
	"regexp"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"auth/common"
	log "auth/log"
	"auth/proto"
	"auth/store/cache"
	"auth/store/models"
	"auth/utils/calltable"
)

var RegUname = regexp.MustCompile(`^[a-zA-Z0-9_]{4,16}$`)

type AuthOptions struct {
	PK        *rsa.PrivateKey
	PublicKey []byte
	DB        *gorm.DB
	Cache     cache.AuthCache
	CT        *calltable.CallTable
}

func NewAuth(opts AuthOptions) *Auth {
	ret := &Auth{
		AuthOptions: opts,
	}
	return ret
}

type Auth struct {
	AuthOptions

	proto.UnimplementedAuthServer
}

func (*Auth) Captcha(ctx context.Context, in *proto.CaptchaRequest) (*proto.CaptchaResponse, error) {
	return &proto.CaptchaResponse{}, nil
}

func (h *Auth) Login(ctx context.Context, in *proto.LoginRequest) (*proto.LoginResponse, error) {
	out := &proto.LoginResponse{}

	if len(in.Uname) < 4 {
		out.Flag = proto.LoginResponse_UNAME_ERROR
		out.Msg = "please input right uname"
		return out, nil
	}

	if len(in.Passwd) < 6 {
		out.Flag = proto.LoginResponse_PASSWD_ERROR
		out.Msg = "passwd is required"
		return out, nil
	}

	user := &models.Users{
		Uname: in.Uname,
	}

	res := h.DB.Limit(1).Find(user, user)
	if err := res.Error; err != nil {
		out.Flag = proto.LoginResponse_FAIL
		out.Msg = "user not found"
		return nil, fmt.Errorf("server internal error")
	}

	if res.RowsAffected == 0 {
		out.Flag = proto.LoginResponse_UNAME_ERROR
		out.Msg = "user not exist"
		return out, nil
	}

	if user.Passwd != in.Passwd {
		out.Flag = proto.LoginResponse_PASSWD_ERROR
		return out, nil
	}

	if user.Stat != 0 {
		out.Flag = proto.LoginResponse_STAT_ERROR
		return out, nil
	}

	assess, err := common.GenerateToken(h.PK, user.UID, user.Uname)
	if err != nil {
		return nil, err
	}

	cacheInfo := &cache.AuthCacheInfo{
		User:         user,
		AssessToken:  assess,
		RefreshToken: uuid.NewString(),
	}

	if err = h.Cache.StoreUser(ctx, cacheInfo, time.Hour); err != nil {
		log.Error(err)
	}

	out.AssessToken = assess
	out.RefreshToken = cacheInfo.RefreshToken
	out.UserInfo = &proto.UserInfo{
		Uid:     user.UID,
		Uname:   user.Uname,
		Stat:    int32(user.Stat),
		Created: user.CreateAt.Unix(),
	}
	return out, nil
}

func (h *Auth) Logout(ctx context.Context, in *proto.LogoutRequest) (*proto.LogoutResponse, error) {

	return nil, nil
}

func (*Auth) RefreshToken(ctx context.Context, in *proto.RefreshTokenRequest) (*proto.RefreshTokenResponse, error) {
	//TODO
	return nil, nil
}

func (h *Auth) UserInfo(ctx context.Context, in *proto.UserInfoRequest) (*proto.UserInfoResponse, error) {
	user := &models.Users{
		UID: in.Uid,
	}
	uc := h.Cache.FetchUser(ctx, in.Uid)
	if uc != nil {
		user = uc.User
	} else {
		res := h.DB.Limit(1).Find(user, user)
		if res.Error != nil {
			return nil, fmt.Errorf("server internal error: %v", res.Error)
		}
		if res.RowsAffected == 0 {
			return nil, fmt.Errorf("user no found")
		}
		//TODO:
		// h.Cache.StoreUser(ctx, &cache.AuthCacheInfo{User: user}, time.Hour)
	}

	out := &proto.UserInfoResponse{}
	out.Info = &proto.UserInfo{
		Uid:     user.UID,
		Uname:   user.Uname,
		Stat:    int32(user.Stat),
		Created: user.CreateAt.Unix(),
	}
	return out, nil
}

func (h *Auth) Register(ctx context.Context, in *proto.RegisterRequest) (*proto.RegisterResponse, error) {
	if len(in.Uname) < 4 {
		return nil, fmt.Errorf("please input right account")
	}
	if len(in.Passwd) < 6 {
		return nil, fmt.Errorf("passwd is required")
	}

	user := &models.Users{
		Uname:    in.Uname,
		Passwd:   in.Passwd,
		Nickname: in.Nickname,
		Gender:   'X',
	}

	f := &models.Users{Uname: in.Uname}

	if res := h.DB.Find(f, f); res.RowsAffected > 0 {
		return nil, fmt.Errorf("user alread exist")
	}

	res := h.DB.Create(user)

	if res.Error != nil {
		log.Error(res.Error)
		return nil, fmt.Errorf("server internal error")
	}

	if res.RowsAffected == 0 {
		return nil, fmt.Errorf("create user error")
	}

	return &proto.RegisterResponse{Msg: "ok"}, nil
}

func (h *Auth) PublicKeys(ctx context.Context, in *proto.PublicKeysRequest) (*proto.PublicKeysResponse, error) {
	return &proto.PublicKeysResponse{Keys: h.PublicKey}, nil
}

func (h *Auth) AnonymousLogin(ctx context.Context, in *proto.AnonymousLoginRequest) (*proto.LoginResponse, error) {
	return nil, nil
}

func (h *Auth) ModifyPasswd(ctx context.Context, in *proto.ModifyPasswdRequest) (*proto.ModifyPasswdResponse, error) {
	//TODO:
	return nil, nil
}

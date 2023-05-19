package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"reflect"

	"github.com/urfave/cli/v2"
	"google.golang.org/protobuf/encoding/protojson"
	protobuf "google.golang.org/protobuf/proto"
	"gorm.io/driver/mysql"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"

	"auth/handler"
	log "auth/log"
	"auth/proto"
	"auth/store/cache"
	"auth/store/models"
	"auth/utils/calltable"
	"auth/utils/rsagen"
	utilSignal "auth/utils/signal"
)

var Version string = "unknown"
var GitCommit string = "unknown"
var BuildAt string = "unknown"
var BuildBy string = "unknown"
var Name string = "auth"

var ConfigPath string = ""
var ListenAddr string = ""
var PrintConf bool = false

const PrivateKeyFile = "private.pem"
const PublicKeyFile = "public.pem"

func ReadRSAKey() ([]byte, []byte, error) {
	privateRaw, err := os.ReadFile(PrivateKeyFile)
	if err != nil {
		privateKey, publicKey, err := rsagen.GenerateRsaPem(2048)
		if err != nil {
			return nil, nil, err
		}
		privateRaw = []byte(privateKey)
		os.WriteFile(PrivateKeyFile, []byte(privateKey), 0644)
		os.WriteFile(PublicKeyFile, []byte(publicKey), 0644)
	}
	publicRaw, err := os.ReadFile(PublicKeyFile)
	if err != nil {
		return nil, nil, err
	}
	return privateRaw, publicRaw, nil
}

func main() {
	cli.VersionPrinter = func(c *cli.Context) {
		fmt.Println("project:", Name)
		fmt.Println("version:", Version)
		fmt.Println("git commit:", GitCommit)
		fmt.Println("build at:", BuildAt)
		fmt.Println("build by:", BuildBy)
	}

	app := cli.NewApp()
	app.Name = Name
	app.Version = Version
	app.Flags = []cli.Flag{
		&cli.StringFlag{
			Name:        "config",
			Aliases:     []string{"c"},
			Value:       "config.yaml",
			Destination: &ConfigPath,
		}, &cli.StringFlag{
			Name:        "listen",
			Aliases:     []string{"l"},
			Value:       ":30020",
			Destination: &ListenAddr,
		}, &cli.BoolFlag{
			Name:        "print-config",
			Destination: &PrintConf,
			Hidden:      true,
		},
	}

	app.Action = RealMain

	if err := app.Run(os.Args); err != nil {
		log.Error(err)
		os.Exit(-1)
	}
}

func CreateMysqlClient(dsn string) *gorm.DB {
	dbc, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		DisableNestedTransaction: true,
		NamingStrategy: schema.NamingStrategy{
			SingularTable: true,
		},
	})
	if err != nil {
		panic(err)
	}
	return dbc
}

func CreateSQLiteClient(dsn string) *gorm.DB {
	dbc, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	if err != nil {
		panic(err)
	}
	err = dbc.AutoMigrate(
		&models.Users{},
	)
	if err != nil {
		panic(err)
	}
	return dbc
}

func RealMain(c *cli.Context) error {
	privateRaw, publicRaw, err := ReadRSAKey()
	if err != nil {
		panic(err)
	}
	pk, err := rsagen.ParseRsaPrivateKeyFromPem(privateRaw)
	if err != nil {
		return err
	}

	authHandler := handler.NewAuth(handler.AuthOptions{
		PK:        pk,
		PublicKey: publicRaw,
		DB:        CreateSQLiteClient("auth.db"),
		Cache:     cache.NewMemory(),
	})

	authHandler.CT = calltable.ExtractParseGRpcMethod(proto.File_proto_auth_proto.Services(), authHandler)

	ServerCallTable(http.DefaultServeMux, authHandler, authHandler.CT)

	fmt.Println("start http server at:", ListenAddr)

	go func() {
		http.ListenAndServe(ListenAddr, nil)
	}()

	signal := utilSignal.WaitShutdown()
	log.Infof("recv signal: %v", signal.String())
	return nil
}

func AuthMiddleWrap(handler interface{}) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {

	}
}

func ServerCallTable(mux *http.ServeMux, handler interface{}, ct *calltable.CallTable) {
	respWithError := func(w http.ResponseWriter, data interface{}, err error) {
		type HttpRespType struct {
			Data    interface{} `json:"data"`
			Code    int         `json:"code"`
			Message string      `json:"message"`
		}
		respWrap := &HttpRespType{
			Data:    data,
			Message: "ok",
		}
		if err != nil {
			respWrap.Code = -1
			respWrap.Message = err.Error()
		}
		raw, _ := json.Marshal(respWrap)
		w.Write(raw)
	}

	ct.Range(func(key string, method *calltable.Method) bool {
		pattern := "/" + key

		fmt.Println("register http method:", pattern)

		cb := func(w http.ResponseWriter, r *http.Request) {
			raw, err := io.ReadAll(r.Body)
			if err != nil {
				respWithError(w, nil, fmt.Errorf("read body error: %s", err.Error()))
				return
			}

			w.Header().Set("Content-Type", "application/json; charset=utf-8")
			req := reflect.New(method.RequestType).Interface().(protobuf.Message)

			// todo : get marshaler
			unmarshal := &protojson.UnmarshalOptions{}
			if err := unmarshal.Unmarshal(raw, req); err != nil {
				respWithError(w, nil, fmt.Errorf("unmarshal request error: %s", err.Error()))
				return
			}

			// here call method
			respArgs := method.Call(handler, r.Context(), req)

			if len(respArgs) != 2 {
				return
			}

			var respErr error
			if !respArgs[1].IsNil() {
				respErr = respArgs[1].Interface().(error)
			}

			var respData json.RawMessage
			if !respArgs[0].IsNil() {
				if resp, ok := respArgs[0].Interface().(protobuf.Message); ok {
					marshal := &protojson.MarshalOptions{
						EmitUnpopulated: true,
						UseProtoNames:   true,
					}
					data, err := marshal.Marshal(resp)
					if err == nil {
						respData = data
					} else {
						respWithError(w, nil, respErr)
					}
				}
			}
			respWithError(w, respData, respErr)
		}

		mux.HandleFunc(pattern, cb)

		return true
	})
}

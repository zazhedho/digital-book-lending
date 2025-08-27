package utils

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/spf13/viper"
	_ "github.com/spf13/viper/remote"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

// Logging
const (
	LogLevelPanic = 0
	LogLevelError = 1
	LogLevelFail  = 2
	LogLevelInfo  = 3
	LogLevelData  = 4
	LogLevelDebug = 5
)

var logLevelMap = map[int]string{
	LogLevelPanic: "PANIC",
	LogLevelError: "ERROR",
	LogLevelFail:  "FAIL ",
	LogLevelInfo:  "INFO ",
	LogLevelData:  "DATA ",
	LogLevelDebug: "DEBUG",
}

func WriteLog(level int, msg ...any) {
	if _, ok := logLevelMap[level]; !ok {
		return
	}

	if logLevel, _ := strconv.Atoi(GetEnv("LOG_LEVEL", "5").(string)); logLevel < level {
		return
	}

	logPrefix := fmt.Sprintf("[%s][%s][%s]", os.Getenv("ServerIP"), os.Getenv("NODE"), logLevelMap[level])
	switch level {
	case LogLevelPanic, LogLevelError:
		log.Println(time.Now().Format(".000000"), logPrefix, msg)
	case LogLevelFail, LogLevelInfo, LogLevelData, LogLevelDebug:
		fmt.Println(time.Now().Format("2006/01/02 15:04:05 .000000"), logPrefix, msg)
	}
}

func GetEnv(key string, def interface{}) interface{} {
	v, isset := os.LookupEnv(key)
	if !isset {
		return def
	}

	switch def.(type) {
	case int:
		if i, err := strconv.Atoi(v); err == nil {
			return i
		}
		break
	case time.Duration:
		if v == "eod" {
			eodTime, _ := time.Parse(time.RFC3339, time.Now().Format("2006-01-02")+"T23:59:59+07:00")
			return time.Duration(int(eodTime.Sub(time.Now()).Seconds()))
		}

		if i, err := strconv.Atoi(v); err == nil && i > 0 {
			return time.Duration(i)
		} else {
			return time.Minute //default
		}
	case bool:
		if b, err := strconv.ParseBool(v); err == nil {
			return b
		}
		break
	case string:
		return v
	}
	return def
}

func GetAppConf(key string, def interface{}, rdbCache *redis.Client) interface{} {
	var (
		err   error
		cache bool
	)

	cacheKey := RedisAppConf
	appConf := make(map[string]string)
	var getNewConfig bool
	if consul := os.Getenv("CONSUL"); consul != "" {
		cache = strings.ToLower(os.Getenv("CACHE")) == "on" && rdbCache != nil

		if cache {
			if jsonAppConf, err := rdbCache.Get(context.Background(), cacheKey).Result(); err == nil {
				if err = json.Unmarshal([]byte(jsonAppConf), &appConf); err != nil {
					WriteLog(LogLevelError, fmt.Sprintf("utils.GetAppConf; Unmarshal conf from cache; %s; error: %+v;", jsonAppConf, err))
					getNewConfig = true
				}
			} else if err == redis.Nil {
				getNewConfig = true
			}
		} else {
			getNewConfig = true
		}

		if getNewConfig {
			consulPath := fmt.Sprintf("%s/%s", os.Getenv("CONSUL_PATH"), os.Getenv("APP_ENV"))
			runtimeViper := viper.New()
			runtimeViper.AddRemoteProvider("consul", consul, consulPath)
			runtimeViper.SetConfigType("json") // Need to explicitly set this to json
			if err = runtimeViper.ReadRemoteConfig(); err != nil {
				WriteLog(LogLevelError, fmt.Sprintf("utils.GetAppConf; Loading config: %s/%s; error: %+v;", consul, consulPath, err))
			} else if err = runtimeViper.Unmarshal(&appConf); err != nil {
				WriteLog(LogLevelError, fmt.Sprintf("utils.GetAppConf; Loading congig: %s/%s; error: unable to decode into map, %+v;", consul, consulPath, err))
			}
		}
	} else {
		configName := "app"
		pathConfig := os.Getenv("APP_CONFIG")
		if pathConfig == "" {
			pathConfig = "config"
			configName = os.Getenv("APP_ENV")
		}

		viper.AddConfigPath(pathConfig)
		viper.SetConfigType("env")
		viper.SetConfigName(configName)

		err = viper.ReadInConfig()
		if err != nil {
			WriteLog(LogLevelError, fmt.Sprintf("utils.GetAppConf; Loading config: %s - %s.env; error:  %+v;", pathConfig, configName, err))
		} else {
			_ = viper.Unmarshal(&appConf)
		}
	}

	if len(appConf) > 0 {
		if _, ok := appConf["config_id"]; !ok {
			appConf["config_id"] = uuid.NewString()
		}

		if appConf["config_id"] != os.Getenv("CONFIG_ID") {
			for k, v := range appConf {
				//Reset all env variable
				os.Setenv(strings.ToUpper(k), v)
			}
		}

		if cache && getNewConfig {
			go func(rdbCache *redis.Client, cacheKey string, data map[string]string) {
				if cacheData, err := json.Marshal(data); err == nil {
					ttl := GetEnv("TTL_CACHE_CONFIG_APP", time.Duration(60*60*24)).(time.Duration) * time.Second
					_ = rdbCache.Set(context.Background(), cacheKey, string(cacheData), ttl).Err()
				}
			}(rdbCache, cacheKey, appConf)
		}
	}

	return GetEnv(key, def)
}

func GenerateLogId(ctx *gin.Context) uuid.UUID {
	if logId, ok := ctx.Value(CtxKeyId).(uuid.UUID); ok {
		return logId
	}

	logId, err := uuid.NewV7()
	if err != nil {
		logId = uuid.New()
	}

	return logId
}

func StructToMapUpdate(data interface{}) map[string]interface{} {
	ret := map[string]interface{}{}
	v := reflect.ValueOf(data)
	for i := 0; i < v.NumField(); i++ {
		field := v.Type().Field(i).Name
		value := v.Field(i).Interface()
		ret[field] = value
	}
	return ret
}

func SplitAndTrim(data string) []string {
	if data == "" {
		return []string{}
	}
	parts := strings.Split(data, "|")
	for i := range parts {
		parts[i] = strings.TrimSpace(parts[i])
	}
	return parts
}

func CreateUUID() string {
	var id string
	if uuid7, err := uuid.NewV7(); err == nil {
		id = uuid7.String()
	} else {
		id = uuid.NewString()
	}

	return id
}

func TitleCase(s string) string {
	titleCaser := cases.Title(language.English)
	return titleCaser.String(s)
}

type ValidateMessage struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

func mapValidateMessage(fe validator.FieldError) string {
	switch fe.Tag() {
	case "required":
		return "This field is required"
	case "email":
		return "Invalid email"
	case "alphanum":
		return "Should be alphanumeric"
	case "min":
		return "Minimum " + fe.Param()
	case "max":
		return "Maximum " + fe.Param()
	case "lte":
		return "Should be less than " + fe.Param()
	case "gte":
		return "Should be greater than " + fe.Param()
	case "ltefield":
		return "Should be less than " + fe.Param()
	case "gtefield":
		return "Should be greater than " + fe.Param()
	}

	return "Invalid value"
}

func ValidateError(err error, reflectType reflect.Type, tagName string) []ValidateMessage {
	var ve validator.ValidationErrors
	if errors.As(err, &ve) {
		out := make([]ValidateMessage, len(ve))
		for i, fe := range ve {
			field := fe.Field()
			if structField, ok := reflectType.FieldByName(fe.Field()); ok {
				field = structField.Tag.Get(tagName)
			}
			out[i] = ValidateMessage{field, mapValidateMessage(fe)}
		}
		return out
	}
	return []ValidateMessage{{"", err.Error()}}
}

func InterfaceString(data interface{}) string {
	if data == nil {
		return ""
	}
	switch v := data.(type) {
	case string:
		return v
	case []byte:
		return string(v)
	default:
		bytes, _ := json.Marshal(data)
		return string(bytes)
	}
}

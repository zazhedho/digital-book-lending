package utils

const (
	InvalidRequest = "Invalid request format. Please ensure the structure is correct and matches the expected data format."
	InvalidHeader  = "Invalid header format. Please ensure the structure is correct and matches the expected data format."
	MsgErr         = "Something Went Wrong"
	MsgFail        = "Something Went Wrong"
	MsgDenied      = "Access Denied"
	MsgCredential  = "Invalid Credentials. Please input the correct credentials and try again."
	MsgRequired    = "Please fill the %s field."
	MsgExists      = "Already exists."
	MsgNotFound    = "Data Not Found"
	NotFound       = "The requested resource could not be found"
	MsgSuccess     = "Success"
	MsgUpdated     = "Updated"
	NoProperties   = "No properties to update has been provided in request. Please specify at least one property which needs to be updated."
	InvalidCred    = "Invalid email or password"
)

// Redis Key
const (
	RedisAppConf = "cache:config:app"
	RedisDbConf  = "cache:config:db"
)

// CtxKeyId Context KEYS
const (
	CtxKeyId       = "CTX_ID"
	CtxKeyAuthData = "auth_data"
)

const (
	LayoutMonth       = "2006_01"
	LayoutYearMonth   = "200601"
	LayoutDate        = "2006-01-02"
	LayoutTime        = "15:04:05"
	LayoutDateTime    = "2006-01-02 15:04:05"
	LayoutDateTimeDot = "2006-01-02 15.04.05"
	LayoutTimestamp   = "2006-01-02 15:04:05.000000000"
	LayoutTimeStamp   = "2006-01-02 15:04:05.000"
	LayoutDateTimeH   = "2006-01-02 15"
)

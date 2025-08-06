package auth

const (
	evtLoginAttempt                = "login_attempt"
	evtLoginInvalidEmail           = "login_invalid_email"
	evtLoginInvalidCreds           = "login_invalid_credentials"
	evtLoginCredsValid             = "login_credentials_valid"
	evtLoginMissingReq             = "login_missing_request"
	evtLoginInvalidDeviceInfo      = "login_invalid_device_info"
	evtLoginGenerateTokensFailed   = "login_generate_tokens_failed"
	evtLoginHashRefTokenFailed     = "login_hash_ref_token_failed"
	evtLoginBeginSqlTxFailed       = "login_begin_sql_tx_failed"
	evtLoginHandleDeviceInfoFailed = "login_handle_device_info_failed"
	evtLoginRefTokInsertDBFailed   = "login_ref_token_db_insert_failed"
	evtLoginSqlCommitTxFailed      = "login_sql_commit_tx_failed"
	evtLoginSuccess                = "login_success"
)

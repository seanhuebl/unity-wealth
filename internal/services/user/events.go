package user

const (
	evtSignUpAttempt            = "sign_up_attempt"
	evtSignUpInvalidEmail       = "sign_up_invalid_email"
	evtSignUpInvalidPassword    = "sign_up_invalid_pwd"
	evtSignUpValidInput         = "sign_up_valid_input"
	evtSignUpPwdHashFailed      = "sign_up_pwd_hashing_failed"
	evtSignUpUserInsertDBFailed = "sign_up_user_insert_db_failed"
	evtSignUpSuccess            = "sign_up_success"
)

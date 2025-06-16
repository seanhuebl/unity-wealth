package transaction

const (
	evtCreateTxAttempt           = "create_tx_attempt"
	evtCreateTxInvalidDateFormat = "create_tx_invalid_date_format"
	evtCreateTxDBExecFailed      = "create_tx_db_insert_failed"
	evtCreateTxSuccess           = "create_tx_success"

	evtUpdateTxAttempt           = "update_tx_attempt"
	evtUpdateTxInvalidDateFormat = "update_tx_invalid_date_format"
	evtUpdateTxNotFound          = "update_tx_not_found"
	evtUpdateTxDBExecFailed      = "update_tx_db_failed"
	evtUpdateTxSuccess           = "update_tx_success"

	evtDeleteTxAttempt      = "delete_tx_attempt"
	evtDeleteTxNotFound     = "delete_tx_not_found"
	evtDeleteTxDBExecFailed = "delete_tx_db_failed"
	evtDeleteTxSuccess      = "delete_tx_success"

	evtGetTxAttempt      = "get_tx_attempt"
	evtGetTxNotFound     = "get_tx_not_found"
	evtGetTxDBExecFailed = "get_tx_db_failed"
	evtGetTxSuccess      = "get_tx_success"

	evtListTxsAttempt        = "list_txs_attempt"
	evtListTxsPageSize       = "list_txs_page_size_error"
	evtListTxsFetchAttempt   = "list_tcs_fetch_attempt"
	evtListTxsNotFound       = "list_txs_not_found"
	evtListTxsPaginatedempty = "list_txs_paginated_empty"
	evtListTxsDBExecFailed   = "list_txs_db_failed"
	evtListTxsSuccess        = "list_txs_success"
)

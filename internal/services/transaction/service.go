package transaction

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/seanhuebl/unity-wealth/internal/constants"
	"github.com/seanhuebl/unity-wealth/internal/cursor"
	"github.com/seanhuebl/unity-wealth/internal/database"
	"github.com/seanhuebl/unity-wealth/internal/helpers"
	"github.com/seanhuebl/unity-wealth/internal/models"
	"github.com/seanhuebl/unity-wealth/internal/sentinels"
	"go.uber.org/zap"
)

type TransactionService struct {
	txQueries    database.TransactionQuerier
	cursorSigner *cursor.Signer
	logger       *zap.Logger
}

func NewTransactionService(txQueries database.TransactionQuerier, cs *cursor.Signer, logger *zap.Logger) *TransactionService {
	if logger == nil {
		logger = zap.NewNop()
	}
	if cs == nil {
		logger.Fatal("cursor signer is required")
	}
	return &TransactionService{txQueries: txQueries, cursorSigner: cs, logger: logger}
}

func (s *TransactionService) CreateTransaction(ctx context.Context, userID uuid.UUID, req models.NewTxRequest) (*models.Tx, error) {
	start := time.Now()
	reqID := ctx.Value(constants.RequestIDKey).(uuid.UUID)
	txID := uuid.New()
	logger := s.logger.With(
		zap.Stringer("request_id", reqID),
		zap.Stringer("user_id", userID),
		zap.Stringer("tx_id", txID),
	)

	logger.Debug(evtCreateTxAttempt,
		zap.String("raw_date", req.Date),
		zap.String("merchant", req.Merchant),
		zap.Float64("amount", req.Amount),
		zap.Int32("detailed_category_id", req.DetailedCategory),
	)

	parsedDate, err := time.Parse("2006-01-02", req.Date)

	if err != nil {
		baseErr := errors.Join(ErrInvalidDateFormat, err)
		wrapped := fmt.Errorf("create tx: %w", baseErr)
		logger.Warn(evtCreateTxInvalidDateFormat,
			zap.String("raw_date", req.Date),
			zap.Error(wrapped),
		)
		return nil, wrapped
	}
	tx := models.NewTransaction(
		txID,
		userID,
		parsedDate,
		req.Merchant,
		req.Amount,
		req.DetailedCategory,
	)

	dbStart := time.Now()
	err = s.txQueries.CreateTransaction(ctx, database.CreateTransactionParams{
		ID:                 tx.ID,
		UserID:             tx.UserID,
		TransactionDate:    parsedDate,
		Merchant:           tx.Merchant,
		AmountCents:        helpers.ConvertToCents(req.Amount),
		DetailedCategoryID: tx.DetailedCategory,
	})
	dbDuration := time.Since(dbStart)
	if err != nil {
		baseErr := errors.Join(sentinels.ErrDBExecFailed, err)
		wrapped := fmt.Errorf("create tx: %w", baseErr)
		logger.Error(evtCreateTxDBExecFailed,
			zap.Duration("db_duration_ms", dbDuration),
			zap.Error(wrapped),
		)
		return nil, wrapped
	}
	totalDuration := time.Since(start)
	logger.Info(evtCreateTxSuccess,
		zap.Time("parsed_date", parsedDate),
		zap.String("merchant", req.Merchant),
		zap.Float64("amount", req.Amount),
		zap.Int64("amount_cents", helpers.ConvertToCents(req.Amount)),
		zap.Int32("detailed_category_id", req.DetailedCategory),
		zap.Duration("db_duration_ms", dbDuration),
		zap.Duration("total_duration_ms", totalDuration),
	)

	return tx, nil

}

func (s *TransactionService) UpdateTransaction(ctx context.Context, txID, userID uuid.UUID, req models.NewTxRequest) (*models.Tx, error) {
	start := time.Now()
	reqID, _ := ctx.Value(constants.RequestIDKey).(uuid.UUID)

	logger := s.logger.With(
		zap.Stringer("request_id", reqID),
		zap.Stringer("user_id", userID),
		zap.Stringer("tx_id", txID),
	)

	logger.Debug(evtUpdateTxAttempt,
		zap.String("raw_date", req.Date),
		zap.String("merchant", req.Merchant),
		zap.Float64("amount", req.Amount),
		zap.Int32("detailed_category_id", req.DetailedCategory),
	)

	parsedDate, err := time.Parse("2006-01-02", req.Date)
	if err != nil {
		baseErr := errors.Join(ErrInvalidDateFormat, err)
		wrapped := fmt.Errorf("update tx: %w", baseErr)

		logger.Warn(evtUpdateTxInvalidDateFormat,
			zap.String("raw_date", req.Date),
			zap.Error(wrapped),
		)
		return nil, wrapped
	}

	dbStart := time.Now()

	txRow, err := s.txQueries.UpdateTransactionByID(ctx, database.UpdateTransactionByIDParams{
		TransactionDate:    parsedDate,
		Merchant:           req.Merchant,
		AmountCents:        helpers.ConvertToCents(req.Amount),
		DetailedCategoryID: req.DetailedCategory,
		UpdatedAt:          time.Now().UTC(),
		ID:                 txID,
	})
	dbDuration := time.Since(dbStart)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			baseErr := errors.Join(ErrTxNotFound, err)
			wrapped := fmt.Errorf("update tx: %w", baseErr)

			logger.Warn(evtUpdateTxNotFound,

				zap.Duration("db_duration_ms", dbDuration),
				zap.Error(wrapped),
			)
			return nil, wrapped
		}
		baseErr := errors.Join(sentinels.ErrDBExecFailed, err)
		wrapped := fmt.Errorf("update tx: %w", baseErr)
		logger.Error(evtUpdateTxDBExecFailed,
			zap.Duration("db_duration_ms", dbDuration),
			zap.Error(wrapped),
		)
		return nil, wrapped
	}
	tx := &models.Tx{
		ID:               txRow.ID,
		UserID:           userID,
		Date:             parsedDate,
		Merchant:         txRow.Merchant,
		Amount:           helpers.CentsToDollars(txRow.AmountCents),
		DetailedCategory: txRow.DetailedCategoryID,
	}

	totalDuration := time.Since(start)
	logger.Info(evtUpdateTxSuccess,
		zap.Time("parsed_date", tx.Date),
		zap.String("merchant", tx.Merchant),
		zap.Float64("amount", tx.Amount),
		zap.Int64("amount_cents", txRow.AmountCents),
		zap.Int32("detailed_category_id", tx.DetailedCategory),
		zap.Duration("db_duration_ms", dbDuration),
		zap.Duration("total_duration_ms", totalDuration),
	)
	return tx, nil
}

func (s *TransactionService) DeleteTransaction(ctx context.Context, userID, txID uuid.UUID) error {
	start := time.Now()
	reqID, _ := ctx.Value(constants.RequestIDKey).(uuid.UUID)

	logger := s.logger.With(
		zap.Stringer("request_id", reqID),
		zap.Stringer("user_id", userID),
		zap.Stringer("tx_id", txID),
	)

	logger.Debug(evtDeleteTxAttempt)

	dbStart := time.Now()
	_, err := s.txQueries.DeleteTransactionByID(ctx, database.DeleteTransactionByIDParams{
		ID:     txID,
		UserID: userID,
	})

	dbDuration := time.Since(dbStart)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			baseErr := errors.Join(ErrTxNotFound, err)
			wrapped := fmt.Errorf("delete tx: %w", baseErr)

			logger.Warn(evtDeleteTxNotFound,
				zap.Duration("db_duration_ms", dbDuration),
				zap.Error(wrapped),
			)
			return wrapped
		}
		baseErr := errors.Join(sentinels.ErrDBExecFailed, err)
		wrapped := fmt.Errorf("delete tx: %w", baseErr)
		logger.Error(evtCreateTxDBExecFailed,
			zap.Duration("db_duration_ms", dbDuration),
			zap.Error(wrapped),
		)
		return wrapped
	}
	totalDuration := time.Since(start)
	logger.Info(evtDeleteTxSuccess,
		zap.Duration("db_duration_ms", dbDuration),
		zap.Duration("total_duration_ms", totalDuration),
	)
	return nil
}

func (s *TransactionService) GetTransactionByID(ctx context.Context, userID, txID uuid.UUID) (*models.Tx, error) {
	start := time.Now()
	reqID, _ := ctx.Value(constants.RequestIDKey).(uuid.UUID)

	logger := s.logger.With(
		zap.Stringer("request_id", reqID),
		zap.Stringer("user_id", userID),
		zap.Stringer("tx_id", txID),
	)

	if userID == uuid.Nil || txID == uuid.Nil {
		wrapped := fmt.Errorf("get tx by ID: zero UUID: %w", sentinels.ErrInvalidID)
		logger.Warn(evtGetTxZeroUUID,
			zap.Error(wrapped),
		)
		return nil, wrapped
	}

	logger.Debug(evtGetTxAttempt)

	dbStart := time.Now()
	row, err := s.txQueries.GetUserTransactionByID(ctx, database.GetUserTransactionByIDParams{UserID: userID, ID: txID})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			baseErr := errors.Join(ErrTxNotFound, err)
			wrapped := fmt.Errorf("get tx by ID: %w", baseErr)

			logger.Warn(evtGetTxNotFound,
				zap.Error(wrapped),
			)
			return nil, wrapped
		}
		baseErr := errors.Join(sentinels.ErrDBExecFailed, err)
		wrapped := fmt.Errorf("get tx by ID: %w", baseErr)
		logger.Error(evtGetTxDBExecFailed,
			zap.Error(wrapped),
		)
		return nil, wrapped
	}
	dbDuration := time.Since(dbStart)
	txn := &models.Tx{
		ID:               row.ID,
		UserID:           row.UserID,
		Date:             row.TransactionDate,
		Merchant:         row.Merchant,
		Amount:           helpers.CentsToDollars(row.AmountCents),
		DetailedCategory: row.DetailedCategoryID,
	}
	totalDuration := time.Since(start)

	logger.Info(evtGetTxSuccess,
		zap.Time("parsed_date", txn.Date),
		zap.String("merchant", txn.Merchant),
		zap.Float64("amount", txn.Amount),
		zap.Int64("amount_cents", row.AmountCents),
		zap.Int32("detailed_category_id", txn.DetailedCategory),
		zap.Duration("db_duration_ms", dbDuration),
		zap.Duration("total_duration_ms", totalDuration),
	)

	return txn, nil
}

func (s *TransactionService) ListUserTransactions(ctx context.Context, userID uuid.UUID, cursorToken string, pageSize int32) (ListTxResult, error) {

	start := time.Now()
	reqID, _ := ctx.Value(constants.RequestIDKey).(uuid.UUID)
	logger := s.logger.With(
		zap.Stringer("request_id", reqID),
		zap.Stringer("user_id", userID),
		zap.Int32("page_size", pageSize),
	)

	logger.Debug(evtListTxsDecodeOpaque,
		zap.String("cursor", cursorToken),
	)

	datePtr, curID, err := s.cursorSigner.DecodeCursorSigned(cursorToken)
	if err != nil {
		baseErr := errors.Join(sentinels.ErrInvalidCursor, err)
		wrapped := fmt.Errorf("list txs: %w", baseErr)
		logger.Error(evtListTxsDecodeFailed,
			zap.String("cursor", cursorToken),
			zap.Error(wrapped),
		)
		return ListTxResult{}, wrapped
	}
	var dateCursor time.Time
	if datePtr != nil {
		dateCursor = *datePtr
	}
	if !dateCursor.IsZero() {
		logger = logger.With(
			zap.Time("cursor_date", dateCursor),
			zap.Stringer("cursor_id", curID),
		)
	}
	if (datePtr == nil) != (curID == uuid.Nil) {
		wrapped := fmt.Errorf("list txs: %w", ErrInconsistentToken)
		logger.Error(evtListTxsInconToken,
			zap.Stringer("cursor_id", curID),
			zap.Time("cursor_date", dateCursor),
			zap.Error(wrapped),
		)
		return ListTxResult{}, wrapped
	}

	idCursor := curID

	fetchType := constants.FTFirst

	if datePtr != nil && curID != uuid.Nil {
		fetchType = constants.FTPag
	}

	logger.Info(evtListTxsAttempt,
		zap.String("fetch_type", fetchType),
	)

	if pageSize <= 0 {
		wrapped := fmt.Errorf("list txs: %w", ErrInvalidPageSizeNonPositive)
		logger.Warn(evtListTxsPageSize,
			zap.String("fetch_type", fetchType),
			zap.Error(wrapped),
		)
		return ListTxResult{}, wrapped
	}

	clamped := false

	if pageSize > constants.MaxPageSize {
		// to identify potential dos attacks
		logger.Warn(evtListTxsMaxPageSize,
			zap.Int64("requested_page_size", int64(pageSize)),
			zap.Int32("clamped_to", constants.MaxPageSize),
			zap.Error(ErrPageSizeTooLarge),
		)
		clamped = true
		pageSize = constants.MaxPageSize
	}

	fetchSize := pageSize + 1

	logger.Info(evtListTxsFetchAttempt,
		zap.String("fetch_type", fetchType),
		zap.Int32("page_size", pageSize))

	var rowsFirst []database.GetUserTransactionsFirstPageRow
	var rowsNext []database.GetUserTransactionsPaginatedRow
	dbStart := time.Now()

	if fetchType == constants.FTFirst {
		rowsFirst, err = s.txQueries.GetUserTransactionsFirstPage(ctx, database.GetUserTransactionsFirstPageParams{
			UserID: userID,
			Limit:  fetchSize,
		})
	} else {
		rowsNext, err = s.txQueries.GetUserTransactionsPaginated(ctx, database.GetUserTransactionsPaginatedParams{
			UserID:          userID,
			TransactionDate: dateCursor,
			ID:              idCursor,
			Limit:           fetchSize,
		})
	}
	dbDuration := time.Since(dbStart)

	if errors.Is(err, sql.ErrNoRows) {
		if fetchType == constants.FTFirst {
			logger.Debug(evtListTxsNotFound,
				zap.String("fetch_type", fetchType),
				zap.Duration("db_duration_ms", dbDuration), // custom zap encoder outputs ms
			)
			return ListTxResult{}, nil
		}
		logger.Info(evtListTxsPaginatedempty,
			zap.String("fetch_type", fetchType),
			zap.Duration("db_duration_ms", dbDuration),
		)
		return ListTxResult{}, nil
	}
	if err != nil {
		baseErr := errors.Join(sentinels.ErrDBExecFailed, err)
		wrapped := fmt.Errorf("list txs: %w", baseErr)
		logger.Error(evtListTxsDBExecFailed,
			zap.String("fetch_type", fetchType),
			zap.Duration("db_duration_ms", dbDuration),
			zap.Error(wrapped),
		)
		return ListTxResult{}, wrapped
	}
	txs := make([]models.Tx, 0, pageSize)

	if fetchType == constants.FTFirst {
		txs = helpers.AppendTxs(txs, helpers.SliceToTxRows(rowsFirst))
	} else {
		txs = helpers.AppendTxs(txs, helpers.SliceToTxRows(rowsNext))
	}

	var hasMoreData bool
	var nextCursor string

	if int32(len(txs)) > pageSize {
		hasMoreData = true
		lastTxn := txs[pageSize-1]
		nextCursor, err = s.cursorSigner.EncodeCursorSigned(lastTxn.Date, lastTxn.ID)
		if err != nil {
			hasMoreData = false
			baseErr := errors.Join(ErrEncodingFailed, err)
			wrapped := fmt.Errorf("list txs: %w", baseErr)
			logger.Error("cursor_encode_failed", zap.Error(wrapped))
			return ListTxResult{}, wrapped
		}
		txs = txs[:pageSize]
	} else {
		hasMoreData = false
	}

	totalDuration := time.Since(start)
	logger.Info(evtListTxsSuccess,
		zap.String("fetch_type", fetchType),
		zap.Int("number_of_records", len(txs)),
		zap.String("next_cursor", nextCursor),
		zap.Bool("has_more_data", hasMoreData),
		zap.Duration("db_duration_ms", dbDuration),
		zap.Duration("total_duration_ms", totalDuration), // custom zap encoder outputs ms
	)
	return ListTxResult{
		Transactions:  txs,
		NextCursor:    nextCursor,
		HasMoreData:   hasMoreData,
		Clamped:       clamped,
		EffectiveSize: pageSize,
	}, nil
}

package transaction

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/seanhuebl/unity-wealth/internal/constants"
	"github.com/seanhuebl/unity-wealth/internal/database"
	"github.com/seanhuebl/unity-wealth/internal/helpers"
	"github.com/seanhuebl/unity-wealth/internal/models"
	"github.com/seanhuebl/unity-wealth/internal/sentinels"
	"go.uber.org/zap"
)

type TransactionService struct {
	txQueries database.TransactionQuerier
	logger    *zap.Logger
}

func NewTransactionService(txQueries database.TransactionQuerier, logger *zap.Logger) *TransactionService {
	return &TransactionService{txQueries: txQueries, logger: logger}
}

func (s *TransactionService) CreateTransaction(ctx context.Context, userID string, req models.NewTxRequest) (*models.Tx, error) {
	start := time.Now()
	reqID, _ := ctx.Value(constants.RequestIDKey).(string)
	txID := uuid.NewString()

	logger := s.logger.With(
		zap.String("request_id", reqID),
		zap.String("user_id", userID),
		zap.String("tx_id", txID),
	)

	logger.Info(evtCreateTxAttempt,
		zap.String("raw_date", req.Date),
		zap.String("merchant", req.Merchant),
		zap.Float64("amount", req.Amount),
		zap.Int64("detailed_category_id", req.DetailedCategory),
	)

	parsedDate, err := time.Parse("2006-01-02", req.Date)

	if err != nil {
		baseErr := fmt.Errorf("%w: %v", ErrInvalidDateFormat, err)
		wrapped := fmt.Errorf("create tx: %w", baseErr)
		logger.Warn(evtCreateTxInvalidDateFormat,
			zap.String("raw_date", req.Date),
			zap.Error(wrapped),
		)
		return nil, wrapped
	}
	dateStr := parsedDate.Format("2006-01-02")
	tx := models.NewTransaction(
		txID,
		userID,
		dateStr,
		req.Merchant,
		req.Amount,
		req.DetailedCategory,
	)

	dbStart := time.Now()
	err = s.txQueries.CreateTransaction(ctx, database.CreateTransactionParams{
		ID:                 tx.ID,
		UserID:             tx.UserID,
		TransactionDate:    dateStr,
		Merchant:           tx.Merchant,
		AmountCents:        helpers.ConvertToCents(req.Amount),
		DetailedCategoryID: tx.DetailedCategory,
	})
	dbDuration := time.Since(dbStart)
	if err != nil {
		baseErr := fmt.Errorf("%w: %v", sentinels.ErrDBExecFailed, err)
		wrapped := fmt.Errorf("create tx: %w", baseErr)
		logger.Error(evtCreateTxDBExecFailed,
			zap.Duration("db_duration_ms", dbDuration),
			zap.Error(wrapped),
		)
		return nil, wrapped
	}
	totalDuration := time.Since(start)
	logger.Info(evtCreateTxSuccess,
		zap.String("parsed_date", dateStr),
		zap.String("merchant", req.Merchant),
		zap.Float64("amount", req.Amount),
		zap.Int64("amount_cents", helpers.ConvertToCents(req.Amount)),
		zap.Int64("detailed_category_id", req.DetailedCategory),
		zap.Duration("db_duration_ms", dbDuration),
		zap.Duration("total_duration_ms", totalDuration),
	)

	return tx, nil

}

func (s *TransactionService) UpdateTransaction(ctx context.Context, txID, userID string, req models.NewTxRequest) (*models.Tx, error) {
	start := time.Now()
	reqID, _ := ctx.Value(constants.RequestIDKey).(string)

	logger := s.logger.With(
		zap.String("request_id", reqID),
		zap.String("user_id", userID),
		zap.String("tx_id", txID),
	)

	logger.Info(evtUpdateTxAttempt,
		zap.String("raw_date", req.Date),
		zap.String("merchant", req.Merchant),
		zap.Float64("amount", req.Amount),
		zap.Int64("detailed_category_id", req.DetailedCategory),
	)

	parsedDate, err := time.Parse("2006-01-02", req.Date)
	if err != nil {
		baseErr := fmt.Errorf("%w: %v", ErrInvalidDateFormat, err)
		wrapped := fmt.Errorf("update tx: %w", baseErr)

		logger.Warn(evtUpdateTxInvalidDateFormat,
			zap.String("raw_date", req.Date),
			zap.Error(wrapped),
		)
		return nil, wrapped
	}
	dateStr := parsedDate.Format("2006-01-02")
	dbStart := time.Now()
	txRow, err := s.txQueries.UpdateTransactionByID(ctx, database.UpdateTransactionByIDParams{
		TransactionDate:    dateStr,
		Merchant:           req.Merchant,
		AmountCents:        helpers.ConvertToCents(req.Amount),
		DetailedCategoryID: req.DetailedCategory,
		UpdatedAt:          sql.NullTime{Time: time.Now(), Valid: true},
		ID:                 txID,
	})
	dbDuration := time.Since(dbStart)
	if err != nil {
		if err == sql.ErrNoRows {
			baseErr := fmt.Errorf("%w: %v", ErrTxNotFound, err)
			wrapped := fmt.Errorf("update tx: %w", baseErr)

			logger.Warn(evtUpdateTxNotFound,

				zap.Duration("db_duration_ms", dbDuration),
				zap.Error(wrapped),
			)
			return nil, wrapped
		}
		baseErr := fmt.Errorf("%w: %v", sentinels.ErrDBExecFailed, err)
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
		Date:             dateStr,
		Merchant:         txRow.Merchant,
		Amount:           helpers.CentsToDollars(txRow.AmountCents),
		DetailedCategory: txRow.DetailedCategoryID,
	}

	totalDuration := time.Since(start)
	logger.Info(evtUpdateTxSuccess,
		zap.String("parsed_date", tx.Date),
		zap.String("merchant", tx.Merchant),
		zap.Float64("amount", tx.Amount),
		zap.Int64("amount_cents", txRow.AmountCents),
		zap.Int64("detailed_category_id", tx.DetailedCategory),
		zap.Duration("db_duration_ms", dbDuration),
		zap.Duration("total_duration_ms", totalDuration),
	)
	return tx, nil
}

func (s *TransactionService) DeleteTransaction(ctx context.Context, userID, txID string) error {
	start := time.Now()
	reqID, _ := ctx.Value(constants.RequestIDKey).(string)

	logger := s.logger.With(
		zap.String("request_id", reqID),
		zap.String("user_id", userID),
		zap.String("tx_id", txID),
	)

	logger.Info(evtDeleteTxAttempt)

	dbStart := time.Now()
	_, err := s.txQueries.DeleteTransactionByID(ctx, database.DeleteTransactionByIDParams{
		ID:     txID,
		UserID: userID,
	})

	dbDuration := time.Since(dbStart)

	if err != nil {
		if err == sql.ErrNoRows {
			baseErr := fmt.Errorf("%w: %v", ErrTxNotFound, err)
			wrapped := fmt.Errorf("delete tx: %w", baseErr)

			logger.Warn(evtDeleteTxNotFound,
				zap.Duration("db_duration_ms", dbDuration),
				zap.Error(wrapped),
			)
			return wrapped
		}
		baseErr := fmt.Errorf("%w: %v", sentinels.ErrDBExecFailed, err)
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

func (s *TransactionService) GetTransactionByID(ctx context.Context, userID, txID string) (*models.Tx, error) {
	start := time.Now()
	reqID, _ := ctx.Value(constants.RequestIDKey).(string)

	logger := s.logger.With(
		zap.String("request_id", reqID),
		zap.String("user_id", userID),
		zap.String("tx_id", txID),
	)

	logger.Info(evtGetTxAttempt)

	dbStart := time.Now()
	row, err := s.txQueries.GetUserTransactionByID(ctx, database.GetUserTransactionByIDParams{UserID: userID, ID: txID})
	if err != nil {
		if err == sql.ErrNoRows {
			baseErr := fmt.Errorf("%w: %v", ErrTxNotFound, err)
			wrapped := fmt.Errorf("get tx: %w", baseErr)

			logger.Warn(evtGetTxNotFound,
				zap.Error(wrapped),
			)
			return nil, wrapped
		}
		baseErr := fmt.Errorf("%w: %v", sentinels.ErrDBExecFailed, err)
		wrapped := fmt.Errorf("get tx: %w", baseErr)
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
		zap.String("parsed_date", txn.Date),
		zap.String("merchant", txn.Merchant),
		zap.Float64("amount", txn.Amount),
		zap.Int64("amount_cents", row.AmountCents),
		zap.Int64("detailed_category_id", txn.DetailedCategory),
		zap.Duration("db_duration_ms", dbDuration),
		zap.Duration("total_duration_ms", totalDuration),
	)

	return txn, nil
}

func (s *TransactionService) ListUserTransactions(
	ctx context.Context,
	userID uuid.UUID,
	cursorDate *string,
	cursorID *string,
	pageSize int64,
) (transactions []models.Tx, nextCursorDate, nextCursorID string, hasMoreData bool, e error) {
	start := time.Now()
	reqID, _ := ctx.Value(constants.RequestIDKey).(string)
	dateCursor, idCursor := "", ""
	fetchType := "first_page"

	if cursorDate != nil && cursorID != nil {
		fetchType = "paginated"
		dateCursor = *cursorDate
		idCursor = *cursorID

	}

	logger := s.logger.With(
		zap.String("request_id", reqID),
		zap.String("user_id", userID.String()),
		zap.Int64("page_size", pageSize),
	)

	logger.Info(evtListTxsAttempt,
		zap.String("fetch_type", fetchType),
		zap.String("cursor_date", dateCursor),
		zap.String("cursor_id", idCursor))

	if pageSize <= 0 {
		baseErr := fmt.Errorf("%w: %v", ErrInvalidPageSize, errors.New("pageSize <= 0"))
		wrapped := fmt.Errorf("list txs: %w", baseErr)
		logger.Warn(evtListTxsPageSize,
			zap.String("fetch_type", fetchType),
			zap.String("cursor_date", dateCursor),
			zap.String("cursor_id", idCursor),
			zap.Error(wrapped),
		)
		return nil, "", "", false, wrapped
	}
	fetchSize := pageSize + 1

	logger.Info(evtListTxsFetchAttempt,
		zap.String("fetch_type", fetchType),
		zap.String("cursor_date", dateCursor),
		zap.String("cursor_id", idCursor))

	var rowsFirst []database.GetUserTransactionsFirstPageRow
	var rowsNext []database.GetUserTransactionsPaginatedRow
	var err error

	dbStart := time.Now()

	if fetchType == "first_page" {
		rowsFirst, err = s.txQueries.GetUserTransactionsFirstPage(ctx, database.GetUserTransactionsFirstPageParams{
			UserID: userID.String(),
			Limit:  fetchSize,
		})
	} else {
		rowsNext, err = s.txQueries.GetUserTransactionsPaginated(ctx, database.GetUserTransactionsPaginatedParams{
			UserID:          userID.String(),
			TransactionDate: dateCursor,
			ID:              idCursor,
			Limit:           fetchSize,
		})
	}
	dbDuration := time.Since(dbStart)

	if err == sql.ErrNoRows {
		if fetchType == "first_page" {
			logger.Warn(evtListTxsNotFound,
				zap.String("fetch_type", fetchType),
				zap.String("cursor_date", dateCursor),
				zap.String("cursor_id", idCursor),
				zap.Duration("db_duration_ms", dbDuration),
			)
			return []models.Tx{}, "", "", false, nil
		}
		logger.Info(evtListTxsPaginatedempty,
			zap.String("fetch_type", fetchType),
			zap.String("cursor_date", dateCursor),
			zap.String("cursor_id", idCursor),
			zap.Duration("db_duration_ms", dbDuration),
		)
		return []models.Tx{}, "", "", false, nil
	}
	if err != nil {
		baseErr := fmt.Errorf("%w: %v", sentinels.ErrDBExecFailed, err)
		wrapped := fmt.Errorf("list txs: %w", baseErr)
		logger.Error(evtListTxsDBExecFailed,
			zap.String("fetch_type", fetchType),
			zap.String("cursor_date", dateCursor),
			zap.String("cursor_id", idCursor),
			zap.Duration("db_duration_ms", dbDuration),
			zap.Error(wrapped),
		)
		return nil, "", "", false, wrapped
	}
	txs := make([]models.Tx, 0, pageSize)

	if fetchType == "first_page" {
		for _, r := range rowsFirst {
			txs = append(txs, MapToTx(
				r.ID, r.UserID,
				r.TransactionDate,
				r.Merchant,
				r.AmountCents,
				r.DetailedCategoryID,
			))
		}
	} else {
		for _, r := range rowsNext {
			txs = append(txs, MapToTx(
				r.ID, r.UserID,
				r.TransactionDate,
				r.Merchant,
				r.AmountCents,
				r.DetailedCategoryID),
			)
		}
	}

	if int64(len(txs)) > pageSize {
		hasMoreData = true
		lastTxn := txs[pageSize-1]
		nextCursorDate = lastTxn.Date
		nextCursorID = lastTxn.ID
		txs = txs[:pageSize]

	} else {
		hasMoreData = false
	}
	totalDuration := time.Since(start)
	logger.Info(evtListTxsSuccess,
		zap.String("fetch_type", fetchType),
		zap.Int("number_of_records", len(txs)),
		zap.String("next_cursor_date", nextCursorDate),
		zap.String("next_cursor_id", nextCursorID),
		zap.Bool("has_more_data", hasMoreData),
		zap.Duration("db_duration_ms", dbDuration),
		zap.Duration("total_duration_ms", totalDuration),
	)
	return txs, nextCursorDate, nextCursorID, hasMoreData, nil
}

func MapToTx(id, userID, date, merchant string,
	amountCents, detailedCategoryID int64,
) models.Tx {
	return models.Tx{
		ID:               id,
		UserID:           userID,
		Date:             date,
		Merchant:         merchant,
		Amount:           helpers.CentsToDollars(amountCents),
		DetailedCategory: detailedCategoryID,
	}
}

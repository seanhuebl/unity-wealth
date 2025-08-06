package testhelpers

import (
	"context"
	"database/sql"
	"os"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	_ "github.com/jackc/pgx/v5/stdlib"
	httpauth "github.com/seanhuebl/unity-wealth/handlers/auth"
	txhandler "github.com/seanhuebl/unity-wealth/handlers/transaction"
	httpuser "github.com/seanhuebl/unity-wealth/handlers/user"
	"github.com/seanhuebl/unity-wealth/internal/constants"
	"github.com/seanhuebl/unity-wealth/internal/cursor"
	"github.com/seanhuebl/unity-wealth/internal/database"
	"github.com/seanhuebl/unity-wealth/internal/helpers"
	"github.com/seanhuebl/unity-wealth/internal/interfaces"
	"github.com/seanhuebl/unity-wealth/internal/middleware"
	"github.com/seanhuebl/unity-wealth/internal/models"
	"github.com/seanhuebl/unity-wealth/internal/services/auth"
	"github.com/seanhuebl/unity-wealth/internal/services/transaction"
	"github.com/seanhuebl/unity-wealth/internal/services/user"
	"github.com/seanhuebl/unity-wealth/internal/testmodels"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func CreateTestingSchema(t *testing.T, db *sql.DB) {
	_, err := db.Exec(constants.CreateUsersTable)
	require.NoError(t, err)
	_, err = db.Exec(constants.CreatePrimCatTable)
	require.NoError(t, err)
	_, err = db.Exec(constants.CreateDetCatTable)
	require.NoError(t, err)
	_, err = db.Exec(constants.CreateTxTable)
	require.NoError(t, err)
	_, err = db.Exec(constants.CreateDeviceInfoTable)
	require.NoError(t, err)
	_, err = db.Exec(constants.CreateRefrTokenTable)
	require.NoError(t, err)
}

func SeedTestUser(t *testing.T, userQ database.UserQuerier, userID uuid.UUID, requiresHash bool) {
	var hashedPassword string
	if requiresHash {
		hashedPassword, _ = auth.NewRealPwdHasher().HashPassword("Validpass1!")
	} else {
		hashedPassword = "hashedpwd"
	}
	email := "user@example.com"

	err := userQ.CreateUser(context.Background(), database.CreateUserParams{
		ID:             userID,
		Email:          email,
		HashedPassword: hashedPassword,
	})
	require.NoError(t, err)
}

func SeedTestCategories(t *testing.T, db *sql.DB) {
	_, err := db.Exec(`
	INSERT INTO primary_categories (id, name)
	VALUES ($1, $2)
	`, 7, "Food")
	require.NoError(t, err)

	_, err = db.Exec(`
	INSERT INTO detailed_categories (id, name, description, primary_category_id)
	VALUES ($1, $2, $3, $4)
	`, 40, "Groceries", "Purchases for fresh produce and groceries, including farmers' markets", 7)
	require.NoError(t, err)
}

func SeedTestTransaction(t *testing.T, txQ database.TransactionQuerier, userID, txID uuid.UUID, req *models.NewTxRequest) {
	ctx := context.Background()
	date, err := time.Parse(constants.LayoutDate, req.Date)
	if err != nil {
		t.Fatalf("error parsing time")
	}
	err = txQ.CreateTransaction(ctx, database.CreateTransactionParams{
		ID:                 txID,
		UserID:             userID,
		TransactionDate:    date,
		Merchant:           req.Merchant,
		AmountCents:        helpers.ConvertToCents(req.Amount),
		DetailedCategoryID: req.DetailedCategory,
	})
	require.NoError(t, err)
}

func SeedMultipleTestTransactions[T interfaces.TxPageRow](t *testing.T, txQ database.TransactionQuerier, rows []T) {
	ctx := context.Background()
	for _, row := range rows {
		err := txQ.CreateTransaction(ctx, database.CreateTransactionParams{
			ID:                 row.GetTxID(),
			UserID:             row.GetUserID(),
			TransactionDate:    row.GetTxDate(),
			Merchant:           row.GetMerchant(),
			AmountCents:        row.GetAmountCents(),
			DetailedCategoryID: row.GetDetailedCatID(),
		})
		require.NoError(t, err)
	}
}

func SeedCreateTxTestData(t *testing.T, db *sql.DB, userQ database.UserQuerier, userID uuid.UUID) {
	SeedTestUser(t, userQ, userID, false)
	SeedTestCategories(t, db)
}

func SetupTestEnv(t *testing.T) *testmodels.TestEnv {

	t.Helper()

	db := NewDB(t)

	transactionalQ := database.NewRealTransactionalQuerier(database.New(db))

	txQ := database.NewRealTransactionQuerier(transactionalQ)
	userQ := database.NewRealUserQuerier(transactionalQ)
	tokenQ := database.NewRealTokenQuerier(transactionalQ)
	deviceQ := database.NewRealDevicequerier(transactionalQ)
	sqlTxQ := database.NewRealSqlTxQuerier(transactionalQ)
	pwdHasher := auth.NewRealPwdHasher()
	tokenGen := auth.NewRealTokenGenerator("your-secret-key", "your-issuer")
	tokenExtractor := auth.NewRealTokenExtractor()
	mw := middleware.NewMiddleware(tokenGen, tokenExtractor)
	secretB64 := os.Getenv("ENCODE_CURSOR_SECRET")
	if secretB64 == "" {
		secretB64 = "dGVzdC1zZWNyZXQ=" // "test-secret"
	}
	cs, _ := cursor.NewSigner(secretB64)

	testLogger := zap.NewNop()

	authSvc := auth.NewAuthService(sqlTxQ, userQ, tokenGen, tokenExtractor, pwdHasher, testLogger)
	txSvc := transaction.NewTransactionService(txQ, cs, testLogger)
	userSvc := user.NewUserService(userQ, pwdHasher, testLogger)

	txH := txhandler.NewHandler(txSvc)
	authH := httpauth.NewHandler(authSvc)
	userH := httpuser.NewHandler(userSvc)

	r := gin.New()
	return &testmodels.TestEnv{
		Router:         r,
		Db:             db,
		UserQ:          userQ,
		TxQ:            txQ,
		TokenQ:         tokenQ,
		DeviceQ:        deviceQ,
		SqlTxQ:         sqlTxQ,
		TransactionalQ: transactionalQ,
		Logger:         testLogger,
		Services: &testmodels.Services{
			AuthService: authSvc,
			TxService:   txSvc,
			UserService: userSvc,
		},
		Middleware: mw,
		Handlers: &testmodels.Handlers{
			AuthHandler: authH,
			TxHandler:   txH,
			UserHandler: userH,
		},
	}
}

func IsTxFound(t *testing.T, tc testmodels.BaseHTTPTestCase, txID uuid.UUID, env *testmodels.TestEnv) {
	if tc.Name == "not found" {
		SeedTestTransaction(t, env.TxQ, tc.UserID, uuid.New(), &models.NewTxRequest{
			Date:             "2025-03-05",
			Merchant:         "costco",
			Amount:           125.98,
			DetailedCategory: 40,
		})
	} else {
		SeedTestTransaction(t, env.TxQ, tc.UserID, txID, &models.NewTxRequest{
			Date:             "2025-03-05",
			Merchant:         "costco",
			Amount:           125.98,
			DetailedCategory: 40,
		})
	}
}

func SeedTestDeviceInfo(t *testing.T, deviceQ database.DeviceQuerier, userID uuid.UUID) {
	_, err := deviceQ.CreateDeviceInfo(context.Background(), database.CreateDeviceInfoParams{
		ID:             uuid.New(),
		UserID:         userID,
		DeviceType:     "Mobile",
		Browser:        "Chrome",
		BrowserVersion: "100.0",
		Os:             "Android",
		OsVersion:      "11",
	})
	require.NoError(t, err)
}

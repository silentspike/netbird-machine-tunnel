package storetest

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"slices"
	"strings"
	"time"

	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	storepkg "github.com/netbirdio/netbird/management/server/store"
	"github.com/netbirdio/netbird/management/server/testutil"
	"github.com/netbirdio/netbird/management/server/types"
)

const (
	storeSqliteFileName = "store.db"

	postgresDsnEnv       = "NB_STORE_ENGINE_POSTGRES_DSN"
	postgresDsnEnvLegacy = "NETBIRD_STORE_ENGINE_POSTGRES_DSN"
	mysqlDsnEnv          = "NB_STORE_ENGINE_MYSQL_DSN"
	mysqlDsnEnvLegacy    = "NETBIRD_STORE_ENGINE_MYSQL_DSN"
)

var supportedEngines = []types.Engine{types.SqliteStoreEngine, types.PostgresStoreEngine, types.MysqlStoreEngine}

// NewTestStoreFromSQL creates a store for tests outside the store package.
func NewTestStoreFromSQL(ctx context.Context, filename string, dataDir string) (storepkg.Store, func(), error) {
	kind := getStoreEngineFromEnv()
	if kind == "" {
		kind = types.SqliteStoreEngine
	}

	storeStr := fmt.Sprintf("%s?cache=shared", storeSqliteFileName)
	if runtime.GOOS == "windows" {
		// To avoid `The process cannot access the file because it is being used by another process` on Windows
		storeStr = storeSqliteFileName
	}

	file := filepath.Join(dataDir, storeStr)
	db, err := gorm.Open(sqlite.Open(file), getGormConfig())
	if err != nil {
		return nil, nil, err
	}

	if filename != "" {
		err = LoadSQL(db, filename)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to load SQL file: %v", err)
		}
	}

	store, err := storepkg.NewSqlStore(ctx, db, types.SqliteStoreEngine, nil, false)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create test store: %v", err)
	}

	err = addAllGroupToAccount(ctx, store)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to add all group to account: %v", err)
	}

	var sqlStore storepkg.Store
	var cleanup func()

	maxRetries := 2
	for i := 0; i < maxRetries; i++ {
		sqlStore, cleanup, err = getSqlStoreEngine(ctx, store, kind)
		if err == nil {
			return sqlStore, cleanup, nil
		}
		if i < maxRetries-1 {
			time.Sleep(100 * time.Millisecond)
		}
	}
	return nil, nil, fmt.Errorf("failed to create test store after %d attempts: %v", maxRetries, err)
}

func getGormConfig() *gorm.Config {
	return &gorm.Config{
		Logger:          logger.Default.LogMode(logger.Silent),
		CreateBatchSize: 400,
	}
}

func getStoreEngineFromEnv() types.Engine {
	kind, ok := os.LookupEnv("NETBIRD_STORE_ENGINE")
	if !ok {
		return ""
	}

	value := types.Engine(strings.ToLower(kind))
	if slices.Contains(supportedEngines, value) {
		return value
	}

	return types.SqliteStoreEngine
}

func lookupDSNEnv(nbKey, legacyKey string) (string, bool) {
	if v, ok := os.LookupEnv(nbKey); ok {
		return v, true
	}
	return os.LookupEnv(legacyKey)
}

func addAllGroupToAccount(ctx context.Context, store storepkg.Store) error {
	allAccounts := store.GetAllAccounts(ctx)
	for _, account := range allAccounts {
		shouldSave := false

		_, err := account.GetGroupAll()
		if err != nil {
			if err := account.AddAllGroup(false); err != nil {
				return err
			}
			shouldSave = true
		}

		if shouldSave {
			err = store.SaveAccount(ctx, account)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func getSqlStoreEngine(ctx context.Context, store *storepkg.SqlStore, kind types.Engine) (storepkg.Store, func(), error) {
	var cleanup func()
	var err error
	switch kind {
	case types.PostgresStoreEngine:
		store, cleanup, err = newReusedPostgresStore(ctx, store, kind)
	case types.MysqlStoreEngine:
		store, cleanup, err = newReusedMysqlStore(ctx, store, kind)
	default:
		cleanup = func() {
			// sqlite doesn't need to be cleaned up
		}
	}
	if err != nil {
		return nil, cleanup, fmt.Errorf("failed to create test store: %v", err)
	}

	closeConnection := func() {
		cleanup()
		store.Close(ctx)
	}

	return store, closeConnection, nil
}

func newReusedPostgresStore(ctx context.Context, store *storepkg.SqlStore, kind types.Engine) (*storepkg.SqlStore, func(), error) {
	dsn, ok := lookupDSNEnv(postgresDsnEnv, postgresDsnEnvLegacy)
	if !ok || dsn == "" {
		var err error
		_, dsn, err = testutil.CreatePostgresTestContainer()
		if err != nil {
			return nil, nil, err
		}
	}

	if dsn == "" {
		return nil, nil, fmt.Errorf("%s is not set", postgresDsnEnv)
	}

	db, err := openDBWithRetry(dsn, kind, 5)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to open postgres connection: %v", err)
	}

	dsn, cleanup, err := createRandomDB(dsn, db, kind)

	sqlDB, _ := db.DB()
	if sqlDB != nil {
		sqlDB.Close()
	}

	if err != nil {
		return nil, nil, err
	}

	store, err = storepkg.NewPostgresqlStoreFromSqlStore(ctx, store, dsn, nil)
	if err != nil {
		return nil, nil, err
	}

	return store, cleanup, nil
}

func newReusedMysqlStore(ctx context.Context, store *storepkg.SqlStore, kind types.Engine) (*storepkg.SqlStore, func(), error) {
	dsn, ok := lookupDSNEnv(mysqlDsnEnv, mysqlDsnEnvLegacy)
	if !ok || dsn == "" {
		var err error
		_, dsn, err = testutil.CreateMysqlTestContainer()
		if err != nil {
			return nil, nil, err
		}
	}

	if dsn == "" {
		return nil, nil, fmt.Errorf("%s is not set", mysqlDsnEnv)
	}

	db, err := openDBWithRetry(dsn, kind, 5)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to open mysql connection: %v", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get underlying sql.DB: %v", err)
	}
	sqlDB.SetMaxOpenConns(1)
	sqlDB.SetMaxIdleConns(1)

	dsn, cleanup, err := createRandomDB(dsn, db, kind)

	sqlDB.Close()

	if err != nil {
		return nil, nil, err
	}

	store, err = storepkg.NewMysqlStoreFromSqlStore(ctx, store, dsn, nil)
	if err != nil {
		return nil, nil, err
	}

	return store, cleanup, nil
}

func openDBWithRetry(dsn string, engine types.Engine, maxRetries int) (*gorm.DB, error) {
	var db *gorm.DB
	var err error

	for i := range maxRetries {
		switch engine {
		case types.PostgresStoreEngine:
			db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
		case types.MysqlStoreEngine:
			db, err = gorm.Open(mysql.Open(dsn+"?charset=utf8&parseTime=True&loc=Local"), &gorm.Config{})
		}

		if err == nil {
			return db, nil
		}

		if i < maxRetries-1 {
			waitTime := time.Duration(100*(i+1)) * time.Millisecond
			time.Sleep(waitTime)
		}
	}

	return nil, err
}

func createRandomDB(dsn string, db *gorm.DB, engine types.Engine) (string, func(), error) {
	dbName := fmt.Sprintf("test_db_%s", strings.ReplaceAll(uuid.New().String(), "-", "_"))

	if err := db.Exec(fmt.Sprintf("CREATE DATABASE %s", dbName)).Error; err != nil {
		return "", nil, fmt.Errorf("failed to create database: %v", err)
	}

	originalDSN := dsn

	cleanup := func() {
		var dropDB *gorm.DB
		var err error

		switch engine {
		case types.PostgresStoreEngine:
			dropDB, err = gorm.Open(postgres.Open(originalDSN), &gorm.Config{
				SkipDefaultTransaction: true,
				PrepareStmt:            false,
			})
			if err != nil {
				log.Errorf("failed to connect for dropping database %s: %v", dbName, err)
				return
			}
			defer func() {
				if sqlDB, _ := dropDB.DB(); sqlDB != nil {
					sqlDB.Close()
				}
			}()

			if sqlDB, _ := dropDB.DB(); sqlDB != nil {
				sqlDB.SetMaxOpenConns(1)
				sqlDB.SetMaxIdleConns(0)
				sqlDB.SetConnMaxLifetime(time.Second)
			}

			err = dropDB.Exec(fmt.Sprintf("DROP DATABASE IF EXISTS %s WITH (FORCE)", dbName)).Error

		case types.MysqlStoreEngine:
			dropDB, err = gorm.Open(mysql.Open(originalDSN+"?charset=utf8&parseTime=True&loc=Local"), &gorm.Config{
				SkipDefaultTransaction: true,
				PrepareStmt:            false,
			})
			if err != nil {
				log.Errorf("failed to connect for dropping database %s: %v", dbName, err)
				return
			}
			defer func() {
				if sqlDB, _ := dropDB.DB(); sqlDB != nil {
					sqlDB.Close()
				}
			}()

			if sqlDB, _ := dropDB.DB(); sqlDB != nil {
				sqlDB.SetMaxOpenConns(1)
				sqlDB.SetMaxIdleConns(0)
				sqlDB.SetConnMaxLifetime(time.Second)
			}

			err = dropDB.Exec(fmt.Sprintf("DROP DATABASE IF EXISTS %s", dbName)).Error
		}

		if err != nil {
			log.Errorf("failed to drop database %s: %v", dbName, err)
		}
	}

	return replaceDBName(dsn, dbName), cleanup, nil
}

func replaceDBName(dsn, newDBName string) string {
	re := regexp.MustCompile(`(?P<pre>[:/@])(?P<dbname>[^/?]+)(?P<post>\?|$)`)
	return re.ReplaceAllString(dsn, `${pre}`+newDBName+`${post}`)
}

func LoadSQL(db *gorm.DB, filepath string) error {
	sqlContent, err := os.ReadFile(filepath)
	if err != nil {
		return err
	}

	queries := strings.Split(string(sqlContent), ";")

	for _, query := range queries {
		query = strings.TrimSpace(query)
		if query != "" {
			err := db.Exec(query).Error
			if err != nil {
				return err
			}
		}
	}

	return nil
}

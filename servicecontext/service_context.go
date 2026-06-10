package servicecontext

import (
	"context"
	"log/slog"
	"sync"
	"time"

	"github.com/KitHub/DataManagementPlatform_ComputeEngineAPI/component"
	"github.com/KitHub/DataManagementPlatform_ComputeEngineAPI/config"
	"github.com/KitHub/DataManagementPlatform_ComputeEngineAPI/dao"
	"github.com/KitHub/DataManagementPlatform_ComputeEngineAPI/logic"
	"github.com/KitHub/DataManagementPlatform_ComputeEngineAPI/service"
	"gopkg.in/natefinch/lumberjack.v2"
	"xorm.io/xorm"
)

type ServiceContext struct {
	Logger               *slog.Logger
	DBEngine             *xorm.Engine
	OSSClient            component.OSSComponent
	ShutdownLogic        *logic.ShutdownLogic
	PackageDAO           *dao.PackageDAO
	PackageLogic         *logic.PackageLogic
	ComputeEngineService *service.ComputeEngineService
}

var gServiceCtx *ServiceContext
var once sync.Once

func InitServiceContext(ctx context.Context, configEntity *config.ConfigEntity) (
	serviceCtx *ServiceContext, err error) {
	slog.InfoContext(ctx, "init service context")

	once.Do(func() {
		logger, innerErr := initLog(ctx, configEntity.LogConfig)
		if innerErr != nil {
			slog.ErrorContext(ctx, "init log failed", slog.Any("error", innerErr))
			err = innerErr
			return
		}

		shutdownLogic := logic.NewShutdownLogic()
		dbEngine, innerErr := initDB(ctx, configEntity.DBConfig, shutdownLogic)
		if innerErr != nil {
			slog.ErrorContext(ctx, "init database failed", slog.Any("error", innerErr))
			err = innerErr
			return
		}
		packageDAO := dao.NewPackageDAO(ctx)
		packageLogic := logic.NewPackageLogic(dbEngine, packageDAO)
		computeEngineService := service.NewComputeEngineService(packageLogic)
		ossClient, innerErr := initOSSClient(ctx, configEntity.OSSConfig)
		if innerErr != nil {
			slog.ErrorContext(ctx, "init oss client failed", slog.Any("error", innerErr))
			err = innerErr
			return
		}

		gServiceCtx = &ServiceContext{
			ShutdownLogic:        shutdownLogic,
			PackageDAO:           packageDAO,
			PackageLogic:         packageLogic,
			ComputeEngineService: computeEngineService,
			Logger:               logger,
			DBEngine:             dbEngine,
			OSSClient:            ossClient,
		}
	})

	slog.InfoContext(ctx, "init service context done")
	return gServiceCtx, err
}

func initLog(ctx context.Context, logConfig *config.LogConfigEntity) (
	*slog.Logger, error) {
	log := &lumberjack.Logger{
		Filename:   logConfig.Filename,   // 日志文件路径
		MaxSize:    logConfig.MaxSize,    // 每个日志文件的最大大小（以MB为单位）
		MaxBackups: logConfig.MaxBackups, // 保留旧文件的最大数量
		MaxAge:     logConfig.MaxAge,     // 保留旧文件的最大天数
		Compress:   logConfig.Compress,   // 是否压缩旧文件
		LocalTime:  logConfig.LocalTime,  // 是否使用本地时间戳
	}
	serviceLogger := slog.New(slog.NewTextHandler(log, nil))
	slog.SetDefault(serviceLogger)
	return serviceLogger, nil
}

func initDB(ctx context.Context, dbConfig *config.DBConfigEntity,
	shutdownLogic *logic.ShutdownLogic) (*xorm.Engine, error) {
	engine, err := xorm.NewEngine(dbConfig.DriverName, dbConfig.DataSourceName)
	if err != nil {
		slog.ErrorContext(ctx, "Failed to initialize database connection",
			slog.Any("error", err))
		return nil, err
	}

	engine.SetMaxIdleConns(dbConfig.MaxIdleConns)
	engine.SetMaxOpenConns(dbConfig.MaxOpenConns)
	engine.SetConnMaxLifetime(
		time.Duration(dbConfig.ConnMaxLifetime) * time.Second)

	err = engine.PingContext(ctx)
	if err != nil {
		slog.ErrorContext(ctx, "Failed to ping database", slog.Any("error", err))
		return nil, err
	}

	shutdownLogic.RegisterShutdownCallback(func(ctx context.Context) error {
		return engine.Close()
	})

	slog.InfoContext(ctx, "Database connection initialized successfully")
	return engine, nil
}

func initOSSClient(ctx context.Context, config *config.OSSConfigEntity) (component.OSSComponent, error) {
	slog.InfoContext(ctx, "init oss client", slog.Any("config", config))
	switch config.OSSClientType {
	case "aliyun":
		client := component.NewAliyunOSSComponent(ctx, config.AliyunOSSConfig.OSSAccessKeyId, config.AliyunOSSConfig.OSSAccessKeySecret, config.AliyunOSSConfig.OSSRegion)
		return client, nil
	case "tencent":
		client := component.NewTencentCOSComponent(ctx, config.TencentCSSConfig.COSSecretId, config.TencentCSSConfig.COSSecretKey, config.TencentCSSConfig.COSBucketUrl)
		return client, nil
	default:
		slog.ErrorContext(ctx, "unsupported oss client", slog.Any("type", config.OSSClientType))
		return nil, nil
	}
}

func GetServiceContext() *ServiceContext {
	return gServiceCtx
}

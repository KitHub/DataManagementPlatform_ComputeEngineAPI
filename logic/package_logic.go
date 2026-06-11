package logic

import (
	"context"
	"log/slog"
	"sync"
	"time"

	"github.com/KitHub/DataManagementPlatform_ComputeEngineAPI/dao"
	"github.com/KitHub/DataManagementPlatform_ComputeEngineAPI/entity"
	"github.com/KitHub/DataManagementPlatform_ComputeEngineAPI/wrapper"
	"xorm.io/xorm"
)

var packageLogicInstance *PackageLogic
var onceForPackageLogicInstance sync.Once = sync.Once{}

type PackageLogic struct {
	dbEngine   *xorm.Engine
	packageDAO *dao.PackageDAO
}

func NewPackageLogic(dbEngine *xorm.Engine,
	packageDAO *dao.PackageDAO) *PackageLogic {
	onceForPackageLogicInstance.Do(func() {
		packageLogicInstance = &PackageLogic{
			dbEngine:   dbEngine,
			packageDAO: packageDAO,
		}
	})
	return packageLogicInstance
}

func (logic *PackageLogic) InsertPackage(ctx context.Context, originId string, comment string, platform string, bucketName string, keyName string) (packageEntity *entity.PackageEntity, err error) {
	now := time.Now()
	packageEntity = &entity.PackageEntity{
		OriginId:     originId,
		Comment:      comment,
		BucketName:   bucketName,
		KeyName:      keyName,
		Platform:     platform,
		RegisterTime: now,
		CreateTime:   now,
		UpdateTime:   now,
	}
	slog.InfoContext(ctx, "register package", slog.Any("package", packageEntity))

	err = wrapper.TransactionWrapper(ctx, logic.dbEngine,
		func(session *xorm.Session) error {
			packageEntity, err = logic.packageDAO.InsertPackage(ctx, session, packageEntity)
			return err
		})
	slog.InfoContext(ctx, "register package success", slog.Any("package", packageEntity))
	return packageEntity, nil
}

func (logic *PackageLogic) QueryPackageById(ctx context.Context, id int64) (packageEntity *entity.PackageEntity, err error) {
	slog.InfoContext(ctx, "query package", slog.Any("id", id))
	packageEntity = &entity.PackageEntity{
		ID: id,
	}
	err = wrapper.TransactionWrapper(ctx, logic.dbEngine,
		func(session *xorm.Session) error {
			packageEntity, err = logic.packageDAO.QueryPackageById(ctx, session, id)
			return err
		})
	slog.InfoContext(ctx, "query package success", slog.Any("package", packageEntity))
	return packageEntity, err
}

func (logic *PackageLogic) QueryPackageByOriginId(ctx context.Context, originId string) (packageEntity *entity.PackageEntity, err error) {
	slog.InfoContext(ctx, "query package", slog.Any("originId", originId))
	packageEntity = &entity.PackageEntity{
		OriginId: originId,
	}
	err = wrapper.TransactionWrapper(ctx, logic.dbEngine,
		func(session *xorm.Session) error {
			packageEntity, err = logic.packageDAO.QueryPackageByOriginId(ctx, session, originId)
			return err
		})
	slog.InfoContext(ctx, "query package success", slog.Any("package", packageEntity))
	return packageEntity, err
}

func (logic *PackageLogic) QueryPackageListASCById(ctx context.Context, id int64, limit int32) (packages []*entity.PackageEntity, err error) {
	slog.InfoContext(ctx, "query packages", slog.Any("id", id), slog.Any("limit", limit))
	packages = make([]*entity.PackageEntity, 0)
	err = wrapper.TransactionWrapper(ctx, logic.dbEngine,
		func(session *xorm.Session) error {
			packages, err = logic.packageDAO.QueryPackageListByIdASC(ctx, session, id, limit)
			return err
		})
	slog.InfoContext(ctx, "query package success", slog.Any("packages", packages))
	return packages, err
}

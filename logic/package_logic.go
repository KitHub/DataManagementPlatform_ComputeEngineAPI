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

func (logic *PackageLogic) InsertPackage(ctx context.Context, name string, description string, url string, platform string) (packageEntity *entity.PackageEntity, err error) {
	now := time.Now()
	packageEntity = &entity.PackageEntity{
		Name:        name,
		Description: description,
		Url:         url,
		Platform:    platform,
		CreateTime:  now,
		UpdateTime:  now,
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

func (logic *PackageLogic) QueryPackageById(ctx context.Context, packageID string) (packageEntity *entity.PackageEntity, err error) {
	slog.InfoContext(ctx, "query package", slog.Any("packageID", packageID))
	packageEntity = &entity.PackageEntity{
		ID: packageID,
	}
	err = wrapper.TransactionWrapper(ctx, logic.dbEngine,
		func(session *xorm.Session) error {
			packageEntity, err = logic.packageDAO.QueryPackageById(ctx, session, packageID)
			return err
		})
	slog.InfoContext(ctx, "query package success", slog.Any("package", packageEntity))
	return packageEntity, err
}

func (logic *PackageLogic) QueryPackageName(ctx context.Context, packageName string) (packageEntity *entity.PackageEntity, err error) {
	slog.InfoContext(ctx, "query package", slog.Any("packageName", packageName))
	packageEntity = &entity.PackageEntity{
		Name: packageName,
	}
	err = wrapper.TransactionWrapper(ctx, logic.dbEngine,
		func(session *xorm.Session) error {
			packageEntity, err = logic.packageDAO.QueryPackageByName(ctx, session, packageName)
			return err
		})
	slog.InfoContext(ctx, "query package success", slog.Any("package", packageEntity))
	return packageEntity, err
}

func (logic *PackageLogic) QueryPackageListByIdASC(ctx context.Context, packageID int64, limit int32) (packages []*entity.PackageEntity, err error) {
	slog.InfoContext(ctx, "query packages", slog.Any("packageID", packageID), slog.Any("limit", limit))
	packages = make([]*entity.PackageEntity, 0)
	err = wrapper.TransactionWrapper(ctx, logic.dbEngine,
		func(session *xorm.Session) error {
			packages, err = logic.packageDAO.QueryPackageListByIdASC(ctx, session, packageID, limit)
			return err
		})
	slog.InfoContext(ctx, "query package success", slog.Any("packages", packages))
	return packages, err
}

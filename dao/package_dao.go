package dao

import (
	"context"
	"errors"
	"log/slog"
	"sync"

	"github.com/KitHub/DataManagementPlatform_ComputeEngineAPI/entity"
	"xorm.io/xorm"
)

var packageDAOInstance *PackageDAO
var onceForPackageDAOInstance sync.Once = sync.Once{}

type PackageDAO struct {
}

func NewPackageDAO(ctx context.Context) *PackageDAO {
	onceForPackageDAOInstance.Do(func() {
		packageDAOInstance = &PackageDAO{}
	})
	return packageDAOInstance
}

func (dao *PackageDAO) InsertPackage(ctx context.Context,
	session *xorm.Session, packageEntity *entity.PackageEntity) (*entity.PackageEntity, error) {

	slog.InfoContext(ctx, "Inserting package", slog.Any("package", packageEntity))
	rowsEffected, err := session.Insert(packageEntity)
	if err != nil {
		slog.ErrorContext(ctx, "insert package failed",
			slog.Any("package", packageEntity), slog.Any("error", err))
		return nil, err
	}
	if rowsEffected == 0 {
		errMsg := "no rows affected when inserting package"
		err = errors.New(errMsg)
		slog.ErrorContext(ctx, errMsg,
			slog.Any("package_id", packageEntity.ID))
		return nil, err
	}
	slog.InfoContext(ctx, "Package inserted", slog.Any("package", packageEntity),
		slog.Any("rows_affected", rowsEffected))

	return packageEntity, nil
}

func (dao *PackageDAO) QueryPackageById(ctx context.Context,
	session *xorm.Session, id int64) (*entity.PackageEntity, error) {
	packageEntity := &entity.PackageEntity{}
	has, err := session.Where("id = ?", id).Get(packageEntity)
	if err != nil {
		slog.ErrorContext(ctx, "query package by id failed", slog.Int64("package_id", id), slog.Any("error", err))
		return nil, err
	}
	if !has {
		slog.InfoContext(ctx, "package not found", slog.Int64("package_id", id))
		return nil, nil
	}
	slog.InfoContext(ctx, "package found", slog.Any("package", packageEntity))
	return packageEntity, nil
}

func (dao *PackageDAO) QueryPackageByOriginId(ctx context.Context,
	session *xorm.Session, originId string) (*entity.PackageEntity, error) {
	packageEntity := &entity.PackageEntity{}
	has, err := session.Where("origin_id = ?", originId).Get(packageEntity)
	if err != nil {
		slog.ErrorContext(ctx, "query package by name failed", slog.String("origin_id", originId), slog.Any("error", err))
		return nil, err
	}
	if !has {
		slog.InfoContext(ctx, "package not found", slog.String("origin_id", originId))
		return nil, nil
	}
	slog.InfoContext(ctx, "package found", slog.Any("package", packageEntity))
	return packageEntity, nil
}

func (dao *PackageDAO) QueryPackageListByIdASC(ctx context.Context,
	session *xorm.Session, packageID int64, limit int32) ([]*entity.PackageEntity, error) {
	var packageEntities []*entity.PackageEntity
	err := session.Where("id > ?", packageID).Asc("id").Limit(int(limit)).Find(&packageEntities)
	if err != nil {
		slog.ErrorContext(ctx, "query package list by id asc failed", slog.Any("error", err))
		return nil, err
	}
	return packageEntities, nil
}

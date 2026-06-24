package logic

import (
	"bufio"
	"context"
	"log/slog"
	"os"
	"strconv"
	"sync"

	"github.com/KitHub/DataManagementPlatform_ComputeEngineAPI/component"
	"github.com/KitHub/DataManagementPlatform_ComputeEngineAPI/config"
	"github.com/KitHub/DataManagementPlatform_ComputeEngineAPI/entity"
	"github.com/RoaringBitmap/roaring/roaring64"
)

var computeLogicInstance *ComputeLogic
var onceForComputeLogicInstance sync.Once = sync.Once{}

type ComputeLogic struct {
	computeConfig *config.ComputeConfigEntity
	packageLogic  *PackageLogic
	packageMap    map[string]*roaring64.Bitmap
}

func NewComputeLogic(ctx context.Context, computeConfig *config.ComputeConfigEntity, packageLogic *PackageLogic, ossClient component.OSSComponent) (*ComputeLogic, error) {
	var err error = nil
	onceForComputeLogicInstance.Do(func() {
		err = os.MkdirAll(computeConfig.LocalBitmapDir, 0755)
		if err != nil {
			slog.ErrorContext(ctx, "create local bitmap dir failed", slog.String("localBitMapDir", computeConfig.LocalBitmapDir), slog.Any("error", err))
		}

		computeLogicInstance = &ComputeLogic{
			computeConfig: computeConfig,
			packageLogic:  packageLogic,
			packageMap:    make(map[string]*roaring64.Bitmap),
		}
	})
	return computeLogicInstance, err
}

func (logic *ComputeLogic) LoadPackages(ctx context.Context) error {
	var tmpId int64 = -1
	var limit int32 = 10
	var packages []*entity.PackageEntity = nil

	for {
		tmpPackages, err := logic.packageLogic.QueryPackageListASCById(ctx, tmpId, limit)
		if err != nil {
			slog.ErrorContext(ctx, "load packages failed", slog.Any("lastId", tmpId), slog.Any("limit", limit))
			return err
		}

		if len(tmpPackages) == 0 {
			break
		}

		packages = append(packages, tmpPackages...)
	}

	for _, tmpPackage := range packages {
		bitmap, err := logic.loadPackageToBitmap(ctx, tmpPackage, logic.packageLogic.ossClient)
		if err != nil {
			return err
		}

		logic.packageMap[tmpPackage.OriginId] = bitmap
	}
	slog.InfoContext(ctx, "load packages to bitmap done")
	return nil
}

func (logic *ComputeLogic) loadPackageToBitmap(ctx context.Context, packageEntity *entity.PackageEntity, ossClient component.OSSComponent) (*roaring64.Bitmap, error) {
	tmpPackageFileName := "ProjectGenerator-" + packageEntity.OriginId + "-*"
	tmpPackageFilePath, err := os.MkdirTemp("", tmpPackageFileName)
	if err != nil {
		slog.ErrorContext(ctx, "load package failed, create tmp package file faile", slog.String("tmpFileName", tmpPackageFileName), slog.Any("error", err))
		return nil, err
	}
	err = ossClient.GetDataToFile(ctx, packageEntity.BucketName, packageEntity.KeyName, tmpPackageFilePath)
	if err != nil {
		slog.ErrorContext(ctx, "load package failed", slog.String("bucket", packageEntity.BucketName), slog.String("key", packageEntity.KeyName), slog.Any("error", err))
		return nil, err
	}

	bitmap := roaring64.New()
	{
		tmpPackageFile, err := os.Open(tmpPackageFilePath)
		if err != nil {
			slog.ErrorContext(ctx, "reading package file failed", slog.String("packageFile", tmpPackageFilePath), slog.Any("error", err))
			return nil, err
		}
		defer tmpPackageFile.Close()

		scanner := bufio.NewScanner(tmpPackageFile)
		lineNo := 0
		for scanner.Scan() {
			lineNo++
			content := scanner.Text()
			// idmapping, deviceNo to deviceId
			tmpContentId, err := idMappingContentToId(ctx, content)
			if err != nil {
				slog.ErrorContext(ctx, "idmapping content2Id failed", slog.String("content", content), slog.Any("error", err))
				return nil, err
			}
			bitmap.Add(uint64(tmpContentId))
		}

		if err := scanner.Err(); err != nil {
			slog.ErrorContext(ctx, "reading package file failed", slog.String("packageFile", tmpPackageFilePath), slog.Any("error", err))
			return nil, err
		}
		slog.InfoContext(ctx, "read content lines done", slog.String("pacakgeFile", tmpPackageFilePath), slog.Int64("lineCount", int64(lineNo)))
	}

	{
		localBitmapFile := logic.computeConfig.LocalBitmapDir + "/" + packageEntity.OriginId + ".bitmap." + strconv.FormatInt(packageEntity.DataVersion, 10)
		err = serializeBitMapToLocalFile(ctx, bitmap, localBitmapFile)
		if err != nil {
			slog.ErrorContext(ctx, "serialize bitmap to local file failed", slog.String("packageFile", tmpPackageFilePath), slog.Any("error", err))
			return nil, err
		}
		slog.InfoContext(ctx, "serialize bitmap done", slog.Any("package", packageEntity), slog.String("packageBitmapFile", tmpPackageFilePath))
	}

	slog.InfoContext(ctx, "load package done", slog.String("pacakgeOriginId", packageEntity.OriginId))
	return bitmap, nil
}

func idMappingContentToId(ctx context.Context, content string) (int64, error) {
	// todo, add implmentation
	panic("not implmented")
}

func serializeBitMapToLocalFile(ctx context.Context, bitmap *roaring64.Bitmap, filePath string) error {
	// todo, add implmentation
	panic("not implmented")
}

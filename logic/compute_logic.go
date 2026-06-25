package logic

import (
	"bufio"
	"context"
	"errors"
	"log/slog"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/KitHub/DataManagementPlatform_ComputeEngineAPI/component"
	"github.com/KitHub/DataManagementPlatform_ComputeEngineAPI/config"
	"github.com/KitHub/DataManagementPlatform_ComputeEngineAPI/entity"
	devicemanagementplatformapi "github.com/KitHub/protocols/devicemanagementplatformapi"
	"github.com/RoaringBitmap/roaring/roaring64"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const bitmapLocalFileNameInfix = "bitmap"
const bitmapLocalFileNameSeparator = "."
const removeUselessBitmapLocalFileCronTaskName = "RemoveUselessBitmapLocalFileCronTask"

var computeLogicInstance *ComputeLogic
var onceForComputeLogicInstance sync.Once = sync.Once{}

type packageWithBitmapTuple struct {
	packageEntity       *entity.PackageEntity
	bitmap              *roaring64.Bitmap
	bitmapLocalFilePath string
}

type ComputeLogic struct {
	computeConfig                           *config.ComputeConfigEntity
	cronComponent                           *component.CronComponent
	packageLogic                            *PackageLogic
	packagesMap                             *component.SyncMap[string, *packageWithBitmapTuple] // map[string]*packageWithBitmapTuple // key = packageEntity.OriginId
	deviceManagementPlatformAPIClientConfig *config.ClientConfigEntity
	deviceManagementPlatformAPIClient       devicemanagementplatformapi.DeviceManagementPlatformAPIClient
}

type removeUselessBitmapLocalFileCronTask struct {
	packagesMap        *component.SyncMap[string, *packageWithBitmapTuple]
	bitmapLocalFileDir string
}

func (t *removeUselessBitmapLocalFileCronTask) Run() {
	ctx := context.Background()
	slog.InfoContext(ctx, "begin task, "+removeUselessBitmapLocalFileCronTaskName)
	err := removeUselessBitmapLocalFile(ctx, t.packagesMap, t.bitmapLocalFileDir)
	if err != nil {
		slog.ErrorContext(ctx, "task failed"+removeUselessBitmapLocalFileCronTaskName, slog.Any("error", err))
	}
	slog.InfoContext(ctx, "task done, "+removeUselessBitmapLocalFileCronTaskName)
}
func (t *removeUselessBitmapLocalFileCronTask) GetCronSpec() string {
	return "0 */2 * * *"
}
func (t *removeUselessBitmapLocalFileCronTask) GetName() string {
	return removeUselessBitmapLocalFileCronTaskName
}

func NewComputeLogic(ctx context.Context, computeConfig *config.ComputeConfigEntity, deviceManagementPlatformAPIClientConfig *config.ClientConfigEntity, cronComponent *component.CronComponent, packageLogic *PackageLogic, ossClient component.OSSComponent) (*ComputeLogic, error) {
	var err error = nil
	onceForComputeLogicInstance.Do(func() {
		err = os.MkdirAll(computeConfig.LocalBitmapDir, 0755)
		if err != nil {
			slog.ErrorContext(ctx, "create local bitmap dir failed", slog.String("localBitMapDir", computeConfig.LocalBitmapDir), slog.Any("error", err))
		}

		packagesMap := &component.SyncMap[string, *packageWithBitmapTuple]{} // make(map[string]*packageWithBitmapTuple)

		cronTask := &removeUselessBitmapLocalFileCronTask{
			packagesMap:        packagesMap,
			bitmapLocalFileDir: computeConfig.LocalBitmapDir,
		}

		err = computeLogicInstance.cronComponent.Register(ctx, cronTask)
		if err != nil {
			slog.ErrorContext(ctx, "register cron task failed", slog.String("taskName", cronTask.GetName()), slog.Any("error", err))
		}

		// create client to visit DeviceManagementPlatformAPI
		conn, err := grpc.NewClient(deviceManagementPlatformAPIClientConfig.Addr,
			grpc.WithTransportCredentials(insecure.NewCredentials()),
			grpc.WithConnectParams(grpc.ConnectParams{
				MinConnectTimeout: time.Duration(deviceManagementPlatformAPIClientConfig.TimemoutInMilliSeconds) * time.Millisecond,
			}))
		if err != nil {
			slog.ErrorContext(ctx, "create DeviceManagementPlatformAPI client failed", slog.Any("clientConfig", deviceManagementPlatformAPIClientConfig), slog.Any("error", err))
		}

		deviceManagementPlatformAPIClient := devicemanagementplatformapi.NewDeviceManagementPlatformAPIClient(conn)

		computeLogicInstance = &ComputeLogic{
			cronComponent:                           cronComponent,
			computeConfig:                           computeConfig,
			packageLogic:                            packageLogic,
			packagesMap:                             packagesMap,
			deviceManagementPlatformAPIClientConfig: deviceManagementPlatformAPIClientConfig,
			deviceManagementPlatformAPIClient:       deviceManagementPlatformAPIClient,
		}

	})

	return computeLogicInstance, err
}

// compute packages methods =======================================================

// load packages methods =======================================================

func (logic *ComputeLogic) ReloadPackage(ctx context.Context, packageEntity *entity.PackageEntity, forceReloadFromOSS bool) error {
	slog.InfoContext(ctx, "start reloading package", slog.Any("packageEntity", packageEntity))

	bitmap, err := loadPackage(ctx, packageEntity, logic.computeConfig.LocalBitmapDir, logic.packageLogic.ossClient, forceReloadFromOSS)
	if err != nil {
		slog.ErrorContext(ctx, "force reloading package failed", slog.String("packageOriginId", packageEntity.OriginId), slog.Any("error", err))
		return err
	}

	bitmapLocalFilePath, _, err := serializeBitMapToLocalFile(ctx, logic.computeConfig.LocalBitmapDir, packageEntity, bitmap)
	if err != nil {
		slog.ErrorContext(ctx, "force reloading package failed, serializing failed", slog.String("packageOriginId", packageEntity.OriginId), slog.Any("error", err))
		return err
	}

	logic.packagesMap.Store(packageEntity.OriginId, &packageWithBitmapTuple{
		packageEntity:       packageEntity,
		bitmap:              bitmap,
		bitmapLocalFilePath: bitmapLocalFilePath,
	})

	slog.InfoContext(ctx, "start reloading package done", slog.String("packageOriginId", packageEntity.OriginId))
	return nil
}

func (logic *ComputeLogic) LoadAllPackages(ctx context.Context, forceReload bool) error {
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
		err := logic.ReloadPackage(ctx, tmpPackage, forceReload)
		if err != nil {
			return err
		}
	}

	slog.InfoContext(ctx, "load packages to bitmap done")
	return nil
}

// private methods =============================================================================

func idMappingContentToId(ctx context.Context, deviceManagementPlatformAPIClient devicemanagementplatformapi.DeviceManagementPlatformAPIClient, content string) (int64, error) {
	rsp, err := deviceManagementPlatformAPIClient.QueryDeviceByNo(ctx, &devicemanagementplatformapi.QueryDeviceByNoRequest{
		DeviceNo: content,
	})
	if err != nil {
		slog.ErrorContext(ctx, "idMappingContentToId failed", slog.String("content", content), slog.Any("error", err))
		return 0, err
	}
	return rsp.GetData().GetDeviceInfo().DeviceId, nil
}

func serializeBitMapToLocalFile(ctx context.Context, bitmapLocalFileDir string, packageEntity *entity.PackageEntity, bitmap *roaring64.Bitmap) (localFilePath string, bytesCount int64, err error) {
	bitmapLocalFile := composeBitmapLocalFilePath(ctx, bitmapLocalFileDir, packageEntity)
	file, err := os.OpenFile(bitmapLocalFile, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		slog.ErrorContext(ctx, "open/create local bitmap file failed ", slog.Any("packageEntity", packageEntity), slog.String("bitmapLocalFile", bitmapLocalFile), slog.Any("error", err))
		return "", 0, err
	}
	defer func() {
		err := file.Close()
		if err != nil {
			slog.ErrorContext(ctx, "close bitmap local file failed", slog.Any("packageEntity", packageEntity), slog.String("bitmapLocalFile", bitmapLocalFile), slog.Any("error", err))
		}
	}()

	count, err := bitmap.WriteTo(file)
	if err != nil {
		slog.ErrorContext(ctx, "write local bitmap file failed ", slog.Any("packageEntity", packageEntity), slog.String("bitmapLocalFile", bitmapLocalFile), slog.Any("error", err))
		return "", 0, err
	}

	slog.InfoContext(ctx, "serialize bitmap done", slog.Any("packageEntity", packageEntity), slog.String("bitmapLocalFile", bitmapLocalFile), slog.Any("bytesCount", count))
	return bitmapLocalFile, count, nil
}

func composeBitmapLocalFilePath(ctx context.Context, bitmapLocalDir string, packageEntity *entity.PackageEntity) string {
	return bitmapLocalDir + string(os.PathSeparator) + composeBitmapLocalFileName(ctx, packageEntity)
}

func composeBitmapLocalFileName(ctx context.Context, packageEntity *entity.PackageEntity) string {
	return packageEntity.OriginId + bitmapLocalFileNameSeparator + bitmapLocalFileNameInfix + bitmapLocalFileNameSeparator + strconv.FormatInt(packageEntity.DataVersion, 10)
}

func decomposeBitmapLocalFileName(ctx context.Context, name string) (packageOriginId string, packageDataVersion string, err error) {
	parts := strings.Split(name, bitmapLocalFileNameSeparator)
	if len(parts) != 3 || parts[1] != bitmapLocalFileNameInfix {
		errMsg := "decompose bitmap file name failed, invalid name format"
		err = errors.New(errMsg)
		slog.ErrorContext(ctx, errMsg, slog.String("bitmapLocalFileName", name))
		return "", "", err
	}

	return parts[0], parts[2], nil
}

func loadBitmapFromOSS(ctx context.Context, packageEntity *entity.PackageEntity, ossClient component.OSSComponent) (*roaring64.Bitmap, error) {
	tmpPackageFileName := "bitmap-" + packageEntity.OriginId + "-*"
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
		defer func() {
			err := tmpPackageFile.Close()
			if err != nil {
				slog.ErrorContext(ctx, "close oss file failed", slog.String("ossFilePath", tmpPackageFilePath), slog.Any("error", err))
			}
			err = os.Remove(tmpPackageFilePath)
			if err != nil {
				slog.ErrorContext(ctx, "remove oss file failed", slog.String("ossFilePath", tmpPackageFilePath), slog.Any("error", err))
			}
			slog.InfoContext(ctx, "remove oss file done", slog.String("ossFilePath", tmpPackageFilePath))
		}()

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

	slog.InfoContext(ctx, "load package done", slog.String("pacakgeOriginId", packageEntity.OriginId))
	return bitmap, nil
}

func loadBitmapFromLocalFile(ctx context.Context, bitmapLocalFilePath string) (*roaring64.Bitmap, error) {
	file, err := os.OpenFile(bitmapLocalFilePath, os.O_RDONLY, 0644)
	if err != nil {
		slog.ErrorContext(ctx, "open local bitmap file failed ", slog.String("localBitmapFile", bitmapLocalFilePath), slog.Any("error", err))
		return nil, err
	}
	defer func() {
		err := file.Close()
		if err != nil {
			slog.ErrorContext(ctx, "close bitmap local file failed", slog.String("localBitmapFile", bitmapLocalFilePath), slog.Any("error", err))
		}
	}()
	bitmap := roaring64.New()
	bytesCount, err := bitmap.ReadFrom(file)
	if err != nil {
		slog.ErrorContext(ctx, "deserialize bitmap from local file failed", slog.String("localBitmapFile", bitmapLocalFilePath), slog.Any("error", err))
		return nil, err
	}
	slog.InfoContext(ctx, "deserialize bitmap from local file done", slog.String("localBitmapFile", bitmapLocalFilePath), slog.Any("bytesCount", bytesCount))
	return bitmap, nil
}

// loadPackage, load pacakge, if the bitmap local file existed and not expired, load bitmap from local file
// forceReloadFromOSS, if it is true, reload from oss, skipping local file
func loadPackage(ctx context.Context, targetPacakgeEntity *entity.PackageEntity, bitmapLocalFileDir string, ossClient component.OSSComponent, forceReloadFromOSS bool) (*roaring64.Bitmap, error) {
	bitmapLocalFilePath := composeBitmapLocalFilePath(ctx, bitmapLocalFileDir, targetPacakgeEntity)

	var retval *roaring64.Bitmap
	if !forceReloadFromOSS {
		retval, err := loadBitmapFromLocalFile(ctx, bitmapLocalFilePath)
		if err == nil {
			return retval, err
		}

		// if err not null, load bitmap from oss
	}

	// load bitmap from oss, and write content to local file
	retval, err := loadBitmapFromOSS(ctx, targetPacakgeEntity, ossClient)
	return retval, err
}

// removeUselessBitmapLocalFile, with the process running and package changed, old-versioned bitmap file stayed in `bitmapLocalFileDir`
// iterate SyncMap, mark useful file, and remove useless ones
func removeUselessBitmapLocalFile(ctx context.Context, packagesMap *component.SyncMap[string, *packageWithBitmapTuple], bitmapLocalFileDir string) error {
	entries, err := os.ReadDir(bitmapLocalFileDir)
	if err != nil {
		slog.ErrorContext(ctx, "read dir failed", slog.String("bitmapLocalFileDir", bitmapLocalFileDir), slog.Any("error", err))
		return err
	}

	filesToBeRemoved := make([]string, 0)
	for _, entry := range entries {
		name := entry.Name()
		if entry.IsDir() {
			slog.DebugContext(ctx, "removeUselessBitmapLocalFile, skip dir", slog.String("bitmapLocalFileDir", bitmapLocalFileDir), slog.Any("dirName", name))
		}

		packageOriginId, packageDataVersoin, err := decomposeBitmapLocalFileName(ctx, name)
		if err != nil {
			slog.DebugContext(ctx, "removeUselessBitmapLocalFile, skip non-bitmap file", slog.String("bitmapLocalFileDir", bitmapLocalFileDir), slog.Any("dirName", name))
			return err
		}

		tuple, ok := packagesMap.Load(packageOriginId)
		if !ok {
			filesToBeRemoved = append(filesToBeRemoved, bitmapLocalFileDir+string(os.PathSeparator)+name)
			continue
		}

		if packageDataVersoin != strconv.FormatInt(tuple.packageEntity.DataVersion, 10) {
			filesToBeRemoved = append(filesToBeRemoved, bitmapLocalFileDir+string(os.PathSeparator)+name)
			continue
		}
	}

	for _, filePath := range filesToBeRemoved {
		err := os.Remove(filePath)
		if err != nil {
			slog.ErrorContext(ctx, "remove useless bitmap local file failed", slog.Any("error", err))
		}
	}

	return nil
}
